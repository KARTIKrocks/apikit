package sqlbuilder

import (
	"testing"
)

func TestUpdateSimple(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Bob").
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2", sql)
	expectArgs(t, []any{"Bob", 1}, args)
}

func TestUpdateMultipleSet(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Bob").
		Set("email", "bob@example.com").
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE users SET name = $1, email = $2 WHERE id = $3", sql)
	expectArgs(t, []any{"Bob", "bob@example.com", 1}, args)
}

func TestUpdateSetExpr(t *testing.T) {
	sql, args := Update("users").
		SetExpr("updated_at", Raw("NOW()")).
		Set("name", "Bob").
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE users SET updated_at = NOW(), name = $1 WHERE id = $2", sql)
	expectArgs(t, []any{"Bob", 1}, args)
}

func TestUpdateSetMap(t *testing.T) {
	sql, args := Update("users").
		SetMap(map[string]any{
			"email": "bob@example.com",
			"name":  "Bob",
		}).
		Where("id = $1", 1).
		Build()
	// Keys sorted: email, name
	expectSQL(t, "UPDATE users SET email = $1, name = $2 WHERE id = $3", sql)
	expectArgs(t, []any{"bob@example.com", "Bob", 1}, args)
}

func TestUpdateFrom(t *testing.T) {
	sql, args := Update("users u").
		Set("name", "Updated").
		From("profiles p").
		Where("p.user_id = u.id").
		Where("p.type = $1", "premium").
		Build()
	expectSQL(t, "UPDATE users u SET name = $1 FROM profiles p WHERE p.user_id = u.id AND p.type = $2", sql)
	expectArgs(t, []any{"Updated", "premium"}, args)
}

func TestUpdateWhereIn(t *testing.T) {
	sql, args := Update("users").
		Set("active", false).
		WhereIn("id", 1, 2, 3).
		Build()
	expectSQL(t, "UPDATE users SET active = $1 WHERE id IN ($2, $3, $4)", sql)
	expectArgs(t, []any{false, 1, 2, 3}, args)
}

func TestUpdateWhereNull(t *testing.T) {
	sql, args := Update("users").
		Set("verified", true).
		WhereNull("deleted_at").
		Build()
	expectSQL(t, "UPDATE users SET verified = $1 WHERE deleted_at IS NULL", sql)
	expectArgs(t, []any{true}, args)
}

func TestUpdateWhereNotNull(t *testing.T) {
	sql, args := Update("users").
		Set("notified", true).
		WhereNotNull("email").
		Build()
	expectSQL(t, "UPDATE users SET notified = $1 WHERE email IS NOT NULL", sql)
	expectArgs(t, []any{true}, args)
}

func TestUpdateReturning(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Bob").
		Where("id = $1", 1).
		Returning("id", "name", "updated_at").
		Build()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2 RETURNING id, name, updated_at", sql)
	expectArgs(t, []any{"Bob", 1}, args)
}

func TestUpdateWithCTE(t *testing.T) {
	cteQuery := Select("id").From("users").Where("active = $1", false).Query()
	sql, args := Update("users").
		With("inactive", cteQuery).
		Set("archived", true).
		Where("id IN (SELECT id FROM inactive)").
		Build()
	expectSQL(t, "WITH inactive AS (SELECT id FROM users WHERE active = $1) UPDATE users SET archived = $2 WHERE id IN (SELECT id FROM inactive)", sql)
	expectArgs(t, []any{false, true}, args)
}

func TestUpdateWhereEq(t *testing.T) {
	sql, args := Update("users").Set("name", "Bob").WhereEq("id", 1).Build()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2", sql)
	expectArgs(t, []any{"Bob", 1}, args)
}

func TestUpdateWhereNeq(t *testing.T) {
	sql, args := Update("users").Set("active", false).WhereNeq("role", "admin").Build()
	expectSQL(t, "UPDATE users SET active = $1 WHERE role != $2", sql)
	expectArgs(t, []any{false, "admin"}, args)
}

func TestUpdateWhereGt(t *testing.T) {
	sql, args := Update("users").Set("tier", "senior").WhereGt("age", 30).Build()
	expectSQL(t, "UPDATE users SET tier = $1 WHERE age > $2", sql)
	expectArgs(t, []any{"senior", 30}, args)
}

func TestUpdateWhereGte(t *testing.T) {
	sql, args := Update("users").Set("tier", "senior").WhereGte("age", 30).Build()
	expectSQL(t, "UPDATE users SET tier = $1 WHERE age >= $2", sql)
	expectArgs(t, []any{"senior", 30}, args)
}

func TestUpdateWhereLt(t *testing.T) {
	sql, args := Update("users").Set("status", "junior").WhereLt("experience", 2).Build()
	expectSQL(t, "UPDATE users SET status = $1 WHERE experience < $2", sql)
	expectArgs(t, []any{"junior", 2}, args)
}

func TestUpdateWhereLte(t *testing.T) {
	sql, args := Update("users").Set("status", "junior").WhereLte("experience", 2).Build()
	expectSQL(t, "UPDATE users SET status = $1 WHERE experience <= $2", sql)
	expectArgs(t, []any{"junior", 2}, args)
}

