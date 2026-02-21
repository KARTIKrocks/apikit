package httpclient

import "fmt"

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

// Error implements error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}

// IsClientError returns true if status is 4xx
func (e *HTTPError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if status is 5xx
func (e *HTTPError) IsServerError() bool {
	return e.StatusCode >= 500
}
