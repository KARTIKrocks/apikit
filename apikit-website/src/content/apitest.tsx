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
      apiTable={[
        { name: 'NewRequest(method, path)', description: 'Start building a test request' },
        { name: '.WithBody(v)', description: 'Set JSON request body' },
        { name: '.WithBearerToken(token)', description: 'Set Authorization header' },
        { name: '.WithHeader(k, v)', description: 'Set a request header' },
        { name: '.WithQuery(k, v)', description: 'Add query parameter' },
        { name: '.WithPathValue(k, v)', description: 'Set path parameter (Go 1.22+)' },
        { name: '.Build()', description: 'Build the http.Request' },
        { name: 'RecordHandler(handler, req)', description: 'Record handler response' },
        { name: '.AssertStatus(t, code)', description: 'Assert HTTP status code' },
        { name: '.AssertSuccess(t)', description: 'Assert success: true in envelope' },
        { name: '.AssertError(t, code)', description: 'Assert error code in response' },
        { name: '.Decode(v)', description: 'Decode response data into struct' },
      ]}
    >
      <CodeBlock code={`// Build a request
req := apitest.NewRequest("POST", "/users").
    WithBody(map[string]string{"name": "Alice"}).
    WithBearerToken("valid-token").
    WithHeader("X-Request-ID", "test-123").
    WithQuery("notify", "true").
    WithPathValue("id", "42").
    Build()

// Record handler response
resp := apitest.RecordHandler(createUser, req)

// Fluent assertions
resp.AssertStatus(t, 201)
resp.AssertSuccess(t)
resp.AssertHeader(t, "X-Request-ID", "test-123")
resp.AssertBodyContains(t, "Alice")
resp.AssertError(t, "NOT_FOUND")
resp.AssertValidationError(t, "email")

// Decode response
var user User
resp.Decode(&user)

// Access envelope directly
env, _ := resp.Envelope()
fmt.Println(env.Success, env.Message)`} />
    </ModuleSection>
  );
}
