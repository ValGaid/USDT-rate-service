package tracer

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func InitTracer(jaegerURL string, serviceName string) (trace.Tracer, error) {
	exporter, err := newJaegerExporter(jaegerURL)
	if err != nil {
		return nil, fmt.Errorf("initialize Jaeger exporter: %w", err)
	}

	tp, err := newTraceProvider(exporter, serviceName)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)

	return tp.Tracer("USDT_service_tracer"), nil
}

func newJaegerExporter(url string) (sdktrace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

}

func newTraceProvider(exp sdktrace.SpanExporter, ServiceName string) (*sdktrace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("merge resource: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	), nil
}
