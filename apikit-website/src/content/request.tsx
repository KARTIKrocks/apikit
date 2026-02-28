import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function RequestDocs() {
  return (
    <ModuleSection
      id="request"
      title="request"
      description="Generic request binding and parameter parsing."
      importPath="github.com/KARTIKrocks/apikit/request"
      features={[
        'Generic body binding: JSON, form, multipart with Bind[T]',
        'Struct tag validation (required, email, min, max, oneof, etc.)',
        'Path parameter parsing (Go 1.22+ stdlib routing)',
        'Query parameter helpers with type conversion and defaults',
        'Pagination (offset + cursor), sorting, and filtering',
        'Programmatic validation with fluent API',
        'File upload helpers',
      ]}
    >
      <h3 id="request-binding" className="text-lg font-semibold text-text-heading mt-8 mb-2">Body Binding</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Bind[T](r)</td><td className="py-2 text-text-muted">Auto-detect content type and bind to struct</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BindJSON[T](r)</td><td className="py-2 text-text-muted">Bind JSON request body</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BindForm[T](r)</td><td className="py-2 text-text-muted">Bind application/x-www-form-urlencoded body</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BindMultipart[T](r)</td><td className="py-2 text-text-muted">Bind multipart/form-data body</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">FormFile(r, key)</td><td className="py-2 text-text-muted">Get a single uploaded file</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">FormFiles(r)</td><td className="py-2 text-text-muted">Get all uploaded files</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`type CreatePostReq struct {
    Title string \`json:"title" validate:"required,min=3"\`
    Body  string \`json:"body"  validate:"required"\`
}

post, err := request.Bind[CreatePostReq](r)     // auto-detect
post, err := request.BindJSON[CreatePostReq](r)  // explicit JSON

// Form binding
type ContactForm struct {
    Name    string \`form:"name"    validate:"required"\`
    Email   string \`form:"email"   validate:"required,email"\`
    Message string \`form:"message" validate:"required,min=10"\`
}
form, err := request.BindForm[ContactForm](r)

// Multipart with file uploads
meta, err := request.BindMultipart[UploadForm](r)
fh, err := request.FormFile(r, "avatar")
allFiles := request.FormFiles(r)`} />

      <h3 id="request-params" className="text-lg font-semibold text-text-heading mt-8 mb-2">Path & Query Params</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">PathParam(r, key)</td><td className="py-2 text-text-muted">Get path parameter as string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">PathParamInt(r, key)</td><td className="py-2 text-text-muted">Get path parameter as int</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">QueryFrom(r)</td><td className="py-2 text-text-muted">Create a query helper from request</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">q.String(key, default)</td><td className="py-2 text-text-muted">Get query param as string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">q.Int(key, default)</td><td className="py-2 text-text-muted">Get query param as int</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">q.Bool(key, default)</td><td className="py-2 text-text-muted">Get query param as bool</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">q.IntRange(key, default, min, max)</td><td className="py-2 text-text-muted">Get int clamped to [min, max]</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">q.StringSlice(key)</td><td className="py-2 text-text-muted">Get comma-separated values as slice</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`id := request.PathParam(r, "id")
id, err := request.PathParamInt(r, "id")

q := request.QueryFrom(r)
search := q.String("search", "")
page, err := q.Int("page", 1)
active, err := q.Bool("active", true)
tags := q.StringSlice("tags")                  // ?tags=go,api
limit, err := q.IntRange("limit", 20, 1, 100)  // clamped to [1, 100]`} />

      <h3 id="request-headers" className="text-lg font-semibold text-text-heading mt-8 mb-2">Headers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BearerToken(r)</td><td className="py-2 text-text-muted">Extract Bearer token from Authorization header</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ClientIP(r)</td><td className="py-2 text-text-muted">Get client IP (checks X-Forwarded-For, X-Real-IP)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">RequestID(r)</td><td className="py-2 text-text-muted">Get request ID from context</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`token := request.BearerToken(r)
