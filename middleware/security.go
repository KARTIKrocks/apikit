package middleware

import (
	"net/http"
)

// SecurityHeadersConfig configures security headers.
type SecurityHeadersConfig struct {
	// ContentTypeNosniff sets X-Content-Type-Options: nosniff
	// Default: true
	ContentTypeNosniff bool

	// XFrameOptions sets X-Frame-Options. Common values: DENY, SAMEORIGIN
	// Default: "DENY"
	XFrameOptions string

	// XSSProtection sets X-XSS-Protection.
	// Default: "1; mode=block"
	XSSProtection string

	// HSTSMaxAge sets Strict-Transport-Security max-age in seconds.
	// Set to 0 to disable. Default: 31536000 (1 year)
	HSTSMaxAge int

	// HSTSIncludeSubdomains adds includeSubDomains to HSTS.
	// Default: true
	HSTSIncludeSubdomains bool

	// ReferrerPolicy sets Referrer-Policy header.
	// Default: "strict-origin-when-cross-origin"
	ReferrerPolicy string

	// ContentSecurityPolicy sets Content-Security-Policy header.
	// Default: "" (not set)
	ContentSecurityPolicy string

	// PermissionsPolicy sets Permissions-Policy header.
	// Default: "" (not set)
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns recommended security headers.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		XFrameOptions:         "DENY",
		XSSProtection:         "1; mode=block",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// SecureHeaders adds security headers to responses.
func SecureHeaders() Middleware {
	return SecureHeadersWithConfig(DefaultSecurityHeadersConfig())
}

// SecureHeadersWithConfig adds security headers with custom configuration.
func SecureHeadersWithConfig(cfg SecurityHeadersConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.ContentTypeNosniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			if cfg.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}

			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}

			if cfg.HSTSMaxAge > 0 {
				w.Header().Set("Strict-Transport-Security",
					formatHSTS(cfg.HSTSMaxAge, cfg.HSTSIncludeSubdomains))
			}

			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func formatHSTS(maxAge int, includeSubdomains bool) string {
	s := "max-age=" + itoa(maxAge)
	if includeSubdomains {
		s += "; includeSubDomains"
	}
	return s
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
