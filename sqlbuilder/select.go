package sqlbuilder

import (
	"slices"
	"strconv"
	"strings"

	"github.com/KARTIKrocks/apikit/request"
)

// SelectBuilder builds SELECT queries.
type SelectBuilder struct {
	dialect      Dialect
	distinct     bool
	distinctOn   []string
	columns      []string
	columnExpr   []Expr
	from         string
	fromSubquery *SelectBuilder
	fromSubAlias string
	joins        []joinClause
	conditions   []condition
	groupBy      []string
	groupByExpr  []Expr
	having       []condition
	orderBy      []string
	limit        int
	hasLimit     bool
	offsetVal    int64
	hasOffset    bool
	lockMode     string
	skipLocked   bool
	noWait       bool
	setOps       []setOp
	ctes         []cte
}

type joinClause struct {
	kind     string // "JOIN", "LEFT JOIN", etc.
	table    string
	on       string
	args     []any
	subquery *SelectBuilder
	subAlias string
}

// Select creates a new SelectBuilder with the given columns.
//
//	sqlbuilder.Select("id", "name", "email")
func Select(columns ...string) *SelectBuilder {
	return &SelectBuilder{columns: columns}
}

// SelectExpr creates a new SelectBuilder with expression columns.
//
//	sqlbuilder.SelectExpr(sqlbuilder.Raw("COUNT(*)"))
func SelectExpr(exprs ...Expr) *SelectBuilder {
	return &SelectBuilder{columnExpr: exprs}
}

// SetDialect sets the SQL dialect for placeholder conversion at Build time.
func (s *SelectBuilder) SetDialect(d Dialect) *SelectBuilder {
	s.dialect = d
	return s
}

// Distinct adds DISTINCT to the SELECT.
func (s *SelectBuilder) Distinct() *SelectBuilder {
	s.distinct = true
	return s
}

// DistinctOn adds DISTINCT ON (cols) to the SELECT (PostgreSQL).
func (s *SelectBuilder) DistinctOn(cols ...string) *SelectBuilder {
	s.distinctOn = append(s.distinctOn, cols...)
	return s
}

// Column adds a column to the SELECT list.
func (s *SelectBuilder) Column(col string) *SelectBuilder {
	s.columns = append(s.columns, col)
	return s
}

// Columns adds multiple columns to the SELECT list.
func (s *SelectBuilder) Columns(cols ...string) *SelectBuilder {
	s.columns = append(s.columns, cols...)
	return s
}

// ColumnExpr adds an expression column to the SELECT list.
func (s *SelectBuilder) ColumnExpr(expr Expr) *SelectBuilder {
	s.columnExpr = append(s.columnExpr, expr)
	return s
}

// From sets the FROM table.
func (s *SelectBuilder) From(table string) *SelectBuilder {
	s.from = table
	return s
}

// FromAlias sets the FROM table with an alias.
func (s *SelectBuilder) FromAlias(table, alias string) *SelectBuilder {
	s.from = table + " " + alias
	return s
}

// FromSubquery sets the FROM clause to a subquery.
// The subquery is built at Build() time so placeholders are correctly rebased.
func (s *SelectBuilder) FromSubquery(sub *SelectBuilder, alias string) *SelectBuilder {
	s.fromSubquery = sub
	s.fromSubAlias = alias
	return s
}

// Join adds an INNER JOIN.
func (s *SelectBuilder) Join(table, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "JOIN", table: table, on: on, args: args})
	return s
}

// LeftJoin adds a LEFT JOIN.
func (s *SelectBuilder) LeftJoin(table, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "LEFT JOIN", table: table, on: on, args: args})
	return s
}

// RightJoin adds a RIGHT JOIN.
func (s *SelectBuilder) RightJoin(table, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "RIGHT JOIN", table: table, on: on, args: args})
	return s
}

// FullJoin adds a FULL JOIN.
func (s *SelectBuilder) FullJoin(table, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "FULL JOIN", table: table, on: on, args: args})
	return s
}

