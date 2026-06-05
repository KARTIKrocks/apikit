package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// JSON unmarshals response body into v
func (r *Response) JSON(v any) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// String returns response body as string
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess returns true if status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true if status code is 4xx
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if status code is 5xx
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500
}
