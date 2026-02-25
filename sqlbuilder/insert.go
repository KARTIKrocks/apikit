package sqlbuilder

import (
	"slices"
	"strconv"
	"strings"
)

// InsertBuilder builds INSERT queries.
type InsertBuilder struct {
	table      string
	columns    []string
	rows       [][]any
	fromSelect *SelectBuilder
	onConflict string
	returning  []string
	ctes       []cte
}

// Insert creates a new InsertBuilder for the given table.
//
//	sqlbuilder.Insert("users").Columns("name", "email").Values("Alice", "alice@example.com")
func Insert(table string) *InsertBuilder {
	return &InsertBuilder{table: table}
}

// Columns sets the column names for the INSERT.
func (b *InsertBuilder) Columns(cols ...string) *InsertBuilder {
	b.columns = cols
	return b
}

// Values adds a row of values. Can be called multiple times for batch inserts.
func (b *InsertBuilder) Values(vals ...any) *InsertBuilder {
	b.rows = append(b.rows, vals)
	return b
}

// ValueMap adds a row from a map. Keys are sorted for deterministic output.
// If columns are already set (via Columns or a prior ValueMap), values are
// extracted in column order to ensure correct alignment.
func (b *InsertBuilder) ValueMap(m map[string]any) *InsertBuilder {
	if len(b.columns) == 0 {
		b.columns = sortedKeys(m)
	}
	row := make([]any, len(b.columns))
	for i, col := range b.columns {
		row[i] = m[col]
	}
	b.rows = append(b.rows, row)
	return b
}

// BatchValues adds multiple rows at once.
func (b *InsertBuilder) BatchValues(rows [][]any) *InsertBuilder {
	b.rows = append(b.rows, rows...)
	return b
}

// FromSelect sets the INSERT to use a SELECT as the data source.
//
//	Insert("archive").Columns("id", "name").FromSelect(
//	    Select("id", "name").From("users").Where("archived = $1", true),
//	)
func (b *InsertBuilder) FromSelect(sel *SelectBuilder) *InsertBuilder {
	b.fromSelect = sel
	return b
}

// OnConflictDoNothing adds ON CONFLICT DO NOTHING.
// target is optional conflict target columns (e.g., "id" or "email").
func (b *InsertBuilder) OnConflictDoNothing(target ...string) *InsertBuilder {
	if len(target) > 0 {
		b.onConflict = "ON CONFLICT (" + strings.Join(target, ", ") + ") DO NOTHING"
	} else {
		b.onConflict = "ON CONFLICT DO NOTHING"
	}
	return b
}

// OnConflictUpdate adds ON CONFLICT ... DO UPDATE SET for upsert.
// target specifies the conflict columns, updates maps column names to values.
func (b *InsertBuilder) OnConflictUpdate(target []string, updates map[string]any) *InsertBuilder {
	keys := sortedKeys(updates)
	var sb strings.Builder
	sb.Grow(64)
	sb.WriteString("ON CONFLICT (")
	writeJoined(&sb, target, ", ")
	sb.WriteString(") DO UPDATE SET ")

	var setArgs []any
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(k)
		sb.WriteString(" = $")
		sb.WriteString(strconv.Itoa(i + 1))
		setArgs = append(setArgs, updates[k])
	}
	b.onConflict = sb.String()
	// Store args in the last row position to be picked up during build
	b.rows = append(b.rows, setArgs)
	return b
}

// OnConflictUpdateExpr adds ON CONFLICT ... DO UPDATE SET with expressions.
func (b *InsertBuilder) OnConflictUpdateExpr(target []string, updates map[string]Expr) *InsertBuilder {
	keys := sortedKeys(updates)
	var sb strings.Builder
	sb.Grow(64)
	sb.WriteString("ON CONFLICT (")
	writeJoined(&sb, target, ", ")
	sb.WriteString(") DO UPDATE SET ")

	var setArgs []any
	n := 1
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		expr := updates[k]
		sb.WriteString(k)
		sb.WriteString(" = ")
		rebased := rebasePlaceholders(expr.SQL, n-1)
		sb.WriteString(rebased)
		setArgs = append(setArgs, expr.Args...)
		n += len(expr.Args)
	}
	b.onConflict = sb.String()
	// Always append conflict args row (even if empty) so Build() can split correctly
	b.rows = append(b.rows, setArgs)
	return b
}

