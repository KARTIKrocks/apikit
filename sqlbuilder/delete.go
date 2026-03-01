package sqlbuilder

import (
	"slices"
	"strings"
)

// DeleteBuilder builds DELETE queries.
type DeleteBuilder struct {
	dialect    Dialect
	table      string
	using      []string
	conditions    []condition
	returning     []string
	returningExpr []Expr
	ctes          []cte
}

// Delete creates a new DeleteBuilder for the given table.
//
//	sqlbuilder.Delete("users").Where("id = $1", 1)
func Delete(table string) *DeleteBuilder {
	return &DeleteBuilder{table: table}
}

// SetDialect sets the SQL dialect for placeholder conversion at Build time.
func (b *DeleteBuilder) SetDialect(d Dialect) *DeleteBuilder {
	b.dialect = d
	return b
}

// Using adds tables for multi-table DELETE (PostgreSQL USING clause).
func (b *DeleteBuilder) Using(tables ...string) *DeleteBuilder {
	b.using = append(b.using, tables...)
	return b
}

// Where adds a WHERE condition.
func (b *DeleteBuilder) Where(sql string, args ...any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: sql, args: args})
	return b
}

// WhereIn adds a "col IN (...)" condition.
func (b *DeleteBuilder) WhereIn(col string, vals ...any) *DeleteBuilder {
	b.conditions = append(b.conditions, buildWhereIn(col, vals))
	return b
}

// WhereNull adds a "col IS NULL" condition.
func (b *DeleteBuilder) WhereNull(col string) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " IS NULL"})
	return b
}

// WhereNotNull adds a "col IS NOT NULL" condition.
func (b *DeleteBuilder) WhereNotNull(col string) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " IS NOT NULL"})
	return b
}

// WhereNotIn adds a "col NOT IN (...)" condition.
func (b *DeleteBuilder) WhereNotIn(col string, vals ...any) *DeleteBuilder {
	b.conditions = append(b.conditions, buildWhereNotIn(col, vals))
	return b
}

// WhereBetween adds a "col BETWEEN low AND high" condition.
func (b *DeleteBuilder) WhereBetween(col string, low, high any) *DeleteBuilder {
	b.conditions = append(b.conditions, buildWhereBetween(col, low, high))
	return b
}

// WhereEq adds a "col = val" condition.
func (b *DeleteBuilder) WhereEq(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " = $1", args: []any{val}})
	return b
}

// WhereNeq adds a "col != val" condition.
func (b *DeleteBuilder) WhereNeq(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " != $1", args: []any{val}})
	return b
}

// WhereGt adds a "col > val" condition.
func (b *DeleteBuilder) WhereGt(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " > $1", args: []any{val}})
	return b
}

// WhereGte adds a "col >= val" condition.
func (b *DeleteBuilder) WhereGte(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " >= $1", args: []any{val}})
	return b
}

// WhereLt adds a "col < val" condition.
func (b *DeleteBuilder) WhereLt(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " < $1", args: []any{val}})
	return b
}

// WhereLte adds a "col <= val" condition.
func (b *DeleteBuilder) WhereLte(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " <= $1", args: []any{val}})
	return b
}

// WhereLike adds a "col LIKE val" condition.
func (b *DeleteBuilder) WhereLike(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " LIKE $1", args: []any{val}})
	return b
}

// WhereILike adds a "col ILIKE val" condition (case-insensitive LIKE, PostgreSQL).
func (b *DeleteBuilder) WhereILike(col string, val any) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " ILIKE $1", args: []any{val}})
	return b
}

// WhereColumn adds a column-to-column comparison condition (e.g., "a.id = b.id").
func (b *DeleteBuilder) WhereColumn(col1, op, col2 string) *DeleteBuilder {
	b.conditions = append(b.conditions, condition{sql: col1 + " " + op + " " + col2})
	return b
}

// WhereExists adds a "EXISTS (subquery)" condition.
func (b *DeleteBuilder) WhereExists(sub *SelectBuilder) *DeleteBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  "EXISTS (" + sql + ")",
		args: args,
	})
	return b
}

