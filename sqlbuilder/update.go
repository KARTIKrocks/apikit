package sqlbuilder

import (
	"slices"
	"strings"
)

// UpdateBuilder builds UPDATE queries.
type UpdateBuilder struct {
	dialect    Dialect
	table      string
	setClauses []setClause
	fromTables []string
	conditions []condition
	returning  []string
	ctes       []cte
}

type setClause struct {
	col  string
	val  any
	expr *Expr // if non-nil, use expression instead of val
}

// Update creates a new UpdateBuilder for the given table.
//
//	sqlbuilder.Update("users").Set("name", "Bob").Where("id = $1", 1)
func Update(table string) *UpdateBuilder {
	return &UpdateBuilder{table: table}
}

// SetDialect sets the SQL dialect for placeholder conversion at Build time.
func (b *UpdateBuilder) SetDialect(d Dialect) *UpdateBuilder {
	b.dialect = d
	return b
}

// Set adds a column = value assignment.
func (b *UpdateBuilder) Set(col string, val any) *UpdateBuilder {
	b.setClauses = append(b.setClauses, setClause{col: col, val: val})
	return b
}

// SetExpr adds a column = expression assignment.
//
//	.SetExpr("updated_at", sqlbuilder.Raw("NOW()"))
func (b *UpdateBuilder) SetExpr(col string, expr Expr) *UpdateBuilder {
	b.setClauses = append(b.setClauses, setClause{col: col, expr: &expr})
	return b
}

// SetMap adds multiple column = value assignments from a map.
// Keys are sorted for deterministic output.
func (b *UpdateBuilder) SetMap(m map[string]any) *UpdateBuilder {
	keys := sortedKeys(m)
	for _, k := range keys {
		b.setClauses = append(b.setClauses, setClause{col: k, val: m[k]})
	}
	return b
}

// From adds tables for multi-table UPDATE (PostgreSQL FROM clause).
func (b *UpdateBuilder) From(tables ...string) *UpdateBuilder {
	b.fromTables = append(b.fromTables, tables...)
	return b
}

// Where adds a WHERE condition.
func (b *UpdateBuilder) Where(sql string, args ...any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: sql, args: args})
	return b
}

// WhereIn adds a "col IN (...)" condition.
func (b *UpdateBuilder) WhereIn(col string, vals ...any) *UpdateBuilder {
	b.conditions = append(b.conditions, buildWhereIn(col, vals))
	return b
}

// WhereNull adds a "col IS NULL" condition.
func (b *UpdateBuilder) WhereNull(col string) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " IS NULL"})
	return b
}

// WhereNotNull adds a "col IS NOT NULL" condition.
func (b *UpdateBuilder) WhereNotNull(col string) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " IS NOT NULL"})
	return b
}

// WhereNotIn adds a "col NOT IN (...)" condition.
func (b *UpdateBuilder) WhereNotIn(col string, vals ...any) *UpdateBuilder {
	b.conditions = append(b.conditions, buildWhereNotIn(col, vals))
	return b
}

// WhereBetween adds a "col BETWEEN low AND high" condition.
func (b *UpdateBuilder) WhereBetween(col string, low, high any) *UpdateBuilder {
	b.conditions = append(b.conditions, buildWhereBetween(col, low, high))
	return b
}

// WhereEq adds a "col = val" condition.
func (b *UpdateBuilder) WhereEq(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " = $1", args: []any{val}})
	return b
}

// WhereNeq adds a "col != val" condition.
func (b *UpdateBuilder) WhereNeq(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " != $1", args: []any{val}})
	return b
}

// WhereGt adds a "col > val" condition.
func (b *UpdateBuilder) WhereGt(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " > $1", args: []any{val}})
	return b
}

// WhereGte adds a "col >= val" condition.
func (b *UpdateBuilder) WhereGte(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " >= $1", args: []any{val}})
	return b
}

// WhereLt adds a "col < val" condition.
func (b *UpdateBuilder) WhereLt(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " < $1", args: []any{val}})
	return b
}

// WhereLte adds a "col <= val" condition.
func (b *UpdateBuilder) WhereLte(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " <= $1", args: []any{val}})
	return b
}

// WhereLike adds a "col LIKE val" condition.
func (b *UpdateBuilder) WhereLike(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " LIKE $1", args: []any{val}})
	return b
}

// WhereILike adds a "col ILIKE val" condition (case-insensitive LIKE, PostgreSQL).
func (b *UpdateBuilder) WhereILike(col string, val any) *UpdateBuilder {
	b.conditions = append(b.conditions, condition{sql: col + " ILIKE $1", args: []any{val}})
	return b
}

// WhereExists adds a "EXISTS (subquery)" condition.
func (b *UpdateBuilder) WhereExists(sub *SelectBuilder) *UpdateBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  "EXISTS (" + sql + ")",
		args: args,
	})
	return b
}

