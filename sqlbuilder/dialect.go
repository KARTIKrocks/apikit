package sqlbuilder

import "strings"

// Dialect represents the target SQL database dialect.
// The zero value is Postgres.
type Dialect int

const (
	// Postgres uses $1, $2, ... numbered placeholders.
	Postgres Dialect = iota
	// MySQL uses ? positional placeholders.
	MySQL
	// SQLite uses ? positional placeholders.
	SQLite
)

// convertPlaceholders rewrites $N placeholders to ? for non-Postgres dialects.
// For Postgres it returns the input unchanged (zero allocation).
func convertPlaceholders(sql string, d Dialect) string {
	if d == Postgres {
		return sql
	}

	// Fast path: if there's no '$' at all, return unchanged (zero allocation).
	if !strings.Contains(sql, "$") {
		return sql
	}

	var sb strings.Builder
	sb.Grow(len(sql))
	i := 0
	for i < len(sql) {
		if sql[i] == '$' && i+1 < len(sql) && sql[i+1] >= '1' && sql[i+1] <= '9' {
			sb.WriteByte('?')
			i += 2 // skip '$' and first digit
			for i < len(sql) && sql[i] >= '0' && sql[i] <= '9' {
				i++ // skip remaining digits
			}
		} else {
			sb.WriteByte(sql[i])
			i++
		}
	}
	return sb.String()
}

// buildSelectPostgres builds a SelectBuilder using Postgres placeholders,
// regardless of the builder's dialect. This is used internally when embedding
// subqueries so that rebasePlaceholders can adjust $N numbering correctly.
// The outermost Build() handles the final dialect conversion.
//
// This function is safe for concurrent use: it does not mutate sub.
func buildSelectPostgres(sub *SelectBuilder) (string, []any) {
	if sub.dialect == Postgres {
		return sub.Build()
	}
	// Build a shallow copy with Postgres dialect to avoid mutating the original.
	cp := *sub
	cp.dialect = Postgres
	return cp.Build()
}

// SelectWith creates a new SelectBuilder with the given dialect and columns.
func SelectWith(d Dialect, columns ...string) *SelectBuilder {
	return &SelectBuilder{dialect: d, columns: columns}
}

// InsertWith creates a new InsertBuilder with the given dialect and table.
func InsertWith(d Dialect, table string) *InsertBuilder {
	return &InsertBuilder{dialect: d, table: table}
}

// UpdateWith creates a new UpdateBuilder with the given dialect and table.
func UpdateWith(d Dialect, table string) *UpdateBuilder {
	return &UpdateBuilder{dialect: d, table: table}
}

// DeleteWith creates a new DeleteBuilder with the given dialect and table.
func DeleteWith(d Dialect, table string) *DeleteBuilder {
	return &DeleteBuilder{dialect: d, table: table}
}
