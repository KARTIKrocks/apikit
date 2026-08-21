# Style and pattern rationale

Context for the scoped rules in `config.json`. This file is freeform prose read
alongside the diff; `config.json` is what actually gates comment scope and
severity.

## Set-op composition — worked example

`set-op-composition-integrity` exists because of a real, subtle bug fixed in
`sqlbuilder` (see CHANGELOG 0.27.0 / commit `dbede9c`): `Union`, `UnionAll`,
`Intersect`, and `Except` mishandled two separate things at once, and both are
easy to reintroduce by "simplifying" the code:

- **Branch scope.** Each branch used to be concatenated with no parentheses,
  so a branch's own `ORDER BY`/`LIMIT`/`OFFSET` leaked onto the enclosing
  statement — a union of two independently-ordered-and-limited branches
  emitted two `ORDER BY`s and two `LIMIT`s (a hard syntax error), and the
  right branch's clauses silently bound to the whole statement instead of
  that branch.
- **Chain grouping and precedence.** A compound branch like
  `a.UnionAll(b.Union(c))` was flattened into `(a UNION ALL b) UNION c` — a
  different result set than intended. And because SQL binds `INTERSECT`
  tighter than `UNION`/`EXCEPT`, a flat chain like `a.Union(b).Intersect(c)`
  rendered as `a UNION b INTERSECT c`, which SQL parses as
  `a UNION (b INTERSECT c)` — not the fluent left-to-right order the caller
  wrote.

The fix snapshots each branch's own clauses and parenthesizes both compound
branches and precedence-rising points in a flat chain. A future "cleanup" that
flattens the parenthesization logic back down, or that treats all set-op kinds
as commutative/associative for rendering purposes, reintroduces this bug.
Note the fix is Postgres/MySQL-only — SQLite doesn't accept a parenthesized or
per-branch-ordered compound select as a set-op term, so don't suggest
extending the parenthesized form to the SQLite dialect.

## Identifiers are the caller's responsibility — and there's a gap

`sqlbuilder` parameterizes every *value* (numbered `$N` for Postgres, `?` for
MySQL/SQLite via `convertPlaceholders`) — that part is solid. What it does
**not** do is quote or validate *identifiers* (table/column names): those are
written into the SQL string as given. That's a reasonable design for a
builder driven by Go code, but `request.SortConfig`/`FilterConfig`
(`request/sort.go`, `request/filter.go`) exist specifically to turn
query-string input into field names, and both only enforce their
`AllowedFields` allowlist when it's non-empty. An empty/unset `AllowedFields`
— which reads as "no restriction" rather than "no fields allowed" — lets any
query-param value through untouched. If that value later reaches a
`sqlbuilder` column/table position or a raw SQL string, it's a straightforward
identifier-injection vector. There's no bug on record for this today; it's a
gap between two packages that individually look fine and jointly don't. Treat
any new integration point between `request`'s sort/filter parsing and SQL
construction as security-relevant.

## Error metadata has to reach every renderer

`errors.Error` carries `Code`, `Message`, `Fields`, `Details`, and `Stack`.
Commit `d2b24e1` fixed exactly this class of bug: `Details` had existed on
`errors.Error` (with `WithDetail`/`WithDetails` constructors) and in the
`response` package's envelope for a while before anyone noticed `router`'s own
`errorBody`/`handleError` never read it — so details attached to an error
returned from a router handler were silently dropped before reaching the
client. There was no compile error, no test failure in `errors/` or
`response/` — only a router-level test would have caught the gap, which is
exactly why `errors/**`, `router/**`, and `response/**` are all in this rule's
scope together.

## Package decoupling is a stated design goal, not a suggestion

AGENTS.md is explicit: "a user importing only `errors` should not pull in
`sqlbuilder`." Each top-level directory is meant to be usable standalone. A PR
that adds a convenience import across these boundaries (e.g. a helper in
`response/` that imports `errors/` types more deeply than the public
constructors, or `middleware/` reaching into `dbx/`) should be flagged even if
it's otherwise correct code — the cost is paid by every consumer who now pulls
in a package they didn't ask for.

## dbx's scanner fails silently by design — know when that's not okay

`dbx/scan.go`'s `buildColumnMapping` discards any SQL result column that has
no matching `db`-tagged field: `targets[i] = discard` (a shared `sql.RawBytes`
scratch value), no error. This is the right behavior for `SELECT *` against a
struct that only maps a subset of columns. It is indistinguishable, though,
from a `db:"usre_id"` typo silently leaving a field at its zero value forever.
Nothing in the current design distinguishes "intentionally partial" from
"broken mapping" — don't assume a future caller will notice; if a new API
wants strict mode, it needs to be its own explicit option, not inferred from
existing behavior.
