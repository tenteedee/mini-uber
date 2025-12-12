package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
}

func InitTracer(cfg Config) (func(context.Context) error, error) {
	// Exporter
	traceExporter, err := newExporter(cfg.JaegerEndpoint)
	if err != nil {
		return nil, err
	}

	// Trace provider
	traceProvider, err := newTraceProvider(cfg, traceExporter)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace provider: %w", err)
	}
	otel.SetTracerProvider(traceProvider)

	// Propagator
	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	return traceProvider.Shutdown, nil
}

func newExporter(endpoint string) (sdktrace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
}

func newTraceProvider(cfg Config, traceExporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	return traceProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
