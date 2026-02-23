# apikit

[![Go Reference](https://pkg.go.dev/badge/github.com/KARTIKrocks/apikit.svg)](https://pkg.go.dev/github.com/KARTIKrocks/apikit)
[![Go Report Card](https://goreportcard.com/badge/github.com/KARTIKrocks/apikit)](https://goreportcard.com/report/github.com/KARTIKrocks/apikit)

A production-ready Go toolkit for building REST APIs. Zero mandatory dependencies. Works with any `net/http` compatible router.

## Features

- **`errors`** — Structured API errors with `errors.Is`/`errors.As` support, error codes, and sentinel errors
- **`request`** — Generic body binding (`Bind[T]`), query/path/header parsing, pagination, sorting, filtering
- **`response`** — Consistent JSON envelope, fluent builder, pagination helpers, SSE streaming
- **`middleware`** — Request ID, logging, panic recovery, CORS, rate limiting, auth, security headers, timeout
- **`httpclient`** — HTTP client with retries, exponential backoff, circuit breaker, and `HTTPClient` interface for mocking
- **`router`** — Route grouping with `.Get()`/`.Post()` method helpers, prefix groups, and per-group middleware on top of `http.ServeMux`
- **`server`** — Graceful shutdown wrapper with signal handling and lifecycle hooks
- **`apitest`** — Fluent test helpers for recording and asserting HTTP handler responses

## Install

```bash
go get github.com/KARTIKrocks/apikit
```

Requires **Go 1.22+**.

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/KARTIKrocks/apikit/errors"
    "github.com/KARTIKrocks/apikit/middleware"
    "github.com/KARTIKrocks/apikit/request"
    "github.com/KARTIKrocks/apikit/response"
    "github.com/KARTIKrocks/apikit/router"
    "github.com/KARTIKrocks/apikit/server"
)

type CreateUserReq struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func createUser(w http.ResponseWriter, r *http.Request) error {
    req, err := request.Bind[CreateUserReq](r)
    if err != nil {
        return err // Automatically sends 400 with structured error
    }

    // Your business logic...
    if req.Email == "taken@example.com" {
        return errors.Conflict("Email already in use")
    }

    user := map[string]string{"id": "123", "name": req.Name}
    response.Created(w, "User created", user)
    return nil
}

func main() {
    r := router.New()
    r.Use(
        middleware.RequestID(),
        middleware.Recover(),
        middleware.Timeout(30 * time.Second),
    )

    r.Post("/users", createUser)

    // Graceful shutdown with signal handling (SIGINT/SIGTERM)
    srv := server.New(r, server.WithAddr(":8080"))
    srv.OnStart(func() error {
        log.Println("server started...")
        return nil
    })
    srv.OnShutdown(func(ctx context.Context) error {
        log.Println("closing database...")
        return nil // db.Close()
    })
    if err := srv.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Packages

### errors

Structured API errors that integrate with Go's standard `errors` package.

```go
import "github.com/KARTIKrocks/apikit/errors"

// Create errors
err := errors.NotFound("User")            // 404
err := errors.BadRequest("Invalid input") // 400
err := errors.Validation("Invalid", map[string]string{
    "email": "is required",
    "name":  "too short",
}) // 422

// Wrap underlying errors
err := errors.Internal("Database failed").Wrap(dbErr)

// Add details
err := errors.Conflict("Duplicate").
    WithField("email", "already exists").
    WithDetail("existing_id", "abc123")

// Check errors anywhere in your code
if errors.Is(err, errors.ErrNotFound) { ... }
if errors.Is(err, errors.ErrValidation) { ... }

// Extract error details
var apiErr *errors.Error
if errors.As(err, &apiErr) {
    log.Printf("Status: %d, Code: %s", apiErr.StatusCode, apiErr.Code)
}

// Custom error codes
errors.RegisterCode("SUBSCRIPTION_EXPIRED", 402)
err := errors.New("SUBSCRIPTION_EXPIRED", "Your subscription has expired")
```

### request

Generic request binding and parameter parsing.

```go
import "github.com/KARTIKrocks/apikit/request"

// --- Body binding ---
type CreatePostReq struct {
    Title string `json:"title"`
    Body  string `json:"body"`
}

// Generic binding (returns typed value, not pointer)
post, err := request.Bind[CreatePostReq](r) // Checks content-type, size limits, decodes JSON

// --- Path parameters (Go 1.22+ stdlib routing) ---
// Route: "GET /posts/{id}"
id := request.PathParam(r, "id")
id, err := request.PathParamInt(r, "id")

// --- Query parameters ---
q := request.QueryFrom(r)
search := q.String("search", "")                // Default: ""
page, err := q.Int("page", 1)                   // Default: 1
active, err := q.Bool("active", true)            // Accepts: true/false/1/0/yes/no
tags := q.StringSlice("tags")                    // ?tags=go,api → ["go", "api"]
limit, err := q.IntRange("limit", 20, 1, 100)   // Clamped to [1, 100]

// --- Headers ---
token := request.BearerToken(r)                  // Extract Bearer token
ip := request.ClientIP(r)                        // Respects X-Forwarded-For
reqID := request.RequestID(r)                     // X-Request-ID or X-Trace-ID

// --- Pagination ---
pg, err := request.Paginate(r)                   // ?page=2&per_page=25
// pg.Page=2, pg.PerPage=25, pg.Offset=25
// Also detects ?page_size=25 and ?limit=25 automatically

// Navigation helpers (need total count from your DB)
pg.HasNext(total)                                // true if more pages exist
pg.HasPrevious()                                 // true if not first page
pg.TotalPages(total)                             // total number of pages
pg.NextPage()                                    // next page number
pg.PreviousPage()                                // previous page number (min 1)

// SQL helpers
pg.SQLClause()                                   // "LIMIT 25 OFFSET 25"
pg.SQLClauseMySQL()                              // "LIMIT 25, 25"

cursor, err := request.PaginateCursor(r)         // ?cursor=abc&limit=25

// --- Sorting ---
sorts, err := request.ParseSort(r, request.SortConfig{
    AllowedFields: []string{"name", "created_at"},
})
// ?sort=name,-created_at → [{name, asc}, {created_at, desc}]

// --- Filtering ---
filters, err := request.ParseFilters(r, request.FilterConfig{
    AllowedFields: []string{"status", "role"},
})
// ?filter[status]=active&filter[age][gte]=18

// --- Struct tag validation (automatic in Bind[T]) ---
type CreateUserReq struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"oneof=admin user mod"`
}
// Bind[T] automatically validates tags before returning.
// Supported: required, email, url, min, max, len, oneof, alpha,
//            alphanum, numeric, uuid, contains, startswith, endswith

// --- Programmatic validation (for cross-field logic) ---
v := request.NewValidation()
v.RequireString("name", req.Name)
v.RequireEmail("email", req.Email)
v.RequireURL("website", req.Website)
v.UUID("id", req.ID)
v.MinLength("name", req.Name, 2)
v.OneOf("role", req.Role, []string{"admin", "user", "mod"})
v.MatchesPattern("code", req.Code, `^[A-Z]{3}-\d{4}$`, "must match format XXX-0000")
v.Custom("end_date", func() bool {
    return req.EndDate.After(req.StartDate)
}, "must be after start_date")
if err := v.Error(); err != nil {
    return err // Returns structured 422 with field errors
}
```

### response

Consistent JSON responses with a standard envelope.

```go
import "github.com/KARTIKrocks/apikit/response"

// --- Success responses ---
response.OK(w, "Success", data)          // 200
response.Created(w, "Created", data)     // 201
response.Accepted(w, "Accepted", nil)    // 202
response.NoContent(w)                    // 204

// --- Error responses ---
response.BadRequest(w, "Invalid input")
response.Unauthorized(w, "Login required")
response.NotFound(w, "User not found")
response.ValidationError(w, map[string]string{"email": "required"})

// --- From errors package (recommended) ---
response.Err(w, errors.NotFound("User"))
response.Err(w, err) // Any error — *errors.Error gets proper status, others get 500

// --- Builder pattern ---
response.New().
    Status(201).
    Message("Created").
    Data(user).
    Header("X-Resource-ID", user.ID).
    Pagination(1, 20, 150).
    Send(w)

// --- Paginated ---
response.Paginated(w, users, response.NewPageMeta(page, perPage, total))
// NewPageMeta auto-computes TotalPages, HasNext, and HasPrevious

// Link header (RFC 5988) for API clients
response.SetLinkHeader(w, "https://api.example.com/users", page, perPage, total)
// Link: <...?page=2&per_page=20>; rel="next", <...?page=1&per_page=20>; rel="first", ...

response.CursorPaginated(w, events, response.CursorMeta{
    NextCursor: "eyJpZCI6MTAwfQ==",
    HasMore:    true,
})

// --- Streaming (SSE) ---
response.StreamJSON(w, func(send func(event string, data any) error) error {
    for msg := range messages {
        send("update", msg)
    }
    return nil
})

// --- Handler wrapper ---
// Converts func(w, r) error → http.HandlerFunc
mux.HandleFunc("GET /users/{id}", response.Handle(getUser))
```

**Response envelope format:**

```json
{
  "success": true,
  "message": "User created",
  "data": { "id": "123", "name": "Alice" },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  },
  "timestamp": 1700000000
}
```

### router

Route grouping and method helpers on top of `http.ServeMux`.

```go
import "github.com/KARTIKrocks/apikit/router"

// Create a router (implements http.Handler)
r := router.New()

// Global middleware
r.Use(middleware.RequestID(), middleware.Logger(slog.Default()))

// Method helpers — handlers return error
r.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
    response.OK(w, "OK", nil)
    return nil
})

