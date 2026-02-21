package request

import (
	"net/http"
)

// Pagination holds parsed pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
	Offset  int
}

// PaginationConfig configures pagination defaults and limits.
type PaginationConfig struct {
	DefaultPage    int
	DefaultPerPage int
	MaxPerPage     int
	PageParam      string
	PerPageParam   string
}

// DefaultPaginationConfig returns sensible defaults.
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPage:    1,
		DefaultPerPage: 20,
		MaxPerPage:     100,
		PageParam:      "page",
		PerPageParam:   "per_page",
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
	if cfg.PerPageParam == "" {
		cfg.PerPageParam = "per_page"
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

	perPage, err := q.IntRange(cfg.PerPageParam, cfg.DefaultPerPage, 1, cfg.MaxPerPage)
	if err != nil {
		return Pagination{}, err
	}

	return Pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
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
