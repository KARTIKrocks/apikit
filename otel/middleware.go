package otel

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/KARTIKrocks/apikit/middleware"
)

const instrumentationName = "github.com/KARTIKrocks/apikit/otel"

// Middleware returns HTTP middleware that starts a span and records
// duration/count metrics for every request.
//
// Spans and metrics are named using the matched route pattern (via
// WithSpanNameFormatter's default, r.Pattern), not the raw request path, to
// avoid the classic high-cardinality mistake of one span/metric series per
// distinct resource ID. Register it with router.Router.Use so it runs inside
// the mux's dispatch, where r.Pattern has already been set by the match.
//
// If the incoming request carries an apikit middleware.RequestID (see
// apikit/middleware.RequestID), it's attached to the span as an attribute,
// giving free correlation between logs and traces.
func Middleware(opts ...Option) func(http.Handler) http.Handler {
	cfg := newConfig(opts)
	tracer := cfg.tracerProvider.Tracer(instrumentationName)
	meter := cfg.meterProvider.Meter(instrumentationName)

	duration, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of inbound HTTP requests."),
	)
	requestCount, _ := meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Count of inbound HTTP requests."),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := cfg.propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			name := cfg.spanName(r)

			ctx, span := tracer.Start(ctx, name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.URLPath(r.URL.Path),
					semconv.ServerAddress(r.Host),
				),
			)
			defer span.End()

			if reqID := middleware.GetRequestID(ctx); reqID != "" {
				span.SetAttributes(attribute.String("apikit.request_id", reqID))
			}

			rw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rw, r.WithContext(ctx))
			elapsed := time.Since(start).Seconds()

			attrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.HTTPResponseStatusCode(rw.statusCode),
			}
			span.SetAttributes(semconv.HTTPResponseStatusCode(rw.statusCode))
			if rw.statusCode >= 500 {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			}

			opt := metric.WithAttributes(attrs...)
			duration.Record(ctx, elapsed, opt)
			requestCount.Add(ctx, 1, opt)
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the response status
// code for metrics/span attributes without altering response behavior.
type statusWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *statusWriter) WriteHeader(code int) {
	if w.written {
		return
	}
	w.statusCode = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(b)
}

// Unwrap allows http.ResponseController and interface probes (Flusher,
// Hijacker) to reach the underlying ResponseWriter through this wrapper.
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
