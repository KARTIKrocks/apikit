package openapi

// Info describes the API metadata block of the document.
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Operation describes a single method+path entry. Request and the values in
// Responses are sample instances (usually zero values, e.g. CreateUserReq{})
// used only for their type — schemas are derived via reflection, never from
// the values themselves.
type Operation struct {
	Summary     string
	Description string
	Tags        []string

	// Request, when non-nil, becomes the JSON request body schema.
	Request any

	// Params declares path/query/header parameters explicitly. Path
	// parameters present in the route pattern (e.g. "{id}") but missing here
	// are added automatically as required string parameters.
	Params []Param

	// Responses maps HTTP status code to a response description and body type.
	Responses map[int]Response
}

// Param describes a single path, query, or header parameter.
type Param struct {
	Name        string
	In          string // "path", "query", or "header"
	Required    bool
	Description string

	// Schema, when nil, defaults to a plain string schema.
	Schema *Schema
}

// Response describes one entry in an operation's Responses map.
type Response struct {
	Description string

	// Body, when non-nil, becomes the JSON response body schema.
	Body any
}

// Schema is an OpenAPI 3.1 (JSON Schema 2020-12) schema object. Only the
// subset of keywords the reflection engine in schema.go actually produces is
// represented.
type Schema struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`

	Description string `json:"description,omitempty"`
	Format      string `json:"format,omitempty"`
	Pattern     string `json:"pattern,omitempty"`

	Items      *Schema            `json:"items,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`

	AdditionalProperties *Schema `json:"additionalProperties,omitempty"`

	// AnyOf is populated only by MarshalJSON, to express a nullable $ref
	// (see MarshalJSON's doc comment) — it's not set directly by schema.go.
	AnyOf []*Schema `json:"anyOf,omitempty"`

	Enum []string `json:"enum,omitempty"`

	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	MinLength        *int     `json:"minLength,omitempty"`
	MaxLength        *int     `json:"maxLength,omitempty"`

	Nullable bool `json:"-"` // folded into Type as ["<type>", "null"] at marshal time
}
