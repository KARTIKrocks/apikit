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
    >
      <h3 id="response-success" className="text-lg font-semibold text-text-heading mt-8 mb-2">Success Responses</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">OK(w, msg, data)</td><td className="py-2 text-text-muted">200 success response</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Created(w, msg, data)</td><td className="py-2 text-text-muted">201 created response</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Accepted(w, msg, data)</td><td className="py-2 text-text-muted">202 accepted response</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NoContent(w)</td><td className="py-2 text-text-muted">204 no content</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`response.OK(w, "Users retrieved", users)   // 200
response.Created(w, "User created", user)  // 201
response.Accepted(w, "Job queued", job)     // 202
response.NoContent(w)                       // 204`} />

      <h3 id="response-error" className="text-lg font-semibold text-text-heading mt-8 mb-2">Error Responses</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Err(w, err)</td><td className="py-2 text-text-muted">Send error with correct status code</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BadRequest(w, msg)</td><td className="py-2 text-text-muted">400 error</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Unauthorized(w, msg)</td><td className="py-2 text-text-muted">401 error</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Forbidden(w, msg)</td><td className="py-2 text-text-muted">403 error</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NotFound(w, msg)</td><td className="py-2 text-text-muted">404 error</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ValidationError(w, fields)</td><td className="py-2 text-text-muted">422 with field errors</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`response.BadRequest(w, "Invalid input")
response.NotFound(w, "User not found")
response.ValidationError(w, map[string]string{"email": "required"})

// From errors package (recommended)
response.Err(w, errors.NotFound("User"))
response.Err(w, err) // *errors.Error gets proper status, others get 500`} />

      <h3 id="response-builder" className="text-lg font-semibold text-text-heading mt-8 mb-2">Builder Pattern</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New()</td><td className="py-2 text-text-muted">Create a response builder</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Status(code)</td><td className="py-2 text-text-muted">Set HTTP status</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Message(msg) / .Data(d)</td><td className="py-2 text-text-muted">Set envelope fields</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Header(k, v)</td><td className="py-2 text-text-muted">Set response header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Send(w)</td><td className="py-2 text-text-muted">Write the response</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`response.New().
    Status(201).
    Message("User created").
    Data(user).
    Header("X-Resource-ID", user.ID).
    Pagination(1, 20, 150).
    Send(w)`} />

      <h3 id="response-pagination" className="text-lg font-semibold text-text-heading mt-8 mb-2">Pagination</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Paginated(w, data, meta)</td><td className="py-2 text-text-muted">Paginated response with page meta</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NewPageMeta(page, perPage, total)</td><td className="py-2 text-text-muted">Create offset pagination metadata</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SetLinkHeader(w, baseURL, page, perPage, total)</td><td className="py-2 text-text-muted">Set RFC 5988 Link header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">CursorPaginated(w, data, meta)</td><td className="py-2 text-text-muted">Cursor-paginated response</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`response.Paginated(w, users, response.NewPageMeta(page, perPage, total))

response.SetLinkHeader(w, "https://api.example.com/users", page, perPage, total)

response.CursorPaginated(w, events, response.CursorMeta{
    NextCursor: "eyJpZCI6MTAwfQ==",
    HasMore:    true,
})`} />

      <h3 id="response-streaming" className="text-lg font-semibold text-text-heading mt-8 mb-2">Streaming & Formats</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StreamJSON(w, fn)</td><td className="py-2 text-text-muted">Server-Sent Events streaming</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">XML(w, status, data)</td><td className="py-2 text-text-muted">XML response</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">IndentedJSON / PureJSON / JSONP</td><td className="py-2 text-text-muted">Alternative JSON formats</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">HTML / Text / Raw / Reader</td><td className="py-2 text-text-muted">HTML, plain text, raw bytes, io.Reader</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`response.StreamJSON(w, func(send func(data any) error) error {
    for item := range ch {
        if err := send(item); err != nil { return err }
    }
    return nil
})

response.XML(w, 200, xmlData)
response.IndentedJSON(w, 200, data)
response.HTML(w, 200, "<h1>Hello</h1>")
response.Text(w, 200, "plain text")
response.Raw(w, 200, "application/pdf", pdfBytes)
response.Reader(w, 200, "image/png", imageReader)`} />

      <h3 id="response-handler" className="text-lg font-semibold text-text-heading mt-8 mb-2">Handler Wrapper</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Handle(fn)</td><td className="py-2 text-text-muted">Wrap error-returning handler into http.HandlerFunc</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`mux.HandleFunc("GET /users", response.Handle(func(w http.ResponseWriter, r *http.Request) error {
    users, err := db.ListUsers(r.Context())
    if err != nil { return err }
    response.OK(w, "Success", users)
    return nil
}))`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Response Envelope</h3>
      <CodeBlock lang="json" code={`{
  "success": true,
  "message": "User created",
  "data": { "id": "123", "name": "Alice" },
  "meta": {
    "page": 1, "per_page": 20, "total": 150,
    "total_pages": 8, "has_next": true, "has_previous": false
  },
  "timestamp": 1700000000
}`} />
    </ModuleSection>
  );
}
