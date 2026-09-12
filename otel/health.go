package otel

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/KARTIKrocks/apikit/health"
)

// WrapCheck wraps a health.CheckFunc in its own span named "health.check <name>",
// so a slow or failing dependency check shows up in traces instead of only in
// the health endpoint's response body.
//
//	h := health.NewChecker()
//	h.AddCheck("postgres", apikitotel.WrapCheck("postgres", func(ctx context.Context) error {
//	    return db.PingContext(ctx)
//	}))
func WrapCheck(name string, fn health.CheckFunc, opts ...Option) health.CheckFunc {
	cfg := newConfig(opts)
	tracer := cfg.tracerProvider.Tracer(instrumentationName)

	return func(ctx context.Context) error {
		ctx, span := tracer.Start(ctx, "health.check "+name, trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		if err := fn(ctx); err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return err
		}
		return nil
	}
}