// CrossJoin adds a CROSS JOIN.
func (s *SelectBuilder) CrossJoin(table string) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "CROSS JOIN", table: table})
	return s
}

// JoinSubquery adds an INNER JOIN with a subquery.
func (s *SelectBuilder) JoinSubquery(sub *SelectBuilder, alias, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "JOIN", subquery: sub, subAlias: alias, on: on, args: args})
	return s
}

// LeftJoinSubquery adds a LEFT JOIN with a subquery.
func (s *SelectBuilder) LeftJoinSubquery(sub *SelectBuilder, alias, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "LEFT JOIN", subquery: sub, subAlias: alias, on: on, args: args})
	return s
}

// RightJoinSubquery adds a RIGHT JOIN with a subquery.
func (s *SelectBuilder) RightJoinSubquery(sub *SelectBuilder, alias, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "RIGHT JOIN", subquery: sub, subAlias: alias, on: on, args: args})
	return s
}

// FullJoinSubquery adds a FULL JOIN with a subquery.
func (s *SelectBuilder) FullJoinSubquery(sub *SelectBuilder, alias, on string, args ...any) *SelectBuilder {
	s.joins = append(s.joins, joinClause{kind: "FULL JOIN", subquery: sub, subAlias: alias, on: on, args: args})
	return s
}

// Where adds a WHERE condition. Placeholders use $1-relative numbering
// and are rebased at Build time.
func (s *SelectBuilder) Where(sql string, args ...any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: sql, args: args})
	return s
}

// WhereIn adds a "col IN (...)" condition.
func (s *SelectBuilder) WhereIn(col string, vals ...any) *SelectBuilder {
	s.conditions = append(s.conditions, buildWhereIn(col, vals))
	return s
}

// WhereNotIn adds a "col NOT IN (...)" condition.
func (s *SelectBuilder) WhereNotIn(col string, vals ...any) *SelectBuilder {
	s.conditions = append(s.conditions, buildWhereNotIn(col, vals))
	return s
}

// WhereBetween adds a "col BETWEEN low AND high" condition.
func (s *SelectBuilder) WhereBetween(col string, low, high any) *SelectBuilder {
	s.conditions = append(s.conditions, buildWhereBetween(col, low, high))
	return s
}

// WhereNull adds a "col IS NULL" condition.
func (s *SelectBuilder) WhereNull(col string) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " IS NULL"})
	return s
}

// WhereNotNull adds a "col IS NOT NULL" condition.
func (s *SelectBuilder) WhereNotNull(col string) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " IS NOT NULL"})
	return s
}

// WhereExists adds a "EXISTS (subquery)" condition.
func (s *SelectBuilder) WhereExists(sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.conditions = append(s.conditions, condition{
		sql:  "EXISTS (" + sql + ")",
		args: args,
	})
	return s
}

// WhereNotExists adds a "NOT EXISTS (subquery)" condition.
func (s *SelectBuilder) WhereNotExists(sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.conditions = append(s.conditions, condition{
		sql:  "NOT EXISTS (" + sql + ")",
		args: args,
	})
	return s
}

// WhereInSubquery adds a "col IN (subquery)" condition.
func (s *SelectBuilder) WhereInSubquery(col string, sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.conditions = append(s.conditions, condition{
		sql:  col + " IN (" + sql + ")",
		args: args,
	})
	return s
}

// WhereNotInSubquery adds a "col NOT IN (subquery)" condition.
func (s *SelectBuilder) WhereNotInSubquery(col string, sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.conditions = append(s.conditions, condition{
		sql:  col + " NOT IN (" + sql + ")",
		args: args,
	})
	return s
}

// WhereColumn adds a column-to-column comparison condition (e.g., "a.id = b.id").
func (s *SelectBuilder) WhereColumn(col1, op, col2 string) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col1 + " " + op + " " + col2})
	return s
}

