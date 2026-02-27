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
      apiTable={[
        { name: 'NewChecker(opts...)', description: 'Create a health checker' },
        { name: 'WithTimeout(d)', description: 'Set per-check timeout' },
        { name: 'AddCheck(name, fn)', description: 'Add a critical dependency check' },
        { name: 'AddNonCriticalCheck(name, fn)', description: 'Add a non-critical check (degraded on failure)' },
        { name: 'Handler()', description: 'HTTP handler for readiness probe' },
        { name: 'LiveHandler()', description: 'HTTP handler for liveness probe (always 200)' },
        { name: 'Check(ctx)', description: 'Run checks programmatically' },
      ]}
    >
      <CodeBlock code={`h := health.NewChecker(health.WithTimeout(3 * time.Second))

// Critical checks — failure → "unhealthy" (503)
h.AddCheck("postgres", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

// Non-critical checks — failure → "degraded" (200)
h.AddNonCriticalCheck("redis", func(ctx context.Context) error {
    return rdb.Ping(ctx).Err()
})

// Register with your router
r.Get("/health", h.Handler())          // Full check (readiness)
r.Get("/health/live", h.LiveHandler()) // Always 200 (liveness)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Response Format</h3>
      <CodeBlock lang="json" code={`{
  "success": true,
  "message": "Health check",
  "data": {
    "status": "healthy",
    "checks": {
      "postgres": { "status": "healthy", "duration_ms": 2 },
      "redis": { "status": "healthy", "duration_ms": 1 }
    },
    "timestamp": 1700000000
  }
}`} />
    </ModuleSection>
  );
}