// WhereNotExists adds a "NOT EXISTS (subquery)" condition.
func (b *DeleteBuilder) WhereNotExists(sub *SelectBuilder) *DeleteBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  "NOT EXISTS (" + sql + ")",
		args: args,
	})
	return b
}

// WhereOr adds a group of OR conditions wrapped in parentheses.
func (b *DeleteBuilder) WhereOr(conditions ...condition) *DeleteBuilder {
	b.conditions = append(b.conditions, buildOrGroup(conditions))
	return b
}

// WhereInSubquery adds a "col IN (subquery)" condition.
func (b *DeleteBuilder) WhereInSubquery(col string, sub *SelectBuilder) *DeleteBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  col + " IN (" + sql + ")",
		args: args,
	})
	return b
}

// WhereNotInSubquery adds a "col NOT IN (subquery)" condition.
func (b *DeleteBuilder) WhereNotInSubquery(col string, sub *SelectBuilder) *DeleteBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  col + " NOT IN (" + sql + ")",
		args: args,
	})
	return b
}

// When conditionally applies a function to the builder.
func (b *DeleteBuilder) When(cond bool, fn func(*DeleteBuilder)) *DeleteBuilder {
	if cond {
		fn(b)
	}
	return b
}

// Clone creates a deep copy of the builder.
func (b *DeleteBuilder) Clone() *DeleteBuilder {
	c := *b
	c.using = slices.Clone(b.using)
	c.conditions = make([]condition, len(b.conditions))
	copy(c.conditions, b.conditions)
	for i, cond := range c.conditions {
		c.conditions[i].args = slices.Clone(cond.args)
	}
	c.returning = slices.Clone(b.returning)
	c.returningExpr = slices.Clone(b.returningExpr)
	c.ctes = slices.Clone(b.ctes)
	return &c
}

// Returning adds a RETURNING clause.
func (b *DeleteBuilder) Returning(cols ...string) *DeleteBuilder {
	b.returning = cols
	return b
}

// ReturningExpr adds expression columns to the RETURNING clause.
func (b *DeleteBuilder) ReturningExpr(exprs ...Expr) *DeleteBuilder {
	b.returningExpr = append(b.returningExpr, exprs...)
	return b
}

// With adds a CTE (Common Table Expression).
func (b *DeleteBuilder) With(name string, q Query) *DeleteBuilder {
	b.ctes = append(b.ctes, cte{name: name, query: q})
	return b
}

// WithSelect adds a CTE from a SelectBuilder. This is dialect-safe: the
// subquery is always built with Postgres placeholders internally.
func (b *DeleteBuilder) WithSelect(name string, sub *SelectBuilder) *DeleteBuilder {
	sql, args := buildSelectPostgres(sub)
	b.ctes = append(b.ctes, cte{name: name, query: Query{SQL: sql, Args: args}})
	return b
}

// Build assembles the SQL string and arguments.
func (b *DeleteBuilder) Build() (string, []any) {
	var sb strings.Builder
	sb.Grow(128)
	args := make([]any, 0, 8)
	ac := &argCounter{}

	// CTEs
	cteArgs := writeCTEs(&sb, b.ctes, ac)
	args = append(args, cteArgs...)

	sb.WriteString("DELETE FROM ")
	sb.WriteString(b.table)

	// USING
	if len(b.using) > 0 {
		sb.WriteString(" USING ")
		writeJoined(&sb, b.using, ", ")
	}

	// WHERE
	whereArgs := writeWhereClause(&sb, b.conditions, ac)
	args = append(args, whereArgs...)

	writeReturningWithExpr(&sb, b.returning, b.returningExpr, ac, &args)

	return convertPlaceholders(sb.String(), b.dialect), args
}

// MustBuild calls Build and panics if the builder is in an invalid state.
func (b *DeleteBuilder) MustBuild() (string, []any) {
	return b.Build()
}

// Query builds and returns a Query struct.
func (b *DeleteBuilder) Query() Query {
	sql, args := b.Build()
	return Query{SQL: sql, Args: args}
}

// String returns the SQL string only, for debugging.
func (b *DeleteBuilder) String() string {
	sql, _ := b.Build()
	return sql
}
