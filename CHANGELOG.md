# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-21

### Added

- **errors** — Structured API errors with `errors.Is`/`errors.As` support, error codes, sentinel errors, and custom code registration
- **request** — Generic body binding (`Bind[T]`), query/path/header parsing, pagination, sorting, filtering, and validation
- **response** — Consistent JSON envelope, fluent builder, pagination helpers, SSE streaming, and handler wrapper (`Handle`)
- **middleware** — Request ID, structured logging, panic recovery, CORS, rate limiting, auth, security headers, timeout, and body limit
- **httpclient** — HTTP client with retries, exponential backoff, circuit breaker, `HTTPClient` interface, `MockClient` for testing, and request builder
- **server** — Graceful shutdown wrapper with signal handling (`SIGINT`/`SIGTERM`) and lifecycle hooks (`OnStart`/`OnShutdown`)
- **apitest** — Fluent test request builder, response recorder, and assertion helpers (`AssertStatus`, `AssertSuccess`, `AssertError`, `AssertValidationError`)
