package sqlbuilder

import (
	"testing"
)

func TestDeleteSimple(t *testing.T) {
	sql, args := Delete("users").Where("id = $1", 1).Build()
	expectSQL(t, "DELETE FROM users WHERE id = $1", sql)
	expectArgs(t, []any{1}, args)
}

func TestDeleteMultipleWhere(t *testing.T) {
	sql, args := Delete("users").
		Where("active = $1", false).
		Where("created_at < $1", "2023-01-01").
		Build()
	expectSQL(t, "DELETE FROM users WHERE active = $1 AND created_at < $2", sql)
	expectArgs(t, []any{false, "2023-01-01"}, args)
}

func TestDeleteUsing(t *testing.T) {
	sql, args := Delete("users u").
		Using("blacklist b").
		Where("u.email = b.email").
		Where("b.reason = $1", "spam").
		Build()
	expectSQL(t, "DELETE FROM users u USING blacklist b WHERE u.email = b.email AND b.reason = $1", sql)
	expectArgs(t, []any{"spam"}, args)
}

func TestDeleteWhereIn(t *testing.T) {
	sql, args := Delete("users").WhereIn("id", 1, 2, 3).Build()
	expectSQL(t, "DELETE FROM users WHERE id IN ($1, $2, $3)", sql)
	expectArgs(t, []any{1, 2, 3}, args)
}

func TestDeleteWhereNull(t *testing.T) {
	sql, _ := Delete("users").WhereNull("email").Build()
	expectSQL(t, "DELETE FROM users WHERE email IS NULL", sql)
}

func TestDeleteWhereNotNull(t *testing.T) {
	sql, _ := Delete("sessions").WhereNotNull("expired_at").Build()
	expectSQL(t, "DELETE FROM sessions WHERE expired_at IS NOT NULL", sql)
}

func TestDeleteReturning(t *testing.T) {
	sql, args := Delete("users").
		Where("id = $1", 1).
		Returning("id", "name").
		Build()
	expectSQL(t, "DELETE FROM users WHERE id = $1 RETURNING id, name", sql)
	expectArgs(t, []any{1}, args)
}

func TestDeleteWithCTE(t *testing.T) {
	cteQuery := Select("id").From("users").Where("active = $1", false).Query()
	sql, args := Delete("users").
		With("to_delete", cteQuery).
		Where("id IN (SELECT id FROM to_delete)").
		Build()
	expectSQL(t, "WITH to_delete AS (SELECT id FROM users WHERE active = $1) DELETE FROM users WHERE id IN (SELECT id FROM to_delete)", sql)
	expectArgs(t, []any{false}, args)
}

func TestDeleteWhereEq(t *testing.T) {
	sql, args := Delete("users").WhereEq("id", 1).Build()
	expectSQL(t, "DELETE FROM users WHERE id = $1", sql)
	expectArgs(t, []any{1}, args)
}

func TestDeleteWhereNeq(t *testing.T) {
	sql, args := Delete("users").WhereNeq("role", "admin").Build()
	expectSQL(t, "DELETE FROM users WHERE role != $1", sql)
	expectArgs(t, []any{"admin"}, args)
}

func TestDeleteWhereGt(t *testing.T) {
	sql, args := Delete("logs").WhereGt("age_days", 90).Build()
	expectSQL(t, "DELETE FROM logs WHERE age_days > $1", sql)
	expectArgs(t, []any{90}, args)
}

func TestDeleteWhereGte(t *testing.T) {
	sql, args := Delete("logs").WhereGte("age_days", 90).Build()
	expectSQL(t, "DELETE FROM logs WHERE age_days >= $1", sql)
	expectArgs(t, []any{90}, args)
}

func TestDeleteWhereLt(t *testing.T) {
	sql, args := Delete("sessions").WhereLt("last_active", "2024-01-01").Build()
	expectSQL(t, "DELETE FROM sessions WHERE last_active < $1", sql)
	expectArgs(t, []any{"2024-01-01"}, args)
}

func TestDeleteWhereLte(t *testing.T) {
	sql, args := Delete("sessions").WhereLte("last_active", "2024-01-01").Build()
	expectSQL(t, "DELETE FROM sessions WHERE last_active <= $1", sql)
	expectArgs(t, []any{"2024-01-01"}, args)
}

func TestDeleteWhereLike(t *testing.T) {
	sql, args := Delete("users").WhereLike("email", "%@spam.com").Build()
	expectSQL(t, "DELETE FROM users WHERE email LIKE $1", sql)
	expectArgs(t, []any{"%@spam.com"}, args)
}

func TestDeleteWhereILike(t *testing.T) {
	sql, args := Delete("users").WhereILike("email", "%@spam.com").Build()
	expectSQL(t, "DELETE FROM users WHERE email ILIKE $1", sql)
	expectArgs(t, []any{"%@spam.com"}, args)
}

