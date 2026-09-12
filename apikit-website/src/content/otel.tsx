import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function OtelDocs() {
  return (
    <ModuleSection
      id="otel"
      title="otel"
      description="Real OpenTelemetry tracing and metrics — a separate Go module with its own go.mod, so the dependency never reaches the core apikit module. Tagged and versioned independently as otel/vX.Y.Z."
      importPath="github.com/KARTIKrocks/apikit/otel"
      features={[
        'Server middleware: spans + duration/count metrics per request',
        'Spans/metrics named after the matched route pattern (e.g. "GET /users/{id}"), not the raw path — keeps cardinality low',
        'httpclient-compatible Transport for outbound trace propagation',
        'health.CheckFunc wrapper — slow/failing dependency checks show up as spans',
        'Correlates with apikit middleware.RequestID automatically',
        'Requires Go 1.25+ (pinned by the OpenTelemetry SDK itself)',
      ]}
    >
      <h3 id="otel-setup" className="text-lg font-semibold text-text-heading mt-8 mb-2">Install &amp; Setup</h3>
      <p className="text-text-muted mb-3 text-sm">
        This is a separate module — install it on top of the core package.
      </p>
      <CodeBlock lang="bash" code={`go get github.com/KARTIKrocks/apikit/otel`} />
      <p className="text-text-muted mb-3 text-sm mt-4">
        OpenTelemetry's global TracerProvider, MeterProvider, and propagator all default to
        <strong className="text-text"> no-ops</strong> until you configure real ones — this package follows that
        convention rather than silently overriding it. At minimum, set a real propagator:
      </p>
      <CodeBlock code={`import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{}, propagation.Baggage{},
))
// otel.SetTracerProvider(...) / otel.SetMeterProvider(...) with a real SDK
// exporter (OTLP, stdout, etc.) — see go.opentelemetry.io/otel/sdk.`} />

      <h3 id="otel-middleware" className="text-lg font-semibold text-text-heading mt-8 mb-2">Server Middleware</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Middleware(opts...)</td><td className="py-2 text-text-muted">HTTP middleware that starts a span and records request metrics</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`import (
    apikitotel "github.com/KARTIKrocks/apikit/otel"
    "github.com/KARTIKrocks/apikit/router"
)

r := router.New()
r.Use(apikitotel.Middleware(
    apikitotel.WithTracerProvider(tp),
    apikitotel.WithMeterProvider(mp),
))`} />

      <h3 id="otel-transport" className="text-lg font-semibold text-text-heading mt-8 mb-2">Outbound Requests</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Transport(base, opts...)</td><td className="py-2 text-text-muted">http.RoundTripper with client spans + propagated trace context</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`import "github.com/KARTIKrocks/apikit/httpclient"

client := httpclient.New("https://api.example.com",
    httpclient.WithTransport(apikitotel.Transport(nil)),
)`} />

      <h3 id="otel-health" className="text-lg font-semibold text-text-heading mt-8 mb-2">Health Checks</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WrapCheck(name, fn, opts...)</td><td className="py-2 text-text-muted">Wraps a health.CheckFunc in its own span</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`import "github.com/KARTIKrocks/apikit/health"

h := health.NewChecker()
h.AddCheck("postgres", apikitotel.WrapCheck("postgres", func(ctx context.Context) error {
    return db.PingContext(ctx)
}))`} />

      <h3 id="otel-options" className="text-lg font-semibold text-text-heading mt-8 mb-2">Options</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Option</th><th className="py-2 text-text-heading font-semibold">Default</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithTracerProvider(tp)</td><td className="py-2 text-text-muted">otel.GetTracerProvider()</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithMeterProvider(mp)</td><td className="py-2 text-text-muted">otel.GetMeterProvider()</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithPropagator(p)</td><td className="py-2 text-text-muted">otel.GetTextMapPropagator()</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithSpanNameFormatter(fn)</td><td className="py-2 text-text-muted">r.Pattern, falling back to r.Method</td></tr>
        </tbody></table>
      </div>
    </ModuleSection>
  );
}
