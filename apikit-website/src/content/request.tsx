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
      <h3 className="text-lg font-semibold text-text-heading mb-2">Body Binding</h3>
      <CodeBlock code={`type CreatePostReq struct {
    Title string \`json:"title"\`
    Body  string \`json:"body"\`
}

// Generic binding — auto-detects JSON, form, or multipart
post, err := request.Bind[CreatePostReq](r)

// Explicit JSON binding
post, err := request.BindJSON[CreatePostReq](r)

// Form binding (application/x-www-form-urlencoded)
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

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Parameters & Query</h3>
      <CodeBlock code={`// Path parameters (Go 1.22+ stdlib routing)
// Route: "GET /posts/{id}"
id := request.PathParam(r, "id")
id, err := request.PathParamInt(r, "id")

// Query parameters with type conversion and defaults
q := request.QueryFrom(r)
search := q.String("search", "")
page, err := q.Int("page", 1)
active, err := q.Bool("active", true)
tags := q.StringSlice("tags")                   // ?tags=go,api
limit, err := q.IntRange("limit", 20, 1, 100)   // Clamped to [1, 100]

// Headers
token := request.BearerToken(r)
ip := request.ClientIP(r)
reqID := request.RequestID(r)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Pagination, Sorting & Filtering</h3>
      <CodeBlock code={`// Offset pagination
pg, err := request.Paginate(r)  // ?page=2&per_page=25
pg.HasNext(total)
pg.TotalPages(total)
pg.SQLClause()  // "LIMIT 25 OFFSET 25"

// Cursor pagination
cursor, err := request.PaginateCursor(r)  // ?cursor=abc&limit=25

// Sorting — allowlist-based
sorts, err := request.ParseSort(r, request.SortConfig{
    AllowedFields: []string{"name", "created_at"},
})
// ?sort=name,-created_at → [{name, asc}, {created_at, desc}]

// Filtering
filters, err := request.ParseFilters(r, request.FilterConfig{
    AllowedFields: []string{"status", "role"},
})
// ?filter[status]=active&filter[age][gte]=18`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Validation</h3>
      <CodeBlock code={`// Struct tags (automatic in Bind[T])
type CreateUserReq struct {
    Name  string \`json:"name" validate:"required,min=2,max=100"\`
    Email string \`json:"email" validate:"required,email"\`
    Role  string \`json:"role" validate:"oneof=admin user mod"\`
}

// Programmatic validation (for cross-field logic)
v := request.NewValidation()
v.RequireString("name", req.Name)
v.RequireEmail("email", req.Email)
v.MinLength("name", req.Name, 2)
v.OneOf("role", req.Role, []string{"admin", "user", "mod"})
v.Custom("end_date", func() bool {
    return req.EndDate.After(req.StartDate)
}, "must be after start_date")
if err := v.Error(); err != nil {
    return err // Returns structured 422
}`} />
    </ModuleSection>
  );
}
