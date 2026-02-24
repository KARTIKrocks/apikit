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
	"sync/atomic"
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
	doneCh          chan struct{} // closed when Start() returns
	tlsCertFile     string
	tlsKeyFile      string
	listenAddr      atomic.Value  // stores net.Addr after listening
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
		doneCh:          make(chan struct{}),
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

// WithTLS enables HTTPS with the given certificate and key files.
func WithTLS(certFile, keyFile string) Option {
	return func(s *Server) {
		s.tlsCertFile = certFile
		s.tlsKeyFile = keyFile
	}
}

// Addr returns the listener address after the server has started.
// Returns nil if the server has not started yet.
func (s *Server) Addr() net.Addr {
	v := s.listenAddr.Load()
	if v == nil {
		return nil
	}
	return v.(net.Addr)
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

	s.listenAddr.Store(ln.Addr())

	if s.tlsCertFile != "" {
		s.logger.Info("server starting (tls)", "addr", ln.Addr().String())
	} else {
		s.logger.Info("server starting", "addr", ln.Addr().String())
	}

	// Serve in background
	errCh := make(chan error, 1)
	go func() {
		var serveErr error
		if s.tlsCertFile != "" {
			serveErr = s.httpServer.ServeTLS(ln, s.tlsCertFile, s.tlsKeyFile)
		} else {
			serveErr = s.httpServer.Serve(ln)
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
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
	close(s.doneCh)
	return nil
}

// Shutdown triggers a graceful shutdown of the server and blocks until
// the shutdown completes or the context expires.
func (s *Server) Shutdown(ctx context.Context) error {
	select {
	case s.shutdownCh <- struct{}{}:
	default:
	}

	// Wait for Start() to finish or context to expire.
	select {
	case <-s.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
