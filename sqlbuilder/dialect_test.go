package sqlbuilder

import (
	"testing"
)

func TestConvertPlaceholders(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		dialect Dialect
		want    string
	}{
		{"postgres passthrough", "SELECT * FROM users WHERE id = $1", Postgres, "SELECT * FROM users WHERE id = $1"},
		{"mysql single", "SELECT * FROM users WHERE id = $1", MySQL, "SELECT * FROM users WHERE id = ?"},
		{"sqlite single", "SELECT * FROM users WHERE id = $1", SQLite, "SELECT * FROM users WHERE id = ?"},
		{"mysql multi", "SELECT * FROM users WHERE id = $1 AND name = $2", MySQL, "SELECT * FROM users WHERE id = ? AND name = ?"},
		{"high numbers", "WHERE a = $10 AND b = $99", MySQL, "WHERE a = ? AND b = ?"},
		{"no placeholders", "SELECT 1", MySQL, "SELECT 1"},
		{"dollar not placeholder", "SELECT '$$' FROM t", MySQL, "SELECT '$$' FROM t"},
		{"adjacent text", "WHERE x=$1AND y=$2", MySQL, "WHERE x=?AND y=?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPlaceholders(tt.sql, tt.dialect)
			if got != tt.want {
				t.Errorf("convertPlaceholders(%q, %d) = %q, want %q", tt.sql, tt.dialect, got, tt.want)
			}
		})
	}
}

func TestRebasePlaceholdersSkipsDollarZero(t *testing.T) {
	// $0 is not a valid PostgreSQL placeholder — it should be left untouched.
	slowResult := rebasePlaceholders("$0 AND $1", 5)
	if slowResult != "$0 AND $6" {
		t.Errorf("multi-placeholder: got %q, want %q", slowResult, "$0 AND $6")
	}

	fastResult := rebasePlaceholders("x = $0", 5)
	if fastResult != "x = $0" {
		t.Errorf("single-placeholder: got %q, want %q", fastResult, "x = $0")
	}
}

func TestSelectMySQL(t *testing.T) {
	sql, args := Select("id", "name").
		From("users").
		Where("active = $1", true).
		Where("age > $1", 18).
		SetDialect(MySQL).
		Build()

	want := "SELECT id, name FROM users WHERE active = ? AND age > ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != true || args[1] != 18 {
		t.Errorf("args = %v, want [true 18]", args)
	}
}

func TestSelectSQLite(t *testing.T) {
	sql, _ := Select("id").
		From("users").
		WhereEq("status", "active").
		SetDialect(SQLite).
		Build()

	want := "SELECT id FROM users WHERE status = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestInsertMySQL(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		SetDialect(MySQL).
		Build()

	want := "INSERT INTO users (name, email) VALUES (?, ?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 {
		t.Errorf("args = %v, want 2 args", args)
	}
}

func TestUpdateMySQL(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Bob").
		WhereEq("id", 1).
		SetDialect(MySQL).
		Build()

	want := "UPDATE users SET name = ? WHERE id = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 {
		t.Errorf("args = %v, want 2 args", args)
	}
}

func TestDeleteMySQL(t *testing.T) {
	sql, args := Delete("users").
		WhereEq("id", 1).
		SetDialect(MySQL).
		Build()

	want := "DELETE FROM users WHERE id = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 1 {
		t.Errorf("args = %v, want 1 arg", args)
	}
}

func TestSelectWithConstructor(t *testing.T) {
	sql, _ := SelectWith(MySQL, "id", "name").
		From("users").
		WhereEq("active", true).
		Build()

	want := "SELECT id, name FROM users WHERE active = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestInsertWithConstructor(t *testing.T) {
	sql, _ := InsertWith(MySQL, "users").
		Columns("name").
		Values("Alice").
		Build()

	want := "INSERT INTO users (name) VALUES (?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestUpdateWithConstructor(t *testing.T) {
	sql, _ := UpdateWith(MySQL, "users").
		Set("name", "Bob").
		WhereEq("id", 1).
		Build()

	want := "UPDATE users SET name = ? WHERE id = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestDeleteWithConstructor(t *testing.T) {
	sql, _ := DeleteWith(MySQL, "users").
		WhereEq("id", 1).
		Build()

	want := "DELETE FROM users WHERE id = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestDefaultDialectIsPostgres(t *testing.T) {
	sql, _ := Select("id").From("users").WhereEq("id", 1).Build()
	want := "SELECT id FROM users WHERE id = $1"
	if sql != want {
		t.Errorf("default dialect should be Postgres: got %q, want %q", sql, want)
	}
}

func TestClonePreservesDialect(t *testing.T) {
	base := Select("id").From("users").SetDialect(MySQL)
	clone := base.Clone()
	sql, _ := clone.WhereEq("id", 1).Build()
	want := "SELECT id FROM users WHERE id = ?"
	if sql != want {
		t.Errorf("clone should preserve dialect: got %q, want %q", sql, want)
	}
}

