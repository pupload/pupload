package telemetry

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	provider *sdktrace.TracerProvider
	once     sync.Once
)

func Init(cfg TelemetrySettings, name string) error {
	var err error

	once.Do(func() {
		err = initOnce(cfg, name)
	})

	return err
}

func initOnce(cfg TelemetrySettings, name string) error {
	if !cfg.Enabled || cfg.Exporter == ExporterNone {
		return nil
	}

	res, err := resource.New(context.Background(), resource.WithAttributes(
		semconv.ServiceNameKey.String(name),
	))

	if err != nil {
		return err
	}

	var exporter sdktrace.SpanExporter

	switch cfg.Exporter {
	case ExporterOTLP:
		exporter, err = createOTLPExporter(cfg)
		if err != nil {
			return err
		}
	case ExporterStdout:
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown exporter type: %s", cfg.Exporter)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	provider = tp

	return nil
}

func createOTLPExporter(cfg TelemetrySettings) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

func Shutdown(ctx context.Context) error {
	p := provider
	if p == nil {
		return nil
	}

	return p.Shutdown(ctx)
}

func Tracer(name string) trace.Tracer {
	p := provider
	if p == nil {
		return noop.NewTracerProvider().Tracer(name)
	}

	return p.Tracer(name)
}
