package server_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARTIKrocks/apikit/server"
)

func TestNew_Defaults(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestNew_WithOptions(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler,
		server.WithAddr(":0"),
		server.WithReadTimeout(5*time.Second),
		server.WithWriteTimeout(10*time.Second),
		server.WithIdleTimeout(30*time.Second),
		server.WithShutdownTimeout(5*time.Second),
	)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	srv := server.New(handler, server.WithAddr(":0"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServer_OnStartHook(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler, server.WithAddr(":0"))

	var called atomic.Bool
	srv.OnStart(func() error {
		called.Store(true)
		return nil
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	if !called.Load() {
		t.Error("expected OnStart hook to be called")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	<-errCh
}

func TestServer_OnStartHook_Error(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler, server.WithAddr(":0"))

	srv.OnStart(func() error {
		return fmt.Errorf("connection refused")
	})

	err := srv.Start()
	if err == nil {
		t.Fatal("expected error from OnStart hook")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", err.Error())
	}
}

func TestServer_OnShutdownHook(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler, server.WithAddr(":0"))

	called := false
	srv.OnShutdown(func(ctx context.Context) error {
		called = true
		return nil
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop in time")
	}

	if !called {
		t.Error("expected OnShutdown hook to be called")
	}
}

func TestServer_MultipleHooks(t *testing.T) {
	handler := http.NewServeMux()
	srv := server.New(handler, server.WithAddr(":0"))

	var order []int
	srv.OnStart(func() error { order = append(order, 1); return nil })
	srv.OnStart(func() error { order = append(order, 2); return nil })
	srv.OnShutdown(func(ctx context.Context) error { order = append(order, 3); return nil })
	srv.OnShutdown(func(ctx context.Context) error { order = append(order, 4); return nil })

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	<-errCh

	if len(order) != 4 {
		t.Fatalf("expected 4 hooks to run, got %d", len(order))
	}
	for i, v := range order {
		if v != i+1 {
			t.Errorf("expected hook %d at position %d, got %d", i+1, i, v)
		}
	}
}
