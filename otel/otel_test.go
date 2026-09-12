package otel_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/KARTIKrocks/apikit/health"
	apikitotel "github.com/KARTIKrocks/apikit/otel"
	"github.com/KARTIKrocks/apikit/router"
)

func TestMiddlewareRecordsSpanAndMetrics(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	r := router.New()
	r.Use(apikitotel.Middleware(
		apikitotel.WithTracerProvider(tp),
		apikitotel.WithMeterProvider(mp),
	))
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	if span.Name() != "GET /users/{id}" {
		t.Fatalf("expected span name using route pattern, got %q", span.Name())
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(rm.ScopeMetrics) == 0 || len(rm.ScopeMetrics[0].Metrics) == 0 {
		t.Fatalf("expected recorded metrics, got none: %+v", rm)
	}
	names := map[string]bool{}
	for _, m := range rm.ScopeMetrics[0].Metrics {
		names[m.Name] = true
	}
	if !names["http.server.request.duration"] || !names["http.server.request.count"] {
		t.Fatalf("expected duration+count metrics, got %v", names)
	}
}

func TestMiddlewareMarksServerErrorSpans(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))

	r := router.New()
	r.Use(apikitotel.Middleware(apikitotel.WithTracerProvider(tp)))
	r.Get("/boom", func(w http.ResponseWriter, req *http.Request) error {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code.String() != "Error" {
		t.Fatalf("expected Error status for 5xx response, got %v", spans[0].Status())
	}
}

func TestTransportInjectsAndRecordsClientSpan(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))

	var gotTraceParent string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceParent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	client := &http.Client{Transport: apikitotel.Transport(nil,
		apikitotel.WithTracerProvider(tp),
		apikitotel.WithPropagator(propagation.TraceContext{}), // the global default propagator is a no-op
	)}
	resp, err := client.Get(upstream.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if gotTraceParent == "" {
		t.Fatalf("expected traceparent header to be injected by the default propagator")
	}
	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 client span, got %d", len(spans))
	}
}

func TestWrapCheckRecordsErrorSpan(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))

	boom := errors.New("connection refused")
	checked := apikitotel.WrapCheck("postgres", func(ctx context.Context) error {
		return boom
	}, apikitotel.WithTracerProvider(tp))

	h := health.NewChecker()
	h.AddCheck("postgres", checked)

	resp := h.Check(context.Background())
	if resp.Status == "healthy" {
		t.Fatalf("expected unhealthy status, got %q", resp.Status)
	}

	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name() != "health.check postgres" {
		t.Fatalf("unexpected span name: %q", spans[0].Name())
	}
	if spans[0].Status().Code.String() != "Error" {
		t.Fatalf("expected Error status, got %v", spans[0].Status())
	}
}
