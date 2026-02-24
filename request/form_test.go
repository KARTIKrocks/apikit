package request

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/KARTIKrocks/apikit/errors"
)

// --- decodeFormValues tests ---

type simpleForm struct {
	Name  string `form:"name"`
	Email string `form:"email"`
}

type typedForm struct {
	Str   string   `form:"str"`
	Bool  bool     `form:"bool"`
	Int   int      `form:"int"`
	Int8  int8     `form:"int8"`
	Int16 int16    `form:"int16"`
	Int32 int32    `form:"int32"`
	Int64 int64    `form:"int64"`
	Uint  uint     `form:"uint"`
	Uint8 uint8    `form:"uint8"`
	F32   float32  `form:"f32"`
	F64   float64  `form:"f64"`
	Tags  []string `form:"tags"`
}

type tagPriorityForm struct {
	NameForm string `form:"name_form" json:"name_json"`
	OnlyJSON string `json:"only_json"`
	NoTag    string
}

type skipForm struct {
	Name    string `form:"name"`
	Ignored string `form:"-"`
}

func TestDecodeFormValues_Simple(t *testing.T) {
	vals := url.Values{
		"name":  {"Alice"},
		"email": {"alice@example.com"},
	}
	var f simpleForm
	if err := decodeFormValues(vals, &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
	if f.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", f.Email, "alice@example.com")
	}
}

func TestDecodeFormValues_AllTypes(t *testing.T) {
	vals := url.Values{
		"str":   {"hello"},
		"bool":  {"true"},
		"int":   {"-42"},
		"int8":  {"-8"},
		"int16": {"-16"},
		"int32": {"-32"},
		"int64": {"-64"},
		"uint":  {"42"},
		"uint8": {"8"},
		"f32":   {"3.14"},
		"f64":   {"2.718"},
		"tags":  {"go", "api", "kit"},
	}
	var f typedForm
	if err := decodeFormValues(vals, &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Str != "hello" {
		t.Errorf("Str = %q, want %q", f.Str, "hello")
	}
	if f.Bool != true {
		t.Errorf("Bool = %v, want true", f.Bool)
	}
	if f.Int != -42 {
		t.Errorf("Int = %d, want -42", f.Int)
	}
	if f.Int8 != -8 {
		t.Errorf("Int8 = %d, want -8", f.Int8)
	}
	if f.Int16 != -16 {
		t.Errorf("Int16 = %d, want -16", f.Int16)
	}
	if f.Int32 != -32 {
		t.Errorf("Int32 = %d, want -32", f.Int32)
	}
	if f.Int64 != -64 {
		t.Errorf("Int64 = %d, want -64", f.Int64)
	}
	if f.Uint != 42 {
		t.Errorf("Uint = %d, want 42", f.Uint)
	}
	if f.Uint8 != 8 {
		t.Errorf("Uint8 = %d, want 8", f.Uint8)
	}
	if f.F32 != 3.14 {
		t.Errorf("F32 = %f, want 3.14", f.F32)
	}
	if f.F64 != 2.718 {
		t.Errorf("F64 = %f, want 2.718", f.F64)
	}
	if len(f.Tags) != 3 || f.Tags[0] != "go" || f.Tags[1] != "api" || f.Tags[2] != "kit" {
		t.Errorf("Tags = %v, want [go api kit]", f.Tags)
	}
}

func TestDecodeFormValues_TagPriority(t *testing.T) {
	vals := url.Values{
		"name_form": {"from_form"},
		"only_json": {"from_json"},
		"NoTag":     {"from_go"},
	}
	var f tagPriorityForm
	if err := decodeFormValues(vals, &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.NameForm != "from_form" {
		t.Errorf("NameForm = %q, want %q", f.NameForm, "from_form")
	}
	if f.OnlyJSON != "from_json" {
		t.Errorf("OnlyJSON = %q, want %q", f.OnlyJSON, "from_json")
	}
	if f.NoTag != "from_go" {
		t.Errorf("NoTag = %q, want %q", f.NoTag, "from_go")
	}
}

func TestDecodeFormValues_Skip(t *testing.T) {
	vals := url.Values{
		"name": {"Alice"},
		"-":    {"should not bind"},
	}
	var f skipForm
	if err := decodeFormValues(vals, &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
	if f.Ignored != "" {
		t.Errorf("Ignored = %q, want empty", f.Ignored)
	}
}

func TestDecodeFormValues_BoolHTMLValues(t *testing.T) {
	type boolForm struct {
		A bool `form:"a"`
		B bool `form:"b"`
		C bool `form:"c"`
		D bool `form:"d"`
	}
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"on", "on", true},
		{"ON", "ON", true},
		{"off", "off", false},
		{"yes", "yes", true},
		{"no", "no", false},
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f boolForm
			err := decodeFormValues(url.Values{"a": {tt.val}}, &f)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.val, err)
			}
			if f.A != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.val, f.A, tt.want)
			}
		})
	}
}

