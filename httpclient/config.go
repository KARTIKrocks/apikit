package httpclient

import (
	"log/slog"
	"net/http"
	"time"
)

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// WithRetryDelay sets the initial delay between retries.
func WithRetryDelay(d time.Duration) Option {
	return func(c *Client) { c.retryDelay = d }
}

// WithMaxRetryDelay sets the maximum delay between retries.
func WithMaxRetryDelay(d time.Duration) Option {
	return func(c *Client) { c.maxRetryDelay = d }
}

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		} else {
			c.logger = slog.Default()
		}
	}
}

// WithCircuitBreaker enables the circuit breaker with the given threshold and timeout.
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *Client) {
		c.cb = NewCircuitBreaker(threshold, timeout)
	}
}

// WithMaxResponseBody sets the maximum response body size in bytes.
// Responses larger than this are rejected. Default: 10 MB.
func WithMaxResponseBody(n int64) Option {
	return func(c *Client) { c.maxResponseBody = n }
}

// WithTransport sets the HTTP transport.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) { c.transport = t }
}

// WithErrorOnStatus controls whether a non-2xx HTTP status is returned as an
// error. It is enabled by default: requests to a 4xx/5xx endpoint return an
// *HTTPError alongside the response.
//
// Disable it for services that return structured error bodies you want to
// decode directly. Non-2xx responses then return (resp, nil) and you branch on
// resp.IsClientError()/IsServerError() and read the body via resp.JSON(). Retry
// behavior is unchanged — 5xx responses are still retried internally — and
// transport, context, and circuit-breaker failures are still returned as errors.
func WithErrorOnStatus(enabled bool) Option {
	return func(c *Client) { c.errorOnStatus = enabled }
}