// WhereOr adds a group of OR conditions wrapped in parentheses.
//
//	.WhereOr(
//	    sqlbuilder.Or("status = $1", "active"),
//	    sqlbuilder.Or("role = $1", "admin"),
//	)
//	// WHERE ... AND (status = $1 OR role = $2)
func (s *SelectBuilder) WhereOr(conditions ...condition) *SelectBuilder {
	s.conditions = append(s.conditions, buildOrGroup(conditions))
	return s
}

// WhereEq adds a "col = val" condition.
func (s *SelectBuilder) WhereEq(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " = $1", args: []any{val}})
	return s
}

// WhereNeq adds a "col != val" condition.
func (s *SelectBuilder) WhereNeq(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " != $1", args: []any{val}})
	return s
}

// WhereGt adds a "col > val" condition.
func (s *SelectBuilder) WhereGt(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " > $1", args: []any{val}})
	return s
}

// WhereGte adds a "col >= val" condition.
func (s *SelectBuilder) WhereGte(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " >= $1", args: []any{val}})
	return s
}

// WhereLt adds a "col < val" condition.
func (s *SelectBuilder) WhereLt(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " < $1", args: []any{val}})
	return s
}

// WhereLte adds a "col <= val" condition.
func (s *SelectBuilder) WhereLte(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " <= $1", args: []any{val}})
	return s
}

// WhereLike adds a "col LIKE val" condition.
func (s *SelectBuilder) WhereLike(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " LIKE $1", args: []any{val}})
	return s
}

// WhereILike adds a "col ILIKE val" condition (case-insensitive LIKE, PostgreSQL).
func (s *SelectBuilder) WhereILike(col string, val any) *SelectBuilder {
	s.conditions = append(s.conditions, condition{sql: col + " ILIKE $1", args: []any{val}})
	return s
}

// GroupBy adds GROUP BY columns.
func (s *SelectBuilder) GroupBy(cols ...string) *SelectBuilder {
	s.groupBy = append(s.groupBy, cols...)
	return s
}

// GroupByExpr adds a GROUP BY expression.
func (s *SelectBuilder) GroupByExpr(expr Expr) *SelectBuilder {
	s.groupByExpr = append(s.groupByExpr, expr)
	return s
}

// Having adds a HAVING condition.
func (s *SelectBuilder) Having(sql string, args ...any) *SelectBuilder {
	s.having = append(s.having, condition{sql: sql, args: args})
	return s
}

// HavingIn adds a "col IN (...)" HAVING condition.
func (s *SelectBuilder) HavingIn(col string, vals ...any) *SelectBuilder {
	s.having = append(s.having, buildWhereIn(col, vals))
	return s
}

// HavingBetween adds a "col BETWEEN low AND high" HAVING condition.
func (s *SelectBuilder) HavingBetween(col string, low, high any) *SelectBuilder {
	s.having = append(s.having, buildWhereBetween(col, low, high))
	return s
}

// OrderBy adds ORDER BY clauses (e.g., "name ASC", "created_at DESC").
func (s *SelectBuilder) OrderBy(clauses ...string) *SelectBuilder {
	s.orderBy = append(s.orderBy, clauses...)
	return s
}

// OrderByAsc adds ORDER BY columns with ASC direction.
func (s *SelectBuilder) OrderByAsc(cols ...string) *SelectBuilder {
	for _, col := range cols {
		s.orderBy = append(s.orderBy, col+" ASC")
	}
	return s
}

// OrderByDesc adds ORDER BY columns with DESC direction.
func (s *SelectBuilder) OrderByDesc(cols ...string) *SelectBuilder {
	for _, col := range cols {
		s.orderBy = append(s.orderBy, col+" DESC")
	}
	return s
}

// OrderByExpr adds an ORDER BY clause from an expression.
func (s *SelectBuilder) OrderByExpr(expr Expr) *SelectBuilder {
	s.orderBy = append(s.orderBy, expr.SQL)
	return s
}

// Limit sets the LIMIT clause.
func (s *SelectBuilder) Limit(n int) *SelectBuilder {
	s.limit = n
	s.hasLimit = true
	return s
}

// Offset sets the OFFSET clause.
func (s *SelectBuilder) Offset(n int64) *SelectBuilder {
	s.offsetVal = n
	s.hasOffset = true
	return s
}

