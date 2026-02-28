package sqlbuilder

import (
	"testing"
)

func TestInsertSingleRow(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2)", sql)
	expectArgs(t, []any{"Alice", "alice@example.com"}, args)
}

func TestInsertBatchValues(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Values("Bob", "bob@example.com").
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)", sql)
	expectArgs(t, []any{"Alice", "alice@example.com", "Bob", "bob@example.com"}, args)
}

func TestInsertBatchValuesMethod(t *testing.T) {
	rows := [][]any{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}
	sql, args := Insert("users").
		Columns("name", "email").
		BatchValues(rows).
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4), ($5, $6)", sql)
	expectArgs(t, []any{"Alice", "alice@example.com", "Bob", "bob@example.com", "Charlie", "charlie@example.com"}, args)
}

func TestInsertValueMap(t *testing.T) {
	sql, args := Insert("users").
		ValueMap(map[string]any{
			"email": "alice@example.com",
			"name":  "Alice",
		}).
		Build()
	// Keys sorted: email, name
	expectSQL(t, "INSERT INTO users (email, name) VALUES ($1, $2)", sql)
	expectArgs(t, []any{"alice@example.com", "Alice"}, args)
}

func TestInsertFromSelect(t *testing.T) {
	sel := Select("id", "name").From("users").Where("active = $1", true)
	sql, args := Insert("archive").
		Columns("id", "name").
		FromSelect(sel).
		Build()
	expectSQL(t, "INSERT INTO archive (id, name) SELECT id, name FROM users WHERE active = $1", sql)
	expectArgs(t, []any{true}, args)
}

func TestInsertOnConflictDoNothing(t *testing.T) {
	sql, args := Insert("users").
		Columns("email", "name").
		Values("alice@example.com", "Alice").
		OnConflictDoNothing("email").
		Build()
	expectSQL(t, "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING", sql)
	expectArgs(t, []any{"alice@example.com", "Alice"}, args)
}

func TestInsertOnConflictDoNothingNoTarget(t *testing.T) {
	sql, _ := Insert("users").
		Columns("email", "name").
		Values("alice@example.com", "Alice").
		OnConflictDoNothing().
		Build()
	expectSQL(t, "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT DO NOTHING", sql)
}

func TestInsertOnConflictUpdate(t *testing.T) {
	sql, args := Insert("users").
		Columns("email", "name").
		Values("alice@example.com", "Alice").
		OnConflictUpdate(
			[]string{"email"},
			map[string]any{"name": "Alice Updated"},
		).
		Build()
	expectSQL(t, "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = $3", sql)
	expectArgs(t, []any{"alice@example.com", "Alice", "Alice Updated"}, args)
}

func TestInsertReturning(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Returning("id", "created_at").
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at", sql)
	expectArgs(t, []any{"Alice", "alice@example.com"}, args)
}

func TestInsertWithCTE(t *testing.T) {
	cteQuery := Select("id").From("old_users").Where("active = $1", false).Query()
	sql, args := Insert("archive").
		With("to_archive", cteQuery).
		Columns("user_id").
		FromSelect(Select("id").From("to_archive")).
		Build()
	expectSQL(t, "WITH to_archive AS (SELECT id FROM old_users WHERE active = $1) INSERT INTO archive (user_id) SELECT id FROM to_archive", sql)
	expectArgs(t, []any{false}, args)
}

func TestInsertReturningExpr(t *testing.T) {
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Returning("id").
		ReturningExpr(Raw("NOW()").As("created_at")).
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, NOW() AS created_at", sql)
	expectArgs(t, []any{"Alice", "alice@example.com"}, args)
}

func TestInsertReturningExprOnly(t *testing.T) {
	sql, args := Insert("users").
		Columns("name").
		Values("Alice").
		ReturningExpr(
			Raw("id"),
			CoalesceExpr(RawExpr("$1", "default"), Raw("name")).As("display"),
		).
		Build()
	expectSQL(t, "INSERT INTO users (name) VALUES ($1) RETURNING id, COALESCE($2, name) AS display", sql)
	expectArgs(t, []any{"Alice", "default"}, args)
}

func TestInsertWhen(t *testing.T) {
	addReturning := true
	sql, args := Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		When(addReturning, func(b *InsertBuilder) {
			b.Returning("id")
		}).
		Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", sql)
	expectArgs(t, []any{"Alice", "alice@example.com"}, args)
}

func TestInsertWhenFalse(t *testing.T) {
	sql, args := Insert("users").
		Columns("name").
		When(false, func(b *InsertBuilder) {
			b.Columns("email")
		}).
		Values("Alice").
		Build()
	expectSQL(t, "INSERT INTO users (name) VALUES ($1)", sql)
	expectArgs(t, []any{"Alice"}, args)
}

func TestInsertClone(t *testing.T) {
	base := Insert("users").Columns("name", "email").Values("Alice", "alice@example.com")
	clone := base.Clone()
	clone.Values("Bob", "bob@example.com")

	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2)", sqlOrig)
	expectArgs(t, []any{"Alice", "alice@example.com"}, argsOrig)

	sqlClone, argsClone := clone.Build()
	expectSQL(t, "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)", sqlClone)
	expectArgs(t, []any{"Alice", "alice@example.com", "Bob", "bob@example.com"}, argsClone)
}

func TestInsertCloneWithFromSelect(t *testing.T) {
	sel := Select("id", "name").From("users").Where("active = $1", true)
	base := Insert("archive").Columns("id", "name").FromSelect(sel)
	clone := base.Clone()

	// Mutate the clone's fromSelect indirectly by building both
	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "INSERT INTO archive (id, name) SELECT id, name FROM users WHERE active = $1", sqlOrig)
	expectArgs(t, []any{true}, argsOrig)

	sqlClone, argsClone := clone.Build()
	expectSQL(t, "INSERT INTO archive (id, name) SELECT id, name FROM users WHERE active = $1", sqlClone)
	expectArgs(t, []any{true}, argsClone)
}

func TestInsertQuery(t *testing.T) {
	q := Insert("users").Columns("name").Values("Alice").Query()
	expectSQL(t, "INSERT INTO users (name) VALUES ($1)", q.SQL)
	expectArgs(t, []any{"Alice"}, q.Args)
}

func TestInsertString(t *testing.T) {
	s := Insert("users").Columns("name").Values("Alice").String()
	expectSQL(t, "INSERT INTO users (name) VALUES ($1)", s)
}
