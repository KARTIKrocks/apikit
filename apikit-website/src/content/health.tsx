import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function HealthDocs() {
  return (
    <ModuleSection
      id="health"
      title="health"
      description="Health check endpoint builder with dependency checks, timeouts, and liveness/readiness probes."
      importPath="github.com/KARTIKrocks/apikit/health"
      features={[
        'Critical and non-critical dependency checks',
        'Per-check timeout support',
        'Kubernetes-ready liveness and readiness probes',
        'Status: healthy, degraded, or unhealthy',
        'Standard JSON response format',
      ]}
    >
      <h3 id="health-creating" className="text-lg font-semibold text-text-heading mt-8 mb-2">Creating a Checker</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NewChecker(opts...)</td><td className="py-2 text-text-muted">Create a health checker with options</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithTimeout(d)</td><td className="py-2 text-text-muted">Set per-check timeout</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`h := health.NewChecker(health.WithTimeout(3 * time.Second))`} />

      <h3 id="health-checks" className="text-lg font-semibold text-text-heading mt-8 mb-2">Adding Checks</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AddCheck(name, fn)</td><td className="py-2 text-text-muted">Add critical check (failure = unhealthy, 503)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AddNonCriticalCheck(name, fn)</td><td className="py-2 text-text-muted">Add non-critical check (failure = degraded, 200)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`h.AddCheck("postgres", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

h.AddNonCriticalCheck("redis", func(ctx context.Context) error {
    return rdb.Ping(ctx).Err()
})

h.AddNonCriticalCheck("smtp", func(ctx context.Context) error {
    return smtpClient.Noop()
})`} />

      <h3 id="health-handlers" className="text-lg font-semibold text-text-heading mt-8 mb-2">HTTP Handlers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Handler()</td><td className="py-2 text-text-muted">HTTP handler for readiness probe (runs all checks)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.LiveHandler()</td><td className="py-2 text-text-muted">HTTP handler for liveness probe (always 200)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Check(ctx)</td><td className="py-2 text-text-muted">Run checks programmatically</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.GetFunc("/health", h.Handler())          // readiness probe
r.GetFunc("/health/live", h.LiveHandler()) // liveness probe (always 200)

// Programmatic check
result, err := h.Check(ctx)
fmt.Println(result.Status) // "healthy", "degraded", or "unhealthy"`} />

      <h3 id="health-response" className="text-lg font-semibold text-text-heading mt-8 mb-2">Response Format</h3>
      <CodeBlock lang="json" code={`{
  "success": true,
  "message": "Health check",
  "data": {
    "status": "healthy",
    "checks": {
      "postgres": { "status": "healthy", "duration_ms": 2 },
      "redis": { "status": "healthy", "duration_ms": 1 },
      "smtp": { "status": "degraded", "duration_ms": 150, "error": "timeout" }
    },
    "timestamp": 1700000000
  }
}`} />
    </ModuleSection>
  );
}
