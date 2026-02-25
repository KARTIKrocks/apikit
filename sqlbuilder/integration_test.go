package sqlbuilder

import (
	"testing"

	"github.com/KARTIKrocks/apikit/request"
)

func TestApplyPagination(t *testing.T) {
	p := request.Pagination{Page: 3, PerPage: 25, Offset: 50}
	sql, _ := Select("id").From("users").ApplyPagination(p).Build()
	expectSQL(t, "SELECT id FROM users LIMIT 25 OFFSET 50", sql)
}

func TestApplyPaginationFirstPage(t *testing.T) {
	p := request.Pagination{Page: 1, PerPage: 20, Offset: 0}
	sql, _ := Select("id").From("users").ApplyPagination(p).Build()
	expectSQL(t, "SELECT id FROM users LIMIT 20 OFFSET 0", sql)
}

func TestApplySort(t *testing.T) {
	sorts := []request.SortField{
		{Field: "name", Direction: request.SortAsc},
		{Field: "created_at", Direction: request.SortDesc},
	}
	cols := map[string]string{
		"name":       "u.name",
		"created_at": "u.created_at",
	}
	sql, _ := Select("id").From("users u").ApplySort(sorts, cols).Build()
	expectSQL(t, "SELECT id FROM users u ORDER BY u.name ASC, u.created_at DESC", sql)
}

func TestApplySortUnknownField(t *testing.T) {
	sorts := []request.SortField{
		{Field: "name", Direction: request.SortAsc},
		{Field: "unknown", Direction: request.SortDesc},
	}
	cols := map[string]string{"name": "u.name"}
	sql, _ := Select("id").From("users u").ApplySort(sorts, cols).Build()
	expectSQL(t, "SELECT id FROM users u ORDER BY u.name ASC", sql)
}

func TestApplySortEmpty(t *testing.T) {
	sql, _ := Select("id").From("users").ApplySort(nil, nil).Build()
	expectSQL(t, "SELECT id FROM users", sql)
}

func TestApplyFiltersEq(t *testing.T) {
	filters := []request.Filter{
		{Field: "status", Operator: request.FilterOpEq, Value: "active"},
	}
	cols := map[string]string{"status": "u.status"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.status = $1", sql)
	expectArgs(t, []any{"active"}, args)
}

func TestApplyFiltersNeq(t *testing.T) {
	filters := []request.Filter{
		{Field: "status", Operator: request.FilterOpNeq, Value: "banned"},
	}
	cols := map[string]string{"status": "u.status"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.status != $1", sql)
	expectArgs(t, []any{"banned"}, args)
}

func TestApplyFiltersGt(t *testing.T) {
	filters := []request.Filter{
		{Field: "age", Operator: request.FilterOpGt, Value: "18"},
	}
	cols := map[string]string{"age": "u.age"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.age > $1", sql)
	expectArgs(t, []any{"18"}, args)
}

func TestApplyFiltersGte(t *testing.T) {
	filters := []request.Filter{
		{Field: "age", Operator: request.FilterOpGte, Value: "18"},
	}
	cols := map[string]string{"age": "u.age"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.age >= $1", sql)
	expectArgs(t, []any{"18"}, args)
}

func TestApplyFiltersLt(t *testing.T) {
	filters := []request.Filter{
		{Field: "age", Operator: request.FilterOpLt, Value: "65"},
	}
	cols := map[string]string{"age": "u.age"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.age < $1", sql)
	expectArgs(t, []any{"65"}, args)
}

func TestApplyFiltersLte(t *testing.T) {
	filters := []request.Filter{
		{Field: "age", Operator: request.FilterOpLte, Value: "65"},
	}
	cols := map[string]string{"age": "u.age"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.age <= $1", sql)
	expectArgs(t, []any{"65"}, args)
}

func TestApplyFiltersIn(t *testing.T) {
	filters := []request.Filter{
		{Field: "status", Operator: request.FilterOpIn, Value: "active,pending,review"},
	}
	cols := map[string]string{"status": "u.status"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.status IN ($1, $2, $3)", sql)
	expectArgs(t, []any{"active", "pending", "review"}, args)
}

func TestApplyFiltersUnknownField(t *testing.T) {
	filters := []request.Filter{
		{Field: "unknown", Operator: request.FilterOpEq, Value: "val"},
	}
	cols := map[string]string{"status": "u.status"}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u", sql)
	expectArgs(t, nil, args)
}

func TestApplyFiltersMultiple(t *testing.T) {
	filters := []request.Filter{
		{Field: "status", Operator: request.FilterOpEq, Value: "active"},
		{Field: "age", Operator: request.FilterOpGte, Value: "18"},
	}
	cols := map[string]string{
		"status": "u.status",
		"age":    "u.age",
	}
	sql, args := Select("id").From("users u").ApplyFilters(filters, cols).Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.status = $1 AND u.age >= $2", sql)
	expectArgs(t, []any{"active", "18"}, args)
}

func TestApplyFiltersEmpty(t *testing.T) {
	sql, args := Select("id").From("users").ApplyFilters(nil, nil).Build()
	expectSQL(t, "SELECT id FROM users", sql)
	expectArgs(t, nil, args)
}

func TestIntegrationFull(t *testing.T) {
	p := request.Pagination{Page: 2, PerPage: 10, Offset: 10}
	sorts := []request.SortField{
		{Field: "name", Direction: request.SortAsc},
	}
	filters := []request.Filter{
		{Field: "status", Operator: request.FilterOpEq, Value: "active"},
	}
	cols := map[string]string{
		"name":   "u.name",
		"status": "u.status",
	}

	sql, args := Select("u.id", "u.name", "u.email").
		From("users u").
		LeftJoin("profiles p", "p.user_id = u.id").
		Where("u.active = $1", true).
		ApplyFilters(filters, cols).
		ApplySort(sorts, cols).
		ApplyPagination(p).
		Build()

	expectSQL(t, "SELECT u.id, u.name, u.email FROM users u LEFT JOIN profiles p ON p.user_id = u.id WHERE u.active = $1 AND u.status = $2 ORDER BY u.name ASC LIMIT 10 OFFSET 10", sql)
	expectArgs(t, []any{true, "active"}, args)
}

func TestStandaloneFunctions(t *testing.T) {
	sb := Select("id").From("users")
	p := request.Pagination{Page: 1, PerPage: 20, Offset: 0}
	result := ApplyPagination(sb, p)
	sql, _ := result.Build()
	expectSQL(t, "SELECT id FROM users LIMIT 20 OFFSET 0", sql)
}

func TestStandaloneApplySort(t *testing.T) {
	sb := Select("id").From("users u")
	sorts := []request.SortField{{Field: "name", Direction: request.SortAsc}}
	cols := map[string]string{"name": "u.name"}
	result := ApplySort(sb, sorts, cols)
	sql, _ := result.Build()
	expectSQL(t, "SELECT id FROM users u ORDER BY u.name ASC", sql)
}

func TestStandaloneApplyFilters(t *testing.T) {
	sb := Select("id").From("users u")
	filters := []request.Filter{{Field: "status", Operator: request.FilterOpEq, Value: "active"}}
	cols := map[string]string{"status": "u.status"}
	result := ApplyFilters(sb, filters, cols)
	sql, args := result.Build()
	expectSQL(t, "SELECT id FROM users u WHERE u.status = $1", sql)
	expectArgs(t, []any{"active"}, args)
}
