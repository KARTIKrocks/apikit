import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ApitestDocs() {
  return (
    <ModuleSection
      id="apitest"
      title="apitest"
      description="Fluent test helpers for building requests and asserting responses against your handlers."
      importPath="github.com/KARTIKrocks/apikit/apitest"
      features={[
        'Fluent request builder with body, headers, query, path values',
        'Record handler responses via httptest.ResponseRecorder',
        'Assertion helpers: status, success, headers, body, errors',
        'Decode response into typed structs',
        'Access response envelope directly',
      ]}
    >
      <h3 id="apitest-requests" className="text-lg font-semibold text-text-heading mt-8 mb-2">Building Requests</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function / Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NewRequest(method, path)</td><td className="py-2 text-text-muted">Start building a test request</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithBody(v)</td><td className="py-2 text-text-muted">Set JSON body</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithBearerToken(token)</td><td className="py-2 text-text-muted">Set Authorization header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithHeader(k, v)</td><td className="py-2 text-text-muted">Set request header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithQuery(k, v)</td><td className="py-2 text-text-muted">Add query parameter</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithPathValue(k, v)</td><td className="py-2 text-text-muted">Set path parameter (Go 1.22+)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Build()</td><td className="py-2 text-text-muted">Build *http.Request</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`req := apitest.NewRequest("POST", "/users").
    WithBody(map[string]string{"name": "Alice", "email": "alice@example.com"}).
    WithBearerToken("valid-token").
    WithHeader("X-Request-ID", "test-123").
    WithQuery("notify", "true").
    Build()

req := apitest.NewRequest("GET", "/users/{id}").
    WithPathValue("id", "42").Build()`} />

      <h3 id="apitest-recording" className="text-lg font-semibold text-text-heading mt-8 mb-2">Recording Responses</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">RecordHandler(handler, req)</td><td className="py-2 text-text-muted">Execute handler, return recorded response</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`resp := apitest.RecordHandler(createUser, req)`} />

      <h3 id="apitest-assertions" className="text-lg font-semibold text-text-heading mt-8 mb-2">Assertions</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertStatus(t, code)</td><td className="py-2 text-text-muted">Assert HTTP status code</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertSuccess(t)</td><td className="py-2 text-text-muted">Assert success: true</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertError(t, code)</td><td className="py-2 text-text-muted">Assert error code</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertHeader(t, key, val)</td><td className="py-2 text-text-muted">Assert header value</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertBodyContains(t, substr)</td><td className="py-2 text-text-muted">Assert body contains string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.AssertValidationError(t, field)</td><td className="py-2 text-text-muted">Assert validation error for field</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`func TestCreateUser(t *testing.T) {
    req := apitest.NewRequest("POST", "/users").
        WithBody(map[string]string{"name": "Alice", "email": "alice@example.com"}).
        Build()
    resp := apitest.RecordHandler(createUser, req)

    resp.AssertStatus(t, 201)
    resp.AssertSuccess(t)
    resp.AssertHeader(t, "Content-Type", "application/json")
    resp.AssertBodyContains(t, "Alice")
}

func TestCreateUser_Validation(t *testing.T) {
    req := apitest.NewRequest("POST", "/users").WithBody(map[string]string{}).Build()
    resp := apitest.RecordHandler(createUser, req)

    resp.AssertStatus(t, 422)
    resp.AssertError(t, "VALIDATION")
    resp.AssertValidationError(t, "email")
}

func TestGetUser_NotFound(t *testing.T) {
    req := apitest.NewRequest("GET", "/users/{id}").WithPathValue("id", "999").Build()
    resp := apitest.RecordHandler(getUser, req)

    resp.AssertStatus(t, 404)
    resp.AssertError(t, "NOT_FOUND")
}`} />

      <h3 id="apitest-decoding" className="text-lg font-semibold text-text-heading mt-8 mb-2">Decoding</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Decode(v)</td><td className="py-2 text-text-muted">Decode response data into struct</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Envelope()</td><td className="py-2 text-text-muted">Access full response envelope</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`var user User
resp.Decode(&user)
assert.Equal(t, "Alice", user.Name)

env, _ := resp.Envelope()
fmt.Println(env.Success, env.Message)`} />
    </ModuleSection>
  );
}
