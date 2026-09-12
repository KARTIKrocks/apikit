# Changelog

All notable changes to this module (`github.com/KARTIKrocks/apikit/otel`) will
be documented in this file. It is versioned independently from the core
`github.com/KARTIKrocks/apikit` module — see its own
[CHANGELOG.md](../CHANGELOG.md) — and tagged as `otel/vX.Y.Z`.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-12

Initial release.

### Added

- **middleware** — `Middleware(opts...)` HTTP server middleware that starts a span and records `http.server.request.duration`/`http.server.request.count` metrics for every request. Spans and metrics are named after the matched route pattern (e.g. `GET /users/{id}`, via `r.Pattern`) rather than the raw path, avoiding one series per distinct resource ID. A 5xx response marks the span `Error`. When the request carries an apikit `middleware.RequestID`, it's attached to the span as `apikit.request_id`
- **transport** — `Transport(base, opts...)` wraps an `http.RoundTripper` with a client span per outbound request and injects trace context via the configured propagator; designed to drop straight into `httpclient.WithTransport`
- **health** — `WrapCheck(name, fn, opts...)` wraps a `health.CheckFunc` in its own span (`health.check <name>`), so a slow or failing dependency check shows up in traces instead of only in the health endpoint's response body
- `WithTracerProvider`, `WithMeterProvider`, `WithPropagator`, and `WithSpanNameFormatter` options, all defaulting to the corresponding OpenTelemetry global (`otel.GetTracerProvider()`, etc.)

### Notes

- The OpenTelemetry API's global `TracerProvider`/`MeterProvider`/propagator default to no-ops until configured — this module follows that convention. At minimum, set a real propagator (e.g. `propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})`) or `Transport`/`Middleware` won't inject/extract anything
- Requires Go 1.25+, pinned by the OpenTelemetry SDK's own module requirement — higher than the core module's Go 1.22+, since this is an independent module
