package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/michaelw/ext-authz-router/api"
	"github.com/michaelw/ext-authz-router/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	publicURL := os.Getenv("URLS_SELF_PUBLIC")
	if publicURL == "" {
		publicURL = "http://localhost:" + port
	}

	// Set up OpenTelemetry exporters
	shutdownTracer := InitTracer()
	defer shutdownTracer(context.Background())

	shutdownMetrics := InitMeter()
	defer shutdownMetrics(context.Background())

	// Load OpenAPI spec for validation
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("failed to load OpenAPI spec: %v", err)
	}
	swagger.Servers = nil // Clear servers

	r := gin.Default()

	server := server.NewServerHandler(publicURL, swagger)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Use(MetricsMiddleware())
	// All routes registered from this point will have metrics tracking
	server.RegisterRoutes(r)

	r.Use(otelgin.Middleware(os.Getenv("HOSTNAME")))
	// All routes registered from this point will have tracing

	// Add custom non-OpenAPI routes before validation
	r.Use(ginmiddleware.OapiRequestValidator(swagger))
	// All routes registered from this point will be enforced against the OpenAPI spec
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(r, handler)

	log.Printf("Starting server on :%s", port)
	log.Printf("Server public URL: %s", publicURL)
	r.Run(":" + port)
}
