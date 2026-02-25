package sqlbuilder

import (
	"testing"
)

func TestSelectSimple(t *testing.T) {
	sql, args := Select("id", "name").From("users").Build()
	expectSQL(t, "SELECT id, name FROM users", sql)
	expectArgs(t, nil, args)
}

func TestSelectStar(t *testing.T) {
	sql, _ := Select().From("users").Build()
	expectSQL(t, "SELECT * FROM users", sql)
}

func TestSelectDistinct(t *testing.T) {
	sql, _ := Select("status").From("users").Distinct().Build()
	expectSQL(t, "SELECT DISTINCT status FROM users", sql)
}

func TestSelectWhereSingle(t *testing.T) {
	sql, args := Select("id").From("users").Where("active = $1", true).Build()
	expectSQL(t, "SELECT id FROM users WHERE active = $1", sql)
	expectArgs(t, []any{true}, args)
}

func TestSelectWhereMultiple(t *testing.T) {
	sql, args := Select("id").From("users").
		Where("status = $1", "active").
		Where("age > $1", 18).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE status = $1 AND age > $2", sql)
	expectArgs(t, []any{"active", 18}, args)
}

func TestSelectWhereIn(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereIn("status", "active", "pending").
		Build()
	expectSQL(t, "SELECT id FROM users WHERE status IN ($1, $2)", sql)
	expectArgs(t, []any{"active", "pending"}, args)
}

func TestSelectWhereNotIn(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereNotIn("role", "banned", "suspended").
		Build()
	expectSQL(t, "SELECT id FROM users WHERE role NOT IN ($1, $2)", sql)
	expectArgs(t, []any{"banned", "suspended"}, args)
}

func TestSelectWhereBetween(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereBetween("age", 18, 65).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE age BETWEEN $1 AND $2", sql)
	expectArgs(t, []any{18, 65}, args)
}

