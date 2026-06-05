package sqlbuilder

import (
	"strconv"
	"strings"
)

// WindowBuilder builds a window specification for OVER clauses.
type WindowBuilder struct {
	partitionBy []string
	orderBy     []string
}

// Window creates a new WindowBuilder.
func Window() *WindowBuilder {
	return &WindowBuilder{}
}

// PartitionBy sets the PARTITION BY columns.
func (w *WindowBuilder) PartitionBy(cols ...string) *WindowBuilder {
	w.partitionBy = append(w.partitionBy, cols...)
	return w
}

// OrderBy sets the ORDER BY clauses.
func (w *WindowBuilder) OrderBy(clauses ...string) *WindowBuilder {
	w.orderBy = append(w.orderBy, clauses...)
	return w
}

// Build renders the window specification as "(PARTITION BY ... ORDER BY ...)".
func (w *WindowBuilder) Build() string {
	var sb strings.Builder
	sb.WriteByte('(')
	if len(w.partitionBy) > 0 {
		sb.WriteString("PARTITION BY ")
		writeJoined(&sb, w.partitionBy, ", ")
	}
	if len(w.orderBy) > 0 {
		if len(w.partitionBy) > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString("ORDER BY ")
		writeJoined(&sb, w.orderBy, ", ")
	}
	sb.WriteByte(')')
	return sb.String()
}

// RowNumber creates a ROW_NUMBER() expression.
func RowNumber() Expr {
	return Expr{SQL: "ROW_NUMBER()"}
}

// Rank creates a RANK() expression.
func Rank() Expr {
	return Expr{SQL: "RANK()"}
}

// DenseRank creates a DENSE_RANK() expression.
func DenseRank() Expr {
	return Expr{SQL: "DENSE_RANK()"}
}

// Ntile creates an NTILE(n) expression.
func Ntile(n int) Expr {
	return Expr{SQL: "NTILE(" + strconv.Itoa(n) + ")"}
}

// Lag creates a LAG(col) expression.
func Lag(col string) Expr {
	return Expr{SQL: "LAG(" + col + ")"}
}

// Lead creates a LEAD(col) expression.
func Lead(col string) Expr {
	return Expr{SQL: "LEAD(" + col + ")"}
}
