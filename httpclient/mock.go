package httpclient

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Verify interface compliance at compile time.
var _ HTTPClient = (*MockClient)(nil)

// MockClient is a concurrency-safe mock HTTP client for testing.
type MockClient struct {
	mu        sync.Mutex
	responses map[string]*Response
	errors    map[string]error
	calls     []MockCall
}

// MockCall represents a recorded API call.
type MockCall struct {
	Method string
	Path   string
	Body   any
	Time   time.Time
}

// NewMockClient creates a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		responses: make(map[string]*Response),
		errors:    make(map[string]error),
	}
}

// OnGet sets mock response for GET request.
func (mc *MockClient) OnGet(path string, statusCode int, body []byte) *MockClient {
	mc.mu.Lock()
	mc.responses["GET:"+path] = &Response{StatusCode: statusCode, Body: body, Headers: http.Header{}}
	mc.mu.Unlock()
	return mc
}

// OnPost sets mock response for POST request.
func (mc *MockClient) OnPost(path string, statusCode int, body []byte) *MockClient {
	mc.mu.Lock()
	mc.responses["POST:"+path] = &Response{StatusCode: statusCode, Body: body, Headers: http.Header{}}
	mc.mu.Unlock()
	return mc
}

// OnPut sets mock response for PUT request.
func (mc *MockClient) OnPut(path string, statusCode int, body []byte) *MockClient {
	mc.mu.Lock()
	mc.responses["PUT:"+path] = &Response{StatusCode: statusCode, Body: body, Headers: http.Header{}}
	mc.mu.Unlock()
	return mc
}

// OnPatch sets mock response for PATCH request.
func (mc *MockClient) OnPatch(path string, statusCode int, body []byte) *MockClient {
	mc.mu.Lock()
	mc.responses["PATCH:"+path] = &Response{StatusCode: statusCode, Body: body, Headers: http.Header{}}
	mc.mu.Unlock()
	return mc
}

// OnDelete sets mock response for DELETE request.
func (mc *MockClient) OnDelete(path string, statusCode int, body []byte) *MockClient {
	mc.mu.Lock()
	mc.responses["DELETE:"+path] = &Response{StatusCode: statusCode, Body: body, Headers: http.Header{}}
	mc.mu.Unlock()
	return mc
}

// OnError sets error for a request.
func (mc *MockClient) OnError(method, path string, err error) *MockClient {
	mc.mu.Lock()
	mc.errors[method+":"+path] = err
	mc.mu.Unlock()
	return mc
}

// Get mocks GET request.
func (mc *MockClient) Get(ctx context.Context, path string) (*Response, error) {
	return mc.doMockRequest("GET", path, nil)
}

// Post mocks POST request.
func (mc *MockClient) Post(ctx context.Context, path string, body any) (*Response, error) {
	return mc.doMockRequest("POST", path, body)
}

// Put mocks PUT request.
func (mc *MockClient) Put(ctx context.Context, path string, body any) (*Response, error) {
	return mc.doMockRequest("PUT", path, body)
}

// Patch mocks PATCH request.
func (mc *MockClient) Patch(ctx context.Context, path string, body any) (*Response, error) {
	return mc.doMockRequest("PATCH", path, body)
}

// Delete mocks DELETE request.
func (mc *MockClient) Delete(ctx context.Context, path string) (*Response, error) {
	return mc.doMockRequest("DELETE", path, nil)
}

func (mc *MockClient) doMockRequest(method, path string, body any) (*Response, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.calls = append(mc.calls, MockCall{
		Method: method,
		Path:   path,
		Body:   body,
		Time:   time.Now(),
	})

	key := method + ":" + path

	if err, exists := mc.errors[key]; exists {
		return nil, err
	}

	if resp, exists := mc.responses[key]; exists {
		return resp, nil
	}

	return &Response{
		StatusCode: 404,
		Body:       []byte(`{"error":"not found"}`),
		Headers:    http.Header{},
	}, nil
}

// GetCalls returns all recorded calls.
func (mc *MockClient) GetCalls() []MockCall {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	out := make([]MockCall, len(mc.calls))
	copy(out, mc.calls)
	return out
}

// GetCallCount returns the number of calls.
func (mc *MockClient) GetCallCount() int {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return len(mc.calls)
}

// Reset clears all mocks and recorded calls.
func (mc *MockClient) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.responses = make(map[string]*Response)
	mc.errors = make(map[string]error)
	mc.calls = nil
}
