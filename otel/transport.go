package otel

import (
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Transport wraps base (or http.DefaultTransport when base is nil) with a
// RoundTripper that starts a client span per request and injects trace
// context into outbound headers via the configured propagator.
//
// It's designed to drop into httpclient's existing extension point:
//
//	client := httpclient.New("https://api.example.com",
//	    httpclient.WithTransport(apikitotel.Transport(nil)),
//	)
func Transport(base http.RoundTripper, opts ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	cfg := newConfig(opts)
	return &otelTransport{
		base:   base,
		tracer: cfg.tracerProvider.Tracer(instrumentationName),
		prop:   cfg.propagator,
	}
}

type otelTransport struct {
	base   http.RoundTripper
	tracer trace.Tracer
	prop   propagation.TextMapPropagator
}

func (t *otelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.Start(req.Context(), req.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPRequestMethodKey.String(req.Method),
			semconv.URLFull(req.URL.String()),
		),
	)
	defer span.End()

	req = req.Clone(ctx)
	t.prop.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return resp, err
	}

	span.SetAttributes(semconv.HTTPResponseStatusCode(resp.StatusCode))
	if resp.StatusCode >= 500 {
		span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
	}
	return resp, nil
}