// Standard http.HandlerFunc (no error return) — use GetFunc/PostFunc/etc.
r.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
    response.OK(w, "OK", map[string]string{"version": "1.0.0"})
})

// Route groups with prefix and per-group middleware
api := r.Group("/api/v1", authMiddleware)
api.Get("/users", listUsers)
api.Post("/users", createUser)
api.GetFunc("/users/{id}", getUser)     // stdlib handler

// Nested groups accumulate prefix and middleware
admin := api.Group("/admin", adminOnly)
admin.Delete("/users/{id}", deleteUser)
// Registers "DELETE /api/v1/admin/users/{id}" with auth + adminOnly middleware

// Handle/HandleFunc for http.Handler (e.g. file servers)
api.Handle("GET /docs", http.FileServer(http.Dir("./docs")))

// Use with server package
srv := server.New(r, server.WithAddr(":8080"))
srv.Start()
```

Middleware is resolved at registration time. `Use()` only applies to routes registered after the call, matching chi/echo/gin behavior.

### middleware

Production-ready middleware that works with any `net/http` router.

```go
import "github.com/KARTIKrocks/apikit/middleware"

// Chain middleware (applied in order)
stack := middleware.Chain(
    middleware.RequestID(),
    middleware.Logger(slog.Default()),
    middleware.Recover(),
    middleware.SecureHeaders(),
    middleware.CORS(middleware.DefaultCORSConfig()),
    middleware.RateLimit(middleware.RateLimitConfig{Rate: 100, Window: time.Minute}),
    middleware.Timeout(30 * time.Second),
    middleware.BodyLimit(5 << 20), // 5 MB
)
handler := stack(mux)

