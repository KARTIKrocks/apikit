// Package server provides a production-ready HTTP server with graceful shutdown,
// signal handling, and lifecycle hooks.
package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server wraps http.Server with graceful shutdown and lifecycle hooks.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	logger          *slog.Logger
	onStart         []func() error
	onShutdown      []func(ctx context.Context) error
	shutdownCh      chan struct{} // signals Shutdown was called programmatically
}

// Option configures a Server.
type Option func(*Server)

// New creates a new Server with the given handler and options.
func New(handler http.Handler, opts ...Option) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		shutdownTimeout: 30 * time.Second,
		logger:          slog.Default(),
		shutdownCh:      make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithAddr sets the listen address (default ":8080").
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.httpServer.Addr = addr
	}
}

// WithReadTimeout sets the read timeout (default 15s).
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.httpServer.ReadTimeout = d
	}
}

// WithWriteTimeout sets the write timeout (default 60s).
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.httpServer.WriteTimeout = d
	}
}

// WithIdleTimeout sets the idle timeout (default 120s).
func WithIdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.httpServer.IdleTimeout = d
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout (default 30s).
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

// WithLogger sets the structured logger (default slog.Default()).
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// OnStart registers a hook that runs before the server starts listening.
// If any hook returns an error, Start aborts and returns the error.
func (s *Server) OnStart(fn func() error) {
	s.onStart = append(s.onStart, fn)
}

// OnShutdown registers a hook that runs after HTTP connections are drained.
// Errors are logged but all hooks still execute.
func (s *Server) OnShutdown(fn func(ctx context.Context) error) {
	s.onShutdown = append(s.onShutdown, fn)
}

// Start runs the server and blocks until a shutdown signal (SIGINT/SIGTERM)
// is received or Shutdown is called programmatically.
func (s *Server) Start() error {
	// Run OnStart hooks
	for _, fn := range s.onStart {
		if err := fn(); err != nil {
			return err
		}
	}

	// Start listening
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	s.logger.Info("server starting", "addr", ln.Addr().String())

	// Serve in background
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for signal or programmatic shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-s.shutdownCh:
		s.logger.Info("shutdown signal received", "signal", "programmatic")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown error", "error", err)
		return err
	}

	// Run OnShutdown hooks
	for _, fn := range s.onShutdown {
		if err := fn(ctx); err != nil {
			s.logger.Error("shutdown hook error", "error", err)
		}
	}

	s.logger.Info("server stopped")
	return nil
}

// Shutdown triggers a graceful shutdown of the server.
func (s *Server) Shutdown(ctx context.Context) error {
	select {
	case s.shutdownCh <- struct{}{}:
	default:
	}
	return nil
}
