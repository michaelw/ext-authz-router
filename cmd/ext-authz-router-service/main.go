package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/michaelw/ext-authz-router/api"
	"github.com/michaelw/ext-authz-router/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "3001"
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

	// Create handler (shared between HTTP and gRPC)
	authzHandler := server.NewServerHandler(publicURL, swagger)

	var wg sync.WaitGroup

	// Start HTTP server for UI and legacy endpoints
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := gin.Default()

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
		r.Use(MetricsMiddleware())
		// All routes registered from this point will have metrics tracking
		authzHandler.RegisterRoutes(r)

		r.Use(otelgin.Middleware(os.Getenv("HOSTNAME")))
		// All routes registered from this point will have tracing

		// Add custom non-OpenAPI routes before validation
		r.Use(ginmiddleware.OapiRequestValidator(swagger))
		// All routes registered from this point will be enforced against the OpenAPI spec
		handler := api.NewStrictHandler(authzHandler, nil)
		api.RegisterHandlers(r, handler)

		log.Printf("Starting HTTP server on :%s", port)
		log.Printf("Server public URL: %s", publicURL)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server for ext_authz
	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port %s: %v", grpcPort, err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(server.LoggingInterceptor()),
		)
		authzServer := server.NewAuthzGRPCServer(authzHandler)
		envoy_service_auth_v3.RegisterAuthorizationServer(grpcServer, authzServer)
		reflection.Register(grpcServer) // Register reflection service for grpcurl

		log.Printf("Starting gRPC ext_authz server on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	wg.Wait()
}
