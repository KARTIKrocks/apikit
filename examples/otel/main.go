// Command otel demonstrates apikit's optional OpenTelemetry submodule: a
// server middleware that produces spans + metrics named after the matched
// route pattern, and an httpclient Transport that propagates trace context
// on outbound calls. Both providers print to stdout so the demo needs no
// collector — swap NewTracerProvider/NewMeterProvider for OTLP exporters in
// production.
//
// This example lives in its own Go module (examples/otel/go.mod) so its
// OpenTelemetry SDK dependencies never leak into the core apikit module or
// the apikit/otel module itself.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/KARTIKrocks/apikit/health"
	"github.com/KARTIKrocks/apikit/httpclient"
	apikitotel "github.com/KARTIKrocks/apikit/otel"
	"github.com/KARTIKrocks/apikit/response"
	"github.com/KARTIKrocks/apikit/router"
)

func main() {
	// A composite propagator is required — the global default is a no-op.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter))
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)

	metricExporter, err := stdoutmetric.New()
	if err != nil {
		log.Fatal(err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))))
	defer mp.Shutdown(context.Background())
	otel.SetMeterProvider(mp)

	r := router.New()
	r.Use(apikitotel.Middleware())

	// Outbound calls get their own client span + propagated trace context.
	client := httpclient.New("https://example.com",
		httpclient.WithTransport(apikitotel.Transport(nil)),
	)

	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		resp, err := client.Get(req.Context(), "/")
		if err != nil {
			return err
		}
		response.OK(w, "OK", map[string]any{"id": req.PathValue("id"), "upstream_status": resp.StatusCode})
		return nil
	})

	h := health.NewChecker()
	h.AddCheck("upstream", apikitotel.WrapCheck("upstream", func(ctx context.Context) error {
		_, err := client.Get(ctx, "/")
		return err
	}))
	r.Get("/health", h.Handler())

	log.Println("listening on :8080 — try GET /users/42, spans print to stdout")
	log.Fatal(http.ListenAndServe(":8080", r))
}
