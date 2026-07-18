package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName string `mapstructure:"serviceName"`
	HostPort    string `mapstructure:"hostPort"`
	Enable      bool   `mapstructure:"enable"`
	LogSpans    bool   `mapstructure:"logSpans"`
}

var tracer trace.Tracer

func tracerProvider(config Config) (*tracesdk.TracerProvider, error) {
	if !config.Enable {
		tracer = trace.NewNoopTracerProvider().Tracer(config.ServiceName)
		return nil, nil
	}
	url := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if url != "" {
		config.HostPort = url
	}

	client := otlptracegrpc.NewClient(otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(config.HostPort))
	exporter, err := otlptrace.New(context.Background(), client)

	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(0.3))),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			attribute.String("serviceName", config.ServiceName),
		)),
	)
	tracer = tp.Tracer(config.ServiceName)
	return tp, nil
}
