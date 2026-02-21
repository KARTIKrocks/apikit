package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

type authUserKey struct{}

// AuthConfig configures authentication middleware.
type AuthConfig struct {
	// Authenticate is the function that validates credentials and returns user info.
	// The token/key string is extracted from the request based on the scheme.
	// Return any user object to store in context, or error to reject.
	Authenticate func(ctx context.Context, token string) (any, error)

	// Scheme determines where to extract credentials from.
	// Default: "bearer"
	// Options: "bearer", "api-key"
	Scheme string

	// APIKeyHeader is the header name for API key auth.
	// Default: "X-API-Key"
	APIKeyHeader string

	// SkipPaths are paths that bypass authentication.
	SkipPaths map[string]bool

	// ErrorMessage is the message for unauthorized responses.
	// Default: "Authentication required"
	ErrorMessage string
}

// Auth creates an authentication middleware.
//
// Example with bearer token:
//
//	auth := middleware.Auth(middleware.AuthConfig{
//	    Authenticate: func(ctx context.Context, token string) (any, error) {
//	        user, err := verifyJWT(token)
//	        if err != nil { return nil, errors.Unauthorized("Invalid token") }
//	        return user, nil
//	    },
//	    SkipPaths: map[string]bool{"/health": true, "/login": true},
//	})
func Auth(cfg AuthConfig) Middleware {
	if cfg.Authenticate == nil {
		panic("apikit/middleware: Auth requires Authenticate function")
	}
	if cfg.Scheme == "" {
		cfg.Scheme = "bearer"
	}
	if cfg.APIKeyHeader == "" {
		cfg.APIKeyHeader = "X-API-Key"
	}
	if cfg.ErrorMessage == "" {
		cfg.ErrorMessage = "Authentication required"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip configured paths
			if cfg.SkipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token based on scheme
			token, err := extractToken(r, cfg)
			if err != nil {
				response.Err(w, err)
				return
			}

			if token == "" {
				response.Err(w, errors.Unauthorized(cfg.ErrorMessage))
				return
			}

			// Validate token
			user, authErr := cfg.Authenticate(r.Context(), token)
			if authErr != nil {
				response.Err(w, authErr)
				return
			}

			// Store user in context
			ctx := context.WithValue(r.Context(), authUserKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthUser retrieves the authenticated user from the request context.
// Returns nil if no user is set.
func GetAuthUser(ctx context.Context) any {
	return ctx.Value(authUserKey{})
}

// GetAuthUserAs retrieves the authenticated user with type assertion.
//
//	user, ok := middleware.GetAuthUserAs[*User](r.Context())
func GetAuthUserAs[T any](ctx context.Context) (T, bool) {
	val := ctx.Value(authUserKey{})
	if val == nil {
		var zero T
		return zero, false
	}
	user, ok := val.(T)
	return user, ok
}

// RequireRole creates middleware that checks if the authenticated user has a required role.
// The roleExtractor function should return the user's roles from the context.
func RequireRole(role string, roleExtractor func(ctx context.Context) []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			roles := roleExtractor(req.Context())

			for _, rl := range roles {
				if rl == role {
					next.ServeHTTP(w, req)
					return
				}
			}

			response.Err(w, errors.Forbidden("Insufficient permissions"))
		})
	}
}

// extractToken extracts the authentication token from the request.
func extractToken(r *http.Request, cfg AuthConfig) (string, error) {
	switch strings.ToLower(cfg.Scheme) {
	case "bearer":
		auth := r.Header.Get("Authorization")
		if auth == "" {
			return "", nil
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", errors.BadRequest("Invalid Authorization header format. Expected: Bearer <token>")
		}
		return strings.TrimSpace(parts[1]), nil

	case "api-key":
		return r.Header.Get(cfg.APIKeyHeader), nil

	default:
		return "", errors.Internal("Unsupported auth scheme: " + cfg.Scheme)
	}
}
