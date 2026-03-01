package sqlbuilder

import "strings"

// Expr represents a raw SQL expression with optional arguments.
// Use it where a plain string column would not suffice, such as
// function calls or computed expressions.
type Expr struct {
	SQL  string
	Args []any
}

// Raw creates an Expr with no arguments, for use with SQL literals
// like NOW(), DEFAULT, or column references.
//
//	sqlbuilder.Raw("NOW()")
//	sqlbuilder.Raw("DEFAULT")
func Raw(sql string) Expr {
	return Expr{SQL: sql}
}

// RawExpr creates an Expr with arguments, for parameterized expressions.
//
//	sqlbuilder.RawExpr("COUNT(*) FILTER (WHERE status = $1)", "active")
//	sqlbuilder.RawExpr("COALESCE($1, $2)", fallback1, fallback2)
func RawExpr(sql string, args ...any) Expr {
	return Expr{SQL: sql, Args: args}
}

// Over appends an OVER clause from a WindowBuilder to the expression.
//
//	RowNumber().Over(Window().PartitionBy("dept").OrderBy("salary DESC"))
func (e Expr) Over(w *WindowBuilder) Expr {
	return Expr{SQL: e.SQL + " OVER " + w.Build(), Args: e.Args}
}

// OverRaw appends an OVER clause with a raw string to the expression.
//
//	Count("*").OverRaw("ORDER BY created_at")
func (e Expr) OverRaw(clause string) Expr {
	return Expr{SQL: e.SQL + " OVER (" + clause + ")", Args: e.Args}
}

// As appends an alias to the expression.
//
//	Count("*").As("total") → Expr{SQL: "COUNT(*) AS total"}
func (e Expr) As(alias string) Expr {
	return Expr{SQL: e.SQL + " AS " + alias, Args: e.Args}
}

// Count creates a COUNT(col) expression.
func Count(col string) Expr {
	return Expr{SQL: "COUNT(" + col + ")"}
}

// CountDistinct creates a COUNT(DISTINCT col) expression.
func CountDistinct(col string) Expr {
	return Expr{SQL: "COUNT(DISTINCT " + col + ")"}
}

// Sum creates a SUM(col) expression.
func Sum(col string) Expr {
	return Expr{SQL: "SUM(" + col + ")"}
}

// Avg creates an AVG(col) expression.
func Avg(col string) Expr {
	return Expr{SQL: "AVG(" + col + ")"}
}

// Min creates a MIN(col) expression.
func Min(col string) Expr {
	return Expr{SQL: "MIN(" + col + ")"}
}

// Max creates a MAX(col) expression.
func Max(col string) Expr {
	return Expr{SQL: "MAX(" + col + ")"}
}

// Coalesce creates a COALESCE(a, b, c) expression with literal arguments.
func Coalesce(exprs ...string) Expr {
	return Expr{SQL: "COALESCE(" + strings.Join(exprs, ", ") + ")"}
}

// CoalesceExpr creates a COALESCE expression from parameterized Exprs.
// Placeholders in each Expr are rebased so they chain correctly.
func CoalesceExpr(exprs ...Expr) Expr {
	var sb strings.Builder
	sb.WriteString("COALESCE(")
	var allArgs []any
	offset := 0
	for i, e := range exprs {
		if i > 0 {
			sb.WriteString(", ")
		}
		rebased := rebasePlaceholders(e.SQL, offset)
		sb.WriteString(rebased)
		allArgs = append(allArgs, e.Args...)
		offset += len(e.Args)
	}
	sb.WriteByte(')')
	return Expr{SQL: sb.String(), Args: allArgs}
}

// NullIf creates a NULLIF(a, b) expression with literal arguments.
func NullIf(expr1, expr2 string) Expr {
	return Expr{SQL: "NULLIF(" + expr1 + ", " + expr2 + ")"}
}

// NullIfExpr creates a NULLIF expression from parameterized Exprs.
func NullIfExpr(expr1, expr2 Expr) Expr {
	var sb strings.Builder
	sb.WriteString("NULLIF(")
	sb.WriteString(expr1.SQL)
	sb.WriteString(", ")
	rebased := rebasePlaceholders(expr2.SQL, len(expr1.Args))
	sb.WriteString(rebased)
	sb.WriteByte(')')
	var allArgs []any
	allArgs = append(allArgs, expr1.Args...)
	allArgs = append(allArgs, expr2.Args...)
	return Expr{SQL: sb.String(), Args: allArgs}
}

// Cast creates a CAST(expr AS type) expression with a literal expression.
func Cast(expr, typeName string) Expr {
	return Expr{SQL: "CAST(" + expr + " AS " + typeName + ")"}
}

// CastExpr creates a CAST expression from a parameterized Expr.
func CastExpr(expr Expr, typeName string) Expr {
	return Expr{SQL: "CAST(" + expr.SQL + " AS " + typeName + ")", Args: expr.Args}
}
