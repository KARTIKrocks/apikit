package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- XML ---

type xmlUser struct {
	XMLName xml.Name `xml:"user"`
	Name    string   `xml:"name"`
	Age     int      `xml:"age"`
}

func TestXML(t *testing.T) {
	w := httptest.NewRecorder()
	XML(w, http.StatusOK, xmlUser{Name: "Alice", Age: 30})

	if w.Code != 200 {
		t.Errorf("status: expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/xml; charset=utf-8" {
		t.Errorf("content-type: expected application/xml; charset=utf-8, got %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<?xml") {
		t.Error("expected XML header")
	}
	if !strings.Contains(body, "<name>Alice</name>") {
		t.Error("expected <name>Alice</name> in body")
	}
}

func TestXML_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	// Channels cannot be marshalled to XML
	XML(w, http.StatusOK, make(chan int))

	if w.Code != 500 {
		t.Errorf("status: expected 500 on marshal error, got %d", w.Code)
	}
}

// --- IndentedJSON ---

func TestIndentedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"name": "Alice"}
	IndentedJSON(w, http.StatusOK, data)

	if w.Code != 200 {
		t.Errorf("status: expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "  ") {
		t.Error("expected indented output")
	}
	if !strings.Contains(body, `"name": "Alice"`) {
		t.Error("expected name field in body")
	}
}

func TestIndentedJSON_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	IndentedJSON(w, http.StatusOK, make(chan int))

	if w.Code != 500 {
		t.Errorf("status: expected 500 on marshal error, got %d", w.Code)
	}
}

// --- PureJSON ---

func TestPureJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"html": "<b>bold</b> & stuff"}
	PureJSON(w, http.StatusOK, data)

	if w.Code != 200 {
		t.Errorf("status: expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	// PureJSON should NOT escape < > &
	if strings.Contains(body, `\u003c`) {
		t.Error("PureJSON should not escape < to \\u003c")
	}
	if !strings.Contains(body, "<b>bold</b>") {
		t.Error("expected unescaped HTML in body")
	}
	if !strings.Contains(body, "& stuff") {
		t.Error("expected unescaped & in body")
	}
}

func TestPureJSON_VsStandardJSON(t *testing.T) {
	data := map[string]string{"v": "<a>"}

	// Standard json.Marshal escapes HTML
	std, _ := json.Marshal(data)
	if !strings.Contains(string(std), `\u003c`) {
		t.Skip("standard json.Marshal does not escape HTML on this Go version")
	}

	// PureJSON should not
	w := httptest.NewRecorder()
	PureJSON(w, http.StatusOK, data)
	if strings.Contains(w.Body.String(), `\u003c`) {
		t.Error("PureJSON should not escape HTML entities")
	}
}

// --- JSONP ---

func TestJSONP_WithCallback(t *testing.T) {
	r := httptest.NewRequest("GET", "/data?callback=myFunc", nil)
	w := httptest.NewRecorder()
	JSONP(w, r, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != 200 {
		t.Errorf("status: expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/javascript; charset=utf-8" {
		t.Errorf("content-type: expected application/javascript, got %q", ct)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "myFunc(") {
		t.Errorf("expected body to start with myFunc(, got %q", body)
	}
	if !strings.HasSuffix(body, ");") {
		t.Errorf("expected body to end with );, got %q", body)
	}
}

func TestJSONP_WithoutCallback(t *testing.T) {
	r := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	JSONP(w, r, http.StatusOK, map[string]string{"key": "value"})

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("content-type: expected application/json fallback, got %q", ct)
	}
	body := w.Body.String()
	if strings.Contains(body, "(") {
		t.Error("expected no JSONP wrapping without callback")
	}
}

func TestJSONP_MarshalError(t *testing.T) {
	r := httptest.NewRequest("GET", "/data?callback=fn", nil)
	w := httptest.NewRecorder()
	JSONP(w, r, http.StatusOK, make(chan int))

	if w.Code != 500 {
		t.Errorf("status: expected 500 on marshal error, got %d", w.Code)
	}
}

// --- Reader ---

func TestReader(t *testing.T) {
	body := "hello world"
	r := strings.NewReader(body)
	w := httptest.NewRecorder()

	Reader(w, http.StatusOK, "text/plain", int64(len(body)), r)

	if w.Code != 200 {
		t.Errorf("status: expected 200, got %d", w.Code)
	}
	if w.Body.String() != body {
		t.Errorf("body: expected %q, got %q", body, w.Body.String())
	}
	if w.Header().Get("Content-Length") != "11" {
		t.Errorf("content-length: expected 11, got %q", w.Header().Get("Content-Length"))
	}
}

func TestReader_NoContentLength(t *testing.T) {
	r := strings.NewReader("data")
	w := httptest.NewRecorder()

	Reader(w, http.StatusOK, "application/octet-stream", -1, r)

	if w.Header().Get("Content-Length") != "" {
		t.Error("expected no Content-Length header when contentLength is -1")
	}
	if w.Body.String() != "data" {
		t.Errorf("body: expected %q, got %q", "data", w.Body.String())
	}
}

func TestReader_DefaultContentType(t *testing.T) {
	r := strings.NewReader("data")
	w := httptest.NewRecorder()

	Reader(w, http.StatusOK, "", 4, r)

	ct := w.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("content-type: expected application/octet-stream, got %q", ct)
	}
}