// WhereNotExists adds a "NOT EXISTS (subquery)" condition.
func (b *UpdateBuilder) WhereNotExists(sub *SelectBuilder) *UpdateBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  "NOT EXISTS (" + sql + ")",
		args: args,
	})
	return b
}

// WhereOr adds a group of OR conditions wrapped in parentheses.
func (b *UpdateBuilder) WhereOr(conditions ...condition) *UpdateBuilder {
	b.conditions = append(b.conditions, buildOrGroup(conditions))
	return b
}

// WhereInSubquery adds a "col IN (subquery)" condition.
func (b *UpdateBuilder) WhereInSubquery(col string, sub *SelectBuilder) *UpdateBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  col + " IN (" + sql + ")",
		args: args,
	})
	return b
}

// WhereNotInSubquery adds a "col NOT IN (subquery)" condition.
func (b *UpdateBuilder) WhereNotInSubquery(col string, sub *SelectBuilder) *UpdateBuilder {
	sql, args := buildSelectPostgres(sub)
	b.conditions = append(b.conditions, condition{
		sql:  col + " NOT IN (" + sql + ")",
		args: args,
	})
	return b
}

// When conditionally applies a function to the builder.
func (b *UpdateBuilder) When(cond bool, fn func(*UpdateBuilder)) *UpdateBuilder {
	if cond {
		fn(b)
	}
	return b
}

// Clone creates a deep copy of the builder.
func (b *UpdateBuilder) Clone() *UpdateBuilder {
	c := *b
	c.setClauses = slices.Clone(b.setClauses)
	c.fromTables = slices.Clone(b.fromTables)
	c.conditions = slices.Clone(b.conditions)
	c.returning = slices.Clone(b.returning)
	c.ctes = slices.Clone(b.ctes)
	return &c
}

// Increment adds a "col = col + n" assignment.
func (b *UpdateBuilder) Increment(col string, n any) *UpdateBuilder {
	expr := Expr{SQL: col + " + $1", Args: []any{n}}
	b.setClauses = append(b.setClauses, setClause{col: col, expr: &expr})
	return b
}

// Decrement adds a "col = col - n" assignment.
func (b *UpdateBuilder) Decrement(col string, n any) *UpdateBuilder {
	expr := Expr{SQL: col + " - $1", Args: []any{n}}
	b.setClauses = append(b.setClauses, setClause{col: col, expr: &expr})
	return b
}

// Returning adds a RETURNING clause.
func (b *UpdateBuilder) Returning(cols ...string) *UpdateBuilder {
	b.returning = cols
	return b
}

// With adds a CTE (Common Table Expression).
func (b *UpdateBuilder) With(name string, q Query) *UpdateBuilder {
	b.ctes = append(b.ctes, cte{name: name, query: q})
	return b
}

// WithSelect adds a CTE from a SelectBuilder. This is dialect-safe: the
// subquery is always built with Postgres placeholders internally.
func (b *UpdateBuilder) WithSelect(name string, sub *SelectBuilder) *UpdateBuilder {
	sql, args := buildSelectPostgres(sub)
	b.ctes = append(b.ctes, cte{name: name, query: Query{SQL: sql, Args: args}})
	return b
}

// Build assembles the SQL string and arguments.
func (b *UpdateBuilder) Build() (string, []any) {
	var sb strings.Builder
	sb.Grow(256)
	args := make([]any, 0, 8)
	ac := &argCounter{}

	// CTEs
	cteArgs := writeCTEs(&sb, b.ctes, ac)
	args = append(args, cteArgs...)

	sb.WriteString("UPDATE ")
	sb.WriteString(b.table)
	sb.WriteString(" SET ")

	for i, sc := range b.setClauses {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sc.col)
		sb.WriteString(" = ")
		if sc.expr != nil {
			rebased := rebasePlaceholders(sc.expr.SQL, ac.offset())
			sb.WriteString(rebased)
			args = append(args, sc.expr.Args...)
			ac.n += len(sc.expr.Args)
		} else {
			ac.writePlaceholder(&sb)
			args = append(args, sc.val)
		}
	}

	// FROM
	if len(b.fromTables) > 0 {
		sb.WriteString(" FROM ")
		writeJoined(&sb, b.fromTables, ", ")
	}

	// WHERE
	whereArgs := writeWhereClause(&sb, b.conditions, ac)
	args = append(args, whereArgs...)

	writeReturning(&sb, b.returning)

	return convertPlaceholders(sb.String(), b.dialect), args
}

// MustBuild calls Build and panics if the builder is in an invalid state.
func (b *UpdateBuilder) MustBuild() (string, []any) {
	return b.Build()
}

// Query builds and returns a Query struct.
func (b *UpdateBuilder) Query() Query {
	sql, args := b.Build()
	return Query{SQL: sql, Args: args}
}

// String returns the SQL string only, for debugging.
func (b *UpdateBuilder) String() string {
	sql, _ := b.Build()
	return sql
}