// --- Authentication ---
auth := middleware.Auth(middleware.AuthConfig{
    Authenticate: func(ctx context.Context, token string) (any, error) {
        user, err := verifyJWT(token)
        if err != nil {
            return nil, errors.Unauthorized("Invalid token")
        }
        return user, nil
    },
    SkipPaths: map[string]bool{"/health": true, "/login": true},
})

// In handlers, retrieve the user:
user, ok := middleware.GetAuthUserAs[*User](r.Context())

// --- Get request ID anywhere ---
reqID := middleware.GetRequestID(r.Context())

// --- Custom rate limiter backend ---
type RedisLimiter struct { ... }
func (rl *RedisLimiter) Allow(key string) bool { ... }

middleware.RateLimit(middleware.RateLimitConfig{
    Limiter: &RedisLimiter{},
})
```

### httpclient

HTTP client with retries, exponential backoff, circuit breaker, and an interface for easy mocking.

```go
import "github.com/KARTIKrocks/apikit/httpclient"

// Create a client with functional options
client := httpclient.New("https://api.example.com",
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithMaxRetries(3),
    httpclient.WithRetryDelay(500 * time.Millisecond),
    httpclient.WithMaxRetryDelay(5 * time.Second),
    httpclient.WithLogger(slog.Default()),
)

