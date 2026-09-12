# apikit/otel

Real [OpenTelemetry](https://opentelemetry.io/) tracing and metrics for [apikit](https://github.com/KARTIKrocks/apikit) — a **separate Go module**, so importing it never adds a dependency to the zero-dependency core `github.com/KARTIKrocks/apikit` module.

```bash
go get github.com/KARTIKrocks/apikit/otel
```

Requires Go 1.25+ (pinned by the OpenTelemetry SDK's own module requirement — higher than the core module's Go 1.22+, since this is an independent module).

## Why a separate module

Real OTel interop means interoperating with the actual `go.opentelemetry.io/otel` ecosystem — collectors, Jaeger, Grafana Tempo, Datadog, and so on. There's no way to do that without the real SDK types (`trace.Tracer`, `metric.Meter`, propagators). Shipping it as its own module means:

- The core `apikit` module's `go.mod` stays empty — you only pay for this dependency if you `go get` it.
- This module can version independently as the OTel SDK evolves.

## Quick start

The OpenTelemetry API defaults its global `TracerProvider`, `MeterProvider`, and propagator to no-ops until you configure real ones — this package follows that convention rather than silently overriding it. At minimum, set a real propagator:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{}, propagation.Baggage{},
))
// otel.SetTracerProvider(...) / otel.SetMeterProvider(...) with a real SDK
// exporter (OTLP, stdout, etc.) — see go.opentelemetry.io/otel/sdk.
```

### Server middleware

```go
import (
    apikitotel "github.com/KARTIKrocks/apikit/otel"
    "github.com/KARTIKrocks/apikit/router"
)

r := router.New()
r.Use(apikitotel.Middleware())
```

Each request gets a span and a duration/count metric named after the **matched route pattern** (e.g. `GET /users/{id}`), not the raw path — so `/users/1` and `/users/2` collapse into one low-cardinality series instead of one per user ID. If the request carries an apikit `middleware.RequestID`, it's attached to the span as `apikit.request_id`, giving free log↔trace correlation.

Pass explicit providers instead of relying on the globals:

```go
r.Use(apikitotel.Middleware(
    apikitotel.WithTracerProvider(tp),
    apikitotel.WithMeterProvider(mp),
))
```

### Outbound requests (httpclient)

`Transport` drops straight into `httpclient`'s existing `WithTransport` option — the two packages compose without either knowing about the other:

```go
import "github.com/KARTIKrocks/apikit/httpclient"

client := httpclient.New("https://api.example.com",
    httpclient.WithTransport(apikitotel.Transport(nil)),
)
```

Each outbound call gets a client span and has trace context injected into its headers via the configured propagator.

### Health checks

Wrap a `health.CheckFunc` to see slow or failing dependency checks as their own spans:

```go
import "github.com/KARTIKrocks/apikit/health"

h := health.NewChecker()
h.AddCheck("postgres", apikitotel.WrapCheck("postgres", func(ctx context.Context) error {
    return db.PingContext(ctx)
}))
```

## Options

| Option                       | Default                          |
| ----------------------------- | --------------------------------- |
| `WithTracerProvider(tp)`      | `otel.GetTracerProvider()`        |
| `WithMeterProvider(mp)`       | `otel.GetMeterProvider()`         |
| `WithPropagator(p)`           | `otel.GetTextMapPropagator()`     |
| `WithSpanNameFormatter(fn)`   | `r.Pattern`, falling back to `r.Method` |

## Testing

The package's own test suite (`go test ./...` from this directory) uses the OTel SDK's in-memory exporters (`sdk/trace/tracetest`, `sdk/metric`) — those are regular imports of this module's test files, so they never affect what a consuming application needs to download or compile.

## Local development

This module has a `replace github.com/KARTIKrocks/apikit => ../` directive so it always builds against the sibling checkout during development — `go build`/`go test`/`go vet` in this directory work standalone, no extra setup needed.

For cross-module IDE tooling (jump-to-definition into the core module, etc.), create a local (gitignored) workspace at the repo root:

```bash
go work init . ./otel ./examples/otel
```

CI deliberately runs with `GOWORK=off` so each module is still built/tested against its own `go.mod` in isolation, matching what a consumer sees.