// ForUpdate adds FOR UPDATE locking.
func (s *SelectBuilder) ForUpdate() *SelectBuilder {
	s.lockMode = "FOR UPDATE"
	return s
}

// ForShare adds FOR SHARE locking.
func (s *SelectBuilder) ForShare() *SelectBuilder {
	s.lockMode = "FOR SHARE"
	return s
}

// SkipLocked adds SKIP LOCKED to the lock clause.
func (s *SelectBuilder) SkipLocked() *SelectBuilder {
	s.skipLocked = true
	return s
}

// NoWait adds NOWAIT to the lock clause.
func (s *SelectBuilder) NoWait() *SelectBuilder {
	s.noWait = true
	return s
}

// Union adds a UNION with another SELECT.
func (s *SelectBuilder) Union(other *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(other)
	s.setOps = append(s.setOps, setOp{kind: setOpUnion, query: Query{SQL: sql, Args: args}})
	return s
}

// UnionAll adds a UNION ALL with another SELECT.
func (s *SelectBuilder) UnionAll(other *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(other)
	s.setOps = append(s.setOps, setOp{kind: setOpUnionAll, query: Query{SQL: sql, Args: args}})
	return s
}

// Intersect adds an INTERSECT with another SELECT.
func (s *SelectBuilder) Intersect(other *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(other)
	s.setOps = append(s.setOps, setOp{kind: setOpIntersect, query: Query{SQL: sql, Args: args}})
	return s
}

// Except adds an EXCEPT with another SELECT.
func (s *SelectBuilder) Except(other *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(other)
	s.setOps = append(s.setOps, setOp{kind: setOpExcept, query: Query{SQL: sql, Args: args}})
	return s
}

// With adds a CTE (Common Table Expression).
func (s *SelectBuilder) With(name string, q Query) *SelectBuilder {
	s.ctes = append(s.ctes, cte{name: name, query: q})
	return s
}

// WithRecursive adds a recursive CTE.
func (s *SelectBuilder) WithRecursive(name string, q Query) *SelectBuilder {
	s.ctes = append(s.ctes, cte{name: name, query: q, recursive: true})
	return s
}

// WithSelect adds a CTE from a SelectBuilder. This is dialect-safe: the
// subquery is always built with Postgres placeholders internally, so
// placeholder rebasing works correctly regardless of the sub's dialect.
func (s *SelectBuilder) WithSelect(name string, sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.ctes = append(s.ctes, cte{name: name, query: Query{SQL: sql, Args: args}})
	return s
}

// WithRecursiveSelect adds a recursive CTE from a SelectBuilder.
func (s *SelectBuilder) WithRecursiveSelect(name string, sub *SelectBuilder) *SelectBuilder {
	sql, args := buildSelectPostgres(sub)
	s.ctes = append(s.ctes, cte{name: name, query: Query{SQL: sql, Args: args}, recursive: true})
	return s
}

// ApplyPagination applies request.Pagination as LIMIT and OFFSET.
func (s *SelectBuilder) ApplyPagination(p request.Pagination) *SelectBuilder {
	s.limit = p.PerPage
	s.hasLimit = true
	s.offsetVal = p.Offset
	s.hasOffset = true
	return s
}

// ApplySort applies request.SortField slices as ORDER BY clauses.
// allowedColumns maps API field names to SQL column expressions.
// Unknown fields are silently skipped.
func (s *SelectBuilder) ApplySort(sorts []request.SortField, allowedColumns map[string]string) *SelectBuilder {
	for _, sf := range sorts {
		col, ok := allowedColumns[sf.Field]
		if !ok {
			continue
		}
		dir := "ASC"
		if sf.Direction == request.SortDesc {
			dir = "DESC"
		}
		s.orderBy = append(s.orderBy, col+" "+dir)
	}
	return s
}

