import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function OpenapiDocs() {
  return (
    <ModuleSection
      id="openapi"
      title="openapi"
      description="Generates an OpenAPI 3.1 document from the same json/validate struct tags request.Bind[T] already uses — no annotation comments or CLI generator required, and zero third-party dependencies."
      importPath="github.com/KARTIKrocks/apikit/openapi"
      features={[
        'Request/response schemas derived from json + validate struct tags',
        'Document.Sync fills in any route the router knows about that Add hasn’t described yet',
        'Path parameters (e.g. "{id}") detected from the route pattern automatically',
        'Rendered JSON is cached after the first call, invalidated by Add/Sync',
        'Swagger UI handler included (assets load from a CDN, no Go dependency)',
        'Targets OpenAPI 3.1 / JSON Schema 2020-12 — nullable, enum, format, and length/range constraints',
      ]}
    >
      <h3 id="openapi-basics" className="text-lg font-semibold text-text-heading mt-8 mb-2">Describing Operations</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New(Info{'{'}...{'}'})</td><td className="py-2 text-text-muted">Create a document with title/version/description</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Add(method, path, Operation{'{'}...{'}'})</td><td className="py-2 text-text-muted">Describe one operation's summary, tags, request, and responses</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`type CreateUserReq struct {
    Name  string \`json:"name"  validate:"required,min=2,max=100"\`
    Email string \`json:"email" validate:"required,email"\`
    Role  string \`json:"role"  validate:"omitempty,oneof=admin user mod"\`
}

doc := openapi.New(openapi.Info{Title: "My API", Version: "1.0.0"})

doc.Add("POST", "/users", openapi.Operation{
    Summary: "Create a user",
    Tags:    []string{"users"},
    Request: CreateUserReq{}, // schema derived by reflection, not by calling this
    Responses: map[int]openapi.Response{
        201: {Description: "Created", Body: User{}},
        422: {Description: "Validation failed"},
    },
})

// Path params in the route pattern (e.g. "{id}") are added automatically.
doc.Add("GET", "/users/{id}", openapi.Operation{
    Summary:   "Get a user",
    Responses: map[int]openapi.Response{200: {Body: User{}}},
})`} />

      <h3 id="openapi-schema" className="text-lg font-semibold text-text-heading mt-8 mb-2">Struct Tag &rarr; Schema Mapping</h3>
      <p className="text-text-muted mb-3 text-sm">Mirrors <code className="font-mono text-accent">request.ValidateStruct</code>'s rule set exactly, so a type is described once.</p>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">validate tag</th><th className="py-2 text-text-heading font-semibold">Schema effect</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">required</td><td className="py-2 text-text-muted">Added to the parent's required list</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">min=N / max=N</td><td className="py-2 text-text-muted">minLength/maxLength (string/slice/map) or minimum/maximum (numeric)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">gte=N / lte=N</td><td className="py-2 text-text-muted">Alias of min/max</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">gt=N / lt=N</td><td className="py-2 text-text-muted">exclusiveMinimum/exclusiveMaximum</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">oneof=a b c</td><td className="py-2 text-text-muted">enum</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">email / url / uuid</td><td className="py-2 text-text-muted">format</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">e164 / alpha / contains=X / ...</td><td className="py-2 text-text-muted">pattern (regex)</td></tr>
        </tbody></table>
      </div>
      <p className="text-text-muted text-sm">
        Nested structs become named components (<code className="font-mono text-accent">#/components/schemas/...</code>),
        embedded structs are flattened like <code className="font-mono text-accent">encoding/json</code> promotes them, and a
        nullable pointer-to-struct field is encoded as <code className="font-mono text-accent">anyOf: [$ref, {'{'}type: null{'}'}]</code> rather
        than a sibling type keyword next to <code className="font-mono text-accent">$ref</code> &mdash; not every tool honors that combination.
      </p>

      <h3 id="openapi-sync" className="text-lg font-semibold text-text-heading mt-8 mb-2">Staying in Sync with the Router</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Sync(r)</td><td className="py-2 text-text-muted">Add a bare entry for any route the router knows about that Add hasn't described</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`// Fill in any route registered on r that Add hasn't described yet, so the
// spec can never silently drift out of sync with the router.
doc.Sync(r)`} />

      <h3 id="openapi-serving" className="text-lg font-semibold text-text-heading mt-8 mb-2">Serving the Spec</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.JSON()</td><td className="py-2 text-text-muted">Render the document as JSON bytes</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Handler()</td><td className="py-2 text-text-muted">http.HandlerFunc serving the rendered spec (cached after the first call)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.SwaggerUIHandler(specURL)</td><td className="py-2 text-text-muted">Serves a Swagger UI page pointed at specURL (assets load from a CDN)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.GetFunc("/openapi.json", doc.Handler())
r.GetFunc("/docs", doc.SwaggerUIHandler("/openapi.json"))`} />
    </ModuleSection>
  );
}
