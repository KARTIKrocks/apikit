package httpclient

import (
	"context"
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
	rb.headers["Authorization"] = "Bearer " + token
	return rb
}

// Send executes the request
func (rb *RequestBuilder) Send(ctx context.Context) (*Response, error) {
	// Build query string
	path := rb.path
	if len(rb.params) > 0 {
		values := url.Values{}
		for k, v := range rb.params {
			values.Add(k, v)
		}
		path += "?" + values.Encode()
	}

	return rb.client.doRequest(ctx, rb.method, path, rb.body, rb.headers)
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