func TestUpdateWhereLike(t *testing.T) {
	sql, args := Update("users").Set("flagged", true).WhereLike("email", "%spam%").Build()
	expectSQL(t, "UPDATE users SET flagged = $1 WHERE email LIKE $2", sql)
	expectArgs(t, []any{true, "%spam%"}, args)
}

func TestUpdateWhereILike(t *testing.T) {
	sql, args := Update("users").Set("flagged", true).WhereILike("email", "%spam%").Build()
	expectSQL(t, "UPDATE users SET flagged = $1 WHERE email ILIKE $2", sql)
	expectArgs(t, []any{true, "%spam%"}, args)
}

func TestUpdateWhereExists(t *testing.T) {
	sub := Select("1").From("orders").Where("orders.user_id = users.id")
	sql, args := Update("users").
		Set("has_orders", true).
		WhereExists(sub).
		Build()
	expectSQL(t, "UPDATE users SET has_orders = $1 WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)", sql)
	expectArgs(t, []any{true}, args)
}

func TestUpdateWhereNotExists(t *testing.T) {
	sub := Select("1").From("orders").Where("orders.user_id = users.id")
	sql, args := Update("users").
		Set("has_orders", false).
		WhereNotExists(sub).
		Build()
	expectSQL(t, "UPDATE users SET has_orders = $1 WHERE NOT EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)", sql)
	expectArgs(t, []any{false}, args)
}

func TestUpdateWhereOr(t *testing.T) {
	sql, args := Update("users").
		Set("flagged", true).
		WhereOr(
			Or("status = $1", "spam"),
			Or("status = $1", "banned"),
		).
		Build()
	expectSQL(t, "UPDATE users SET flagged = $1 WHERE (status = $2 OR status = $3)", sql)
	expectArgs(t, []any{true, "spam", "banned"}, args)
}

func TestUpdateWhereInSubquery(t *testing.T) {
	sub := Select("user_id").From("orders").Where("total > $1", 1000)
	sql, args := Update("users").
		Set("vip", true).
		WhereInSubquery("id", sub).
		Build()
	expectSQL(t, "UPDATE users SET vip = $1 WHERE id IN (SELECT user_id FROM orders WHERE total > $2)", sql)
	expectArgs(t, []any{true, 1000}, args)
}

func TestUpdateWhereNotInSubquery(t *testing.T) {
	sub := Select("user_id").From("active_sessions")
	sql, args := Update("users").
		Set("online", false).
		WhereNotInSubquery("id", sub).
		Build()
	expectSQL(t, "UPDATE users SET online = $1 WHERE id NOT IN (SELECT user_id FROM active_sessions)", sql)
	expectArgs(t, []any{false}, args)
}

func TestUpdateWhen(t *testing.T) {
	setVIP := true
	sql, args := Update("users").
		Set("name", "Bob").
		When(setVIP, func(b *UpdateBuilder) {
			b.Set("vip", true)
		}).
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE users SET name = $1, vip = $2 WHERE id = $3", sql)
	expectArgs(t, []any{"Bob", true, 1}, args)
}

func TestUpdateWhenFalse(t *testing.T) {
	sql, args := Update("users").
		Set("name", "Bob").
		When(false, func(b *UpdateBuilder) {
			b.Set("vip", true)
		}).
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2", sql)
	expectArgs(t, []any{"Bob", 1}, args)
}

func TestUpdateClone(t *testing.T) {
	base := Update("users").Set("name", "Bob").Where("active = $1", true)
	clone := base.Clone()
	clone.Set("email", "bob@example.com")

	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "UPDATE users SET name = $1 WHERE active = $2", sqlOrig)
	expectArgs(t, []any{"Bob", true}, argsOrig)

	sqlClone, argsClone := clone.Build()
	expectSQL(t, "UPDATE users SET name = $1, email = $2 WHERE active = $3", sqlClone)
	expectArgs(t, []any{"Bob", "bob@example.com", true}, argsClone)
}

func TestUpdateIncrement(t *testing.T) {
	sql, args := Update("products").
		Increment("view_count", 1).
		Where("id = $1", 42).
		Build()
	expectSQL(t, "UPDATE products SET view_count = view_count + $1 WHERE id = $2", sql)
	expectArgs(t, []any{1, 42}, args)
}

func TestUpdateDecrement(t *testing.T) {
	sql, args := Update("products").
		Decrement("stock", 5).
		Where("id = $1", 42).
		Build()
	expectSQL(t, "UPDATE products SET stock = stock - $1 WHERE id = $2", sql)
	expectArgs(t, []any{5, 42}, args)
}

func TestUpdateIncrementDecrement(t *testing.T) {
	sql, args := Update("accounts").
		Increment("credits", 100).
		Decrement("pending", 100).
		Where("id = $1", 1).
		Build()
	expectSQL(t, "UPDATE accounts SET credits = credits + $1, pending = pending - $2 WHERE id = $3", sql)
	expectArgs(t, []any{100, 100, 1}, args)
}

func TestUpdateQuery(t *testing.T) {
	q := Update("users").Set("name", "Bob").Where("id = $1", 1).Query()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2", q.SQL)
	expectArgs(t, []any{"Bob", 1}, q.Args)
}

func TestUpdateString(t *testing.T) {
	s := Update("users").Set("name", "Bob").Where("id = $1", 1).String()
	expectSQL(t, "UPDATE users SET name = $1 WHERE id = $2", s)
}
