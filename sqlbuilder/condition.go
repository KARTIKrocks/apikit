package sqlbuilder

import (
	"strconv"
	"strings"
)

// condition represents a single WHERE or HAVING clause fragment.
type condition struct {
	sql  string
	args []any
}

// Or creates a condition that is used inside WhereOr to combine
// multiple conditions with OR logic.
//
//	Select("id").From("users").
//	    WhereOr(
//	        sqlbuilder.Or("status = $1", "active"),
//	        sqlbuilder.Or("role = $1", "admin"),
//	    ).Build()
//	// WHERE (status = $1 OR role = $2)
func Or(sql string, args ...any) condition {
	return condition{sql: sql, args: args}
}

// buildWhereIn generates "col IN ($1, $2, ...)" with the given values.
func buildWhereIn(col string, vals []any) condition {
	if len(vals) == 0 {
		return condition{sql: "1=0"}
	}
	// Write directly: "col IN ($1, $2, ...)" without intermediate []string.
	var sb strings.Builder
	sb.Grow(len(col) + 5 + len(vals)*4) // "col IN ($N, ...)"
	sb.WriteString(col)
	sb.WriteString(" IN (")
	for i := range vals {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(i + 1))
	}
	sb.WriteByte(')')
	return condition{sql: sb.String(), args: vals}
}

// buildWhereNotIn generates "col NOT IN ($1, $2, ...)" with the given values.
func buildWhereNotIn(col string, vals []any) condition {
	if len(vals) == 0 {
		return condition{sql: "1=1"}
	}
	var sb strings.Builder
	sb.Grow(len(col) + 9 + len(vals)*4)
	sb.WriteString(col)
	sb.WriteString(" NOT IN (")
	for i := range vals {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(i + 1))
	}
	sb.WriteByte(')')
	return condition{sql: sb.String(), args: vals}
}

// buildWhereBetween generates "col BETWEEN $1 AND $2".
func buildWhereBetween(col string, low, high any) condition {
	return condition{
		sql:  col + " BETWEEN $1 AND $2",
		args: []any{low, high},
	}
}

// buildOrGroup builds a group of OR conditions into a single condition
// wrapped in parentheses: (cond1 OR cond2 OR ...).
func buildOrGroup(conditions []condition) condition {
	if len(conditions) == 0 {
		return condition{sql: "1=0"}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}

	// Pre-count total args.
	totalArgs := 0
	for i := range conditions {
		totalArgs += len(conditions[i].args)
	}
	allArgs := make([]any, 0, totalArgs)

	var sb strings.Builder
	sb.Grow(64)
	sb.WriteByte('(')
	offset := 0
	for i := range conditions {
		if i > 0 {
			sb.WriteString(" OR ")
		}
		rebased := rebasePlaceholders(conditions[i].sql, offset)
		sb.WriteString(rebased)
		allArgs = append(allArgs, conditions[i].args...)
		offset += len(conditions[i].args)
	}
	sb.WriteByte(')')
	return condition{
		sql:  sb.String(),
		args: allArgs,
	}
}
