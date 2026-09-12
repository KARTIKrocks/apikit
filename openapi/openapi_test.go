package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/KARTIKrocks/apikit/router"
)

type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"omitempty,oneof=admin user mod"`
	Age   *int   `json:"age,omitempty" validate:"omitempty,gte=18"`
}

type Address struct {
	City string `json:"city"`
}

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Address Address  `json:"address"`
	Tags    []string `json:"tags"`
	Owner   *Address `json:"owner,omitempty"`
}

func decode(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, body)
	}
	return m
}

func TestDocumentBasicShape(t *testing.T) {
	doc := New(Info{Title: "Test API", Version: "1.0.0"})
	doc.Add("get", "/users/{id}", Operation{
		Summary: "Get a user",
		Responses: map[int]Response{
			200: {Body: User{}},
		},
	})

	body, err := doc.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	m := decode(t, body)

	if m["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi 3.1.0, got %v", m["openapi"])
	}
	info := m["info"].(map[string]any)
	if info["title"] != "Test API" {
		t.Fatalf("unexpected info: %v", info)
	}

	paths := m["paths"].(map[string]any)
	path := paths["/users/{id}"].(map[string]any)
	get := path["get"].(map[string]any)
	if get["summary"] != "Get a user" {
		t.Fatalf("expected summary, got %v", get)
	}

	// Path param auto-detected from "{id}".
	params := get["parameters"].([]any)
	if len(params) != 1 {
		t.Fatalf("expected 1 auto-detected path param, got %v", params)
	}
	p := params[0].(map[string]any)
	if p["name"] != "id" || p["in"] != "path" || p["required"] != true {
		t.Fatalf("unexpected param: %v", p)
	}

	// Response body resolves through a $ref into components.schemas.User.
	resp200 := get["responses"].(map[string]any)["200"].(map[string]any)
	schema := resp200["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if schema["$ref"] != "#/components/schemas/User" {
		t.Fatalf("expected $ref to User, got %v", schema)
	}

	components := m["components"].(map[string]any)["schemas"].(map[string]any)
	userSchema := components["User"].(map[string]any)
	props := userSchema["properties"].(map[string]any)
	if _, ok := props["tags"]; !ok {
		t.Fatalf("expected 'tags' property, got %v", props)
	}
	addr := props["address"].(map[string]any)
	if addr["$ref"] != "#/components/schemas/Address" {
		t.Fatalf("expected nested struct to become its own component ref, got %v", addr)
	}
}

func TestRequestBodySchemaFromValidateTags(t *testing.T) {
	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("POST", "/users", Operation{
		Request:   CreateUserReq{},
		Responses: map[int]Response{201: {Body: User{}}},
	})

	body, _ := doc.JSON()
	m := decode(t, body)
	components := m["components"].(map[string]any)["schemas"].(map[string]any)
	reqSchema := components["CreateUserReq"].(map[string]any)

	required := toStringSlice(reqSchema["required"])
	assertContains(t, required, "name")
	assertContains(t, required, "email")
	assertMissing(t, required, "role") // omitempty validate rule
	assertMissing(t, required, "age")  // pointer field with omitempty

	props := reqSchema["properties"].(map[string]any)
	name := props["name"].(map[string]any)
	if name["minLength"] != float64(2) || name["maxLength"] != float64(100) {
		t.Fatalf("expected min/max length on name, got %v", name)
	}
	email := props["email"].(map[string]any)
	if email["format"] != "email" {
		t.Fatalf("expected email format, got %v", email)
	}
	role := props["role"].(map[string]any)
	enum := toStringSlice(role["enum"])
	assertContains(t, enum, "admin")
	assertContains(t, enum, "user")
	assertContains(t, enum, "mod")

	age := props["age"].(map[string]any)
	if age["minimum"] != float64(18) {
		t.Fatalf("expected minimum 18 on age, got %v", age)
	}
	// age is *int -> nullable -> type folded to ["integer","null"]
	ageType := toStringSlice(age["type"])
	assertContains(t, ageType, "integer")
	assertContains(t, ageType, "null")
}

func TestNullablePointerToStructUsesAnyOf(t *testing.T) {
	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("GET", "/users/{id}", Operation{Responses: map[int]Response{200: {Body: User{}}}})

	body, _ := doc.JSON()
	m := decode(t, body)
	components := m["components"].(map[string]any)["schemas"].(map[string]any)
	userSchema := components["User"].(map[string]any)
	owner := userSchema["properties"].(map[string]any)["owner"].(map[string]any)

	if _, hasRef := owner["$ref"]; hasRef {
		t.Fatalf("expected a nullable *Address field to be wrapped in anyOf, not a bare $ref: %v", owner)
	}
	anyOf, ok := owner["anyOf"].([]any)
	if !ok || len(anyOf) != 2 {
		t.Fatalf("expected anyOf with 2 entries, got %v", owner)
	}
	ref := anyOf[0].(map[string]any)
	if ref["$ref"] != "#/components/schemas/Address" {
		t.Fatalf("expected first anyOf entry to ref Address, got %v", ref)
	}
	null := anyOf[1].(map[string]any)
	if null["type"] != "null" {
		t.Fatalf("expected second anyOf entry to be type null, got %v", null)
	}
}

func TestJSONCacheInvalidatedByAdd(t *testing.T) {
	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("GET", "/a", Operation{Summary: "first"})

	first, err := doc.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	if !strings.Contains(string(first), `/a`) {
		t.Fatalf("expected /a in first render, got:\n%s", first)
	}

	// A later Add must invalidate the cached render from the call above.
	doc.Add("GET", "/b", Operation{Summary: "second"})
	second, err := doc.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	if !strings.Contains(string(second), `/b`) {
		t.Fatalf("expected /b in second render after Add, got:\n%s", second)
	}
}

func TestSyncFillsUndescribedRoutes(t *testing.T) {
	r := router.New()
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) error { return nil })
	r.Post("/users", func(w http.ResponseWriter, req *http.Request) error { return nil })

	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("POST", "/users", Operation{Summary: "Create a user"}) // already described
	doc.Sync(r)

	body, _ := doc.JSON()
	m := decode(t, body)
	paths := m["paths"].(map[string]any)

	if _, ok := paths["/health"]; !ok {
		t.Fatalf("expected /health to be added by Sync, got paths: %v", paths)
	}
	users := paths["/users"].(map[string]any)["post"].(map[string]any)
	if users["summary"] != "Create a user" {
		t.Fatalf("Sync must not overwrite an existing Add description, got %v", users)
	}
}

func TestHandlerServesJSON(t *testing.T) {
	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("GET", "/ping", Operation{Responses: map[int]Response{200: {Description: "ok"}}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	doc.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", ct)
	}
	decode(t, rec.Body.Bytes())
}

func TestSwaggerUIHandlerServesHTML(t *testing.T) {
	doc := New(Info{Title: "My <API>", Version: "1"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/docs", nil)
	doc.SwaggerUIHandler("/openapi.json")(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `url: "/openapi.json"`) {
		t.Fatalf("expected spec URL in page, got:\n%s", body)
	}
	if strings.Contains(body, "<API>") {
		t.Fatalf("expected title to be HTML-escaped, got:\n%s", body)
	}
}

func TestNoResponsesGetsDefault(t *testing.T) {
	doc := New(Info{Title: "T", Version: "1"})
	doc.Add("GET", "/noop", Operation{})

	body, _ := doc.JSON()
	m := decode(t, body)
	op := m["paths"].(map[string]any)["/noop"].(map[string]any)["get"].(map[string]any)
	responses := op["responses"].(map[string]any)
	if _, ok := responses["default"]; !ok {
		t.Fatalf("expected a default response entry, got %v", responses)
	}
}

// --- helpers ---

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	arr := v.([]any)
	out := make([]string, len(arr))
	for i, x := range arr {
		out[i] = x.(string)
	}
	return out
}

func assertContains(t *testing.T, list []string, want string) {
	t.Helper()
	if !slices.Contains(list, want) {
		t.Fatalf("expected %v to contain %q", list, want)
	}
}

func assertMissing(t *testing.T, list []string, unwanted string) {
	t.Helper()
	for _, v := range list {
		if v == unwanted {
			t.Fatalf("expected %v not to contain %q", list, unwanted)
		}
	}
}
