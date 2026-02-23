// Package httpclient provides a production-ready HTTP client with automatic retries,
// exponential backoff, circuit breaker, and structured logging via slog.
//
// Use the HTTPClient interface to accept either a real Client or a MockClient in
// your code, making external HTTP calls easy to test.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// HTTPClient is the interface implemented by Client and MockClient.
type HTTPClient interface {
	Get(ctx context.Context, path string) (*Response, error)
	Post(ctx context.Context, path string, body any) (*Response, error)
	Put(ctx context.Context, path string, body any) (*Response, error)
	Patch(ctx context.Context, path string, body any) (*Response, error)
	Delete(ctx context.Context, path string) (*Response, error)
}

// Verify interface compliance at compile time.
var _ HTTPClient = (*Client)(nil)

// Client is an HTTP client with retries, circuit breaker, and structured logging.
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	mu         sync.RWMutex

	// config fields
	timeout      time.Duration
	maxRetries   int
	retryDelay   time.Duration
	maxRetryDelay time.Duration
	logger       *slog.Logger
	cb           *CircuitBreaker
	transport    http.RoundTripper
}

// New creates a new HTTP client with the given base URL and options.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:      baseURL,
		headers:      make(map[string]string),
		timeout:      30 * time.Second,
		maxRetries:   3,
		retryDelay:   time.Second,
		maxRetryDelay: 10 * time.Second,
		logger:       slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := c.transport
	if transport == nil {
		transport = &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
	}

	c.httpClient = &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}

	return c
}

// SetHeader sets a default header for all requests.
func (c *Client) SetHeader(key, value string) *Client {
	c.mu.Lock()
	c.headers[key] = value
	c.mu.Unlock()
	return c
}

// SetHeaders sets multiple default headers.
func (c *Client) SetHeaders(headers map[string]string) *Client {
	c.mu.Lock()
	for k, v := range headers {
		c.headers[k] = v
	}
	c.mu.Unlock()
	return c
}

// SetBearerToken sets Authorization header with Bearer token.
func (c *Client) SetBearerToken(token string) *Client {
	c.mu.Lock()
	c.headers["Authorization"] = "Bearer " + token
	c.mu.Unlock()
	return c
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

// Post performs a POST request with JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, nil)
}

// Put performs a PUT request with JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) (*Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, nil)
}

// Patch performs a PATCH request with JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*Response, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body, nil)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// Request creates a new request builder.
func (c *Client) Request() *RequestBuilder {
	return &RequestBuilder{
		client:  c,
		headers: make(map[string]string),
		params:  make(map[string]string),
	}
}

// doRequest performs the HTTP request with retries and optional circuit breaker.
func (c *Client) doRequest(ctx context.Context, method, path string, body any, headers map[string]string) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateRetryDelay(attempt)
			c.logger.Info("retrying request", "attempt", attempt, "delay", delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		var resp *Response
		var err error

		if c.cb != nil {
			cbErr := c.cb.Call(func() error {
				resp, err = c.executeRequest(ctx, method, path, body, headers)
				if err != nil {
					return err
				}
				return nil
			})
			if cbErr != nil && err == nil {
				// Circuit breaker rejected the call.
				err = cbErr
			}
		} else {
			resp, err = c.executeRequest(ctx, method, path, body, headers)
		}

		if err == nil {
			return resp, nil
		}

		lastErr = err

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Don't retry client errors (4xx).
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return resp, err
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// executeRequest executes a single HTTP request.
func (c *Client) executeRequest(ctx context.Context, method, path string, body any, headers map[string]string) (*Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy default headers under read lock.
	c.mu.RLock()
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	c.mu.RUnlock()

	// Set per-request headers.
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logger.Info("sending request", "method", method, "url", url)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("request failed", "method", method, "url", url, "error", err, "duration", duration)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       responseBody,
		Duration:   duration,
	}

	c.logger.Info("request completed",
		"method", method,
		"url", url,
		"status", resp.StatusCode,
		"duration", duration,
	)

	if resp.StatusCode >= 400 {
		return response, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       responseBody,
		}
	}

	return response, nil
}

// calculateRetryDelay calculates delay with exponential backoff.
func (c *Client) calculateRetryDelay(attempt int) time.Duration {
	shift := attempt - 1
	// Prevent overflow: 1<<63 flips the sign bit on 64-bit integers.
	if shift >= 62 {
		return c.maxRetryDelay
	}
	delay := c.retryDelay * time.Duration(1<<uint(shift))
	if c.maxRetryDelay > 0 && delay > c.maxRetryDelay {
		delay = c.maxRetryDelay
	}
	return delay
}
