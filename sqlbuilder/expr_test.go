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

func TestCoalesce(t *testing.T) {
	e := Coalesce("name", "'unknown'")
	if e.SQL != "COALESCE(name, 'unknown')" {
		t.Errorf("expected COALESCE(name, 'unknown'), got %s", e.SQL)
	}
}

func TestCoalesceExpr(t *testing.T) {
	e := CoalesceExpr(
		RawExpr("$1", "first"),
		RawExpr("$1", "second"),
	)
	if e.SQL != "COALESCE($1, $2)" {
		t.Errorf("expected COALESCE($1, $2), got %s", e.SQL)
	}
	expectArgs(t, []any{"first", "second"}, e.Args)
}

func TestCoalesceExprInSelect(t *testing.T) {
	sql, args := SelectExpr(
		CoalesceExpr(RawExpr("$1", "val"), Raw("0")).As("result"),
	).From("t").Build()
	expectSQL(t, "SELECT COALESCE($1, 0) AS result FROM t", sql)
	expectArgs(t, []any{"val"}, args)
}

func TestNullIf(t *testing.T) {
	e := NullIf("price", "0")
	if e.SQL != "NULLIF(price, 0)" {
		t.Errorf("expected NULLIF(price, 0), got %s", e.SQL)
	}
}

func TestNullIfExpr(t *testing.T) {
	e := NullIfExpr(RawExpr("$1", "a"), RawExpr("$1", "b"))
	if e.SQL != "NULLIF($1, $2)" {
		t.Errorf("expected NULLIF($1, $2), got %s", e.SQL)
	}
	expectArgs(t, []any{"a", "b"}, e.Args)
}

func TestCast(t *testing.T) {
	e := Cast("price", "INTEGER")
	if e.SQL != "CAST(price AS INTEGER)" {
		t.Errorf("expected CAST(price AS INTEGER), got %s", e.SQL)
	}
}

func TestCastExpr(t *testing.T) {
	e := CastExpr(RawExpr("$1", "123"), "INTEGER")
	if e.SQL != "CAST($1 AS INTEGER)" {
		t.Errorf("expected CAST($1 AS INTEGER), got %s", e.SQL)
	}
	expectArgs(t, []any{"123"}, e.Args)
}
