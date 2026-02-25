# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0] - 2026-XX-XX

### Added

- **sqlbuilder** — New `sqlbuilder` package: fluent, type-safe SQL query builder for PostgreSQL with `$1, $2, ...` numbered placeholders
- **sqlbuilder** — `Select`, `SelectExpr` builders with `Distinct`, `Column`, `Columns`, `ColumnExpr`, `From`, `FromAlias`, `FromSubquery`
- **sqlbuilder** — JOIN support: `Join`, `LeftJoin`, `RightJoin`, `FullJoin`, `CrossJoin` with parameterized ON clauses
- **sqlbuilder** — WHERE conditions: `Where`, `WhereIn`, `WhereNotIn`, `WhereBetween`, `WhereNull`, `WhereNotNull`, `WhereExists`, `WhereOr`
- **sqlbuilder** — `GroupBy`, `Having`, `OrderBy`, `OrderByExpr`, `Limit`, `Offset`
- **sqlbuilder** — Row-level locking: `ForUpdate`, `ForShare`, `SkipLocked`, `NoWait`
- **sqlbuilder** — Set operations: `Union`, `UnionAll`, `Intersect`, `Except`
- **sqlbuilder** — CTEs: `With(name, query)` and `WithRecursive(name, query)` on all builders
- **sqlbuilder** — `Insert` builder with `Columns`, `Values`, `ValueMap`, `BatchValues`, `FromSelect`
- **sqlbuilder** — Upsert support: `OnConflictDoNothing`, `OnConflictUpdate`, `OnConflictUpdateExpr`
- **sqlbuilder** — `Update` builder with `Set`, `SetExpr`, `SetMap`, multi-table `From`
- **sqlbuilder** — `Delete` builder with `Using` for PostgreSQL multi-table deletes
- **sqlbuilder** — `Returning` clause on INSERT, UPDATE, and DELETE builders
- **sqlbuilder** — Automatic placeholder rebasing: each `Where` call uses `$1`-relative numbering, globally rebased at `Build()` time
- **sqlbuilder** — `Raw(sql)` and `RawExpr(sql, args...)` for raw SQL expressions in SELECT columns, SET clauses, and ORDER BY
- **sqlbuilder** — `ApplyPagination(request.Pagination)` sets LIMIT and OFFSET from parsed pagination
- **sqlbuilder** — `ApplySort([]request.SortField, allowedColumns)` converts API sort fields to ORDER BY with column allowlist
- **sqlbuilder** — `ApplyFilters([]request.Filter, allowedColumns)` converts API filters to WHERE conditions (eq, neq, gt, gte, lt, lte, in)
- **sqlbuilder** — All builders expose `Build() (string, []any)`, `MustBuild()`, `Query()`, and `String()` terminal methods
- **sqlbuilder** — Convenience `Where` helpers on Select, Update, and Delete builders: `WhereEq`, `WhereNeq`, `WhereGt`, `WhereGte`, `WhereLt`, `WhereLte`, `WhereLike`, `WhereILike`
- **sqlbuilder** — `OrderByAsc(cols...)` and `OrderByDesc(cols...)` helpers on SelectBuilder
- **sqlbuilder** — `WhereNotIn` and `WhereBetween` helpers on Update and Delete builders (were previously Select-only)
- **sqlbuilder** — Aggregate expression helpers: `Count`, `CountDistinct`, `Sum`, `Avg`, `Min`, `Max` return `Expr` for use in `SelectExpr`
- **sqlbuilder** — `Expr.As(alias)` method for aliasing expressions (e.g., `Count("*").As("total")`)
- **sqlbuilder** — `WhereNotExists` on Select, Update, and Delete builders
- **sqlbuilder** — `WhereExists` and `WhereOr` on Update and Delete builders (were previously Select-only)
- **sqlbuilder** — `WhereInSubquery` and `WhereNotInSubquery` on Select, Update, and Delete builders for subquery-based IN conditions
- **sqlbuilder** — `When(cond, fn)` conditional builder on all four builders (Select, Insert, Update, Delete)
- **sqlbuilder** — `Clone()` deep copy on all four builders for safe query composition from a shared base
- **sqlbuilder** — `Increment(col, n)` and `Decrement(col, n)` on UpdateBuilder for `SET col = col + $1` / `SET col = col - $1`

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
