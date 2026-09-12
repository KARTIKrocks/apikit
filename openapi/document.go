package openapi

import (
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/KARTIKrocks/apikit/router"
)

// Document accumulates operations and renders them as an OpenAPI 3.1 JSON
// document. The zero value is not usable; create one with New.
//
// A Document is safe for concurrent use: Add and Sync are typically called
// during startup while routes are registered, and JSON/Handler are safe to
// call concurrently from request goroutines afterward — the rendered JSON is
// cached after the first call and invalidated by Add/Sync, so repeated
// requests to a served spec don't re-run reflection on every hit.
type Document struct {
	mu    sync.Mutex
	info  Info
	paths map[string]map[string]Operation // path -> method -> Operation
	cache []byte                          // rendered JSON; nil when stale
}

// New creates an empty Document with the given metadata.
func New(info Info) *Document {
	return &Document{info: info, paths: make(map[string]map[string]Operation)}
}

// Add describes one method+path operation. Calling Add again for the same
// method and path overwrites the previous description.
func (d *Document) Add(method, path string, op Operation) {
	d.mu.Lock()
	defer d.mu.Unlock()

	m := normalizeMethod(method)
	if d.paths[path] == nil {
		d.paths[path] = make(map[string]Operation)
	}
	d.paths[path][m] = op
	d.cache = nil
}

// Sync walks r's registered routes and adds a bare (undescribed) operation
// for any method+path that Add hasn't already covered, so the document never
// drifts out of sync with the router — every real route shows up, even
// before someone fills in its request/response schema with Add.
//
// Routes registered without a specific method (Handle/HandleFunc, matching
// every method) have no single OpenAPI operation to represent them and are
// skipped.
func (d *Document) Sync(r *router.Router) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ri := range r.Routes() {
		if ri.Method == "" {
			continue
		}
		m := normalizeMethod(ri.Method)
		if d.paths[ri.Pattern] == nil {
			d.paths[ri.Pattern] = make(map[string]Operation)
		}
		if _, exists := d.paths[ri.Pattern][m]; exists {
			continue
		}
		d.paths[ri.Pattern][m] = Operation{Summary: summarizeHandlerName(ri.HandlerName)}
		d.cache = nil
	}
}

// JSON renders the accumulated operations as an OpenAPI 3.1 document.
func (d *Document) JSON() ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cache == nil {
		b, err := json.MarshalIndent(d.build(), "", "  ")
		if err != nil {
			return nil, err
		}
		d.cache = b
	}
	out := make([]byte, len(d.cache))
	copy(out, d.cache)
	return out, nil
}

// Handler serves the document as "application/json".
func (d *Document) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := d.JSON()
		if err != nil {
			http.Error(w, "openapi: failed to render document", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(body)
	}
}

// --- wire format ---

type wireDocument struct {
	OpenAPI    string                               `json:"openapi"`
	Info       Info                                 `json:"info"`
	Paths      map[string]map[string]*wireOperation `json:"paths"`
	Components *wireComponents                      `json:"components,omitempty"`
}

type wireOperation struct {
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	Tags        []string                 `json:"tags,omitempty"`
	Parameters  []*wireParam             `json:"parameters,omitempty"`
	RequestBody *wireRequestBody         `json:"requestBody,omitempty"`
	Responses   map[string]*wireResponse `json:"responses"`
}

type wireParam struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Required    bool    `json:"required,omitempty"`
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type wireRequestBody struct {
	Required bool                      `json:"required"`
	Content  map[string]*wireMediaType `json:"content"`
}

type wireMediaType struct {
	Schema *Schema `json:"schema"`
}

type wireResponse struct {
	Description string                    `json:"description"`
	Content     map[string]*wireMediaType `json:"content,omitempty"`
}

type wireComponents struct {
	Schemas map[string]*Schema `json:"schemas,omitempty"`
}

var pathParamRe = regexp.MustCompile(`\{([A-Za-z0-9_]+)(\.\.\.)?\}`)

func (d *Document) build() wireDocument {
	reg := newRegistry()
	paths := make(map[string]map[string]*wireOperation, len(d.paths))

	for path, methods := range d.paths {
		pathOut := make(map[string]*wireOperation, len(methods))
		pathParams := pathParamRe.FindAllStringSubmatch(path, -1)

		for method, op := range methods {
			pathOut[method] = buildOperation(reg, pathParams, op)
		}
		paths[path] = pathOut
	}

	doc := wireDocument{OpenAPI: "3.1.0", Info: d.info, Paths: paths}
	if len(reg.schemas) > 0 {
		doc.Components = &wireComponents{Schemas: reg.schemas}
	}
	return doc
}

func buildOperation(reg *registry, pathParams [][]string, op Operation) *wireOperation {
	out := &wireOperation{
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
	}

	declared := make(map[string]bool, len(op.Params))
	for _, p := range op.Params {
		declared[p.In+":"+p.Name] = true
		schema := p.Schema
		if schema == nil {
			schema = &Schema{Type: "string"}
		}
		out.Parameters = append(out.Parameters, &wireParam{
			Name: p.Name, In: p.In, Required: p.Required,
			Description: p.Description, Schema: schema,
		})
	}
	for _, m := range pathParams {
		name := m[1]
		if declared["path:"+name] {
			continue
		}
		out.Parameters = append(out.Parameters, &wireParam{
			Name: name, In: "path", Required: true, Schema: &Schema{Type: "string"},
		})
	}
	sort.Slice(out.Parameters, func(i, j int) bool {
		if out.Parameters[i].In != out.Parameters[j].In {
			return out.Parameters[i].In < out.Parameters[j].In
		}
		return out.Parameters[i].Name < out.Parameters[j].Name
	})

	if op.Request != nil {
		out.RequestBody = &wireRequestBody{
			Required: true,
			Content: map[string]*wireMediaType{
				"application/json": {Schema: reg.schemaFor(reflect.TypeOf(op.Request))},
			},
		}
	}

	out.Responses = make(map[string]*wireResponse, len(op.Responses))
	for code, resp := range op.Responses {
		wr := &wireResponse{Description: resp.Description}
		if wr.Description == "" {
			wr.Description = http.StatusText(code)
		}
		if resp.Body != nil {
			wr.Content = map[string]*wireMediaType{
				"application/json": {Schema: reg.schemaFor(reflect.TypeOf(resp.Body))},
			}
		}
		out.Responses[strconv.Itoa(code)] = wr
	}
	if len(out.Responses) == 0 {
		out.Responses["default"] = &wireResponse{Description: "Response"}
	}

	return out
}

func normalizeMethod(method string) string {
	return strings.ToLower(method)
}

func summarizeHandlerName(fn string) string {
	if fn == "" {
		return ""
	}
	// Trim package/receiver qualifiers, e.g. "main.createUser" -> "createUser".
	for i := len(fn) - 1; i >= 0; i-- {
		if fn[i] == '.' {
			return fn[i+1:]
		}
	}
	return fn
}
