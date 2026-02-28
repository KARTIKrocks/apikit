package sqlbuilder

import (
	"strconv"
	"strings"
)

// Query holds a built SQL statement and its arguments.
type Query struct {
	SQL  string
	Args []any
}

// argCounter tracks the current placeholder number for sequential generation.
type argCounter struct {
	n int
}

// offset returns the current total number of placeholders consumed.
func (a *argCounter) offset() int {
	return a.n
}

// writePlaceholder writes "$N" directly into the builder and increments the counter.
func (a *argCounter) writePlaceholder(sb *strings.Builder) {
	a.n++
	sb.WriteByte('$')
	sb.WriteString(strconv.Itoa(a.n))
}

// rebasePlaceholders rewrites $1, $2, ... in sql by adding offset to each number.
// If offset is 0, the string is returned unchanged.
func rebasePlaceholders(sql string, offset int) string {
	if offset == 0 {
		return sql
	}

	// Fast path: single placeholder at known position.
	// Most WHERE fragments look like "col = $1" or "col > $1".
	if idx := strings.IndexByte(sql, '$'); idx >= 0 && idx+1 < len(sql) && sql[idx+1] >= '1' && sql[idx+1] <= '9' {
		// Check if there's only one placeholder and it's $N.
		end := idx + 1
		for end < len(sql) && sql[end] >= '0' && sql[end] <= '9' {
			end++
		}
		// If the rest has no more $, use fast path.
		if !strings.ContainsRune(sql[end:], '$') {
			n, _ := strconv.Atoi(sql[idx+1 : end])
			return sql[:idx] + "$" + strconv.Itoa(n+offset) + sql[end:]
		}
	}

	var b strings.Builder
	b.Grow(len(sql) + 8)
	i := 0
	for i < len(sql) {
		if sql[i] == '$' && i+1 < len(sql) && sql[i+1] >= '1' && sql[i+1] <= '9' {
			j := i + 1
			for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
				j++
			}
			n, _ := strconv.Atoi(sql[i+1 : j])
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n + offset))
			i = j
		} else {
			b.WriteByte(sql[i])
			i++
		}
	}
	return b.String()
}

// writeWhereClause writes the WHERE clause from conditions into the builder.
func writeWhereClause(sb *strings.Builder, conditions []condition, ac *argCounter) []any {
	if len(conditions) == 0 {
		return nil
	}

	// Pre-allocate args slice by counting total args across conditions.
	totalArgs := 0
	for i := range conditions {
		totalArgs += len(conditions[i].args)
	}
	allArgs := make([]any, 0, totalArgs)

	sb.WriteString(" WHERE ")
	for i := range conditions {
		if i > 0 {
			sb.WriteString(" AND ")
		}
		rebased := rebasePlaceholders(conditions[i].sql, ac.offset())
		sb.WriteString(rebased)
		allArgs = append(allArgs, conditions[i].args...)
		ac.n += len(conditions[i].args)
	}
	return allArgs
}

// writeCTEs writes the WITH [RECURSIVE] clause directly into the caller's builder.
func writeCTEs(sb *strings.Builder, ctes []cte, ac *argCounter) []any {
	if len(ctes) == 0 {
		return nil
	}

	// Check for recursive upfront to write the correct keyword.
	recursive := false
	totalArgs := 0
	for i := range ctes {
		if ctes[i].recursive {
			recursive = true
		}
		totalArgs += len(ctes[i].query.Args)
	}
	allArgs := make([]any, 0, totalArgs)

	if recursive {
		sb.WriteString("WITH RECURSIVE ")
	} else {
		sb.WriteString("WITH ")
	}

	for i := range ctes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ctes[i].name)
		sb.WriteString(" AS (")
		rebased := rebasePlaceholders(ctes[i].query.SQL, ac.offset())
		sb.WriteString(rebased)
		sb.WriteByte(')')
		allArgs = append(allArgs, ctes[i].query.Args...)
		ac.n += len(ctes[i].query.Args)
	}
	sb.WriteByte(' ')
	return allArgs
}

// writeReturningWithExpr writes the RETURNING clause supporting both plain columns and Exprs.
func writeReturningWithExpr(sb *strings.Builder, cols []string, exprs []Expr, ac *argCounter, args *[]any) {
	if len(cols) == 0 && len(exprs) == 0 {
		return
	}
	sb.WriteString(" RETURNING ")
	written := 0
	for _, col := range cols {
		if written > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(col)
		written++
	}
	for _, e := range exprs {
		if written > 0 {
			sb.WriteString(", ")
		}
		rebased := rebasePlaceholders(e.SQL, ac.offset())
		sb.WriteString(rebased)
		*args = append(*args, e.Args...)
		ac.n += len(e.Args)
		written++
	}
}

// writeJoined writes string slice elements separated by sep directly into the builder,
// avoiding the intermediate allocation of strings.Join.
func writeJoined(sb *strings.Builder, items []string, sep string) {
	for i, item := range items {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(item)
	}
}

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

// sortStrings sorts a string slice in place using insertion sort (sufficient for small slices).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