// When conditionally applies a function to the builder.
// If cond is true, fn is called with the builder. Always returns the builder for chaining.
func (s *SelectBuilder) When(cond bool, fn func(*SelectBuilder)) *SelectBuilder {
	if cond {
		fn(s)
	}
	return s
}

// Clone creates a deep copy of the builder.
func (s *SelectBuilder) Clone() *SelectBuilder {
	c := *s
	c.columns = slices.Clone(s.columns)
	c.columnExpr = slices.Clone(s.columnExpr)
	c.distinctOn = slices.Clone(s.distinctOn)
	c.joins = make([]joinClause, len(s.joins))
	copy(c.joins, s.joins)
	for i, j := range c.joins {
		if j.subquery != nil {
			c.joins[i].subquery = j.subquery.Clone()
		}
	}
	c.conditions = slices.Clone(s.conditions)
	c.groupBy = slices.Clone(s.groupBy)
	c.groupByExpr = slices.Clone(s.groupByExpr)
	c.having = slices.Clone(s.having)
	c.orderBy = slices.Clone(s.orderBy)
	c.setOps = slices.Clone(s.setOps)
	c.ctes = slices.Clone(s.ctes)
	if s.fromSubquery != nil {
		sub := s.fromSubquery.Clone()
		c.fromSubquery = sub
	}
	return &c
}

// ApplyFilters applies request.Filter slices as WHERE conditions.
// allowedColumns maps API field names to SQL column expressions.
// Unknown fields are silently skipped.
func (s *SelectBuilder) ApplyFilters(filters []request.Filter, allowedColumns map[string]string) *SelectBuilder {
	for _, f := range filters {
		col, ok := allowedColumns[f.Field]
		if !ok {
			continue
		}
		switch f.Operator {
		case request.FilterOpEq:
			s.conditions = append(s.conditions, condition{sql: col + " = $1", args: []any{f.Value}})
		case request.FilterOpNeq:
			s.conditions = append(s.conditions, condition{sql: col + " != $1", args: []any{f.Value}})
		case request.FilterOpGt:
			s.conditions = append(s.conditions, condition{sql: col + " > $1", args: []any{f.Value}})
		case request.FilterOpGte:
			s.conditions = append(s.conditions, condition{sql: col + " >= $1", args: []any{f.Value}})
		case request.FilterOpLt:
			s.conditions = append(s.conditions, condition{sql: col + " < $1", args: []any{f.Value}})
		case request.FilterOpLte:
			s.conditions = append(s.conditions, condition{sql: col + " <= $1", args: []any{f.Value}})
		case request.FilterOpIn:
			vals := strings.Split(f.Value, ",")
			anyVals := make([]any, len(vals))
			for i, v := range vals {
				anyVals[i] = strings.TrimSpace(v)
			}
			s.conditions = append(s.conditions, buildWhereIn(col, anyVals))
		}
	}
	return s
}

