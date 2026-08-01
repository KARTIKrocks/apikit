package sqlbuilder

import "strings"

// setOpKind is the type of set operation.
type setOpKind int

const (
	setOpUnion setOpKind = iota
	setOpUnionAll
	setOpIntersect
	setOpExcept
)

// setOp represents a set operation (UNION, INTERSECT, EXCEPT) with another query.
type setOp struct {
	kind  setOpKind
	query Query
	// parenWrap wraps the branch SQL in parentheses so its own ORDER BY /
	// LIMIT / OFFSET bind to that branch rather than to the whole set-op
	// statement.
	parenWrap bool
}

// branchScope holds the ORDER BY / LIMIT / OFFSET that were set on the left
// (receiver) branch before a set op was added. These are snapshotted so they
// stay bound to that branch instead of leaking onto the enclosing statement.
type branchScope struct {
	orderBy     []string
	orderByExpr []Expr
	limit       int
	hasLimit    bool
	offset      int64
	hasOffset   bool
}

// write renders the captured ORDER BY / LIMIT / OFFSET into sb.
func (b *branchScope) write(sb *strings.Builder, ac *argCounter) []any {
	args := writeOrderByClause(sb, ac, b.orderBy, b.orderByExpr)
	writeLimitOffset(sb, b.hasLimit, b.limit, b.hasOffset, b.offset)
	return args
}

// setOpBranchNeedsParens reports whether b, used as a set-op branch, must be
// wrapped in parentheses so its own scope binds to the branch rather than to
// the whole set-op statement. This is required when b carries ORDER BY / LIMIT
// / OFFSET, or when b is itself a compound (has its own set ops): without the
// wrapping paren, left-to-right associativity silently regroups
// `a OP (b OP2 c)` as `(a OP b) OP2 c`, changing the result set for any
// non-associative combination (e.g. UNION ALL mixed with UNION, or EXCEPT).
//
// A compound b always sets setOpLeft when it captured ordering, so the
// setOps check below also covers the "b's left branch had ORDER BY" case.
//
// Note: SQLite cannot represent any of this — it rejects a parenthesized
// compound select as a set-op term, just as it rejects a per-branch ORDER BY /
// LIMIT — so nested or scoped compound set operations target PostgreSQL/MySQL.
func setOpBranchNeedsParens(b *SelectBuilder) bool {
	return len(b.orderBy) > 0 || len(b.orderByExpr) > 0 ||
		b.hasLimit || b.hasOffset || len(b.setOps) > 0
}

// setOpPrec returns the SQL binding precedence of a set operator. INTERSECT
// binds more tightly than UNION / UNION ALL / EXCEPT, which share a lower
// precedence and are left-associative. Used to decide where a flat operator
// chain must be parenthesized to preserve the caller's left-to-right grouping.
func setOpPrec(k setOpKind) int {
	if k == setOpIntersect {
		return 2
	}
	return 1
}

// setOpKeyword returns the SQL keyword for a set operation.
func setOpKeyword(k setOpKind) string {
	switch k {
	case setOpUnion:
		return "UNION"
	case setOpUnionAll:
		return "UNION ALL"
	case setOpIntersect:
		return "INTERSECT"
	case setOpExcept:
		return "EXCEPT"
	default:
		return "UNION"
	}
}
