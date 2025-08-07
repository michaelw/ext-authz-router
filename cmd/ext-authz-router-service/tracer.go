package main

import (
	"context"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var traceCollectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

func InitTracer() func(context.Context) error {
	if strings.Trim(strings.ToLower(traceCollectorURL), " ") == "" {
		return func(ctx context.Context) error {
			return nil
		}
	}
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
		),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return exporter.Shutdown
}