func TestDecodeFormValues_CheckboxUnchecked(t *testing.T) {
	// Unchecked checkboxes send nothing — field stays zero value (false)
	type checkForm struct {
		Agree bool `form:"agree"`
	}
	var f checkForm
	err := decodeFormValues(url.Values{}, &f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Agree != false {
		t.Errorf("Agree = %v, want false (unchecked)", f.Agree)
	}
}

func TestDecodeFormValues_MultipleCheckboxes(t *testing.T) {
	// Simulates: <input type="checkbox" name="roles" value="admin" />
	//            <input type="checkbox" name="roles" value="editor" />
	//            <input type="checkbox" name="roles" value="viewer" />
	// User checks "admin" and "viewer" only.
	type rolesForm struct {
		Roles []string `form:"roles"`
	}

	t.Run("some checked", func(t *testing.T) {
		vals := url.Values{"roles": {"admin", "viewer"}}
		var f rolesForm
		if err := decodeFormValues(vals, &f); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(f.Roles) != 2 {
			t.Fatalf("Roles length = %d, want 2", len(f.Roles))
		}
		if f.Roles[0] != "admin" || f.Roles[1] != "viewer" {
			t.Errorf("Roles = %v, want [admin viewer]", f.Roles)
		}
	})

	t.Run("all checked", func(t *testing.T) {
		vals := url.Values{"roles": {"admin", "editor", "viewer"}}
		var f rolesForm
		if err := decodeFormValues(vals, &f); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(f.Roles) != 3 {
			t.Fatalf("Roles length = %d, want 3", len(f.Roles))
		}
	})

	t.Run("none checked", func(t *testing.T) {
		vals := url.Values{}
		var f rolesForm
		if err := decodeFormValues(vals, &f); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Roles != nil {
			t.Errorf("Roles = %v, want nil", f.Roles)
		}
	})
}

func TestDecodeFormValues_TypeError(t *testing.T) {
	tests := []struct {
		name string
		vals url.Values
		msg  string
	}{
		{"bad int", url.Values{"int": {"abc"}}, "integer"},
		{"bad uint", url.Values{"uint": {"-1"}}, "positive integer"},
		{"bad float", url.Values{"f64": {"xyz"}}, "number"},
		{"bad bool", url.Values{"bool": {"notbool"}}, "boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f typedForm
			err := decodeFormValues(tt.vals, &f)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var apiErr *errors.Error
			if !stderrors.As(err, &apiErr) {
				t.Fatalf("expected *errors.Error, got %T", err)
			}
			if apiErr.Code != errors.CodeBadRequest {
				t.Errorf("code = %q, want %q", apiErr.Code, errors.CodeBadRequest)
			}
		})
	}
}

// --- BindForm tests ---

