package otel

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(serviceName string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	// Determine the OTLP endpoint based on environment
	endpoint := "http://localhost:4318"
	if os.Getenv("JAEGER_ENDPOINT") != "" {
		endpoint = os.Getenv("JAEGER_ENDPOINT")
	}

	log.Printf("Initializing tracer for service: %s", serviceName)
	log.Printf("Connecting to Jaeger OTLP endpoint: %s", endpoint)

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Required for HTTP (not HTTPS)
		otlptracehttp.WithHeaders(map[string]string{
			"Content-Type": "application/x-protobuf",
		}),
	)
	if err != nil {
		log.Printf("Warning: Could not connect to OTLP: %v", err)
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
