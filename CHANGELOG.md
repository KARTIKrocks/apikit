# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.25.0] - 2026-06-17

### Added

- **request** — `omitempty` modifier for the `validate` tag: when a field is absent or holds its zero value, the remaining rules for that field are skipped, so optional fields can carry rules such as `omitempty,oneof=...` or `omitempty,url` without rejecting the empty/unset case. It is order-sensitive and must appear before the rules it guards (rules apply left-to-right, so a trailing `omitempty` cannot undo an earlier rule's failure)

### Fixed

- **request** — value rules (`min`, `max`, `oneof`, `url`, `e164`, `len`, `contains`, …) now apply to pointer and interface fields by dereferencing the value before the rule runs. Previously they were a silent no-op on a `*T` field (e.g. `*string` with `max=3`), so invalid data behind a pointer passed unchecked. Pointer fields that were effectively unvalidated may now surface validation errors

### Notes

- `required` semantics are unchanged, and are now documented and tested. On a pointer or interface it is a presence (nil) check: a non-nil pointer to a zero value (e.g. `*string` holding `""`) still satisfies `required`, matching go-playground/validator and preserving PATCH semantics (`nil` = absent, `&""` = explicitly provided). On a plain value, `required` still means non-zero

## [0.24.0] - 2026-06-11

### Fixed

- **router / middleware** — `http.Hijacker` is now implemented directly on the response-writer wrappers (`router.probeWriter`, and the `Logger`/`Timeout` middleware writers), so **WebSocket upgrades work behind the router and middleware**. Previously these wrappers exposed the underlying connection only through `Unwrap()`; that satisfies `http.ResponseController` but not libraries like `gorilla/websocket`, which assert `w.(http.Hijacker)` directly — so upgrades failed with `500 "response does not implement http.Hijacker"`. `Hijack()` now delegates down the wrapper chain to the real writer.

### Notes

- The `Timeout` middleware is unsuitable for long-lived hijacked connections (WebSockets): its timer keeps running and will try to write a 503 to the taken-over connection. Mount WebSocket/streaming routes without `Timeout`. After a timeout has fired, `Hijack()` returns `http.ErrHandlerTimeout`.

## [0.23.0] - 2026-06-10

### Added

- **errors** — package-level `Is`, `As`, `Unwrap`, and `Join` that re-export the standard library's `errors` functions. Callers that import this package as `errors` (shadowing the stdlib) can now use them directly, e.g. `errors.Is(err, errors.ErrNotFound)`, as shown in the package docs and README. Because `*Error` implements `Is` (matching on `Code`) and `Unwrap`, sentinel checks and wrapped-cause traversal both work as expected

## [0.22.0] - 2026-06-10

### Added

- **request** — `e164` validation tag for E.164 phone numbers (e.g. `+14155552671`); exported `IsValidE164` helper
- **request** — comparison tags `gte`, `lte`, `gt`, `lt`, `eq`, `ne`. `gte`/`lte` mirror `min`/`max` (value for numbers, length for string/slice/map); `gt`/`lt` are their strict variants; `eq`/`ne` compare string content, numeric value, or item count
- **request** — struct elements of slices, arrays, and maps are now validated recursively (reported as e.g. `items[0].name`, `by_key[a].name`), with no separate `dive` tag required; `nil` pointer elements are skipped and `validate:"-"` skips a field and its elements

### Changed

- **request** — an unrecognized `validate` tag now panics instead of silently passing the value. Validate tags are static, developer-authored config, so an unknown rule (e.g. a typo like `e_164`) is a programmer error that previously let invalid input through unvalidated; it now surfaces immediately. Cross-field rules (`eqfield`, `required_with`, …) remain out of scope for the tag engine — use `request.NewValidation()`

## [0.21.0] - 2026-06-05

### Fixed

- **httpclient** — `RequestBuilder.Header`/`Param`/`BearerToken` no longer panic when used before `Headers`/`Params`; the maps are now lazily initialized. `Send` also guards against a nil builder/client
- **httpclient** — query parameters are merged into any query string already present on the path (via `url.Parse`) instead of being naively appended after a second `?`
- **apitest** — `RequestBuilder.Build` merges query params into the target URL preserving an existing query, and surfaces response-body read errors instead of silently ignoring them
- **dbx** — `QueryAll`/`QueryOne` now report `rows.Close()` errors (previously discarded) by returning them when the query itself otherwise succeeded
- **server** — `Start` guards the `doneCh` close with `sync.Once` so `Done()` always unblocks, including on early-return error paths
- **sqlbuilder** — `CaseBuilder.Else` with a literal now clears any previously bound `ELSE` args, preventing stale arguments from leaking into the built query
- **config** — `Load` skips nil `Option` values instead of panicking

### Changed

- **router** — `URL` now percent-escapes non-catch-all path parameters (`url.PathEscape`); catch-all (`...`) segments are still written verbatim
- **router** — `ValidateParams` panics at setup time if a constraint has a nil `Validate` func, surfacing the misconfiguration immediately rather than on first request
- **httpclient** / **server** — `WithLogger(nil)` now falls back to `slog.Default()` instead of installing a nil logger
- **httpclient** — `NewCircuitBreaker` clamps a non-positive threshold or timeout to safe defaults (`1` / `1s`); `WithMaxRetries` clamps negative values to `0`

## [0.20.0] - 2026-06-05

### Fixed

- **router** — CORS preflight requests are no longer dropped. A browser preflight is an `OPTIONS` request, but handlers register concrete methods (`GET`, `POST`, …), so `ServeMux` answered the preflight with `405` before any handler ran — meaning `middleware.CORS` added via `Router.Use` never saw it and no `Access-Control-*` headers were sent, so the browser blocked the real request. The router now runs the 404/405 fallback through the root group's middleware, so `r.Use(middleware.CORS(...))` handles preflights and the actual cross-origin request works with no extra wiring

### Changed

- **router** — Root middleware registered via `Router.Use` now runs on the 404/405 fallback path (previously it ran only for matched routes). This makes cross-cutting middleware such as CORS, `RequestID`, and `Logger` apply to unmatched requests, so error responses carry the same headers as matched ones. Custom `WithNotFound` / `WithMethodNotAllowed` handlers still take precedence over the `ErrorHandler` and now execute inside the middleware chain

## [0.19.0] - 2026-06-05

### Added

- **router** — `NewErrorHandler(logger *slog.Logger)` constructor returning an `ErrorHandler` that logs to the supplied logger (falls back to `slog.Default()` when nil); inject via `WithErrorHandler(NewErrorHandler(myLogger))`

### Changed

- **router** — `DefaultErrorHandler` now logs server errors via `slog.Default()` so the wrapped cause of an `Internal`/`Internalf` error is no longer silently dropped: status >= 500 logs at Error level, a 4xx carrying a wrapped cause logs at Warn, and plain 4xx are not logged. Only the safe message is still sent to the client

## [0.18.0] - 2026-06-05

### Added

- **httpclient** — `WithErrorOnStatus(enabled bool)` option to opt out of treating a non-2xx status as an error; when disabled, non-2xx responses return `(resp, nil)` so callers branch on `resp.IsClientError()`/`IsServerError()` and decode the body via `resp.JSON()`. Retries on 5xx still occur; transport, context, and circuit-breaker failures still return errors
- **httpclient** — `RequestBuilder.ErrorOnStatus(enabled bool)` to override the client's error-on-status policy for a single request

### Fixed

- **httpclient** — `doRequest` now returns the last `*Response` alongside the error when retries are exhausted (previously returned `nil`), so the status code and error body are inspectable without `errors.As` and `Response`'s status helpers are reachable on every error path

## [0.17.0] - 2026-06-05

### Added

- **response** — `Envelope.MarshalJSON` always includes the `"data"` field on success responses (rendered as `null` when no data is set), while still omitting it on error responses, so clients can rely on the field being present on success

### Changed

- **response** — `JSON(w, statusCode, data)` derives the `success` flag from the status code (2xx is treated as success) instead of hardcoding `true`, keeping the envelope consistent when a non-2xx status is passed and matching `Builder.Status`

## [0.16.0] - 2026-04-06

### Added

- **router** — Named routes via `RouteEntry.Name(name)` and `Router.URL(name, params...)` for reverse URL generation with `{param}` and `{param...}` placeholder substitution
- **router** — Route introspection: `Router.Routes()` returns a snapshot of all registered `RouteInfo` entries; `Router.Walk(fn)` iterates with early-exit support
- **router** — `RouteInfo.HandlerName` captures the original handler's function name via `runtime.FuncForPC` for debugging and documentation
- **router** — Parameter constraints: `ValidateParams(handler, constraints...)` wraps a handler with path-parameter validation; built-in constraint constructors `Int`, `UUID`, `Regex`, `OneOf`
- **router** — `With(middleware...)` returns a child group for per-route inline middleware without affecting sibling routes
- **router** — `Route(prefix, fn, middleware...)` for inline sub-routing — creates a child group, calls `fn` to register routes, and returns the group for further use
- **router** — `Mount(prefix, handler)` attaches an `http.Handler` (or `*Router`) at a prefix with `http.StripPrefix`; sub-router routes and named routes are merged into the parent's route table
- **router** — `Static(prefix, dir)` serves files from a filesystem directory; `File(pattern, filePath)` serves a single file for GET requests
- **router** — `WithNotFound(handler)` and `WithMethodNotAllowed(handler)` options for custom 404/405 handlers, taking precedence over the `ErrorHandler`
- **router** — `WithStripSlash()` silently removes trailing slashes before routing; `WithRedirectSlash()` sends 301 redirects (mutually exclusive, panics if both set)

### Changed

- **router** — All method helpers (`Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options` and their `Func` variants) and `Handle`/`HandleFunc` now return `*RouteEntry` for optional `.Name()` chaining
- **router** — `register()` accepts an additional `origFn` parameter to capture the original handler name before middleware wrapping

## [0.15.0] - 2026-04-04

### Changed

- **middleware** — Renamed `TokenBucket` → `FixedWindow` and `NewTokenBucket` → `NewFixedWindow` to accurately reflect the fixed-window counter algorithm; default limiter in `RateLimit()` updated accordingly

### Added

- **middleware** — `timeoutWriter.Unwrap()` returns the underlying `http.ResponseWriter`, enabling `http.Flusher`/`http.Hijacker` detection via `http.ResponseController`
- **middleware** — `timeoutWriter.Flush()` implements `http.Flusher` with timeout-aware guarding, allowing SSE and chunked responses through the timeout middleware
- **middleware** — Doc comment on `Timeout` clarifying goroutine leak behavior when a handler does not respect context cancellation

### Removed

- **middleware** — `TokenBucket` and `NewTokenBucket` (replaced by `FixedWindow` and `NewFixedWindow`)

## [0.14.1] - 2026-04-04

### Changed

- **request** — `MatchesRegexp` now caches compiled `*regexp.Regexp` patterns in a `sync.Map`, eliminating repeated recompilation
- **request** — `Validation.MatchesPattern` delegates to the cached `MatchesRegexp` instead of calling `regexp.MatchString` directly
- **response** — `write()` marshals JSON to a buffer first and sets the `Content-Length` header before writing, improving client framing behavior
- **response** — JSONP output is now prefixed with `/**/` to mitigate Rosetta Flash and content-type-sniffing attacks

### Fixed

- **response** — JSONP `callback` parameter is validated against `^[a-zA-Z_$][a-zA-Z0-9_$.]*$`; invalid names return 400 to prevent XSS injection
- **response** — JSON encoding errors in `write()` are now caught before headers are sent, returning a clean 500 instead of a truncated body

## [0.14.0] - 2026-04-03

### Added

- **httpclient** — `WithMaxResponseBody(n int64)` option to configure the maximum allowed response body size (default 10 MB via `DefaultMaxResponseBody`)
- **httpclient** — Response reads in `executeRequest` are now capped with `io.LimitReader`; responses exceeding the limit return an error instead of allocating unbounded memory

## [0.13.1] - 2026-04-02

### Fixed

- **health** — Removed unnecessary `indexedResult` wrapper struct in `Check`; goroutines now write `CheckResult` directly into the pre-indexed `[]CheckResult` slice
- **server** — Added `defer signal.Stop(sigCh)` in `Start()` to deregister the signal channel on return, preventing a channel/goroutine leak

## [0.13.0] - 2026-04-01

### Changed

- **errors** — `WithFields` and `WithDetails` now copy the caller-supplied map before storing it, preventing external mutation of the error's internal state after the call

### Removed

- **errors** — `CodeTooManyRequests` constant and its `codeStatusMap` entry removed; rate-limit errors are owned by the `middleware` package

### Fixed

- **errors** — Added clarifying doc comment on `(*Error).Is` explaining that matching is code-based and symmetric between two `*Error` values with the same `Code`

## [0.12.3] - 2026-03-31

### Fixed

- **dbx** — `defaultDB` is now stored in a `sync/atomic.Value`; `SetDefault` is safe for concurrent use and the connection read in `conn()` is race-free

## [0.12.2] - 2026-03-31

### Fixed

- **config** — File I/O errors (missing or unreadable `.env`/JSON files) and JSON parse errors are now classified as `errors.Internal` instead of `errors.BadRequest`; error message updated from `"file not found"` to `"cannot open file"` to better reflect the server-side cause

## [0.12.1] - 2026-03-30

### Fixed

- **router** — `probeWriter` no longer intercepts intentional 404/405 responses from user handlers; only unmatched routes (ServeMux's own 404/405) are routed through the `ErrorHandler`

## [0.12.0] - 2026-03-30

### Added

- **router** — Method helpers: `Head`, `HeadFunc`, `Options`, `OptionsFunc` on both `Router` and `Group` for HEAD and OPTIONS HTTP methods
- **router** — Consistent JSON error responses for unmatched routes: 404 and 405 from `http.ServeMux` now go through the router's `ErrorHandler` instead of returning plain text
- **router** — `joinPath` helper normalizes prefix+pattern concatenation, preventing double-slash bugs (e.g. `Group("/api/")` + `Get("/users")` now correctly registers `/api/users`)

### Fixed

- **router** — `splitPattern` now trims extra whitespace between method and path (e.g. `"GET  /users"` no longer silently creates a broken route)
- **router** — Group prefix concatenation with nested groups no longer produces double slashes when both prefix and child start/end with `/`

## [0.11.1] - 2026-03-29

### Changed

- **ci** — Test matrix now covers Go 1.22–1.26; updated `actions/checkout` and `actions/setup-go` to v6, `golangci-lint-action` to v9 (lint v2.11)
- **ci** — Added dedicated coverage job with Codecov upload and benchmark job on push to main
- **ci** — Added CodeQL security analysis workflow
- **ci** — Added Dependabot configuration for automated dependency updates
- **ci** — Added bug report / feature request issue templates and pull request template
- **build** — Added `fmt` (gofmt + goimports) and `ci` Makefile targets; enabled `-race` in test targets
- **build** — Added `gocyclo` linter (max complexity 15) and simplified `.golangci.yml` exclusion rules
- **config** — Refactored `resolveStruct` by extracting `nestedEnvPrefix` and `nestedDisplayName` helpers to reduce cyclomatic complexity
- **sqlbuilder** — Refactored `rebasePlaceholders` by extracting `tryRebaseSingle` fast-path helper to reduce cyclomatic complexity
- **sqlbuilder** — Decomposed `SelectBuilder.Build` into 8 focused write methods to reduce cyclomatic complexity
- **request** — Simplified `TestDecodeFormValues_AllTypes` using `reflect.DeepEqual`
- Fixed struct field alignment across `health`, `request`, `sqlbuilder` packages
- Fixed import ordering in `request/tag_validator.go` and `request/validate.go`

## [0.11.0] - 2026-03-06

### Added

- **config** — Embedded (anonymous) struct support: fields resolve as if declared directly on the parent, with no extra prefix added
- **config** — `envprefix` struct tag for named struct fields to override the auto-generated nesting prefix (e.g., `envprefix:"DB_"` reads `DB_URL` instead of `DATABASE_URL`)
- **config** — `envprefix:"-"` to skip the nesting prefix entirely, so inner `env` tags are used as-is

## [0.10.0] - 2026-02-28

### Added

- **sqlbuilder** — `Coalesce(exprs ...string)` and `CoalesceExpr(exprs ...Expr)` for `COALESCE(a, b, c)` expressions with placeholder rebasing
- **sqlbuilder** — `NullIf(expr1, expr2 string)` and `NullIfExpr(expr1, expr2 Expr)` for `NULLIF(a, b)` expressions
- **sqlbuilder** — `Cast(expr, typeName string)` and `CastExpr(expr Expr, typeName string)` for `CAST(x AS type)` expressions
- **sqlbuilder** — `CaseBuilder` with `Case`, `When`, `WhenExpr`, `Else`, `ElseExpr`, `End` for building `CASE WHEN ... THEN ... ELSE ... END` expressions
- **sqlbuilder** — `WindowBuilder` with `Window`, `PartitionBy`, `OrderBy`, `Build` for window specifications
- **sqlbuilder** — Window function constructors: `RowNumber`, `Rank`, `DenseRank`, `Ntile`, `Lag`, `Lead`
- **sqlbuilder** — `Expr.Over(w *WindowBuilder)` and `Expr.OverRaw(clause string)` to append `OVER (...)` to any expression
- **sqlbuilder** — `WhereColumn(col1, op, col2 string)` on `SelectBuilder`, `UpdateBuilder`, and `DeleteBuilder` for column-to-column comparisons with no args
- **sqlbuilder** — `DistinctOn(cols ...string)` on `SelectBuilder` for PostgreSQL `DISTINCT ON (cols)` support
- **sqlbuilder** — `GroupByExpr(expr Expr)` on `SelectBuilder` for expression-based GROUP BY with placeholder rebasing
- **sqlbuilder** — `HavingIn(col string, vals ...any)` and `HavingBetween(col string, low, high any)` on `SelectBuilder`
- **sqlbuilder** — `JoinSubquery`, `LeftJoinSubquery`, `RightJoinSubquery`, `FullJoinSubquery` on `SelectBuilder` for joining against subqueries with alias and placeholder rebasing
- **sqlbuilder** — `ReturningExpr(exprs ...Expr)` on `InsertBuilder`, `UpdateBuilder`, and `DeleteBuilder` for expression columns in RETURNING clauses with placeholder rebasing

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
