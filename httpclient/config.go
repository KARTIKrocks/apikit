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
	return func(c *Client) { c.maxRetries = n }
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
	return func(c *Client) { c.logger = logger }
}

// WithCircuitBreaker enables the circuit breaker with the given threshold and timeout.
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *Client) {
		c.cb = NewCircuitBreaker(threshold, timeout)
	}
}

// WithTransport sets the HTTP transport.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) { c.transport = t }
}
