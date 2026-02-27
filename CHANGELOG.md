# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2026-02-27

### Added

- **dbx** — New `dbx` package: lightweight, generic row scanner for `database/sql` — eliminates scan boilerplate while keeping full SQL control
- **dbx** — `SetDefault(db)` sets the package-level connection once at startup; all functions use it automatically
- **dbx** — `WithTx(ctx, tx)` returns a context that overrides the default with a transaction — all dbx calls in that context use the tx
- **dbx** — `QueryAll[T](ctx, query, args...)` scans all rows into `[]T`; returns empty slice for no rows
- **dbx** — `QueryOne[T](ctx, query, args...)` scans the first row into `T`; returns `errors.CodeNotFound` for no rows
- **dbx** — `Exec(ctx, query, args...)` executes non-returning queries with `errors.CodeDatabaseError` wrapping
- **dbx** — Q-suffixed variants (`QueryAllQ`, `QueryOneQ`, `ExecQ`) accept `sqlbuilder.Query` directly
- **dbx** — Struct mapping via `db:"column_name"` tags; `db:"-"` to skip; untagged fields ignored
- **dbx** — Pointer fields (`*string`, `*int`, etc.) handle SQL NULLs naturally
- **dbx** — Column matching is order-independent; unmatched result columns silently discarded
- **dbx** — Embedded struct support (including exported pointer embeds like `*Base`)
- **dbx** — Reflection-based type mapping cached per-type via `sync.Map` (built once per process lifetime)

## [0.8.0] - 2026-02-25

### Added

- **config** — New `config` package: load application configuration from environment variables, `.env` files, and JSON config files into typed Go structs
- **config** — `Load(dst, ...Option)` populates a struct from sources in priority order (env vars > .env file > JSON file > default tags) and validates with `request.ValidateStruct`
- **config** — `MustLoad(dst, ...Option)` calls `Load` and panics on error, for use in `main`/`init`
- **config** — Struct tags: `env:"VAR_NAME"` maps fields to env vars, `default:"value"` sets fallbacks, `validate:"..."` reuses existing request validators
- **config** — Options: `WithPrefix(p)` for prefixed env vars (e.g., `APP_PORT`), `WithEnvFile(path)` for `.env` files, `WithJSONFile(path)` for JSON base config, `WithRequired()` to error on missing files
- **config** — Supported types: `string`, `bool`, `int/int8/16/32/64`, `uint` variants, `float32/64`, `time.Duration`, `[]string`, `[]int`
- **config** — Nested struct support: automatically flattened to env var names (e.g., `DB.Host` → `DB_HOST`)
- **config** — `.env` file parsing: comments, blank lines, quoted values (single/double), inline comments
- **config** — JSON config flattening: nested objects become underscore-separated uppercase keys
- **config** — Structured error messages using `*errors.Error` for parse errors, validation failures, and missing files

## [0.7.0] - 2026-02-25

### Added