func newFormRequest(values url.Values) *http.Request {
	body := values.Encode()
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

type validatedForm struct {
	Name  string `form:"name" validate:"required"`
	Email string `form:"email" validate:"required,email"`
}

func TestBindForm_HappyPath(t *testing.T) {
	r := newFormRequest(url.Values{
		"name":  {"Alice"},
		"email": {"alice@example.com"},
	})
	f, err := BindForm[validatedForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
	if f.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", f.Email, "alice@example.com")
	}
}

func TestBindForm_WrongContentType(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("name=Alice"))
	r.Header.Set("Content-Type", "application/json")
	_, err := BindForm[simpleForm](r)
	if err == nil {
		t.Fatal("expected error for wrong content type")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestBindForm_MissingBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Body = nil
	_, err := BindForm[simpleForm](r)
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestBindForm_BodyTooLarge(t *testing.T) {
	// Create a body larger than 64 bytes
	bigVal := strings.Repeat("x", 100)
	r := newFormRequest(url.Values{"name": {bigVal}})

	cfg := Config{MaxBodySize: 64}
	_, err := BindFormWithConfig[simpleForm](r, cfg)
	if err == nil {
		t.Fatal("expected error for oversized body")
	}
}

func TestBindForm_Validation(t *testing.T) {
	r := newFormRequest(url.Values{
		"name":  {""},
		"email": {"not-email"},
	})
	_, err := BindForm[validatedForm](r)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.Code != errors.CodeValidation {
		t.Errorf("code = %q, want %q", apiErr.Code, errors.CodeValidation)
	}
}

// --- BindMultipart tests ---

func newMultipartRequest(fields map[string]string, files map[string][]byte) *http.Request {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	for name, content := range files {
		fw, _ := w.CreateFormFile(name, name+".txt")
		_, _ = fw.Write(content)
	}
	_ = w.Close()

	r := httptest.NewRequest("POST", "/", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

func TestBindMultipart_HappyPath(t *testing.T) {
	r := newMultipartRequest(
		map[string]string{"name": "Alice", "email": "alice@example.com"},
		nil,
	)
	f, err := BindMultipart[validatedForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
}

func TestBindMultipart_FormFile(t *testing.T) {
	r := newMultipartRequest(
		map[string]string{"name": "Alice", "email": "alice@example.com"},
		map[string][]byte{"avatar": []byte("fake-image-data")},
	)
	_, err := BindMultipart[validatedForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fh, err := FormFile(r, "avatar")
	if err != nil {
		t.Fatalf("FormFile error: %v", err)
	}
	if fh.Filename != "avatar.txt" {
		t.Errorf("Filename = %q, want %q", fh.Filename, "avatar.txt")
	}
	if fh.Size != int64(len("fake-image-data")) {
		t.Errorf("Size = %d, want %d", fh.Size, len("fake-image-data"))
	}

	// Read file content
	f, err := fh.Open()
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != "fake-image-data" {
		t.Errorf("content = %q, want %q", string(data), "fake-image-data")
	}
}

func TestBindMultipart_FormFiles(t *testing.T) {
	r := newMultipartRequest(
		map[string]string{"name": "Alice", "email": "alice@example.com"},
		map[string][]byte{"file1": []byte("a"), "file2": []byte("b")},
	)
	_, err := BindMultipart[validatedForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := FormFiles(r)
	if len(files) != 2 {
		t.Errorf("got %d file fields, want 2", len(files))
	}
}

func TestBindMultipart_Validation(t *testing.T) {
	r := newMultipartRequest(
		map[string]string{"name": "", "email": "bad"},
		nil,
	)
	_, err := BindMultipart[validatedForm](r)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.Code != errors.CodeValidation {
		t.Errorf("code = %q, want %q", apiErr.Code, errors.CodeValidation)
	}
}

func TestBindMultipart_WrongContentType(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("name=Alice"))
	r.Header.Set("Content-Type", "application/json")
	_, err := BindMultipart[simpleForm](r)
	if err == nil {
		t.Fatal("expected error for wrong content type")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestFormFile_Missing(t *testing.T) {
	r := newMultipartRequest(map[string]string{"name": "Alice"}, nil)
	// Parse so MultipartForm is available
	_ = r.ParseMultipartForm(DefaultMaxMultipartMemory)

	_, err := FormFile(r, "avatar")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFormFiles_NilMultipartForm(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	files := FormFiles(r)
	if files != nil {
		t.Errorf("expected nil, got %v", files)
	}
}

// --- Bind[T] auto-detect dispatch tests ---

func TestBind_JSON(t *testing.T) {
	body := `{"name":"Alice","email":"alice@example.com"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	f, err := Bind[simpleForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON uses json tags; simpleForm has form tags but json decoder uses field names
	// simpleForm fields have form tags only, so JSON decoder maps by field name
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
}

type jsonForm struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestBind_JSON_Explicit(t *testing.T) {
	body := `{"name":"Alice","email":"alice@example.com"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	f, err := Bind[jsonForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
	if f.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", f.Email, "alice@example.com")
	}
}

func TestBind_Form(t *testing.T) {
	r := newFormRequest(url.Values{
		"name":  {"Alice"},
		"email": {"alice@example.com"},
	})
	f, err := Bind[simpleForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
}

func TestBind_Multipart(t *testing.T) {
	r := newMultipartRequest(
		map[string]string{"name": "Alice", "email": "alice@example.com"},
		nil,
	)
	f, err := Bind[simpleForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
}

func TestBind_EmptyContentType_FallsBackToJSON(t *testing.T) {
	body := `{"name":"Alice","email":"alice@example.com"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	// No Content-Type header
	f, err := Bind[jsonForm](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "Alice" {
		t.Errorf("Name = %q, want %q", f.Name, "Alice")
	}
}

func TestBind_UnknownContentType_Returns415(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("data"))
	r.Header.Set("Content-Type", "text/plain")
	_, err := Bind[simpleForm](r)
	if err == nil {
		t.Fatal("expected error for unknown content type")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

// --- formFieldName tests ---

func TestFormFieldName(t *testing.T) {
	type S struct {
		A string `form:"a_form" json:"a_json"`
		B string `json:"b_json"`
		C string `form:"c_form"`
		D string
		E string `form:"-"`
	}

	typ := fmt.Sprintf("%T", S{}) // just to use S
	_ = typ

	st := reflect.TypeOf(S{})
	tests := []struct {
		field string
		want  string
	}{
		{"A", "a_form"},
		{"B", "b_json"},
		{"C", "c_form"},
		{"D", "D"},
		{"E", "-"},
	}
	for _, tt := range tests {
		f, _ := st.FieldByName(tt.field)
		got := formFieldName(f)
		if got != tt.want {
			t.Errorf("formFieldName(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}
