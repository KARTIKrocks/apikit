// Package sqlbuilder provides a fluent, type-safe SQL query builder for PostgreSQL,
// MySQL, and SQLite.
//
// By default it uses PostgreSQL $1, $2, ... numbered placeholders. Set a dialect
// to target MySQL or SQLite (? positional placeholders) via .SetDialect() or the
// SelectWith/InsertWith/UpdateWith/DeleteWith constructors.
//
// It integrates with the request package's Pagination, SortField, and Filter types
// for API-driven queries.
//
// # Quick Start
//
//	sql, args := sqlbuilder.Select("id", "name", "email").
//	    From("users").
//	    Where("active = $1", true).
//	    OrderBy("name ASC").
//	    Limit(20).
//	    Build()
//	// sql:  "SELECT id, name, email FROM users WHERE active = $1 ORDER BY name ASC LIMIT 20"
//	// args: [true]
//
// # Placeholder Rebasing
//
// Each Where call uses $1-relative placeholders. At Build time they are
// automatically rebased to globally correct positions:
//
//	sql, args := sqlbuilder.Select("id").From("users").
//	    Where("status = $1", "active").
//	    Where("age > $1", 18).
//	    Build()
//	// sql:  "SELECT id FROM users WHERE status = $1 AND age > $2"
//	// args: ["active", 18]
//
// # Integration with request Package
//
//	pg, _ := request.Paginate(r)
//	sorts, _ := request.ParseSort(r, sortCfg)
//	filters, _ := request.ParseFilters(r, filterCfg)
//
//	cols := map[string]string{"name": "u.name", "created_at": "u.created_at"}
//
//	sql, args := sqlbuilder.Select("u.id", "u.name").
//	    From("users u").
//	    Where("u.active = $1", true).
//	    ApplyFilters(filters, cols).
//	    ApplySort(sorts, cols).
//	    ApplyPagination(pg).
//	    Build()
package sqlbuilder