// Build assembles the SQL string and arguments.
// Order: WITH → SELECT [DISTINCT] → FROM → JOINs → WHERE → GROUP BY → HAVING → set ops → ORDER BY → LIMIT → OFFSET → FOR
func (s *SelectBuilder) Build() (string, []any) {
	var sb strings.Builder
	sb.Grow(256)
	args := make([]any, 0, 8)
	ac := &argCounter{}

	// CTEs
	cteArgs := writeCTEs(&sb, s.ctes, ac)
	args = append(args, cteArgs...)

	// SELECT
	sb.WriteString("SELECT ")
	if len(s.distinctOn) > 0 {
		sb.WriteString("DISTINCT ON (")
		writeJoined(&sb, s.distinctOn, ", ")
		sb.WriteString(") ")
	} else if s.distinct {
		sb.WriteString("DISTINCT ")
	}

	// Columns
	colCount := len(s.columns) + len(s.columnExpr)
	if colCount == 0 {
		sb.WriteString("*")
	} else {
		written := 0
		for _, col := range s.columns {
			if written > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(col)
			written++
		}
		for _, e := range s.columnExpr {
			if written > 0 {
				sb.WriteString(", ")
			}
			rebased := rebasePlaceholders(e.SQL, ac.offset())
			sb.WriteString(rebased)
			args = append(args, e.Args...)
			ac.n += len(e.Args)
			written++
		}
	}

	// FROM
	if s.fromSubquery != nil {
		subSQL, subArgs := buildSelectPostgres(s.fromSubquery)
		sb.WriteString(" FROM (")
		rebased := rebasePlaceholders(subSQL, ac.offset())
		sb.WriteString(rebased)
		sb.WriteString(") ")
		sb.WriteString(s.fromSubAlias)
		args = append(args, subArgs...)
		ac.n += len(subArgs)
	} else if s.from != "" {
		sb.WriteString(" FROM ")
		sb.WriteString(s.from)
	}

	// JOINs
	for _, j := range s.joins {
		sb.WriteByte(' ')
		sb.WriteString(j.kind)
		sb.WriteByte(' ')
		if j.subquery != nil {
			subSQL, subArgs := buildSelectPostgres(j.subquery)
			sb.WriteByte('(')
			rebased := rebasePlaceholders(subSQL, ac.offset())
			sb.WriteString(rebased)
			sb.WriteString(") ")
			sb.WriteString(j.subAlias)
			args = append(args, subArgs...)
			ac.n += len(subArgs)
		} else {
			sb.WriteString(j.table)
		}
		if j.on != "" {
			sb.WriteString(" ON ")
			rebased := rebasePlaceholders(j.on, ac.offset())
			sb.WriteString(rebased)
			args = append(args, j.args...)
			ac.n += len(j.args)
		}
	}

	// WHERE
	whereArgs := writeWhereClause(&sb, s.conditions, ac)
	args = append(args, whereArgs...)

	// GROUP BY
	if len(s.groupBy) > 0 || len(s.groupByExpr) > 0 {
		sb.WriteString(" GROUP BY ")
		writeJoined(&sb, s.groupBy, ", ")
		for i, e := range s.groupByExpr {
			if len(s.groupBy) > 0 || i > 0 {
				sb.WriteString(", ")
			}
			rebased := rebasePlaceholders(e.SQL, ac.offset())
			sb.WriteString(rebased)
			args = append(args, e.Args...)
			ac.n += len(e.Args)
		}
	}

	// HAVING
	if len(s.having) > 0 {
		sb.WriteString(" HAVING ")
		for i := range s.having {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			rebased := rebasePlaceholders(s.having[i].sql, ac.offset())
			sb.WriteString(rebased)
			args = append(args, s.having[i].args...)
			ac.n += len(s.having[i].args)
		}
	}

	// Set operations
	for _, op := range s.setOps {
		sb.WriteByte(' ')
		sb.WriteString(setOpKeyword(op.kind))
		sb.WriteByte(' ')
		rebased := rebasePlaceholders(op.query.SQL, ac.offset())
		sb.WriteString(rebased)
		args = append(args, op.query.Args...)
		ac.n += len(op.query.Args)
	}

	// ORDER BY
	if len(s.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		writeJoined(&sb, s.orderBy, ", ")
	}

	// LIMIT
	if s.hasLimit {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(s.limit))
	}

	// OFFSET
	if s.hasOffset {
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.FormatInt(s.offsetVal, 10))
	}

	// Locking
	if s.lockMode != "" {
		sb.WriteByte(' ')
		sb.WriteString(s.lockMode)
		if s.skipLocked {
			sb.WriteString(" SKIP LOCKED")
		}
		if s.noWait {
			sb.WriteString(" NOWAIT")
		}
	}

	return convertPlaceholders(sb.String(), s.dialect), args
}

// MustBuild calls Build and panics if the builder is in an invalid state.
func (s *SelectBuilder) MustBuild() (string, []any) {
	return s.Build()
}

// Query builds and returns a Query struct.
func (s *SelectBuilder) Query() Query {
	sql, args := s.Build()
	return Query{SQL: sql, Args: args}
}

// String returns the SQL string only, for debugging.
func (s *SelectBuilder) String() string {
	sql, _ := s.Build()
	return sql
}
