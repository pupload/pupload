package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func InjectContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return ""
	}

	prop := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	prop.Inject(ctx, carrier)

	return carrier.Get("traceparent")
}

func ExtractContext(ctx context.Context, traceParent string) context.Context {
	if traceParent == "" {
		return ctx
	}

	carrier := propagation.MapCarrier{
		"traceparent": traceParent,
	}

	prop := propagation.TraceContext{}
	return prop.Extract(ctx, carrier)
}