- **sqlbuilder** — New `sqlbuilder` package: fluent, type-safe SQL query builder with multi-dialect support (PostgreSQL, MySQL, SQLite)
- **sqlbuilder** — `Select`, `Insert`, `Update`, `Delete` builders with chainable API and `Build() (string, []any)` terminal method
- **sqlbuilder** — Dialect support: `Postgres` (default, `$1, $2, ...`), `MySQL` and `SQLite` (`?` placeholders) via `SetDialect(d)` or `SelectWith(d, ...)` / `InsertWith(d, ...)` / `UpdateWith(d, ...)` / `DeleteWith(d, ...)` constructors
- **sqlbuilder** — SELECT: `Distinct`, `Column`, `Columns`, `ColumnExpr`, `From`, `FromAlias`, `FromSubquery`
- **sqlbuilder** — JOIN support: `Join`, `LeftJoin`, `RightJoin`, `FullJoin`, `CrossJoin` with parameterized ON clauses
- **sqlbuilder** — WHERE conditions: `Where`, `WhereEq`, `WhereNeq`, `WhereGt`, `WhereGte`, `WhereLt`, `WhereLte`, `WhereLike`, `WhereILike`, `WhereIn`, `WhereNotIn`, `WhereBetween`, `WhereNull`, `WhereNotNull`, `WhereExists`, `WhereNotExists`, `WhereOr`, `WhereInSubquery`, `WhereNotInSubquery` — available on Select, Update, and Delete builders
- **sqlbuilder** — Automatic placeholder rebasing: each `Where` call uses `$1`-relative numbering, globally rebased at `Build()` time
- **sqlbuilder** — `GroupBy`, `Having`, `OrderBy`, `OrderByAsc`, `OrderByDesc`, `OrderByExpr`, `Limit`, `Offset`
- **sqlbuilder** — Row-level locking: `ForUpdate`, `ForShare`, `SkipLocked`, `NoWait`
- **sqlbuilder** — Set operations: `Union`, `UnionAll`, `Intersect`, `Except`
- **sqlbuilder** — CTEs: `With`, `WithRecursive`, `WithSelect`, `WithRecursiveSelect` on all builders
- **sqlbuilder** — INSERT: `Columns`, `Values`, `ValueMap`, `BatchValues`, `FromSelect`
- **sqlbuilder** — Upsert: `OnConflictDoNothing`, `OnConflictUpdate`, `OnConflictUpdateExpr`
- **sqlbuilder** — UPDATE: `Set`, `SetExpr`, `SetMap`, `Increment`, `Decrement`, multi-table `From`
- **sqlbuilder** — DELETE: `Using` for multi-table deletes
- **sqlbuilder** — `Returning` clause on Insert, Update, and Delete builders
- **sqlbuilder** — Expression helpers: `Raw`, `RawExpr`, `Count`, `CountDistinct`, `Sum`, `Avg`, `Min`, `Max`, `Expr.As(alias)`
- **sqlbuilder** — Request integration: `ApplyPagination`, `ApplySort`, `ApplyFilters` bridge `request` package types to query conditions
- **sqlbuilder** — `When(cond, fn)` conditional builder and `Clone()` deep copy on all builders
- **sqlbuilder** — All builders expose `Build()`, `MustBuild()`, `Query()`, and `String()` terminal methods

## [0.6.0] - 2026-02-24

### Added

- **health** — New `health` package for health check endpoints with dependency checking, timeouts, and standard response formats
- **health** — `NewChecker(opts...)` with functional options pattern and configurable per-check timeout (default 5s)
- **health** — `AddCheck(name, fn)` registers critical checks — failure sets status to `"unhealthy"` (503)
- **health** — `AddNonCriticalCheck(name, fn)` registers non-critical checks — failure sets status to `"degraded"` (200)
- **health** — `Check(ctx)` runs all checks concurrently with per-check timeout and returns aggregated `Response`
- **health** — `Handler()` returns an error-returning HTTP handler for readiness probes (compatible with `router` package)
- **health** — `LiveHandler()` returns a liveness probe handler that always responds 200

## [0.5.0] - 2026-02-24

### Added

- **request** — `BindForm[T]` and `BindFormWithConfig[T]` for `application/x-www-form-urlencoded` body binding with struct tags (`form:"name"`)
- **request** — `BindMultipart[T]` and `BindMultipartWithConfig[T]` for `multipart/form-data` body binding
- **request** — `FormFile(r, field)` returns a single uploaded file header; `FormFiles(r)` returns all uploaded files
- **request** — `Bind[T]` now auto-detects content type: JSON, form, and multipart are dispatched automatically; unknown types return 415
- **request** — `BindJSON[T]` and `BindJSONWithConfig[T]` for explicit JSON binding with content-type enforcement
- **request** — `MaxMultipartMemory` field on `Config` (defaults to 32 MB) controls multipart memory threshold
- **request** — `form` struct tag supported in field name resolution for validation errors (checked before `json` tag)
- **server** — `WithTLS(certFile, keyFile)` option for HTTPS support
- **server** — `Addr()` method returns the listener address after the server starts
- **response** — `XML(w, statusCode, data)` writes XML responses with `<?xml?>` header
- **response** — `IndentedJSON(w, statusCode, data)` writes pretty-printed JSON
- **response** — `PureJSON(w, statusCode, data)` writes JSON without HTML escaping of `<`, `>`, `&`
- **response** — `JSONP(w, r, statusCode, data)` writes JSONP responses (reads `?callback=` query param)
- **response** — `Reader(w, statusCode, contentType, contentLength, reader)` streams from `io.Reader`

## [0.4.1] - 2026-02-23

### Fixed

