package sqlbuilder_test

import (
	"fmt"

	"github.com/KARTIKrocks/apikit/request"
	"github.com/KARTIKrocks/apikit/sqlbuilder"
)

func ExampleSelect() {
	sql, args := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where("active = $1", true).
		OrderBy("name ASC").
		Limit(20).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name, email FROM users WHERE active = $1 ORDER BY name ASC LIMIT 20
	// [true]
}

func ExampleSelect_join() {
	sql, args := sqlbuilder.Select("u.id", "u.name", "p.bio").
		From("users u").
		LeftJoin("profiles p", "p.user_id = u.id").
		Where("u.active = $1", true).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT u.id, u.name, p.bio FROM users u LEFT JOIN profiles p ON p.user_id = u.id WHERE u.active = $1
	// [true]
}

func ExampleSelect_placeholderRebasing() {
	sql, args := sqlbuilder.Select("id").
		From("users").
		Where("status = $1", "active").
		Where("age > $1", 18).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id FROM users WHERE status = $1 AND age > $2
	// [active 18]
}

func ExampleSelect_pagination() {
	p := request.Pagination{Page: 3, PerPage: 25, Offset: 50}
	sql, _ := sqlbuilder.Select("id", "name").
		From("users").
		ApplyPagination(p).
		Build()
	fmt.Println(sql)
	// Output:
	// SELECT id, name FROM users LIMIT 25 OFFSET 50
}

func ExampleInsert() {
	sql, args := sqlbuilder.Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Returning("id").
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id
	// [Alice alice@example.com]
}

func ExampleInsert_batch() {
	sql, args := sqlbuilder.Insert("users").
		Columns("name", "email").
		Values("Alice", "alice@example.com").
		Values("Bob", "bob@example.com").
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)
	// [Alice alice@example.com Bob bob@example.com]
}

func ExampleInsert_upsert() {
	sql, args := sqlbuilder.Insert("users").
		Columns("email", "name").
		Values("alice@example.com", "Alice").
		OnConflictUpdate(
			[]string{"email"},
			map[string]any{"name": "Alice Updated"},
		).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = $3
	// [alice@example.com Alice Alice Updated]
}

func ExampleUpdate() {
	sql, args := sqlbuilder.Update("users").
		Set("name", "Bob").
		SetExpr("updated_at", sqlbuilder.Raw("NOW()")).
		Where("id = $1", 1).
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2
	// [Bob 1]
}

func ExampleDelete() {
	sql, args := sqlbuilder.Delete("users").
		Where("id = $1", 1).
		Returning("id", "name").
		Build()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// DELETE FROM users WHERE id = $1 RETURNING id, name
	// [1]
}
