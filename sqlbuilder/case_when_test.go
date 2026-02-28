package sqlbuilder

import "testing"

func TestCaseWhenSimple(t *testing.T) {
	e := Case().
		When("status = 'active'", "'yes'").
		When("status = 'inactive'", "'no'").
		Else("'unknown'").
		End()
	expectSQL(t, "CASE WHEN status = 'active' THEN 'yes' WHEN status = 'inactive' THEN 'no' ELSE 'unknown' END", e.SQL)
	expectArgs(t, nil, e.Args)
}

func TestCaseWhenWithOperand(t *testing.T) {
	e := Case("status").
		When("'active'", "'yes'").
		When("'inactive'", "'no'").
		End()
	expectSQL(t, "CASE status WHEN 'active' THEN 'yes' WHEN 'inactive' THEN 'no' END", e.SQL)
}

func TestCaseWhenExpr(t *testing.T) {
	e := Case().
		WhenExpr("amount > $1", []any{100}, "$1", []any{"high"}).
		WhenExpr("amount > $1", []any{50}, "$1", []any{"medium"}).
		ElseExpr("$1", "low").
		End()
	expectSQL(t, "CASE WHEN amount > $1 THEN $2 WHEN amount > $3 THEN $4 ELSE $5 END", e.SQL)
	expectArgs(t, []any{100, "high", 50, "medium", "low"}, e.Args)
}

func TestCaseWhenInSelect(t *testing.T) {
	caseExpr := Case().
		When("role = 'admin'", "'full'").
		Else("'limited'").
		End().As("access")
	sql, _ := SelectExpr(Raw("name"), caseExpr).From("users").Build()
	expectSQL(t, "SELECT name, CASE WHEN role = 'admin' THEN 'full' ELSE 'limited' END AS access FROM users", sql)
}

func TestCaseWhenExprInSelect(t *testing.T) {
	caseExpr := Case().
		WhenExpr("score >= $1", []any{90}, "'A'", nil).
		WhenExpr("score >= $1", []any{80}, "'B'", nil).
		ElseExpr("'C'").
		End().As("grade")
	sql, args := SelectExpr(Raw("name"), caseExpr).From("students").Build()
	expectSQL(t, "SELECT name, CASE WHEN score >= $1 THEN 'A' WHEN score >= $2 THEN 'B' ELSE 'C' END AS grade FROM students", sql)
	expectArgs(t, []any{90, 80}, args)
}

func TestCaseWhenNoElse(t *testing.T) {
	e := Case().When("x = 1", "'one'").End()
	expectSQL(t, "CASE WHEN x = 1 THEN 'one' END", e.SQL)
}
