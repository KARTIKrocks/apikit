// Package apitest provides testing helpers for HTTP handlers built with apikit.
package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// RequestBuilder provides a fluent API to construct *http.Request for tests.
type RequestBuilder struct {
	method     string
	target     string
	body       any
	headers    http.Header
	query      url.Values
	pathValues map[string]string
	ctx        context.Context
}

// NewRequest creates a new RequestBuilder with the given method and target path.
func NewRequest(method, target string) *RequestBuilder {
	return &RequestBuilder{
		method:     method,
		target:     target,
		headers:    make(http.Header),
		query:      make(url.Values),
		pathValues: make(map[string]string),
	}
}

// WithBody sets the request body. The value will be JSON-marshaled.
func (b *RequestBuilder) WithBody(v any) *RequestBuilder {
	b.body = v
	return b
}

// WithHeader adds a header to the request.
func (b *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	b.headers.Set(key, value)
	return b
}

// WithQuery adds a query parameter to the request.
func (b *RequestBuilder) WithQuery(key, value string) *RequestBuilder {
	b.query.Set(key, value)
	return b
}

// WithPathValue sets a path parameter value (Go 1.22+ routing).
func (b *RequestBuilder) WithPathValue(key, value string) *RequestBuilder {
	b.pathValues[key] = value
	return b
}

// WithBearerToken sets the Authorization header with a Bearer token.
func (b *RequestBuilder) WithBearerToken(token string) *RequestBuilder {
	b.headers.Set("Authorization", "Bearer "+token)
	return b
}

// WithContext sets the request context.
func (b *RequestBuilder) WithContext(ctx context.Context) *RequestBuilder {
	b.ctx = ctx
	return b
}

// Build constructs the *http.Request. It panics if JSON marshaling fails.
func (b *RequestBuilder) Build() *http.Request {
	target := b.target
	if len(b.query) > 0 {
		target += "?" + b.query.Encode()
	}

	var req *http.Request
	if b.body != nil {
		data, err := json.Marshal(b.body)
		if err != nil {
			panic("apitest: failed to marshal request body: " + err.Error())
		}
		req = httptest.NewRequest(b.method, target, bytes.NewReader(data))
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		req = httptest.NewRequest(b.method, target, nil)
	}

	for key, values := range b.headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	for key, value := range b.pathValues {
		req.SetPathValue(key, value)
	}

	if b.ctx != nil {
		req = req.WithContext(b.ctx)
	}

	return req
}
