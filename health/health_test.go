package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARTIKrocks/apikit/response"
)

func TestCheckAllHealthy(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error { return nil })
	c.AddCheck("cache", func(ctx context.Context) error { return nil })

	resp := c.Check(context.Background())
	if resp.Status != StatusHealthy {
		t.Fatalf("expected %q, got %q", StatusHealthy, resp.Status)
	}
	if len(resp.Checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(resp.Checks))
	}
	for name, cr := range resp.Checks {
		if cr.Status != StatusHealthy {
			t.Errorf("check %q: expected %q, got %q", name, StatusHealthy, cr.Status)
		}
	}
}

func TestCheckCriticalFails(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error {
		return errors.New("connection refused")
	})
	c.AddCheck("ok", func(ctx context.Context) error { return nil })

	resp := c.Check(context.Background())
	if resp.Status != StatusUnhealthy {
		t.Fatalf("expected %q, got %q", StatusUnhealthy, resp.Status)
	}
	if resp.Checks["db"].Error != "connection refused" {
		t.Errorf("expected error message, got %q", resp.Checks["db"].Error)
	}
}

func TestCheckNonCriticalFails(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error { return nil })
	c.AddNonCriticalCheck("cache", func(ctx context.Context) error {
		return errors.New("cache down")
	})

	resp := c.Check(context.Background())
	if resp.Status != StatusDegraded {
		t.Fatalf("expected %q, got %q", StatusDegraded, resp.Status)
	}
}

func TestCheckCriticalOverridesNonCritical(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error {
		return errors.New("db down")
	})
	c.AddNonCriticalCheck("cache", func(ctx context.Context) error {
		return errors.New("cache down")
	})

	resp := c.Check(context.Background())
	if resp.Status != StatusUnhealthy {
		t.Fatalf("expected %q, got %q — critical failure should override degraded", StatusUnhealthy, resp.Status)
	}
}

func TestCheckTimeout(t *testing.T) {
	c := NewChecker(WithTimeout(50 * time.Millisecond))
	c.AddCheck("slow", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	})

	resp := c.Check(context.Background())
	if resp.Status != StatusUnhealthy {
		t.Fatalf("expected %q for timed-out check, got %q", StatusUnhealthy, resp.Status)
	}
	if resp.Checks["slow"].Error == "" {
		t.Error("expected error message for timed-out check")
	}
}

func TestCheckNoChecks(t *testing.T) {
	c := NewChecker()
	resp := c.Check(context.Background())
	if resp.Status != StatusHealthy {
		t.Fatalf("expected %q with no checks, got %q", StatusHealthy, resp.Status)
	}
	if resp.Checks != nil {
		t.Errorf("expected nil checks map, got %v", resp.Checks)
	}
}

func TestHandlerHealthy(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error { return nil })

	handler := c.Handler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	if err := handler(w, req); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var env response.Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if !env.Success {
		t.Error("expected success=true")
	}
}

func TestHandlerUnhealthy(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error {
		return errors.New("down")
	})

	handler := c.Handler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	if err := handler(w, req); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandlerDegraded(t *testing.T) {
	c := NewChecker()
	c.AddCheck("db", func(ctx context.Context) error { return nil })
	c.AddNonCriticalCheck("cache", func(ctx context.Context) error {
		return errors.New("cache down")
	})

	handler := c.Handler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	if err := handler(w, req); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for degraded, got %d", w.Code)
	}
}

func TestLiveHandler(t *testing.T) {
	c := NewChecker()
	// Even with a failing check, liveness should always return 200.
	c.AddCheck("db", func(ctx context.Context) error {
		return errors.New("down")
	})

	handler := c.LiveHandler()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	if err := handler(w, req); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDefaultTimeout(t *testing.T) {
	c := NewChecker()
	if c.timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %s", c.timeout)
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewChecker(WithTimeout(10 * time.Second))
	if c.timeout != 10*time.Second {
		t.Fatalf("expected timeout 10s, got %s", c.timeout)
	}
}
