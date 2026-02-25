package sqlbuilder

import (
	"testing"
)

// Bug 1: ValueMap ignores existing column order.
// If Columns() was called first, ValueMap extracts values in its own sorted key order,
// which may not match the column order, causing silent value/column misalignment.
func TestBug_ValueMapIgnoresColumnOrder(t *testing.T) {
	sql, args := Insert("t").
		Columns("b", "a").
		ValueMap(map[string]any{"a": 1, "b": 2}).
		Build()
	// Columns are (b, a), so values should be (2, 1) to match
	expectSQL(t, "INSERT INTO t (b, a) VALUES ($1, $2)", sql)
	expectArgs(t, []any{2, 1}, args)
}

// Bug 2: OnConflictUpdateExpr with no-arg expressions doesn't append sentinel row.
// Build() assumes last row is conflict args and eats a real data row.
func TestBug_OnConflictUpdateExprNoArgs(t *testing.T) {
	sql, args := Insert("users").
		Columns("email", "name").
		Values("alice@example.com", "Alice").
		OnConflictUpdateExpr(
			[]string{"email"},
			map[string]Expr{"name": Raw("EXCLUDED.name")},
		).
		Build()
	expectSQL(t, "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name", sql)
	expectArgs(t, []any{"alice@example.com", "Alice"}, args)
}

// Bug 3: FromSubquery placeholders are not rebased when CTE args exist.
// The subquery SQL is baked into s.from at chain time, never rebased at Build time.
func TestBug_FromSubqueryWithCTE(t *testing.T) {
	cteQ := Query{SQL: "SELECT id FROM source WHERE x = $1", Args: []any{"a"}}
	sub := Select("id").From("t").Where("y = $1", "b")

	sql, args := Select("id").
		With("c", cteQ).
		FromSubquery(sub, "s").
		Build()
	// CTE uses $1 for "a", subquery should use $2 for "b"
	expectSQL(t, "WITH c AS (SELECT id FROM source WHERE x = $1) SELECT id FROM (SELECT id FROM t WHERE y = $2) s", sql)
	expectArgs(t, []any{"a", "b"}, args)
}

// Bug 4: FromSubquery with ColumnExpr args — subquery placeholders not rebased.
func TestBug_FromSubqueryAfterColumnExprArgs(t *testing.T) {
	sub := Select("id").From("t").Where("x = $1", "val")

	sql, args := SelectExpr(RawExpr("COALESCE($1, 'default')", "test")).
		FromSubquery(sub, "s").
		Build()
	// Column expr uses $1 for "test", subquery should use $2 for "val"
	expectSQL(t, "SELECT COALESCE($1, 'default') FROM (SELECT id FROM t WHERE x = $2) s", sql)
	expectArgs(t, []any{"test", "val"}, args)
}