// Basic requests
resp, err := client.Get(ctx, "/users")
resp, err := client.Post(ctx, "/users", map[string]string{"name": "Alice"})
resp, err := client.Put(ctx, "/users/1", updateBody)
resp, err := client.Patch(ctx, "/users/1", patchBody)
resp, err := client.Delete(ctx, "/users/1")

// Response helpers
var users []User
resp.JSON(&users)
fmt.Println(resp.String(), resp.StatusCode, resp.IsSuccess())

// Set default headers
client.SetBearerToken("my-token")
client.SetHeader("X-API-Key", "key")

// Fluent request builder
resp, err := client.Request().
    Method("POST").
    Path("/search").
    Header("X-Custom", "value").
    Param("q", "golang").
    Body(searchReq).
    Send(ctx)

// Enable circuit breaker (opens after 5 failures, resets after 30s)
client := httpclient.New("https://api.example.com",
    httpclient.WithCircuitBreaker(5, 30 * time.Second),
)

// --- Mocking in tests ---
// HTTPClient interface allows easy substitution
func fetchUsers(client httpclient.HTTPClient) ([]User, error) { ... }

// In tests:
mock := httpclient.NewMockClient()
mock.OnGet("/users", 200, []byte(`[{"id":1,"name":"Alice"}]`))
mock.OnPost("/users", 201, []byte(`{"id":2}`))
mock.OnError("GET", "/fail", fmt.Errorf("connection refused"))

users, err := fetchUsers(mock)
fmt.Println(mock.GetCallCount()) // 1
```

### server

Production-ready HTTP server with graceful shutdown, signal handling, and lifecycle hooks.

```go
import "github.com/KARTIKrocks/apikit/server"

srv := server.New(handler,
    server.WithAddr(":8080"),
    server.WithReadTimeout(15 * time.Second),
    server.WithWriteTimeout(60 * time.Second),
    server.WithIdleTimeout(120 * time.Second),
    server.WithShutdownTimeout(10 * time.Second),
    server.WithLogger(slog.Default()),
)

// Lifecycle hooks
srv.OnStart(func() error {
    slog.Info("connecting to database...")
    return db.Connect()
})
srv.OnShutdown(func(ctx context.Context) error {
    return db.Close()
})

// Blocks until SIGINT/SIGTERM, then drains connections gracefully
if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

### apitest

Fluent test helpers for building requests and asserting responses against your handlers.

```go
import "github.com/KARTIKrocks/apikit/apitest"

// Build a request
req := apitest.NewRequest("POST", "/users").
    WithBody(map[string]string{"name": "Alice"}).
    WithBearerToken("valid-token").
    WithHeader("X-Request-ID", "test-123").
    WithQuery("notify", "true").
    WithPathValue("id", "42").
    Build()

// Record handler response
resp := apitest.RecordHandler(createUser, req)

// Fluent assertions
resp.AssertStatus(t, 201)
resp.AssertSuccess(t)
resp.AssertHeader(t, "X-Request-ID", "test-123")
resp.AssertBodyContains(t, "Alice")
resp.AssertError(t, "NOT_FOUND")
resp.AssertValidationError(t, "email")

// Decode response
var user User
resp.Decode(&user)

// Access envelope directly
env, _ := resp.Envelope()
fmt.Println(env.Success, env.Message)
```

## Design Principles

| Principle             | How                                                            |
| --------------------- | -------------------------------------------------------------- |
| **stdlib compatible** | Works with `http.Handler`, `http.HandlerFunc`, any router      |
| **Zero dependencies** | Core uses only the Go standard library                         |
| **Generics**          | `Bind[T]`, `GetAuthUserAs[T]` for type safety                  |
| **Interface-driven**  | `RateLimiter`, `Logger` interfaces — plug in your own backends |
| **Composable**        | Each package is independently usable                           |
| **Go 1.22+**          | Leverages enhanced `http.ServeMux` routing                     |

## Roadmap

- [ ] `health` — Health check endpoint builder with dependency checks
- [ ] `ctxutil` — Typed context helpers
- [ ] `observe` — OpenTelemetry integration

## License

[MIT](LICENSE)
