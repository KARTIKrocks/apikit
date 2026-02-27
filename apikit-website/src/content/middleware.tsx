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
      ]}
      apiTable={[
        { name: 'RequestID()', description: 'Add unique request ID to each request' },
        { name: 'Logger(logger)', description: 'Log requests with duration, status, etc.' },
        { name: 'Recover()', description: 'Recover from panics with 500 response' },
        { name: 'CORS(config)', description: 'Cross-Origin Resource Sharing' },
        { name: 'RateLimit(config)', description: 'Rate limiting per client IP' },
        { name: 'Auth(config)', description: 'Bearer token authentication' },
        { name: 'Timeout(d)', description: 'Request timeout' },
        { name: 'BodyLimit(n)', description: 'Limit request body size' },
        { name: 'SecureHeaders()', description: 'Add security-related HTTP headers' },
        { name: 'Chain(mw...)', description: 'Compose middleware into a single wrapper' },
      ]}
    >
      <CodeBlock code={`// Chain middleware (applied in order)
stack := middleware.Chain(
    middleware.RequestID(),
    middleware.Logger(slog.Default()),
    middleware.Recover(),
    middleware.SecureHeaders(),
    middleware.CORS(middleware.DefaultCORSConfig()),
    middleware.RateLimit(middleware.RateLimitConfig{
        Rate: 100, Window: time.Minute,
    }),
    middleware.Timeout(30 * time.Second),
    middleware.BodyLimit(5 << 20), // 5 MB
)
handler := stack(mux)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Authentication</h3>
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

// In handlers, retrieve the authenticated user:
user, ok := middleware.GetAuthUserAs[*User](r.Context())`} />
    </ModuleSection>
  );
}
