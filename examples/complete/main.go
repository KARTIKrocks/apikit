package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/KARTIKrocks/apikit/config"
	"github.com/KARTIKrocks/apikit/dbx"
	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/health"
	"github.com/KARTIKrocks/apikit/middleware"
	"github.com/KARTIKrocks/apikit/request"
	"github.com/KARTIKrocks/apikit/response"
	"github.com/KARTIKrocks/apikit/router"
	"github.com/KARTIKrocks/apikit/server"
	"github.com/KARTIKrocks/apikit/sqlbuilder"
)

// --- Configuration ---

// AppConfig is populated from environment variables, .env file, or JSON config.
// Priority: env vars > .env file > JSON file > default tags.
type AppConfig struct {
	Port        int           `env:"PORT"         default:"8080"    validate:"min=1,max=65535"`
	DatabaseURL string        `env:"DATABASE_URL" validate:"required"`
	LogLevel    string        `env:"LOG_LEVEL"    default:"info"`
	ReadTimeout time.Duration `env:"READ_TIMEOUT" default:"15s"`
}

// --- Domain types ---

// User is the database model. Fields use both `json` (for API responses)
// and `db` tags (for dbx row scanning).
type User struct {
	ID    string  `json:"id"    db:"id"`
	Name  string  `json:"name"  db:"name"`
	Email string  `json:"email" db:"email"`
	Role  string  `json:"role"  db:"role"`
	Bio   *string `json:"bio"   db:"bio"` // nullable column → pointer
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// Validate implements request.Validator for cross-field logic.
// Bind[T] calls this automatically after struct tag validation passes.
func (r CreateUserRequest) Validate() error {
	v := request.NewValidation()
	v.Custom("email", func() bool {
		return r.Email != r.Name
	}, "must be different from name")
	return v.Error()
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// --- Handlers ---

// allowedColumns maps API field names to actual DB column names.
// Used by sqlbuilder to safely apply user-provided sort/filter params.
var allowedColumns = map[string]string{
	"name":       "u.name",
	"email":      "u.email",
	"role":       "u.role",
	"created_at": "u.created_at",
}

// listUsers handles GET /api/v1/users
// Demonstrates sqlbuilder + dbx + request pagination/sort/filter working together.
func listUsers(w http.ResponseWriter, r *http.Request) error {
	// Parse pagination, sorting, and filters from query string
	pg, err := request.Paginate(r)
	if err != nil {
		return err
	}

	sorts, err := request.ParseSort(r, request.SortConfig{
		AllowedFields: []string{"name", "email", "created_at"},
		Default:       []request.SortField{{Field: "created_at", Direction: request.SortDesc}},
	})
	if err != nil {
		return err
	}

	filters, err := request.ParseFilters(r, request.FilterConfig{
		AllowedFields: []string{"role"},
	})
	if err != nil {
		return err
	}

	// Build the query with sqlbuilder — pagination, sorting, and filters
	// are applied directly from the parsed request parameters.
	q := sqlbuilder.Select("id", "name", "email", "role", "bio").
		From("users u").
		ApplyFilters(filters, allowedColumns).
		ApplySort(sorts, allowedColumns).
		ApplyPagination(pg).
		Query()

	// Execute with dbx — rows are scanned into []User automatically
	// using the `db` struct tags. No manual Scan() calls needed.
	users, err := dbx.QueryAllQ[User](r.Context(), q)
	if err != nil {
		return err
	}

	// Count total for pagination metadata
	countQ := sqlbuilder.Select("COUNT(*)").From("users u").
		ApplyFilters(filters, allowedColumns).
		Query()

	type countRow struct {
		Count int `db:"count"`
	}
	cr, err := dbx.QueryOneQ[countRow](r.Context(), countQ)
	if err != nil {
		return err
	}

	response.SetLinkHeader(w, "http://localhost:8080/api/v1/users", pg.Page, pg.PerPage, cr.Count)
	response.Paginated(w, users, response.NewPageMeta(pg.Page, pg.PerPage, cr.Count))
	return nil
}

// createUser handles POST /api/v1/users
// Demonstrates sqlbuilder INSERT with RETURNING + dbx.QueryOneQ for scanning the result.
func createUser(w http.ResponseWriter, r *http.Request) error {
	req, err := request.Bind[CreateUserRequest](r)
	if err != nil {
		return err
	}

	// Build an INSERT ... RETURNING query with sqlbuilder
	q := sqlbuilder.Insert("users").
		Columns("name", "email", "role").
		Values(req.Name, req.Email, "user").
		Returning("id", "name", "email", "role", "bio").
		Query()

	// dbx scans the RETURNING row directly into User
	user, err := dbx.QueryOneQ[User](r.Context(), q)
	if err != nil {
		return err
	}

	response.Created(w, "User created successfully", user)
	return nil
}

// getUser handles GET /api/v1/users/{id}
// Demonstrates the standard (w, r) pattern — no error return.
// dbx.QueryOne returns errors.CodeNotFound automatically when no rows match.
func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathParamRequired(r, "id")
	if err != nil {
		response.Err(w, err)
		return
	}

	user, err := dbx.QueryOne[User](r.Context(),
		"SELECT id, name, email, role, bio FROM users WHERE id = $1", id)
	if err != nil {
		response.Err(w, err) // CodeNotFound → 404, CodeDatabaseError → 500
		return
	}

	response.OK(w, "User retrieved", user)
}

// updateUser handles PUT /api/v1/users/{id}
// Demonstrates sqlbuilder UPDATE with conditional SET and RETURNING.
func updateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := request.PathParamRequired(r, "id")
	if err != nil {
		return err
	}

	req, err := request.Bind[UpdateUserRequest](r)
	if err != nil {
		return err
	}

	// Build UPDATE query — When() conditionally adds SET clauses
	// only when the field was provided in the request body.
	q := sqlbuilder.Update("users").
		When(req.Name != "", func(b *sqlbuilder.UpdateBuilder) {
			b.Set("name", req.Name)
		}).
		When(req.Email != "", func(b *sqlbuilder.UpdateBuilder) {
			b.Set("email", req.Email)
		}).
		Set("updated_at", sqlbuilder.Raw("NOW()")).
		Where("id = $1", id).
		Returning("id", "name", "email", "role", "bio").
		Query()

	user, err := dbx.QueryOneQ[User](r.Context(), q)
	if err != nil {
		return err
	}

	response.OK(w, "User updated", user)
	return nil
}