func TestMySQLWhereIn(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereIn("status", "a", "b", "c").
		SetDialect(MySQL).
		Build()

	want := "SELECT id FROM users WHERE status IN (?, ?, ?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 3 {
		t.Errorf("args = %v, want 3 args", args)
	}
}

func TestMySQLWhereBetween(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereBetween("age", 18, 65).
		SetDialect(MySQL).
		Build()

	want := "SELECT id FROM users WHERE age BETWEEN ? AND ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 {
		t.Errorf("args = %v, want 2 args", args)
	}
}

func TestMySQLBatchInsert(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "a@example.com").
		Values("Bob", "b@example.com").
		SetDialect(MySQL).
		Build()

	want := "INSERT INTO users (name, email) VALUES (?, ?), (?, ?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 4 {
		t.Errorf("args = %v, want 4 args", args)
	}
}

func TestMySQLIncrementDecrement(t *testing.T) {
	sql, args := Update("products").
		Increment("views", 1).
		Decrement("stock", 2).
		WhereEq("id", 42).
		SetDialect(MySQL).
		Build()

	want := "UPDATE products SET views = views + ?, stock = stock - ? WHERE id = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 3 {
		t.Errorf("args = %v, want 3 args", args)
	}
}

func TestMySQLSubqueryConversion(t *testing.T) {
	sub := Select("user_id").From("orders").Where("total > $1", 100)
	sql, args := Select("id", "name").From("users").
		WhereInSubquery("id", sub).
		SetDialect(MySQL).
		Build()

	want := "SELECT id, name FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > ?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 1 || args[0] != 100 {
		t.Errorf("args = %v, want [100]", args)
	}
}

func TestMySQLWithCTE(t *testing.T) {
	cteQuery := Select("id", "name").From("users").Where("active = $1", true).Query()
	sql, args := Select("id", "name").
		With("active_users", cteQuery).
		From("active_users").
		WhereEq("name", "Alice").
		SetDialect(MySQL).
		Build()

	want := "WITH active_users AS (SELECT id, name FROM users WHERE active = ?) SELECT id, name FROM active_users WHERE name = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 {
		t.Errorf("args = %v, want 2 args", args)
	}
}

func TestSubqueryWithDialectOnBoth(t *testing.T) {
	// Dialect set on sub should not corrupt the parent's placeholders.
	sub := Select("user_id").From("orders").Where("total > $1", 100).SetDialect(MySQL)
	sql, args := Select("id", "name").From("users").
		Where("active = $1", true).
		WhereInSubquery("id", sub).
		SetDialect(MySQL).
		Build()

	want := "SELECT id, name FROM users WHERE active = ? AND id IN (SELECT user_id FROM orders WHERE total > ?)"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != true || args[1] != 100 {
		t.Errorf("args = %v, want [true 100]", args)
	}
}

func TestUnionWithDialect(t *testing.T) {
	q1 := Select("id").From("users").Where("active = $1", true).SetDialect(MySQL)
	q2 := Select("id").From("admins").Where("level > $1", 5).SetDialect(MySQL)
	sql, args := q1.Union(q2).Build()

	want := "SELECT id FROM users WHERE active = ? UNION SELECT id FROM admins WHERE level > ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != true || args[1] != 5 {
		t.Errorf("args = %v, want [true 5]", args)
	}
}

func TestWithSelectDialectSafe(t *testing.T) {
	// WithSelect forces Postgres placeholders internally, so the CTE
	// rebasing works even when the sub has MySQL dialect set.
	sub := Select("id", "name").From("users").Where("active = $1", true).SetDialect(MySQL)
	sql, args := Select("id", "name").
		WithSelect("active_users", sub).
		From("active_users").
		WhereEq("name", "Alice").
		SetDialect(MySQL).
		Build()

	want := "WITH active_users AS (SELECT id, name FROM users WHERE active = ?) SELECT id, name FROM active_users WHERE name = ?"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != true || args[1] != "Alice" {
		t.Errorf("args = %v, want [true Alice]", args)
	}
}

func TestWithRecursiveSelectDialectSafe(t *testing.T) {
	base := Select("id", "parent_id").From("categories").Where("id = $1", 1)
	recursive := Select("c.id", "c.parent_id").From("categories c").
		Join("tree t", "t.id = c.parent_id")
	combined := base.Union(recursive).SetDialect(MySQL)

	sql, args := Select("id").
		WithRecursiveSelect("tree", combined).
		From("tree").
		SetDialect(MySQL).
		Build()

	want := "WITH RECURSIVE tree AS (SELECT id, parent_id FROM categories WHERE id = ? UNION SELECT c.id, c.parent_id FROM categories c JOIN tree t ON t.id = c.parent_id) SELECT id FROM tree"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("args = %v, want [1]", args)
	}
}
