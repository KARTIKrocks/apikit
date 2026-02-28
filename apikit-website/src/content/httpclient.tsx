import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function HttpclientDocs() {
  return (
    <ModuleSection
      id="httpclient"
      title="httpclient"
      description="HTTP client with retries, exponential backoff, circuit breaker, and an interface for easy mocking."
      importPath="github.com/KARTIKrocks/apikit/httpclient"
      features={[
        'Automatic retries with exponential backoff',
        'Circuit breaker pattern (opens after N failures)',
        'Fluent request builder',
        'Response helpers: JSON decoding, status checks',
        'Default headers and bearer token support',
        'HTTPClient interface for easy mocking in tests',
        'MockClient with call recording and assertions',
      ]}
    >
      <h3 id="httpclient-creating" className="text-lg font-semibold text-text-heading mt-8 mb-2">Creating a Client</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function / Option</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New(baseURL, opts...)</td><td className="py-2 text-text-muted">Create a new HTTP client</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithTimeout(d)</td><td className="py-2 text-text-muted">Request timeout</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithMaxRetries(n)</td><td className="py-2 text-text-muted">Max retry attempts</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithRetryDelay(d) / WithMaxRetryDelay(d)</td><td className="py-2 text-text-muted">Retry delay / cap</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithLogger(logger)</td><td className="py-2 text-text-muted">Structured logger</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`client := httpclient.New("https://api.example.com",
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithMaxRetries(3),
    httpclient.WithRetryDelay(500 * time.Millisecond),
    httpclient.WithMaxRetryDelay(5 * time.Second),
    httpclient.WithLogger(slog.Default()),
)`} />

      <h3 id="httpclient-methods" className="text-lg font-semibold text-text-heading mt-8 mb-2">HTTP Methods</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Get / .Post / .Put / .Patch / .Delete</td><td className="py-2 text-text-muted">HTTP methods with auto JSON body encoding</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`resp, err := client.Get(ctx, "/users")
resp, err := client.Post(ctx, "/users", map[string]string{"name": "Alice"})
resp, err := client.Put(ctx, "/users/1", updateReq)
resp, err := client.Patch(ctx, "/users/1", patchReq)
resp, err := client.Delete(ctx, "/users/1")`} />

      <h3 id="httpclient-response" className="text-lg font-semibold text-text-heading mt-8 mb-2">Response Handling</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">resp.JSON(&v)</td><td className="py-2 text-text-muted">Decode JSON into v</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">resp.String()</td><td className="py-2 text-text-muted">Body as string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">resp.StatusCode / resp.IsSuccess()</td><td className="py-2 text-text-muted">Status code / 2xx check</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`resp, err := client.Get(ctx, "/users")
if !resp.IsSuccess() {
    return fmt.Errorf("API error: %d", resp.StatusCode)
}
var users []User
resp.JSON(&users)`} />

      <h3 id="httpclient-builder" className="text-lg font-semibold text-text-heading mt-8 mb-2">Request Builder</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Request().Method().Path().Header().Param().Body().Send(ctx)</td><td className="py-2 text-text-muted">Fluent request builder chain</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`resp, err := client.Request().
    Method("POST").
    Path("/search").
    Header("X-Custom", "value").
    Param("q", "golang").
    Body(searchReq).
    Send(ctx)`} />

      <h3 id="httpclient-headers" className="text-lg font-semibold text-text-heading mt-8 mb-2">Default Headers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.SetBearerToken(token)</td><td className="py-2 text-text-muted">Set Authorization: Bearer</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.SetHeader(key, val)</td><td className="py-2 text-text-muted">Set default header</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`client.SetBearerToken("my-jwt-token")
client.SetHeader("X-API-Key", "key-123")`} />

      <h3 id="httpclient-circuit" className="text-lg font-semibold text-text-heading mt-8 mb-2">Circuit Breaker</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Option</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithCircuitBreaker(threshold, timeout)</td><td className="py-2 text-text-muted">Open after N failures, reset after timeout</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`client := httpclient.New("https://api.example.com",
    httpclient.WithCircuitBreaker(5, 30 * time.Second),
)
// Open → fail fast → half-open → test request → close/reopen`} />

      <h3 id="httpclient-mocking" className="text-lg font-semibold text-text-heading mt-8 mb-2">Mocking</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function / Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NewMockClient()</td><td className="py-2 text-text-muted">Create mock (implements HTTPClient)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnGet / .OnPost / .OnError</td><td className="py-2 text-text-muted">Mock responses</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.GetCallCount()</td><td className="py-2 text-text-muted">Number of recorded calls</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`mock := httpclient.NewMockClient()
mock.OnGet("/users", 200, []byte(\`[{"id":1,"name":"Alice"}]\`))
mock.OnPost("/users", 201, []byte(\`{"id":2}\`))
mock.OnError("GET", "/fail", fmt.Errorf("connection refused"))

users, err := fetchUsers(mock)
fmt.Println(mock.GetCallCount()) // 1`} />
    </ModuleSection>
  );
}
