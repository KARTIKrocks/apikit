package sqlbuilder

import (
	"testing"
)

func TestCount(t *testing.T) {
	e := Count("*")
	if e.SQL != "COUNT(*)" {
		t.Errorf("expected COUNT(*), got %s", e.SQL)
	}
}

func TestCountDistinct(t *testing.T) {
	e := CountDistinct("user_id")
	if e.SQL != "COUNT(DISTINCT user_id)" {
		t.Errorf("expected COUNT(DISTINCT user_id), got %s", e.SQL)
	}
}

func TestSum(t *testing.T) {
	e := Sum("amount")
	if e.SQL != "SUM(amount)" {
		t.Errorf("expected SUM(amount), got %s", e.SQL)
	}
}

func TestAvg(t *testing.T) {
	e := Avg("score")
	if e.SQL != "AVG(score)" {
		t.Errorf("expected AVG(score), got %s", e.SQL)
	}
}

func TestMin(t *testing.T) {
	e := Min("price")
	if e.SQL != "MIN(price)" {
		t.Errorf("expected MIN(price), got %s", e.SQL)
	}
}

func TestMax(t *testing.T) {
	e := Max("price")
	if e.SQL != "MAX(price)" {
		t.Errorf("expected MAX(price), got %s", e.SQL)
	}
}

func TestExprAs(t *testing.T) {
	e := Count("*").As("total")
	if e.SQL != "COUNT(*) AS total" {
		t.Errorf("expected COUNT(*) AS total, got %s", e.SQL)
	}
}

func TestExprAsPreservesArgs(t *testing.T) {
	e := RawExpr("COALESCE($1, 0)", "default").As("val")
	if e.SQL != "COALESCE($1, 0) AS val" {
		t.Errorf("expected COALESCE($1, 0) AS val, got %s", e.SQL)
	}
	expectArgs(t, []any{"default"}, e.Args)
}

func TestAggregateInSelect(t *testing.T) {
	sql, _ := SelectExpr(Count("*").As("total")).From("users").Build()
	expectSQL(t, "SELECT COUNT(*) AS total FROM users", sql)
}

func TestMultipleAggregatesInSelect(t *testing.T) {
	sql, _ := SelectExpr(
		Min("price").As("min_price"),
		Max("price").As("max_price"),
		Avg("price").As("avg_price"),
	).From("products").Build()
	expectSQL(t, "SELECT MIN(price) AS min_price, MAX(price) AS max_price, AVG(price) AS avg_price FROM products", sql)
}
