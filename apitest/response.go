package apitest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/KARTIKrocks/apikit/response"
)

// Response wraps a recorded HTTP response for easy inspection and assertions.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Header returns the first value for the given header key.
func (r *Response) Header(key string) string {
	return r.Headers.Get(key)
}

// Decode unmarshals the response body into v.
func (r *Response) Decode(v any) error {
	return json.Unmarshal(r.Body, v)
}

// Envelope decodes the response body as a response.Envelope.
func (r *Response) Envelope() (*response.Envelope, error) {
	var env response.Envelope
	if err := r.Decode(&env); err != nil {
		return nil, err
	}
	return &env, nil
}

// Record executes an http.Handler with the given request and returns a *Response.
func Record(handler http.Handler, req *http.Request) *Response {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	result := rec.Result()
	defer func() { _ = result.Body.Close() }()

	body, _ := io.ReadAll(result.Body)

	return &Response{
		StatusCode: result.StatusCode,
		Headers:    result.Header,
		Body:       body,
	}
}

// RecordHandler wraps a response.HandlerFunc with response.Handle and records the result.
func RecordHandler(fn response.HandlerFunc, req *http.Request) *Response {
	return Record(response.Handle(fn), req)
}