// deleteUser handles DELETE /api/v1/users/{id}
// Demonstrates sqlbuilder DELETE + dbx.ExecQ.
func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathParamRequired(r, "id")
	if err != nil {
		response.Err(w, err)
		return
	}

	q := sqlbuilder.Delete("users").Where("id = $1", id).Query()
	if _, err := dbx.ExecQ(r.Context(), q); err != nil {
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

// transferRole demonstrates dbx transaction support.
// All dbx calls using the WithTx context participate in the same transaction.
// Accepts a *sql.DB so it can begin a transaction.
func transferRole(db *sql.DB) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		fromID, err := request.PathParamRequired(r, "from")
		if err != nil {
			return err
		}
		toID, err := request.PathParamRequired(r, "to")
		if err != nil {
			return err
		}

		// Begin a transaction and attach it to the context.
		// All dbx calls with this ctx will use the transaction.
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			return errors.Internal("failed to begin transaction")
		}
		ctx := dbx.WithTx(r.Context(), tx)

		// Both operations use the same transaction
		q1 := sqlbuilder.Update("users").Set("role", "user").Where("id = $1", fromID).Query()
		if _, err := dbx.ExecQ(ctx, q1); err != nil {
			tx.Rollback()
			return err
		}

		q2 := sqlbuilder.Update("users").Set("role", "admin").Where("id = $1", toID).Query()
		if _, err := dbx.ExecQ(ctx, q2); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return errors.Internal("failed to commit transaction")
		}

		response.OK(w, "Role transferred", nil)
		return nil
	}
}

// --- Main ---

func main() {
	// Load configuration from env vars, .env file, and defaults.
	// Priority: env vars > .env file > JSON file > default tags.
	var cfg AppConfig
	config.MustLoad(&cfg,
		config.WithEnvFile(".env"),                // load .env if it exists
		config.WithJSONFile("config.json"),        // load JSON config if it exists
		config.WithPrefix("APP"),                  // APP_PORT, APP_DATABASE_URL, etc.
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Open database connection and register it as the default for dbx.
	// All dbx.QueryAll, dbx.QueryOne, dbx.Exec calls will use this connection.
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	dbx.SetDefault(db)

	// Configure request defaults
	request.SetConfig(request.Config{
		MaxBodySize:           5 << 20, // 5 MB
		DisallowUnknownFields: true,
	})

	// Set up health checker with database and custom checks.
	// Critical checks → unhealthy (503), non-critical → degraded (200).
	hc := health.NewChecker(health.WithTimeout(3 * time.Second))
	hc.AddCheck("database", func(ctx context.Context) error {
		return db.PingContext(ctx)
	})
	hc.AddNonCriticalCheck("cache", func(ctx context.Context) error {
		// In a real app: ping Redis, Memcached, etc.
		return nil
	})

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
			"/health":       true,
			"/api/v1/login": true,
		},
	})

	// Create router with global middleware
	r := router.New()
	r.Use(
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

	// Public health routes — readiness and liveness probes
	r.Get("/health", hc.Handler())      // runs all checks, returns 200 or 503
	r.Get("/livez", hc.LiveHandler())    // always 200, confirms process is alive

	// Protected API routes
	api := r.Group("/api/v1", auth)
	api.Get("/users", listUsers)
	api.Post("/users", createUser)
	api.GetFunc("/users/{id}", getUser)       // stdlib handler (no error return)
	api.Put("/users/{id}", updateUser)        // error-returning pattern
	api.DeleteFunc("/users/{id}", deleteUser) // stdlib handler (no error return)
	api.Get("/events/stream", streamEvents)
	api.Get("/users/{id}/builder", getUserWithBuilder)              // builder pattern
	api.Post("/users/{from}/transfer-role/{to}", transferRole(db))  // transaction example

	srv := server.New(r,
		server.WithAddr(fmt.Sprintf(":%d", cfg.Port)),
		server.WithReadTimeout(cfg.ReadTimeout),
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