func TestSelectWhereNull(t *testing.T) {
	sql, args := Select("id").From("users").WhereNull("deleted_at").Build()
	expectSQL(t, "SELECT id FROM users WHERE deleted_at IS NULL", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereNotNull(t *testing.T) {
	sql, args := Select("id").From("users").WhereNotNull("email").Build()
	expectSQL(t, "SELECT id FROM users WHERE email IS NOT NULL", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereExists(t *testing.T) {
	sub := Select("1").From("orders").Where("orders.user_id = users.id")
	sql, args := Select("id").From("users").WhereExists(sub).Build()
	expectSQL(t, "SELECT id FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereOr(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereOr(
			Or("status = $1", "active"),
			Or("role = $1", "admin"),
		).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE (status = $1 OR role = $2)", sql)
	expectArgs(t, []any{"active", "admin"}, args)
}

func TestSelectJoin(t *testing.T) {
	sql, _ := Select("u.id", "p.bio").
		From("users u").
		Join("profiles p", "p.user_id = u.id").
		Build()
	expectSQL(t, "SELECT u.id, p.bio FROM users u JOIN profiles p ON p.user_id = u.id", sql)
}

func TestSelectLeftJoin(t *testing.T) {
	sql, _ := Select("u.id").
		From("users u").
		LeftJoin("profiles p", "p.user_id = u.id").
		Build()
	expectSQL(t, "SELECT u.id FROM users u LEFT JOIN profiles p ON p.user_id = u.id", sql)
}

func TestSelectRightJoin(t *testing.T) {
	sql, _ := Select("u.id").
		From("users u").
		RightJoin("profiles p", "p.user_id = u.id").
		Build()
	expectSQL(t, "SELECT u.id FROM users u RIGHT JOIN profiles p ON p.user_id = u.id", sql)
}

func TestSelectFullJoin(t *testing.T) {
	sql, _ := Select("u.id").
		From("users u").
		FullJoin("profiles p", "p.user_id = u.id").
		Build()
	expectSQL(t, "SELECT u.id FROM users u FULL JOIN profiles p ON p.user_id = u.id", sql)
}

func TestSelectCrossJoin(t *testing.T) {
	sql, _ := Select("u.id", "r.name").
		From("users u").
		CrossJoin("roles r").
		Build()
	expectSQL(t, "SELECT u.id, r.name FROM users u CROSS JOIN roles r", sql)
}

func TestSelectJoinWithArgs(t *testing.T) {
	sql, args := Select("u.id").
		From("users u").
		Join("profiles p", "p.user_id = u.id AND p.type = $1", "public").
		Build()
	expectSQL(t, "SELECT u.id FROM users u JOIN profiles p ON p.user_id = u.id AND p.type = $1", sql)
	expectArgs(t, []any{"public"}, args)
}

func TestSelectGroupByHaving(t *testing.T) {
	sql, args := Select("status", "COUNT(*)").
		From("users").
		GroupBy("status").
		Having("COUNT(*) > $1", 5).
		Build()
	expectSQL(t, "SELECT status, COUNT(*) FROM users GROUP BY status HAVING COUNT(*) > $1", sql)
	expectArgs(t, []any{5}, args)
}

func TestSelectOrderBy(t *testing.T) {
	sql, _ := Select("id").From("users").OrderBy("name ASC", "created_at DESC").Build()
	expectSQL(t, "SELECT id FROM users ORDER BY name ASC, created_at DESC", sql)
}

func TestSelectLimitOffset(t *testing.T) {
	sql, _ := Select("id").From("users").Limit(10).Offset(20).Build()
	expectSQL(t, "SELECT id FROM users LIMIT 10 OFFSET 20", sql)
}

func TestSelectForUpdate(t *testing.T) {
	sql, _ := Select("id").From("users").Where("id = $1", 1).ForUpdate().Build()
	expectSQL(t, "SELECT id FROM users WHERE id = $1 FOR UPDATE", sql)
}

func TestSelectForUpdateSkipLocked(t *testing.T) {
	sql, _ := Select("id").From("users").ForUpdate().SkipLocked().Build()
	expectSQL(t, "SELECT id FROM users FOR UPDATE SKIP LOCKED", sql)
}

func TestSelectForUpdateNoWait(t *testing.T) {
	sql, _ := Select("id").From("users").ForUpdate().NoWait().Build()
	expectSQL(t, "SELECT id FROM users FOR UPDATE NOWAIT", sql)
}

func TestSelectForShare(t *testing.T) {
	sql, _ := Select("id").From("users").ForShare().Build()
	expectSQL(t, "SELECT id FROM users FOR SHARE", sql)
}

func TestSelectUnion(t *testing.T) {
	q1 := Select("id", "name").From("users")
	q2 := Select("id", "name").From("admins")
	sql, _ := q1.Union(q2).Build()
	expectSQL(t, "SELECT id, name FROM users UNION SELECT id, name FROM admins", sql)
}

func TestSelectUnionAll(t *testing.T) {
	q1 := Select("id").From("users")
	q2 := Select("id").From("admins")
	sql, _ := q1.UnionAll(q2).Build()
	expectSQL(t, "SELECT id FROM users UNION ALL SELECT id FROM admins", sql)
}

func TestSelectIntersect(t *testing.T) {
	q1 := Select("id").From("users")
	q2 := Select("id").From("premium_users")
	sql, _ := q1.Intersect(q2).Build()
	expectSQL(t, "SELECT id FROM users INTERSECT SELECT id FROM premium_users", sql)
}

func TestSelectExcept(t *testing.T) {
	q1 := Select("id").From("users")
	q2 := Select("id").From("banned_users")
	sql, _ := q1.Except(q2).Build()
	expectSQL(t, "SELECT id FROM users EXCEPT SELECT id FROM banned_users", sql)
}

func TestSelectCTE(t *testing.T) {
	cteQuery := Select("id", "name").From("users").Where("active = $1", true).Query()
	sql, args := Select("id", "name").
		With("active_users", cteQuery).
		From("active_users").
		Build()
	expectSQL(t, "WITH active_users AS (SELECT id, name FROM users WHERE active = $1) SELECT id, name FROM active_users", sql)
	expectArgs(t, []any{true}, args)
}

func TestSelectRecursiveCTE(t *testing.T) {
	cteQuery := Query{
		SQL:  "SELECT id, parent_id, name FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.parent_id, c.name FROM categories c JOIN tree t ON c.parent_id = t.id",
		Args: nil,
	}
	sql, _ := Select("id", "name").
		WithRecursive("tree", cteQuery).
		From("tree").
		Build()
	expectSQL(t, "WITH RECURSIVE tree AS (SELECT id, parent_id, name FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.parent_id, c.name FROM categories c JOIN tree t ON c.parent_id = t.id) SELECT id, name FROM tree", sql)
}

func TestSelectSubquery(t *testing.T) {
	sub := Select("user_id").From("orders").Where("total > $1", 100)
	sql, args := Select("id", "name").
		From("users").
		Where("id IN (SELECT user_id FROM orders WHERE total > $1)", 100).
		Build()
	_ = sub
	expectSQL(t, "SELECT id, name FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > $1)", sql)
	expectArgs(t, []any{100}, args)
}

func TestSelectFromSubquery(t *testing.T) {
	sub := Select("user_id", "SUM(total) as total").
		From("orders").
		GroupBy("user_id")
	sql, _ := Select("user_id", "total").
		FromSubquery(sub, "o").
		Build()
	expectSQL(t, "SELECT user_id, total FROM (SELECT user_id, SUM(total) as total FROM orders GROUP BY user_id) o", sql)
}

func TestSelectColumnExpr(t *testing.T) {
	sql, _ := SelectExpr(Raw("COUNT(*)")).From("users").Build()
	expectSQL(t, "SELECT COUNT(*) FROM users", sql)
}

func TestSelectAddColumn(t *testing.T) {
	sql, _ := Select("id").Column("name").Columns("email", "age").From("users").Build()
	expectSQL(t, "SELECT id, name, email, age FROM users", sql)
}

func TestSelectFromAlias(t *testing.T) {
	sql, _ := Select("u.id").FromAlias("users", "u").Build()
	expectSQL(t, "SELECT u.id FROM users u", sql)
}

func TestSelectMultipleWhereWithJoinArgs(t *testing.T) {
	sql, args := Select("u.id").
		From("users u").
		Join("profiles p", "p.user_id = u.id AND p.type = $1", "public").
		Where("u.active = $1", true).
		Where("u.age > $1", 18).
		Build()
	expectSQL(t, "SELECT u.id FROM users u JOIN profiles p ON p.user_id = u.id AND p.type = $1 WHERE u.active = $2 AND u.age > $3", sql)
	expectArgs(t, []any{"public", true, 18}, args)
}

func TestSelectOrderByExpr(t *testing.T) {
	sql, _ := Select("id").From("users").OrderByExpr(Raw("RANDOM()")).Build()
	expectSQL(t, "SELECT id FROM users ORDER BY RANDOM()", sql)
}

func TestSelectQuery(t *testing.T) {
	q := Select("id").From("users").Where("active = $1", true).Query()
	expectSQL(t, "SELECT id FROM users WHERE active = $1", q.SQL)
	expectArgs(t, []any{true}, q.Args)
}

func TestSelectString(t *testing.T) {
	s := Select("id").From("users").String()
	expectSQL(t, "SELECT id FROM users", s)
}

func TestSelectPlaceholderRebasing(t *testing.T) {
	sql, args := Select("id").From("users").
		Where("a = $1", 1).
		Where("b = $1 AND c = $2", 2, 3).
		Where("d = $1", 4).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE a = $1 AND b = $2 AND c = $3 AND d = $4", sql)
	expectArgs(t, []any{1, 2, 3, 4}, args)
}

func TestSelectWhereInEmpty(t *testing.T) {
	sql, args := Select("id").From("users").WhereIn("status").Build()
	expectSQL(t, "SELECT id FROM users WHERE 1=0", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereNotInEmpty(t *testing.T) {
	sql, _ := Select("id").From("users").WhereNotIn("status").Build()
	expectSQL(t, "SELECT id FROM users WHERE 1=1", sql)
}

func TestSelectComplexQuery(t *testing.T) {
	sql, args := Select("u.id", "u.name", "COUNT(o.id) as order_count").
		From("users u").
		LeftJoin("orders o", "o.user_id = u.id").
		Where("u.active = $1", true).
		Where("u.created_at > $1", "2024-01-01").
		GroupBy("u.id", "u.name").
		Having("COUNT(o.id) > $1", 5).
		OrderBy("order_count DESC").
		Limit(10).
		Offset(0).
		Build()
	expectSQL(t, "SELECT u.id, u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON o.user_id = u.id WHERE u.active = $1 AND u.created_at > $2 GROUP BY u.id, u.name HAVING COUNT(o.id) > $3 ORDER BY order_count DESC LIMIT 10 OFFSET 0", sql)
	expectArgs(t, []any{true, "2024-01-01", 5}, args)
}

func TestSelectWhereEq(t *testing.T) {
	sql, args := Select("id").From("users").WhereEq("status", "active").Build()
	expectSQL(t, "SELECT id FROM users WHERE status = $1", sql)
	expectArgs(t, []any{"active"}, args)
}

func TestSelectWhereNeq(t *testing.T) {
	sql, args := Select("id").From("users").WhereNeq("role", "banned").Build()
	expectSQL(t, "SELECT id FROM users WHERE role != $1", sql)
	expectArgs(t, []any{"banned"}, args)
}

func TestSelectWhereGt(t *testing.T) {
	sql, args := Select("id").From("users").WhereGt("age", 18).Build()
	expectSQL(t, "SELECT id FROM users WHERE age > $1", sql)
	expectArgs(t, []any{18}, args)
}

func TestSelectWhereGte(t *testing.T) {
	sql, args := Select("id").From("users").WhereGte("age", 18).Build()
	expectSQL(t, "SELECT id FROM users WHERE age >= $1", sql)
	expectArgs(t, []any{18}, args)
}

func TestSelectWhereLt(t *testing.T) {
	sql, args := Select("id").From("users").WhereLt("age", 65).Build()
	expectSQL(t, "SELECT id FROM users WHERE age < $1", sql)
	expectArgs(t, []any{65}, args)
}

func TestSelectWhereLte(t *testing.T) {
	sql, args := Select("id").From("users").WhereLte("age", 65).Build()
	expectSQL(t, "SELECT id FROM users WHERE age <= $1", sql)
	expectArgs(t, []any{65}, args)
}

func TestSelectWhereLike(t *testing.T) {
	sql, args := Select("id").From("users").WhereLike("name", "%alice%").Build()
	expectSQL(t, "SELECT id FROM users WHERE name LIKE $1", sql)
	expectArgs(t, []any{"%alice%"}, args)
}

func TestSelectWhereILike(t *testing.T) {
	sql, args := Select("id").From("users").WhereILike("name", "%alice%").Build()
	expectSQL(t, "SELECT id FROM users WHERE name ILIKE $1", sql)
	expectArgs(t, []any{"%alice%"}, args)
}

func TestSelectWhereHelpersCombined(t *testing.T) {
	sql, args := Select("id").From("users").
		WhereEq("active", true).
		WhereGt("age", 18).
		WhereLike("name", "A%").
		Build()
	expectSQL(t, "SELECT id FROM users WHERE active = $1 AND age > $2 AND name LIKE $3", sql)
	expectArgs(t, []any{true, 18, "A%"}, args)
}

func TestSelectOrderByAsc(t *testing.T) {
	sql, _ := Select("id").From("users").OrderByAsc("name", "created_at").Build()
	expectSQL(t, "SELECT id FROM users ORDER BY name ASC, created_at ASC", sql)
}

func TestSelectOrderByDesc(t *testing.T) {
	sql, _ := Select("id").From("users").OrderByDesc("created_at").Build()
	expectSQL(t, "SELECT id FROM users ORDER BY created_at DESC", sql)
}

func TestSelectOrderByAscDesc(t *testing.T) {
	sql, _ := Select("id").From("users").OrderByAsc("name").OrderByDesc("created_at").Build()
	expectSQL(t, "SELECT id FROM users ORDER BY name ASC, created_at DESC", sql)
}

func TestSelectWhereNotExists(t *testing.T) {
	sub := Select("1").From("orders").Where("orders.user_id = users.id")
	sql, args := Select("id").From("users").WhereNotExists(sub).Build()
	expectSQL(t, "SELECT id FROM users WHERE NOT EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereInSubquery(t *testing.T) {
	sub := Select("user_id").From("orders").Where("total > $1", 100)
	sql, args := Select("id", "name").From("users").
		WhereInSubquery("id", sub).
		Build()
	expectSQL(t, "SELECT id, name FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > $1)", sql)
	expectArgs(t, []any{100}, args)
}

func TestSelectWhereNotInSubquery(t *testing.T) {
	sub := Select("user_id").From("banned_users")
	sql, args := Select("id").From("users").
		WhereNotInSubquery("id", sub).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE id NOT IN (SELECT user_id FROM banned_users)", sql)
	expectArgs(t, nil, args)
}

func TestSelectWhereInSubqueryWithExistingConditions(t *testing.T) {
	sub := Select("user_id").From("orders").Where("total > $1", 100)
	sql, args := Select("id").From("users").
		Where("active = $1", true).
		WhereInSubquery("id", sub).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE active = $1 AND id IN (SELECT user_id FROM orders WHERE total > $2)", sql)
	expectArgs(t, []any{true, 100}, args)
}

func TestSelectWhen(t *testing.T) {
	includeInactive := true
	sql, args := Select("id").From("users").
		Where("active = $1", true).
		When(includeInactive, func(s *SelectBuilder) {
			s.Where("deleted_at IS NULL")
		}).
		Build()
	expectSQL(t, "SELECT id FROM users WHERE active = $1 AND deleted_at IS NULL", sql)
	expectArgs(t, []any{true}, args)
}

func TestSelectWhenFalse(t *testing.T) {
	sql, args := Select("id").From("users").
		When(false, func(s *SelectBuilder) {
			s.Where("should_not_appear = $1", "nope")
		}).
		Build()
	expectSQL(t, "SELECT id FROM users", sql)
	expectArgs(t, nil, args)
}

func TestSelectClone(t *testing.T) {
	base := Select("id", "name").From("users").Where("active = $1", true)
	clone := base.Clone()
	clone.Where("age > $1", 18)

	// Original should not be affected
	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "SELECT id, name FROM users WHERE active = $1", sqlOrig)
	expectArgs(t, []any{true}, argsOrig)

	// Clone should have the extra condition
	sqlClone, argsClone := clone.Build()
	expectSQL(t, "SELECT id, name FROM users WHERE active = $1 AND age > $2", sqlClone)
	expectArgs(t, []any{true, 18}, argsClone)
}

func TestSelectCloneWithSubquery(t *testing.T) {
	sub := Select("user_id").From("orders")
	base := Select("id").FromSubquery(sub, "o")
	clone := base.Clone()
	clone.Where("id > $1", 5)

	sqlOrig, argsOrig := base.Build()
	expectSQL(t, "SELECT id FROM (SELECT user_id FROM orders) o", sqlOrig)
	expectArgs(t, nil, argsOrig)

	sqlClone, argsClone := clone.Build()
	expectSQL(t, "SELECT id FROM (SELECT user_id FROM orders) o WHERE id > $1", sqlClone)
	expectArgs(t, []any{5}, argsClone)
}

func TestSelectUnionWithArgs(t *testing.T) {
	q1 := Select("id", "name").From("users").Where("status = $1", "active")
	q2 := Select("id", "name").From("admins").Where("level > $1", 3)
	sql, args := q1.Union(q2).Build()
	expectSQL(t, "SELECT id, name FROM users WHERE status = $1 UNION SELECT id, name FROM admins WHERE level > $2", sql)
	expectArgs(t, []any{"active", 3}, args)
}

// test helpers

func expectSQL(t *testing.T, expected, got string) {
	t.Helper()
	if got != expected {
		t.Errorf("\nexpected: %s\n     got: %s", expected, got)
	}
}

func expectArgs(t *testing.T, expected []any, got []any) {
	t.Helper()
	if len(expected) == 0 && len(got) == 0 {
		return
	}
	if len(expected) != len(got) {
		t.Errorf("args length: expected %d, got %d\n  expected: %v\n       got: %v", len(expected), len(got), expected, got)
		return
	}
	for i := range expected {
		if expected[i] != got[i] {
			t.Errorf("arg[%d]: expected %v (%T), got %v (%T)", i, expected[i], expected[i], got[i], got[i])
		}
	}
}
