package sqlbuilder

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
