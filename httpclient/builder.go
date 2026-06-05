package httpclient

import (
	"context"
	"errors"
	"net/url"
)

// RequestBuilder builds HTTP requests with fluent API
type RequestBuilder struct {
	client  *Client
	method  string
	path    string
	body    any
	headers map[string]string
	params  map[string]string

	// errorOnStatus overrides the client's error-on-status policy for this
	// request when non-nil.
	errorOnStatus *bool
}

// Method sets the HTTP method
func (rb *RequestBuilder) Method(method string) *RequestBuilder {
	rb.method = method
	return rb
}

// Path sets the request path
func (rb *RequestBuilder) Path(path string) *RequestBuilder {
	rb.path = path
	return rb
}

// Body sets the request body
func (rb *RequestBuilder) Body(body any) *RequestBuilder {
	rb.body = body
	return rb
}

// Header sets a request header
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	if rb.headers == nil {
		rb.headers = make(map[string]string)
	}
	rb.headers[key] = value
	return rb
}

// Headers sets multiple headers
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// Param sets a query parameter
func (rb *RequestBuilder) Param(key, value string) *RequestBuilder {
	if rb.params == nil {
		rb.params = make(map[string]string)
	}
	rb.params[key] = value
	return rb
}

// Params sets multiple query parameters
func (rb *RequestBuilder) Params(params map[string]string) *RequestBuilder {
	for k, v := range params {
		rb.params[k] = v
	}
	return rb
}

// BearerToken sets Authorization header
func (rb *RequestBuilder) BearerToken(token string) *RequestBuilder {
	if rb.headers == nil {
		rb.headers = make(map[string]string)
	}
	rb.headers["Authorization"] = "Bearer " + token
	return rb
}

// ErrorOnStatus overrides the client's error-on-status policy for this request.
// See WithErrorOnStatus for the semantics. When unset, the client default applies.
func (rb *RequestBuilder) ErrorOnStatus(enabled bool) *RequestBuilder {
	rb.errorOnStatus = &enabled
	return rb
}

// Send executes the request
func (rb *RequestBuilder) Send(ctx context.Context) (*Response, error) {
	if rb == nil || rb.client == nil {
		return nil, errors.New("request builder or client is nil")
	}

	// Build query string
	path := rb.path
	if len(rb.params) > 0 {
		u, err := url.Parse(rb.path)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, v := range rb.params {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
		path = u.String()
	}

	errorOnStatus := rb.client.errorOnStatus
	if rb.errorOnStatus != nil {
		errorOnStatus = *rb.errorOnStatus
	}

	return rb.client.do(ctx, rb.method, path, rb.body, rb.headers, errorOnStatus)
}

// Get is a shorthand for Method("GET").Send()
func (rb *RequestBuilder) Get(ctx context.Context) (*Response, error) {
	rb.method = "GET"
	return rb.Send(ctx)
}

// Post is a shorthand for Method("POST").Send()
func (rb *RequestBuilder) Post(ctx context.Context) (*Response, error) {
	rb.method = "POST"
	return rb.Send(ctx)
}

// Put is a shorthand for Method("PUT").Send()
func (rb *RequestBuilder) Put(ctx context.Context) (*Response, error) {
	rb.method = "PUT"
	return rb.Send(ctx)
}

// Delete is a shorthand for Method("DELETE").Send()
func (rb *RequestBuilder) Delete(ctx context.Context) (*Response, error) {
	rb.method = "DELETE"
	return rb.Send(ctx)
}
