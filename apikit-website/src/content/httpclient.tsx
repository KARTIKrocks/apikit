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
      apiTable={[
        { name: 'New(baseURL, opts...)', description: 'Create a new HTTP client' },
        { name: 'Get/Post/Put/Patch/Delete(ctx, path, ...)', description: 'HTTP methods' },
        { name: 'Request().Method().Path().Body().Send(ctx)', description: 'Fluent request builder' },
        { name: 'SetBearerToken(token)', description: 'Set default Authorization header' },
        { name: 'WithCircuitBreaker(threshold, timeout)', description: 'Enable circuit breaker' },
        { name: 'NewMockClient()', description: 'Create a mock client for testing' },
      ]}
    >
      <CodeBlock code={`client := httpclient.New("https://api.example.com",
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithMaxRetries(3),
    httpclient.WithRetryDelay(500 * time.Millisecond),
    httpclient.WithMaxRetryDelay(5 * time.Second),
    httpclient.WithLogger(slog.Default()),
)

// Basic requests
resp, err := client.Get(ctx, "/users")
resp, err := client.Post(ctx, "/users", map[string]string{"name": "Alice"})

// Response helpers
var users []User
resp.JSON(&users)
fmt.Println(resp.String(), resp.StatusCode, resp.IsSuccess())

// Default headers
client.SetBearerToken("my-token")
client.SetHeader("X-API-Key", "key")`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Fluent Request Builder</h3>
      <CodeBlock code={`resp, err := client.Request().
    Method("POST").
    Path("/search").
    Header("X-Custom", "value").
    Param("q", "golang").
    Body(searchReq).
    Send(ctx)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Circuit Breaker</h3>
      <CodeBlock code={`// Opens after 5 failures, resets after 30s
client := httpclient.New("https://api.example.com",
    httpclient.WithCircuitBreaker(5, 30 * time.Second),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Mocking in Tests</h3>
      <CodeBlock code={`mock := httpclient.NewMockClient()
mock.OnGet("/users", 200, []byte(\`[{"id":1,"name":"Alice"}]\`))
mock.OnPost("/users", 201, []byte(\`{"id":2}\`))
mock.OnError("GET", "/fail", fmt.Errorf("connection refused"))

users, err := fetchUsers(mock)
fmt.Println(mock.GetCallCount()) // 1`} />
    </ModuleSection>
  );
}