func TestDeleteWhereExists(t *testing.T) {
	sub := Select("1").From("banned_users").Where("banned_users.id = users.id")
	sql, args := Delete("users").WhereExists(sub).Build()
	expectSQL(t, "DELETE FROM users WHERE EXISTS (SELECT 1 FROM banned_users WHERE banned_users.id = users.id)", sql)
	expectArgs(t, nil, args)
}

func TestDeleteWhereNotExists(t *testing.T) {
	sub := Select("1").From("orders").Where("orders.user_id = users.id")
	sql, args := Delete("users").WhereNotExists(sub).Build()
	expectSQL(t, "DELETE FROM users WHERE NOT EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)", sql)
	expectArgs(t, nil, args)
}

func TestDeleteWhereOr(t *testing.T) {
	sql, args := Delete("users").
		WhereOr(
			Or("status = $1", "spam"),
			Or("status = $1", "banned"),
		).
		Build()
	expectSQL(t, "DELETE FROM users WHERE (status = $1 OR status = $2)", sql)
	expectArgs(t, []any{"spam", "banned"}, args)
}

func TestDeleteWhereInSubquery(t *testing.T) {
	sub := Select("user_id").From("blacklist")
	sql, args := Delete("users").
		WhereInSubquery("id", sub).
		Build()
	expectSQL(t, "DELETE FROM users WHERE id IN (SELECT user_id FROM blacklist)", sql)
	expectArgs(t, nil, args)
}

func TestDeleteWhereNotInSubquery(t *testing.T) {
	sub := Select("id").From("active_users")
	sql, args := Delete("users").
		WhereNotInSubquery("id", sub).
		Build()
	expectSQL(t, "DELETE FROM users WHERE id NOT IN (SELECT id FROM active_users)", sql)
	expectArgs(t, nil, args)
}

func TestDeleteWhen(t *testing.T) {
	forceAll := false
	sql, args := Delete("sessions").
		When(!forceAll, func(b *DeleteBuilder) {
			b.Where("expired_at < $1", "2024-01-01")
		}).
		Build()
	expectSQL(t, "DELETE FROM sessions WHERE expired_at < $1", sql)
	expectArgs(t, []any{"2024-01-01"}, args)
}

func TestDeleteWhenFalse(t *testing.T) {
	sql, args := Delete("sessions").
		When(false, func(b *DeleteBuilder) {
			b.Where("should_not_appear = $1", true)
		}).
		Build()
	expectSQL(t, "DELETE FROM sessions", sql)
	expectArgs(t, nil, args)
}

func TestDeleteClone(t *testing.T) {
	base := Delete("users").Where("active = $1", false)
	clone := base.Clone()
	clone.Where("created_at < $1", "2023-01-01")

	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "DELETE FROM users WHERE active = $1", sqlOrig)
	expectArgs(t, []any{false}, argsOrig)

	sqlClone, argsClone := clone.Build()
	expectSQL(t, "DELETE FROM users WHERE active = $1 AND created_at < $2", sqlClone)
	expectArgs(t, []any{false, "2023-01-01"}, argsClone)
}

func TestDeleteReturningExpr(t *testing.T) {
	sql, args := Delete("users").
		Where("id = $1", 1).
		Returning("id").
		ReturningExpr(Raw("name")).
		Build()
	expectSQL(t, "DELETE FROM users WHERE id = $1 RETURNING id, name", sql)
	expectArgs(t, []any{1}, args)
}

func TestDeleteReturningExprWithParams(t *testing.T) {
	sql, args := Delete("users").
		Where("active = $1", false).
		ReturningExpr(
			CoalesceExpr(RawExpr("$1", "archived"), Raw("name")).As("label"),
		).
		Build()
	expectSQL(t, "DELETE FROM users WHERE active = $1 RETURNING COALESCE($2, name) AS label", sql)
	expectArgs(t, []any{false, "archived"}, args)
}

func TestDeleteWhereColumn(t *testing.T) {
	sql, args := Delete("users u").
		Using("blacklist b").
		WhereColumn("u.email", "=", "b.email").
		Build()
	expectSQL(t, "DELETE FROM users u USING blacklist b WHERE u.email = b.email", sql)
	expectArgs(t, nil, args)
}

func TestDeleteQuery(t *testing.T) {
	q := Delete("users").Where("id = $1", 1).Query()
	expectSQL(t, "DELETE FROM users WHERE id = $1", q.SQL)
	expectArgs(t, []any{1}, q.Args)
}

func TestDeleteString(t *testing.T) {
	s := Delete("users").Where("id = $1", 1).String()
	expectSQL(t, "DELETE FROM users WHERE id = $1", s)
}
