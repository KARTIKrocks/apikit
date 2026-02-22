package request

import (
	"fmt"
	"net/http"
)

// Pagination holds parsed pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
	Offset  int64
}

// HasNext returns true if there are more pages after the current one.
func (p Pagination) HasNext(total int) bool {
	return p.Page < p.TotalPages(total)
}

// HasPrevious returns true if there are pages before the current one.
func (p Pagination) HasPrevious() bool {
	return p.Page > 1
}

// IsFirstPage returns true if the current page is the first page.
func (p Pagination) IsFirstPage() bool {
	return p.Page == 1
}

// IsLastPage returns true if the current page is the last page.
func (p Pagination) IsLastPage(total int) bool {
	return p.Page >= p.TotalPages(total)
}

// TotalPages returns the total number of pages for the given total item count.
func (p Pagination) TotalPages(total int) int {
	if p.PerPage <= 0 {
		return 0
	}
	tp := total / p.PerPage
	if total%p.PerPage > 0 {
		tp++
	}
	return tp
}

// NextPage returns the next page number (current page + 1).
// It does not exceed math.MaxInt; callers should use HasNext to check bounds.
func (p Pagination) NextPage() int {
	return p.Page + 1
}

// PreviousPage returns the previous page number, minimum 1.
func (p Pagination) PreviousPage() int {
	if p.Page <= 1 {
		return 1
	}
	return p.Page - 1
}

// SQLClause returns a PostgreSQL-style LIMIT/OFFSET clause.
//
//	p.SQLClause() // "LIMIT 20 OFFSET 40"
func (p Pagination) SQLClause() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.PerPage, p.Offset)
}

// SQLClauseMySQL returns a MySQL-style LIMIT clause.
//
//	p.SQLClauseMySQL() // "LIMIT 40, 20"
func (p Pagination) SQLClauseMySQL() string {
	return fmt.Sprintf("LIMIT %d, %d", p.Offset, p.PerPage)
}

// PaginationConfig configures pagination defaults and limits.
type PaginationConfig struct {
	DefaultPage    int
	DefaultPerPage int
	MaxPerPage     int
	PageParam string
	// PerPageParams lists parameter names to try in order for per-page detection.
	// The first matching non-empty parameter wins. If empty, defaults to
	// ["per_page", "page_size", "limit"].
	PerPageParams []string
}

// DefaultPaginationConfig returns sensible defaults.
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPage:    1,
		DefaultPerPage: 20,
		MaxPerPage:     100,
		PageParam:      "page",
		PerPageParams:  []string{"per_page", "page_size", "limit"},
	}
}

var globalPaginationConfig = DefaultPaginationConfig()

// SetPaginationConfig sets the global pagination configuration.
func SetPaginationConfig(cfg PaginationConfig) {
	if cfg.DefaultPage <= 0 {
		cfg.DefaultPage = 1
	}
	if cfg.DefaultPerPage <= 0 {
		cfg.DefaultPerPage = 20
	}
	if cfg.MaxPerPage <= 0 {
		cfg.MaxPerPage = 100
	}
	if cfg.PageParam == "" {
		cfg.PageParam = "page"
	}
	if len(cfg.PerPageParams) == 0 {
		cfg.PerPageParams = []string{"per_page", "page_size", "limit"}
	}
	globalPaginationConfig = cfg
}

// Paginate parses pagination parameters from the request query string.
//
//	GET /users?page=2&per_page=25
//
//	p, err := request.Paginate(r)
//	// p.Page=2, p.PerPage=25, p.Offset=25
func Paginate(r *http.Request) (Pagination, error) {
	return PaginateWithConfig(r, globalPaginationConfig)
}

// PaginateWithConfig parses pagination with custom config.
func PaginateWithConfig(r *http.Request, cfg PaginationConfig) (Pagination, error) {
	q := QueryFrom(r)

	page, err := q.IntRange(cfg.PageParam, cfg.DefaultPage, 1, 999999)
	if err != nil {
		return Pagination{}, err
	}

	// Multi-param detection: try each param name in order.
	params := cfg.PerPageParams
	if len(params) == 0 {
		params = []string{"per_page", "page_size", "limit"}
	}

	perPage := cfg.DefaultPerPage
	for _, param := range params {
		val := q.values.Get(param)
		if val != "" {
			parsed, parseErr := q.IntRange(param, cfg.DefaultPerPage, 1, cfg.MaxPerPage)
			if parseErr != nil {
				return Pagination{}, parseErr
			}
			perPage = parsed
			break
		}
	}

	return Pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  int64(page-1) * int64(perPage),
	}, nil
}

// CursorPagination holds cursor-based pagination parameters.
type CursorPagination struct {
	Cursor string
	Limit  int
}

// CursorPaginationConfig configures cursor-based pagination.
type CursorPaginationConfig struct {
	DefaultLimit int
	MaxLimit     int
	CursorParam  string
	LimitParam   string
}

// DefaultCursorPaginationConfig returns sensible defaults.
func DefaultCursorPaginationConfig() CursorPaginationConfig {
	return CursorPaginationConfig{
		DefaultLimit: 20,
		MaxLimit:     100,
		CursorParam:  "cursor",
		LimitParam:   "limit",
	}
}

// PaginateCursor parses cursor-based pagination from the request.
//
//	GET /events?cursor=eyJpZCI6MTIzfQ&limit=25
func PaginateCursor(r *http.Request) (CursorPagination, error) {
	return PaginateCursorWithConfig(r, DefaultCursorPaginationConfig())
}

// PaginateCursorWithConfig parses cursor pagination with custom config.
func PaginateCursorWithConfig(r *http.Request, cfg CursorPaginationConfig) (CursorPagination, error) {
	q := QueryFrom(r)

	cursor := q.String(cfg.CursorParam, "")
	limit, err := q.IntRange(cfg.LimitParam, cfg.DefaultLimit, 1, cfg.MaxLimit)
	if err != nil {
		return CursorPagination{}, err
	}

	return CursorPagination{
		Cursor: cursor,
		Limit:  limit,
	}, nil
}
