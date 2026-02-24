// Package health provides health check endpoint builders with dependency
// checking, timeouts, and standard response formats. Designed for Kubernetes
// liveness/readiness probes and load balancer health checks.
//
// Usage:
//
//	h := health.NewChecker(health.WithTimeout(3 * time.Second))
//	h.AddCheck("postgres", func(ctx context.Context) error {
//	    return db.PingContext(ctx)
//	})
//	h.AddNonCriticalCheck("redis", func(ctx context.Context) error {
//	    return rdb.Ping(ctx).Err()
//	})
//
//	r.Get("/health", h.Handler())
//	r.Get("/health/live", h.LiveHandler())
package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/KARTIKrocks/apikit/response"
)

// CheckFunc is a health check function. Return nil = healthy, error = unhealthy.
type CheckFunc func(ctx context.Context) error

// Status constants.
const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
)

// CheckResult is the result of a single named check.
type CheckResult struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

// Response is the full health check response.
type Response struct {
	Status    string                 `json:"status"`
	Checks   map[string]CheckResult `json:"checks,omitempty"`
	Timestamp int64                 `json:"timestamp"`
}

// namedCheck is an internal type pairing a name with its check function.
type namedCheck struct {
	name     string
	fn       CheckFunc
	critical bool
}

// Checker runs health checks and exposes HTTP handlers.
type Checker struct {
	timeout time.Duration
	checks  []namedCheck
}

// Option configures a Checker.
type Option func(*Checker)

// WithTimeout sets the per-check timeout. Default is 5 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *Checker) {
		c.timeout = d
	}
}

// NewChecker creates a new health Checker with the given options.
func NewChecker(opts ...Option) *Checker {
	c := &Checker{
		timeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AddCheck registers a critical health check. If it fails the overall status
// is "unhealthy" and the HTTP handler returns 503.
func (c *Checker) AddCheck(name string, fn CheckFunc) {
	c.checks = append(c.checks, namedCheck{name: name, fn: fn, critical: true})
}

// AddNonCriticalCheck registers a non-critical health check. If it fails the
// overall status is "degraded" but the HTTP handler still returns 200.
func (c *Checker) AddNonCriticalCheck(name string, fn CheckFunc) {
	c.checks = append(c.checks, namedCheck{name: name, fn: fn, critical: false})
}

// Check runs all registered checks concurrently and returns the aggregated result.
func (c *Checker) Check(ctx context.Context) Response {
	if len(c.checks) == 0 {
		return Response{
			Status:    StatusHealthy,
			Timestamp: time.Now().Unix(),
		}
	}

	type indexedResult struct {
		index  int
		result CheckResult
	}

	results := make([]indexedResult, len(c.checks))
	var wg sync.WaitGroup
	wg.Add(len(c.checks))

	for i, nc := range c.checks {
		go func(idx int, nc namedCheck) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
			defer cancel()

			start := time.Now()
			err := nc.fn(checkCtx)
			duration := time.Since(start).Milliseconds()

			r := CheckResult{
				Status:   StatusHealthy,
				Duration: duration,
			}
			if err != nil {
				r.Status = StatusUnhealthy
				r.Error = err.Error()
			}

			results[idx] = indexedResult{index: idx, result: r}
		}(i, nc)
	}

	wg.Wait()

	checks := make(map[string]CheckResult, len(c.checks))
	overall := StatusHealthy

	for i, nc := range c.checks {
		checks[nc.name] = results[i].result

		if results[i].result.Status != StatusHealthy {
			if nc.critical {
				overall = StatusUnhealthy
			} else if overall != StatusUnhealthy {
				overall = StatusDegraded
			}
		}
	}

	return Response{
		Status:    overall,
		Checks:   checks,
		Timestamp: time.Now().Unix(),
	}
}

// Handler returns an HTTP handler that runs all health checks and responds
// with 200 (healthy/degraded) or 503 (unhealthy). Compatible with
// router.HandlerFunc (returns error).
func (c *Checker) Handler() func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		resp := c.Check(r.Context())

		if resp.Status == StatusUnhealthy {
			response.New().
				Status(http.StatusServiceUnavailable).
				Message("Health check").
				Data(resp).
				Send(w)
			return nil
		}

		response.OK(w, "Health check", resp)
		return nil
	}
}

// LiveHandler returns an HTTP handler that always responds with 200.
// Use this for Kubernetes liveness probes to confirm the process is running.
func (c *Checker) LiveHandler() func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		response.OK(w, "OK", nil)
		return nil
	}
}
