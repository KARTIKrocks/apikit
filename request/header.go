package request

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// Header returns a request header value, or the default.
func Header(r *http.Request, key, defaultVal string) string {
	val := r.Header.Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// HeaderRequired returns a header value or an error if missing.
func HeaderRequired(r *http.Request, key string) (string, error) {
	val := r.Header.Get(key)
	if val == "" {
		return "", errors.BadRequest(fmt.Sprintf("Header %q is required", key))
	}
	return val, nil
}

// BearerToken extracts the Bearer token from the Authorization header.
// Returns empty string if not present or not a Bearer token.
func BearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// BearerTokenRequired extracts the Bearer token or returns an error.
func BearerTokenRequired(r *http.Request) (string, error) {
	token := BearerToken(r)
	if token == "" {
		return "", errors.Unauthorized("Bearer token is required")
	}
	return token, nil
}

// APIKey extracts an API key from the specified header (commonly "X-API-Key").
func APIKey(r *http.Request, headerName string) string {
	return r.Header.Get(headerName)
}

// APIKeyRequired extracts an API key or returns an error.
func APIKeyRequired(r *http.Request, headerName string) (string, error) {
	key := r.Header.Get(headerName)
	if key == "" {
		return "", errors.Unauthorized(fmt.Sprintf("Header %q is required", headerName))
	}
	return key, nil
}

// ContentType returns the request's Content-Type (media type only, without params).
func ContentType(r *http.Request) string {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return ""
	}
	// Return just the media type, not params like charset
	if idx := strings.IndexByte(ct, ';'); idx != -1 {
		ct = ct[:idx]
	}
	return strings.TrimSpace(strings.ToLower(ct))
}

// AcceptsJSON checks if the client accepts JSON responses.
func AcceptsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true // Default to JSON if no Accept header
	}
	return strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "*/*")
}

// ClientIP returns the client's IP address.
// When trustProxy is true, proxy headers (X-Forwarded-For, X-Real-IP) are checked first.
// When false, only RemoteAddr is used, which is safe against header spoofing.
//
// For production behind a reverse proxy, set trustProxy to true.
// For direct-to-internet deployments, set trustProxy to false.
func ClientIP(r *http.Request, trustProxy ...bool) string {
	trust := false
	if len(trustProxy) > 0 {
		trust = trustProxy[0]
	}

	if trust {
		// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.IndexByte(xff, ','); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// Use net.SplitHostPort to correctly handle both IPv4 and IPv6 addresses.
	// RemoteAddr is typically "ip:port" or "[::1]:port" for IPv6.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port (unlikely but handle gracefully)
		return r.RemoteAddr
	}
	return host
}

// RequestID returns the request ID from common headers.
// Checks: X-Request-ID → X-Trace-ID
func RequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return r.Header.Get("X-Trace-ID")
}

// IdempotencyKey returns the Idempotency-Key header value.
func IdempotencyKey(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}
