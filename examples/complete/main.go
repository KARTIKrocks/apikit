package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/middleware"
	"github.com/KARTIKrocks/apikit/request"
	"github.com/KARTIKrocks/apikit/response"
	"github.com/KARTIKrocks/apikit/server"
)

// --- Domain types ---

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Validate implements request.Validator
func (r CreateUserRequest) Validate() error {
	v := request.NewValidation()
	v.RequireString("name", r.Name)
	v.MinLength("name", r.Name, 2)
	v.MaxLength("name", r.Name, 100)
	v.RequireString("email", r.Email)
	return v.Error()
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// --- Handlers ---

// listUsers handles GET /api/v1/users
func listUsers(w http.ResponseWriter, r *http.Request) error {
	// Parse pagination
	pg, err := request.Paginate(r)
	if err != nil {
		return err
	}

	// Parse sorting
	sorts, err := request.ParseSort(r, request.SortConfig{
		AllowedFields: []string{"name", "email", "created_at"},
		Default:       []request.SortField{{Field: "created_at", Direction: request.SortDesc}},
	})
	if err != nil {
		return err
	}

	// Parse filters
	filters, err := request.ParseFilters(r, request.FilterConfig{
		AllowedFields: []string{"role", "status"},
	})
	if err != nil {
		return err
	}

	_ = sorts   // Use for DB query
	_ = filters // Use for DB query

	// Simulate DB query
	users := []User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", Role: "admin"},
		{ID: "2", Name: "Bob", Email: "bob@example.com", Role: "user"},
	}
	total := 50

	// Set RFC 5988 Link header for API clients
	response.SetLinkHeader(w, "http://localhost:8080/api/v1/users", pg.Page, pg.PerPage, total)

	// Send paginated response (HasNext/HasPrevious computed automatically)
	response.Paginated(w, users, response.NewPageMeta(pg.Page, pg.PerPage, total))
	return nil
}

// createUser handles POST /api/v1/users
func createUser(w http.ResponseWriter, r *http.Request) error {
	// Bind and decode request body
	req, err := request.Bind[CreateUserRequest](r)
	if err != nil {
		return err
	}

	// Validate
	if verr := req.Validate(); verr != nil {
		return verr
	}

	// Simulate duplicate check
	if req.Email == "taken@example.com" {
		return errors.Conflict("A user with this email already exists").
			WithField("email", "already in use")
	}

	// Simulate creation
	user := User{
		ID:    "new-123",
		Name:  req.Name,
		Email: req.Email,
		Role:  "user",
	}

	response.Created(w, "User created successfully", user)
	return nil
}

// getUser handles GET /api/v1/users/{id}
// Demonstrates the standard (w, r) pattern — no return type.
func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathParamRequired(r, "id")
	if err != nil {
		response.Err(w, err)
		return
	}

	// Simulate DB lookup
	if id != "1" {
		response.Err(w, errors.NotFound("User"))
		return
	}

	user := User{ID: "1", Name: "Alice", Email: "alice@example.com", Role: "admin"}
	response.OK(w, "User retrieved", user)
}

// updateUser handles PUT /api/v1/users/{id}
func updateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := request.PathParamRequired(r, "id")
	if err != nil {
		return err
	}

	req, err := request.Bind[UpdateUserRequest](r)
	if err != nil {
		return err
	}

	// Simulate update
	user := User{ID: id, Name: req.Name, Email: req.Email, Role: "user"}
	response.OK(w, "User updated", user)
	return nil
}

// deleteUser handles DELETE /api/v1/users/{id}
// Demonstrates the standard (w, r) pattern — no return type.
func deleteUser(w http.ResponseWriter, r *http.Request) {
	_, err := request.PathParamRequired(r, "id")
	if err != nil {
		response.Err(w, err)
		return
	}

	response.NoContent(w)
}

// streamEvents demonstrates SSE streaming
func streamEvents(w http.ResponseWriter, r *http.Request) error {
	response.StreamJSON(w, func(send func(event string, data any) error) error {
		for i := range 5 {
			msg := map[string]any{
				"id":      i,
				"message": fmt.Sprintf("Event %d", i),
			}
			if err := send("update", msg); err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
		}
		return nil
	})
	return nil
}

// Builder pattern example
func getUserWithBuilder(w http.ResponseWriter, r *http.Request) error {
	user := User{ID: "1", Name: "Alice", Email: "alice@example.com"}

	response.New().
		Status(http.StatusOK).
		Message("User retrieved").
		Data(user).
		Header("X-Resource-Version", "3").
		Send(w)

	return nil
}

// --- Main ---

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Configure request defaults
	request.SetConfig(request.Config{
		MaxBodySize:           5 << 20, // 5 MB
		DisallowUnknownFields: true,
	})

	// Build middleware stack
	stack := middleware.Chain(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Recover(),
		middleware.SecureHeaders(),
		middleware.CORS(middleware.CORSConfig{
			AllowOrigins:     []string{"https://myapp.com", "http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middleware.RateLimit(middleware.RateLimitConfig{
			Rate:       100,
			Window:     time.Minute,
			TrustProxy: true, // enable if behind a reverse proxy
		}),
		middleware.Timeout(30*time.Second),
		middleware.BodyLimit(5<<20),
	)

	// Auth middleware (for protected routes)
	auth := middleware.Auth(middleware.AuthConfig{
		Authenticate: func(ctx context.Context, token string) (any, error) {
			// In real app: verify JWT, lookup user
			if token == "valid-token" {
				return &User{ID: "1", Name: "Alice", Role: "admin"}, nil
			}
			return nil, errors.Unauthorized("Invalid or expired token")
		},
		SkipPaths: map[string]bool{
			"/health":         true,
			"/api/v1/login":   true,
		},
	})

	// Setup routes (Go 1.22+ enhanced routing)
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "OK", map[string]string{"status": "healthy"})
	})

	// Protected API routes
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/users", response.Handle(listUsers))
	api.HandleFunc("POST /api/v1/users", response.Handle(createUser))
	api.HandleFunc("GET /api/v1/users/{id}", getUser)                       // standard (w, r) pattern
	api.HandleFunc("PUT /api/v1/users/{id}", response.Handle(updateUser))    // error-returning pattern
	api.HandleFunc("DELETE /api/v1/users/{id}", deleteUser)                  // standard (w, r) pattern
	api.HandleFunc("GET /api/v1/events/stream", response.Handle(streamEvents))
	api.HandleFunc("GET /api/v1/users/{id}/builder", response.Handle(getUserWithBuilder)) // builder pattern

	// Compose: public routes + auth-protected API routes
	mux.Handle("/api/", auth(api))

	// Apply global middleware stack
	handler := stack(mux)

	srv := server.New(handler,
		server.WithAddr(":8080"),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(60*time.Second),
		server.WithIdleTimeout(120*time.Second),
		server.WithShutdownTimeout(10*time.Second),
		server.WithLogger(logger),
	)

	if err := srv.Start(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
