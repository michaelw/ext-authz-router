package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

var (
	metricsCollectorURL = os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")

	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
)

func InitMeter() func(context.Context) error {
	prom, err := prometheus.New()
	if err != nil {
		log.Fatalf("Failed to create prometheus exporter: %v", err)
	}

	options := []metricsdk.Option{
		metricsdk.WithReader(prom),
	}

	if strings.Trim(strings.ToLower(metricsCollectorURL), " ") != "" {
		exporter, err := otlpmetricgrpc.New(context.Background())
		if err != nil {
			log.Fatalf("Failed to create exporter: %v", err)
		}
		options = append(options, metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter)))
	}

	provider := metricsdk.NewMeterProvider(options...)
	otel.SetMeterProvider(provider)

	meter := otel.Meter("gin-server")

	requestCount, err = meter.Int64Counter(
		"http_server_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)

	if err != nil {
		log.Fatalf("failed to create requestCount instrument: %v", err)
	}

	requestDuration, err = meter.Float64Histogram(
		"http_server_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		// Default buckets are much more than this
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.05, 0.5, 1, 5),
	)
	if err != nil {
		log.Fatalf("failed to create requestDuration instrument: %v", err)
	}

	return provider.Shutdown
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Let the other handlers process now that the time has started
		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()
		if len(route) <= 0 {
			route = "nonconfigured_route"
		}

		// Group the codes
		// TODO: make this a config option later?
		code := int(c.Writer.Status()/100) * 100

		requestCount.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.Int("http.status_code", code),
				attribute.String("http.path", route),
				attribute.String("instance", os.Getenv("HOSTNAME")),
			),
		)

		requestDuration.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.Int("http.status_code", code),
				attribute.String("http.path", route),
				attribute.String("instance", os.Getenv("HOSTNAME")),
			),
		)
	}
}
