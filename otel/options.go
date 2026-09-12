package otel

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	propagator     propagation.TextMapPropagator
	spanName       func(*http.Request) string
}

// Option configures Middleware and Transport.
type Option func(*config)

// WithTracerProvider sets the TracerProvider used to start spans.
// Defaults to otel.GetTracerProvider() (the global provider) when unset.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *config) { c.tracerProvider = tp }
}

// WithMeterProvider sets the MeterProvider used to record metrics.
// Defaults to otel.GetMeterProvider() (the global provider) when unset.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(c *config) { c.meterProvider = mp }
}

// WithPropagator sets the propagator used to extract/inject trace context.
// Defaults to otel.GetTextMapPropagator() (the global propagator) when unset.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return func(c *config) { c.propagator = p }
}

// WithSpanNameFormatter overrides how a request is named for its span and
// metric attributes. The default uses r.Pattern (the route pattern matched
// by http.ServeMux/router.Router, e.g. "GET /users/{id}" — ServeMux includes
// the method in the pattern itself) to keep span and metric cardinality low;
// it falls back to r.Method alone when r.Pattern is empty (routes registered
// without going through the mux's own matching, e.g. a custom 404 handler).
func WithSpanNameFormatter(fn func(*http.Request) string) Option {
	return func(c *config) { c.spanName = fn }
}

func newConfig(opts []Option) config {
	c := config{
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  otel.GetMeterProvider(),
		propagator:     otel.GetTextMapPropagator(),
		spanName:       defaultSpanName,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func defaultSpanName(r *http.Request) string {
	switch {
	case r.Pattern == "":
		return r.Method
	case strings.HasPrefix(r.Pattern, r.Method+" "):
		return r.Pattern // e.g. "GET /users/{id}" — ServeMux patterns already embed the method
	default:
		return r.Method + " " + r.Pattern // pattern registered without a method (matches every method)
	}
}
