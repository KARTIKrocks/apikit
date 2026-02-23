# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
