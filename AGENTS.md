# AGENTS.md

Guidance for AI coding agents working in the **apikit** repository.

## What this is

`apikit` is a production-ready Go toolkit for building REST APIs, published as
`github.com/KARTIKrocks/apikit`. It targets **Go 1.22+** and has **zero mandatory
runtime dependencies** — it works with any `net/http`-compatible router. Keeping
the dependency footprint at zero is a core design goal: do not add third-party
dependencies to `go.mod` without explicit approval.

## Repository layout

The module is a collection of independent, importable packages. Each top-level
directory is one package:

- `errors/` — Structured API errors with `errors.Is`/`errors.As` support, error codes, sentinels
- `request/` — Generic body binding (`Bind[T]`), query/path/header parsing, pagination, sorting, filtering, validation
- `response/` — JSON envelope, fluent builder, pagination helpers, SSE streaming, XML, JSONP
- `middleware/` — Request ID, logging, panic recovery, CORS, rate limiting, auth, security headers, timeout, body limit
- `httpclient/` — HTTP client with retries, backoff, circuit breaker, and mockable `HTTPClient` interface
- `router/` — Route grouping, named routes, URL generation, param constraints, sub-router mounting, static files (on top of `http.ServeMux`)
- `server/` — Graceful shutdown wrapper with signal handling, lifecycle hooks, TLS
- `health/` — Health check endpoint builder with dependency checks and liveness/readiness probes
- `config/` — Load config from env vars, `.env`, and JSON into typed structs with validation
- `sqlbuilder/` — Fluent SQL query builder for PostgreSQL, MySQL, SQLite (JOINs, CTEs, UNION, upsert)
- `dbx/` — Generic row scanner for `database/sql` that maps rows to structs via `db` tags
- `apitest/` — Fluent test helpers for recording and asserting HTTP handler responses
- `examples/`, `playground/` — Example and scratch code (excluded from strict linting)

Packages should stay decoupled — a user importing only `errors` should not pull
in `sqlbuilder`. Avoid introducing cross-package imports that break this.

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
  Record notable changes in `CHANGELOG.md` (Keep a Changelog style).
- **Docs:** Keep package `doc.go`, `README.md` examples, and code in sync when
  behavior changes.

## Git / PR conventions

- Commit messages follow Conventional Commits (e.g. `fix(tests): ...`,
  `ci: ...`, `feat(router): ...`).
- The `benchmark/` directory is intentionally untracked — never stage or commit
  it. Prefer `git add -u` over `git add -A` so untracked scratch dirs stay out.
- Commit or push only when explicitly asked; branch off `main` first if needed.
- A PR template lives at `.github/pull_request_template.md` — fill it out.
