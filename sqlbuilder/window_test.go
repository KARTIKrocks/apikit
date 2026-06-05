package sqlbuilder

import "testing"

func TestWindowPartitionByOrderBy(t *testing.T) {
	w := Window().PartitionBy("dept").OrderBy("salary DESC")
	got := w.Build()
	expected := "(PARTITION BY dept ORDER BY salary DESC)"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestWindowOrderByOnly(t *testing.T) {
	w := Window().OrderBy("created_at ASC")
	got := w.Build()
	expected := "(ORDER BY created_at ASC)"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestWindowPartitionByOnly(t *testing.T) {
	w := Window().PartitionBy("dept", "region")
	got := w.Build()
	expected := "(PARTITION BY dept, region)"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestWindowEmpty(t *testing.T) {
	w := Window()
	got := w.Build()
	if got != "()" {
		t.Errorf("expected (), got %s", got)
	}
}

func TestRowNumber(t *testing.T) {
	e := RowNumber()
	if e.SQL != "ROW_NUMBER()" {
		t.Errorf("expected ROW_NUMBER(), got %s", e.SQL)
	}
}

func TestRank(t *testing.T) {
	e := Rank()
	if e.SQL != "RANK()" {
		t.Errorf("expected RANK(), got %s", e.SQL)
	}
}

func TestDenseRank(t *testing.T) {
	e := DenseRank()
	if e.SQL != "DENSE_RANK()" {
		t.Errorf("expected DENSE_RANK(), got %s", e.SQL)
	}
}

func TestNtile(t *testing.T) {
	e := Ntile(4)
	if e.SQL != "NTILE(4)" {
		t.Errorf("expected NTILE(4), got %s", e.SQL)
	}
}

func TestLag(t *testing.T) {
	e := Lag("salary")
	if e.SQL != "LAG(salary)" {
		t.Errorf("expected LAG(salary), got %s", e.SQL)
	}
}

func TestLead(t *testing.T) {
	e := Lead("salary")
	if e.SQL != "LEAD(salary)" {
		t.Errorf("expected LEAD(salary), got %s", e.SQL)
	}
}

func TestExprOver(t *testing.T) {
	e := RowNumber().Over(Window().PartitionBy("dept").OrderBy("salary DESC"))
	if e.SQL != "ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC)" {
		t.Errorf("expected ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC), got %s", e.SQL)
	}
}

func TestExprOverRaw(t *testing.T) {
	e := Count("*").OverRaw("ORDER BY created_at")
	if e.SQL != "COUNT(*) OVER (ORDER BY created_at)" {
		t.Errorf("expected COUNT(*) OVER (ORDER BY created_at), got %s", e.SQL)
	}
}

func TestWindowFunctionInSelect(t *testing.T) {
	sql, _ := SelectExpr(
		Raw("name"),
		Raw("salary"),
		RowNumber().Over(Window().PartitionBy("dept").OrderBy("salary DESC")).As("rn"),
	).From("employees").Build()
	expectSQL(t, "SELECT name, salary, ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC) AS rn FROM employees", sql)
}

func TestRankInSelect(t *testing.T) {
	sql, _ := SelectExpr(
		Raw("name"),
		Rank().Over(Window().OrderBy("score DESC")).As("rank"),
	).From("players").Build()
	expectSQL(t, "SELECT name, RANK() OVER (ORDER BY score DESC) AS rank FROM players", sql)
}

func TestWindowWithWhere(t *testing.T) {
	sql, args := SelectExpr(
		Raw("name"),
		DenseRank().Over(Window().PartitionBy("dept").OrderBy("salary DESC")).As("dr"),
	).From("employees").Where("active = $1", true).Build()
	expectSQL(t, "SELECT name, DENSE_RANK() OVER (PARTITION BY dept ORDER BY salary DESC) AS dr FROM employees WHERE active = $1", sql)
	expectArgs(t, []any{true}, args)
}
