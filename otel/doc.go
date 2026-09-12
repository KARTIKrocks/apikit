// Package otel provides optional, real OpenTelemetry integration for apikit:
// an HTTP server middleware (traces + metrics), an httpclient-compatible
// RoundTripper for outbound propagation, and a health-check span wrapper.
//
// It is a separate Go module (github.com/KARTIKrocks/apikit/otel) precisely
// because it needs the actual go.opentelemetry.io/otel SDK to be understood
// by real OTel backends — importing it never adds a dependency to the core
// github.com/KARTIKrocks/apikit module.
//
// The OpenTelemetry API defaults its global TracerProvider, MeterProvider,
// and propagator to no-ops until you configure real ones — this package
// follows that convention rather than silently overriding it. At minimum,
// set a real propagator (the global default does not inject/extract
// anything):
//
//	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
//	    propagation.TraceContext{}, propagation.Baggage{},
//	))
//
//	tp := otel.GetTracerProvider() // wire up a real SDK TracerProvider in production
//	mp := otel.GetMeterProvider()  // and a real MeterProvider
//
//	r := router.New()
//	r.Use(apikitotel.Middleware(
//	    apikitotel.WithTracerProvider(tp),
//	    apikitotel.WithMeterProvider(mp),
//	))
//
//	client := httpclient.New("https://api.example.com",
//	    httpclient.WithTransport(apikitotel.Transport(nil,
//	        apikitotel.WithTracerProvider(tp),
//	    )),
//	)
package otel
