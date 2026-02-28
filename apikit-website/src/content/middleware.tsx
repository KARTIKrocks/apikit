import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function MiddlewareDocs() {
  return (
    <ModuleSection
      id="middleware"
      title="middleware"
      description="Production-ready middleware that works with any net/http router."
      importPath="github.com/KARTIKrocks/apikit/middleware"
      features={[
        'Request ID generation and propagation',
        'Structured logging with slog',
        'Panic recovery with stack traces',
        'CORS with configurable origins',
        'Rate limiting with pluggable backends',
        'Bearer token authentication with context injection',
        'Security headers, timeout, body limit',
        'Middleware composition with Chain',
      ]}
    >
      <h3 id="middleware-core" className="text-lg font-semibold text-text-heading mt-8 mb-2">Core Middleware</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">RequestID()</td><td className="py-2 text-text-muted">Generate unique request ID, set X-Request-ID header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Logger(logger)</td><td className="py-2 text-text-muted">Log request method, path, status, duration via slog</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Recover()</td><td className="py-2 text-text-muted">Recover from panics, return 500, log stack trace</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SecureHeaders()</td><td className="py-2 text-text-muted">Add security headers (X-Content-Type-Options, etc.)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Timeout(d)</td><td className="py-2 text-text-muted">Cancel request context after duration</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BodyLimit(n)</td><td className="py-2 text-text-muted">Limit request body to n bytes</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">GetRequestID(r)</td><td className="py-2 text-text-muted">Retrieve request ID from context</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.Use(
    middleware.RequestID(),
    middleware.Logger(slog.Default()),
    middleware.Recover(),
    middleware.SecureHeaders(),
    middleware.Timeout(30 * time.Second),
    middleware.BodyLimit(5 << 20), // 5 MB
)

// Read request ID in handlers
reqID := middleware.GetRequestID(r)`} />

      <h3 id="middleware-auth" className="text-lg font-semibold text-text-heading mt-8 mb-2">Authentication</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Auth(config)</td><td className="py-2 text-text-muted">Bearer token authentication middleware</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">GetAuthUserAs[T](ctx)</td><td className="py-2 text-text-muted">Retrieve authenticated user from context (generic)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`auth := middleware.Auth(middleware.AuthConfig{
    Authenticate: func(ctx context.Context, token string) (any, error) {
        user, err := verifyJWT(token)
        if err != nil {
            return nil, errors.Unauthorized("Invalid token")
        }
        return user, nil
    },
    SkipPaths: map[string]bool{"/health": true, "/login": true},
})

api := r.Group("/api", auth)

// Retrieve the authenticated user in handlers (type-safe)
user, ok := middleware.GetAuthUserAs[*User](r.Context())`} />

      <h3 id="middleware-ratelimit" className="text-lg font-semibold text-text-heading mt-8 mb-2">Rate Limiting</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">RateLimit(config)</td><td className="py-2 text-text-muted">Per-IP rate limiting middleware</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Rate:   100,           // max requests
    Window: time.Minute,   // per window
}))
// Returns 429 Too Many Requests when limit exceeded`} />

      <h3 id="middleware-cors" className="text-lg font-semibold text-text-heading mt-8 mb-2">CORS</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">CORS(config)</td><td className="py-2 text-text-muted">Cross-Origin Resource Sharing middleware</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">DefaultCORSConfig()</td><td className="py-2 text-text-muted">Sensible default CORS config</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`// Default CORS (allows all origins)
r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

// Custom CORS
r.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"https://app.example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
    MaxAge:       3600,
}))`} />

      <h3 id="middleware-composition" className="text-lg font-semibold text-text-heading mt-8 mb-2">Composition</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Chain(mw...)</td><td className="py-2 text-text-muted">Compose multiple middleware into a single wrapper</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`stack := middleware.Chain(
    middleware.RequestID(),
    middleware.Logger(slog.Default()),
    middleware.Recover(),
    middleware.SecureHeaders(),
    middleware.CORS(middleware.DefaultCORSConfig()),
    middleware.RateLimit(middleware.RateLimitConfig{
        Rate: 100, Window: time.Minute,
    }),
    middleware.Timeout(30 * time.Second),
    middleware.BodyLimit(5 << 20),
)
handler := stack(mux)`} />
    </ModuleSection>
  );
}
