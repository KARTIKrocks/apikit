# AGENTS.md

Guidance for AI coding agents working in the **apikit** repository.

## What this is

`apikit` is a production-ready Go toolkit for building REST APIs, published as
`github.com/KARTIKrocks/apikit`. It targets **Go 1.22+** and has **zero mandatory
runtime dependencies** — it works with any `net/http`-compatible router. Keeping
the dependency footprint at zero is a core design goal for this module: do not
add third-party dependencies to its `go.mod` without explicit approval.

The one deliberate exception is `otel/` (see [Submodules](#submodules) below):
it needs the real OpenTelemetry SDK, so it ships as a **separate Go module**
with its own `go.mod` — that dependency never reaches this one.

## Repository layout

The module is a collection of independent, importable packages. Each top-level
directory is one package:

- `errors/` — Structured API errors with `errors.Is`/`errors.As` support, error codes, sentinels
- `request/` — Generic body binding (`Bind[T]`), query/path/header parsing, pagination, sorting, filtering, validation, file uploads with size and content-type limits
- `response/` — JSON envelope, fluent builder, pagination helpers, SSE streaming, XML, JSONP, file/media serving with HTTP Range support
- `middleware/` — Request ID, logging, panic recovery, CORS, rate limiting, auth, security headers, timeout, body limit
- `httpclient/` — HTTP client with retries, backoff, circuit breaker, and mockable `HTTPClient` interface
- `router/` — Route grouping, named routes, URL generation, param constraints, sub-router mounting, static files (on top of `http.ServeMux`)
- `server/` — Graceful shutdown wrapper with signal handling, lifecycle hooks, TLS
- `health/` — Health check endpoint builder with dependency checks and liveness/readiness probes
- `config/` — Load config from env vars, `.env`, and JSON into typed structs with validation
- `sqlbuilder/` — Fluent SQL query builder for PostgreSQL, MySQL, SQLite (JOINs, CTEs, UNION, upsert)
- `dbx/` — Generic row scanner for `database/sql` that maps rows to structs via `db` tags
- `apitest/` — Fluent test helpers for recording and asserting HTTP handler responses
- `openapi/` — OpenAPI 3.1 document generation from request/response struct tags, plus a Swagger UI handler
- `examples/`, `playground/` — Example and scratch code (excluded from strict linting)

Packages should stay decoupled — a user importing only `errors` should not pull
in `sqlbuilder`. Avoid introducing cross-package imports that break this.
`openapi/` is the one established exception: it imports `router/` (its `Sync`
method walks a registered `router.Router`) — that's intentional, not a
violation to flag.

## Submodules

Some needs can't be met without a third-party dependency — most notably real
OpenTelemetry interop. Rather than pull that into the core module, those ship
as **separate Go modules** under this repo, each with its own `go.mod`,
`go.sum`, `README.md`, `CHANGELOG.md`, and version history (tagged as
`<module-dir>/vX.Y.Z`, e.g. `otel/v0.1.0` — not a plain `vX.Y.Z` tag).

- `otel/` (`github.com/KARTIKrocks/apikit/otel`) — real OpenTelemetry tracing
  and metrics middleware, an `httpclient` transport for outbound propagation,
  and a `health` check span wrapper. Requires Go 1.25+ (higher than the core
  module's Go 1.22+, pinned by the OpenTelemetry SDK's own requirement).

Each submodule has a `replace github.com/KARTIKrocks/apikit => ../` directive
in its `go.mod` for local development — that's intentional and permanent, not
a leftover to clean up before release: `replace` directives are only honored
when that module is the *main* module being built, so they're ignored (and
harmless) once the submodule is fetched as a real dependency. Its `require`
line still needs to point at a real, already-published core version, though
— that's what a build actually resolves to for anyone who isn't building the
submodule as the main module (verify with the `replace` line commented out
before tagging a release).

For cross-module local development (jump-to-definition across the core module
and a submodule, building/testing both at once), create a local workspace —
`go work init . ./otel` — rather than relying on each module's own `replace`.
It's gitignored (`go.work`/`go.work.sum`) and never committed: CI builds and
tests each module independently via its own `go.mod` (with `GOWORK=off`),
matching what a real consumer sees.

## Build, test, and lint

Use the `Makefile` targets:

```bash
make test    # go test -race -count=1 ./...   (always run before finishing)
make vet     # go vet ./...
make lint    # golangci-lint run ./...   (installs golangci-lint via `make setup`)
make build   # go build ./...
make fmt     # gofmt -w . && goimports -w .
make ci      # vet + lint + test (mirrors CI)
make cover   # coverage report -> coverage.html
```

Run `make fmt` after editing and `make test` before considering work done. Tests
run with the **race detector** — new concurrent code must be race-clean. CI runs
`vet`, `lint`, and `test` (see `.github/workflows/ci.yml`), plus CodeQL.

These targets operate on the core module only. A submodule (`otel/`) has its
own `go.mod` and isn't reached by `./...` from the repo root — run the same
`go build`/`go vet`/`go test -race`/`golangci-lint run` commands from inside
it directly (CI does this with `GOWORK=off`; see [Submodules](#submodules)).

## Conventions

- **Formatting:** `gofmt` + `goimports`. Never hand-format; run `make fmt`.
- **Linters** (`.golangci.yml`): `errcheck`, `govet`, `staticcheck`, `unused`,
  `ineffassign`, `misspell`, `gocritic`, `gocyclo` (max complexity **15**),
  `revive`. Keep functions under the cyclomatic-complexity limit; refactor
  rather than suppress. `errcheck` is relaxed only in `_test.go` and `examples/`.
- **Tests:** Table-driven where it fits; co-located `_test.go` files. Many tests
  use `t.Parallel()` — preserve parallel-safety when editing them. Use the
  `apitest` package for handler-level assertions. Add tests for new behavior;
  this library is coverage-tracked (codecov).
- **Errors:** Return structured errors from the `errors` package (e.g.
  `errors.Conflict(...)`, `errors.NotFound(...)`) from handlers rather than
  plain `fmt.Errorf`, so the framework can render consistent responses.
- **Public API:** These packages are consumed by external users. Treat exported
  identifiers as a public contract — avoid breaking changes; document additions.
  Record notable changes in `CHANGELOG.md` (Keep a Changelog style) — a
  submodule keeps its own (e.g. `otel/CHANGELOG.md`), since it's versioned
  independently.
- **Docs:** Keep package `doc.go`, `README.md` examples, and code in sync when
  behavior changes. When a package's public API changes, also update the
  documentation website (see below) so it does not drift from the code.

## Documentation website

A separate **`website` branch** holds the public documentation site (deployed to
GitHub Pages via the `gh-pages` branch). It is **not** merged into `main` — check
it out with `git checkout website` to work on it.

- Location: `apikit-website/` — a React + TypeScript + Vite app styled with
  Tailwind CSS v4, with syntax highlighting via Shiki.
- Structure: a single-page site assembled in `src/App.tsx`. Each Go package has
  one page component in `src/content/<package>.tsx` (e.g. `errors.tsx`,
  `router.tsx`), rendered through the shared `ModuleSection` and `CodeBlock`
  components. Navigation lives in `src/components/Sidebar.tsx`.
- When you add or change a package's public API, add/update the matching
  `src/content/<package>.tsx` page and its entry in `Sidebar.tsx`.
- Commands (run inside `apikit-website/`): `npm run dev` (local server),
  `npm run build` (`tsc -b && vite build`), `npm run lint` (eslint), and
  `npm run deploy` (build + publish to the `gh-pages` branch).

## Git / PR conventions

- Commit messages follow Conventional Commits (e.g. `fix(tests): ...`,
  `ci: ...`, `feat(router): ...`).
- The `benchmark/` directory is intentionally untracked — never stage or commit
  it. Prefer `git add -u` over `git add -A` so untracked scratch dirs stay out.
- Commit or push only when explicitly asked; branch off `main` first if needed.
- A PR template lives at `.github/pull_request_template.md` — fill it out.