- **request** — Race condition on `globalConfig` and `globalPaginationConfig` — added `sync.RWMutex` guards
- **middleware** — Logger `responseWriter` forwarded duplicate `WriteHeader` calls to underlying writer
- **middleware** — `Retry-After` header was hardcoded to `"60"` regardless of configured rate limit window
- **middleware** — `itoa` helper caused infinite recursion on `math.MinInt` — replaced with `strconv.Itoa`
- **httpclient** — Integer overflow in retry delay calculation when attempt count exceeded 62
- **httpclient** — Circuit breaker allowed multiple concurrent probes in half-open state (TOCTOU)
- **response** — `File()` Content-Disposition header injection on non-ASCII filenames — added RFC 5987 encoding
- **server** — `Shutdown()` returned immediately instead of waiting for graceful shutdown to complete
- **request** — `IsValidEmail` accepted display-name formats like `"Alice" <a@b.com>` — now requires bare address
- **apitest** — `RequestBuilder.Build()` used `Header.Set` in loop, dropping multi-value headers — changed to `Header.Add`

## [0.4.0] - 2026-02-23

### Added

- **router** — New `router` package with route grouping and method helpers on top of `http.ServeMux`
- **router** — Method helpers: `Get`, `Post`, `Put`, `Patch`, `Delete` — accept `func(w, r) error` handlers
- **router** — Stdlib variants: `GetFunc`, `PostFunc`, `PutFunc`, `PatchFunc`, `DeleteFunc` — accept `http.HandlerFunc`
- **router** — `Group(prefix, ...middleware)` for prefix-based route grouping with per-group middleware
- **router** — Nested groups accumulate prefix and middleware (root → parent → child ordering)
- **router** — `Handle`/`HandleFunc` stdlib escape hatches for `http.Handler`/`http.HandlerFunc`
- **router** — `Use()` adds middleware to subsequently registered routes (matches chi/echo/gin behavior)
- **router** — `DefaultErrorHandler` writes JSON error envelope matching `response.Err` format, extracts `*errors.Error` via `errors.As`
- **router** — `WithErrorHandler` option for custom error handling
- **router** — Implements `http.Handler` — drop-in replacement for `http.ServeMux` with `server.New(r)`

### Removed

- **response** — Removed `Wrap` function (was a no-op pass-through; use `router.GetFunc`/`PostFunc` etc. instead)

## [0.3.0] - 2026-02-23

### Added

- **request** — Struct tag validation (`validate:"required,email,min=3"`) — zero dependencies, runs automatically in `Bind[T]` and `DecodeJSON`
- **request** — Supported tags: `required`, `email`, `url`, `min`, `max`, `len`, `oneof`, `alpha`, `alphanum`, `numeric`, `uuid`, `contains`, `startswith`, `endswith`
- **request** — `ValidateStruct(v any) error` for standalone struct tag validation
- **request** — Convenience methods on `Validation` builder: `RequireEmail`, `RequireURL`, `UUID`, `MatchesPattern`
- **request** — Shared helpers: `IsValidEmail`, `IsValidURL`, `IsValidUUID`, `MatchesRegexp`

## [0.2.0] - 2026-02-22

### Added

- **request** — Multi-param detection for page size (`per_page`, `page_size`, `limit` tried in order)
- **request** — Navigation helpers on `Pagination`: `HasNext`, `HasPrevious`, `IsFirstPage`, `IsLastPage`, `TotalPages`, `NextPage`, `PreviousPage`
- **request** — SQL clause helpers: `SQLClause()` (PostgreSQL) and `SQLClauseMySQL()` (MySQL)
- **response** — `SetLinkHeader` generates RFC 5988 Link headers with `next`, `prev`, `first`, `last` rels
- **response** — `HasNext` and `HasPrevious` boolean fields on `PageMeta`, auto-computed by `NewPageMeta`

### Changed

- **request** — `Pagination.Offset` is now `int64` for overflow safety on large page numbers
- **request** — `PaginationConfig.PerPageParams []string` replaces singular `PerPageParam`

## [0.1.0] - 2026-02-21

### Added

- **errors** — Structured API errors with `errors.Is`/`errors.As` support, error codes, sentinel errors, and custom code registration
- **request** — Generic body binding (`Bind[T]`), query/path/header parsing, pagination, sorting, filtering, and validation
- **response** — Consistent JSON envelope, fluent builder, pagination helpers, SSE streaming, and handler wrapper (`Handle`)
- **middleware** — Request ID, structured logging, panic recovery, CORS, rate limiting, auth, security headers, timeout, and body limit
- **httpclient** — HTTP client with retries, exponential backoff, circuit breaker, `HTTPClient` interface, `MockClient` for testing, and request builder
- **server** — Graceful shutdown wrapper with signal handling (`SIGINT`/`SIGTERM`) and lifecycle hooks (`OnStart`/`OnShutdown`)
- **apitest** — Fluent test request builder, response recorder, and assertion helpers (`AssertStatus`, `AssertSuccess`, `AssertError`, `AssertValidationError`)