// Returning adds a RETURNING clause.
func (b *InsertBuilder) Returning(cols ...string) *InsertBuilder {
	b.returning = cols
	return b
}

// With adds a CTE (Common Table Expression).
func (b *InsertBuilder) With(name string, q Query) *InsertBuilder {
	b.ctes = append(b.ctes, cte{name: name, query: q})
	return b
}

// When conditionally applies a function to the builder.
func (b *InsertBuilder) When(cond bool, fn func(*InsertBuilder)) *InsertBuilder {
	if cond {
		fn(b)
	}
	return b
}

// Clone creates a deep copy of the builder.
func (b *InsertBuilder) Clone() *InsertBuilder {
	c := *b
	c.columns = slices.Clone(b.columns)
	c.rows = make([][]any, len(b.rows))
	for i, row := range b.rows {
		c.rows[i] = slices.Clone(row)
	}
	c.returning = slices.Clone(b.returning)
	c.ctes = slices.Clone(b.ctes)
	if b.fromSelect != nil {
		c.fromSelect = b.fromSelect.Clone()
	}
	return &c
}

// writeValuesRows writes "(val, val), (val, val)" placeholders directly into sb.
func writeValuesRows(sb *strings.Builder, rows [][]any, ac *argCounter, args *[]any) {
	for i, row := range rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('(')
		for j := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			ac.writePlaceholder(sb)
		}
		sb.WriteByte(')')
		*args = append(*args, row...)
	}
}

// Build assembles the SQL string and arguments.
func (b *InsertBuilder) Build() (string, []any) {
	var sb strings.Builder
	sb.Grow(256)
	args := make([]any, 0, 8)
	ac := &argCounter{}

	// CTEs
	cteArgs := writeCTEs(&sb, b.ctes, ac)
	args = append(args, cteArgs...)

	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.table)

	if len(b.columns) > 0 {
		sb.WriteString(" (")
		writeJoined(&sb, b.columns, ", ")
		sb.WriteByte(')')
	}

	if b.fromSelect != nil {
		sb.WriteByte(' ')
		selSQL, selArgs := b.fromSelect.Build()
		rebased := rebasePlaceholders(selSQL, ac.offset())
		sb.WriteString(rebased)
		args = append(args, selArgs...)
		ac.n += len(selArgs)
	} else if b.onConflict != "" && strings.Contains(b.onConflict, "DO UPDATE") {
		// For upsert: last row in b.rows is the conflict set args
		dataRows := b.rows[:len(b.rows)-1]
		conflictArgs := b.rows[len(b.rows)-1]

		sb.WriteString(" VALUES ")
		writeValuesRows(&sb, dataRows, ac, &args)

		// ON CONFLICT ... DO UPDATE SET — rebase placeholders
		sb.WriteByte(' ')
		rebased := rebasePlaceholders(b.onConflict, ac.offset())
		sb.WriteString(rebased)
		args = append(args, conflictArgs...)
		ac.n += len(conflictArgs)
	} else {
		sb.WriteString(" VALUES ")
		writeValuesRows(&sb, b.rows, ac, &args)

		if b.onConflict != "" {
			sb.WriteByte(' ')
			sb.WriteString(b.onConflict)
		}
	}

	writeReturning(&sb, b.returning)

	return sb.String(), args
}

// MustBuild calls Build and panics if the builder is in an invalid state.
func (b *InsertBuilder) MustBuild() (string, []any) {
	return b.Build()
}

// Query builds and returns a Query struct.
func (b *InsertBuilder) Query() Query {
	sql, args := b.Build()
	return Query{SQL: sql, Args: args}
}

// String returns the SQL string only, for debugging.
func (b *InsertBuilder) String() string {
	sql, _ := b.Build()
	return sql
}