ip := request.ClientIP(r)
reqID := request.RequestID(r)`} />

      <h3 id="request-pagination" className="text-lg font-semibold text-text-heading mt-8 mb-2">Pagination</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Paginate(r)</td><td className="py-2 text-text-muted">Parse offset pagination (?page=2&per_page=25)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">PaginateCursor(r)</td><td className="py-2 text-text-muted">Parse cursor pagination (?cursor=abc&limit=25)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">pg.HasNext(total)</td><td className="py-2 text-text-muted">Check if there are more pages</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">pg.TotalPages(total)</td><td className="py-2 text-text-muted">Calculate total number of pages</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">pg.SQLClause()</td><td className="py-2 text-text-muted">Generate "LIMIT x OFFSET y" clause</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`pg, err := request.Paginate(r) // ?page=2&per_page=25
pg.HasNext(total)
pg.TotalPages(total)
pg.SQLClause() // "LIMIT 25 OFFSET 25"

cursor, err := request.PaginateCursor(r) // ?cursor=abc&limit=25`} />

      <h3 id="request-sorting" className="text-lg font-semibold text-text-heading mt-8 mb-2">Sorting</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ParseSort(r, config)</td><td className="py-2 text-text-muted">Parse sort fields from ?sort= (allowlist-based)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sorts, err := request.ParseSort(r, request.SortConfig{
    AllowedFields: []string{"name", "created_at", "email"},
})
// ?sort=name,-created_at → [{name, asc}, {created_at, desc}]`} />

      <h3 id="request-filtering" className="text-lg font-semibold text-text-heading mt-8 mb-2">Filtering</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ParseFilters(r, config)</td><td className="py-2 text-text-muted">Parse filters from ?filter[field][op]=value</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`filters, err := request.ParseFilters(r, request.FilterConfig{
    AllowedFields: []string{"status", "role", "age"},
})
// ?filter[status]=active&filter[age][gte]=18`} />

      <h3 id="request-struct-validation" className="text-lg font-semibold text-text-heading mt-8 mb-2">Struct Validation</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ValidateStruct(v)</td><td className="py-2 text-text-muted">Validate struct using validate tags</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`type CreateUserReq struct {
    Name  string \`json:"name"  validate:"required,min=2,max=100"\`
    Email string \`json:"email" validate:"required,email"\`
    Role  string \`json:"role"  validate:"oneof=admin user mod"\`
    Age   int    \`json:"age"   validate:"gte=0,lte=150"\`
}
// Supported: required, email, url, min, max, len, gte, lte, oneof, uuid, alpha, numeric`} />

      <h3 id="request-programmatic-validation" className="text-lg font-semibold text-text-heading mt-8 mb-2">Programmatic Validation</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NewValidation()</td><td className="py-2 text-text-muted">Create a validation builder</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.RequireString(field, val)</td><td className="py-2 text-text-muted">Require non-empty string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.RequireEmail(field, val)</td><td className="py-2 text-text-muted">Require valid email</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.MinLength / .MaxLength</td><td className="py-2 text-text-muted">String length constraints</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OneOf(field, val, options)</td><td className="py-2 text-text-muted">Value must be in allowed list</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Custom(field, fn, msg)</td><td className="py-2 text-text-muted">Custom validation function</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Error()</td><td className="py-2 text-text-muted">Return structured 422 error (nil if valid)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`v := request.NewValidation()
v.RequireString("name", req.Name)
v.RequireEmail("email", req.Email)
v.MinLength("name", req.Name, 2)
v.OneOf("role", req.Role, []string{"admin", "user", "mod"})
v.Custom("end_date", func() bool {
    return req.EndDate.After(req.StartDate)
}, "must be after start_date")
if err := v.Error(); err != nil {
    return err // structured 422
}`} />
    </ModuleSection>
  );
}
