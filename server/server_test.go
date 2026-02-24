package server_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

// generateTestCert creates a self-signed certificate and returns paths to
// the cert and key PEM files in a temporary directory.
func generateTestCert(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	// Self-sign
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}

	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")

	certOut, _ := os.Create(certFile)
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	certOut.Close()

	keyBytes, _ := x509.MarshalECPrivateKey(key)
	keyOut, _ := os.Create(keyFile)
	_ = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyOut.Close()

	return certFile, keyFile
}

func TestServer_TLS(t *testing.T) {
	certFile, keyFile := generateTestCert(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	srv := server.New(handler,
		server.WithAddr(":0"),
		server.WithTLS(certFile, keyFile),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	addr := srv.Addr()
	if addr == nil {
		t.Fatal("expected non-nil address after start")
	}

	// Make HTTPS request with InsecureSkipVerify (self-signed cert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/", addr.String()))
	if err != nil {
		t.Fatalf("https request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body %q, got %q", "ok", string(body))
	}
	if resp.TLS == nil {
		t.Error("expected TLS connection")
	}

	// Shutdown
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
