import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ResponseDocs() {
  return (
    <ModuleSection
      id="response"
      title="response"
      description="Consistent JSON responses with a standard envelope."
      importPath="github.com/KARTIKrocks/apikit/response"
      features={[
        'Standard JSON envelope with success, message, data, meta',
        'One-liner success and error responses',
        'Fluent builder pattern for complex responses',
        'Offset and cursor pagination helpers',
        'SSE streaming, XML, JSONP, HTML, plain text, raw bytes',
        'response.Handle wraps error-returning handlers into http.HandlerFunc',
        'Link header (RFC 5988) for paginated APIs',
      ]}
      apiTable={[
        { name: 'OK(w, msg, data)', description: '200 success response' },
        { name: 'Created(w, msg, data)', description: '201 created response' },
        { name: 'NoContent(w)', description: '204 no content' },
        { name: 'Err(w, err)', description: 'Send error with correct status code' },
        { name: 'BadRequest(w, msg)', description: '400 error response' },
        { name: 'NotFound(w, msg)', description: '404 error response' },
        { name: 'ValidationError(w, fields)', description: '422 with field errors' },
        { name: 'Paginated(w, data, meta)', description: 'Paginated response with metadata' },
        { name: 'StreamJSON(w, fn)', description: 'Server-Sent Events streaming' },
        { name: 'Handle(fn)', description: 'Wrap error-returning handler' },
      ]}
    >
      <CodeBlock code={`// Success responses
response.OK(w, "Success", data)          // 200
response.Created(w, "Created", data)     // 201
response.NoContent(w)                    // 204

// Error responses
response.BadRequest(w, "Invalid input")
response.NotFound(w, "User not found")
response.ValidationError(w, map[string]string{"email": "required"})

// From errors package (recommended)
response.Err(w, errors.NotFound("User"))
response.Err(w, err) // *errors.Error gets proper status, others get 500`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Builder Pattern</h3>
      <CodeBlock code={`response.New().
    Status(201).
    Message("Created").
    Data(user).
    Header("X-Resource-ID", user.ID).
    Pagination(1, 20, 150).
    Send(w)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Pagination</h3>
      <CodeBlock code={`// Offset pagination
response.Paginated(w, users, response.NewPageMeta(page, perPage, total))

// Link header (RFC 5988)
response.SetLinkHeader(w, "https://api.example.com/users", page, perPage, total)

// Cursor pagination
response.CursorPaginated(w, events, response.CursorMeta{
    NextCursor: "eyJpZCI6MTAwfQ==",
    HasMore:    true,
})`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Response Envelope</h3>
      <CodeBlock lang="json" code={`{
  "success": true,
  "message": "User created",
  "data": { "id": "123", "name": "Alice" },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  },
  "timestamp": 1700000000
}`} />
    </ModuleSection>
  );
}
