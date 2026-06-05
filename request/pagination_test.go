package request

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequest(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

// --- Multi-param detection ---

func TestPaginateDefaultPerPage(t *testing.T) {
	r := newRequest("/items?page=2&per_page=15")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 15 {
		t.Errorf("expected PerPage=15, got %d", p.PerPage)
	}
}

func TestPaginatePageSizeParam(t *testing.T) {
	r := newRequest("/items?page=1&page_size=30")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 30 {
		t.Errorf("expected PerPage=30, got %d", p.PerPage)
	}
}

func TestPaginateLimitParam(t *testing.T) {
	r := newRequest("/items?page=1&limit=10")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 10 {
		t.Errorf("expected PerPage=10, got %d", p.PerPage)
	}
}

func TestPaginatePerPageTakesPrecedence(t *testing.T) {
	// per_page comes first in the fallback list, so it wins.
	r := newRequest("/items?per_page=5&page_size=10&limit=15")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 5 {
		t.Errorf("expected PerPage=5, got %d", p.PerPage)
	}
}

func TestPaginateNoPerPageUsesDefault(t *testing.T) {
	r := newRequest("/items?page=3")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20, got %d", p.PerPage)
	}
}

func TestPaginateWithConfigCustomPerPageParams(t *testing.T) {
	cfg := PaginationConfig{
		DefaultPage:    1,
		DefaultPerPage: 20,
		MaxPerPage:     100,
		PageParam:      "page",
		PerPageParams:  []string{"size", "per_page"},
	}
	r := newRequest("/items?page=1&size=25")
	p, err := PaginateWithConfig(r, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.PerPage != 25 {
		t.Errorf("expected PerPage=25, got %d", p.PerPage)
	}
}

// --- Navigation helpers ---

func TestNavigationHelpers(t *testing.T) {
	p := Pagination{Page: 3, PerPage: 10, Offset: 20}

	if !p.HasNext(50) {
		t.Error("expected HasNext(50) = true")
	}
	if !p.HasPrevious() {
		t.Error("expected HasPrevious = true")
	}
	if p.IsFirstPage() {
		t.Error("expected IsFirstPage = false")
	}
	if p.IsLastPage(50) {
		t.Error("expected IsLastPage(50) = false")
	}
	if got := p.TotalPages(50); got != 5 {
		t.Errorf("TotalPages(50): expected 5, got %d", got)
	}
	if got := p.NextPage(); got != 4 {
		t.Errorf("NextPage: expected 4, got %d", got)
	}
	if got := p.PreviousPage(); got != 2 {
		t.Errorf("PreviousPage: expected 2, got %d", got)
	}
}

func TestNavigationFirstPage(t *testing.T) {
	p := Pagination{Page: 1, PerPage: 10, Offset: 0}

	if !p.IsFirstPage() {
		t.Error("expected IsFirstPage = true")
	}
	if p.HasPrevious() {
		t.Error("expected HasPrevious = false on first page")
	}
	if got := p.PreviousPage(); got != 1 {
		t.Errorf("PreviousPage on first page: expected 1, got %d", got)
	}
}

func TestNavigationLastPage(t *testing.T) {
	p := Pagination{Page: 5, PerPage: 10, Offset: 40}

	if !p.IsLastPage(50) {
		t.Error("expected IsLastPage(50) = true")
	}
	if p.HasNext(50) {
		t.Error("expected HasNext(50) = false on last page")
	}
}

func TestTotalPagesRoundsUp(t *testing.T) {
	p := Pagination{Page: 1, PerPage: 10}
	if got := p.TotalPages(51); got != 6 {
		t.Errorf("TotalPages(51) with PerPage=10: expected 6, got %d", got)
	}
}

func TestTotalPagesZeroPerPage(t *testing.T) {
	p := Pagination{Page: 1, PerPage: 0}
	if got := p.TotalPages(100); got != 0 {
		t.Errorf("TotalPages with PerPage=0: expected 0, got %d", got)
	}
}

// --- Overflow-safe offset ---

func TestOverflowSafeOffset(t *testing.T) {
	r := newRequest("/items?page=999999&per_page=100")
	p, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	expected := int64(999999-1) * int64(100)
	if p.Offset != expected {
		t.Errorf("expected Offset=%d, got %d", expected, p.Offset)
	}
}

func TestOffsetIsInt64(t *testing.T) {
	expected := int64(500000-1) * int64(100)
	r := newRequest("/items?page=500000&per_page=100")
	parsed, err := Paginate(r)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Offset != expected {
		t.Errorf("expected Offset=%d, got %d", expected, parsed.Offset)
	}
}

// --- SQL helpers ---

func TestSQLClause(t *testing.T) {
	p := Pagination{Page: 3, PerPage: 20, Offset: 40}
	if got := p.SQLClause(); got != "LIMIT 20 OFFSET 40" {
		t.Errorf("SQLClause: expected %q, got %q", "LIMIT 20 OFFSET 40", got)
	}
}

func TestSQLClauseMySQL(t *testing.T) {
	p := Pagination{Page: 3, PerPage: 20, Offset: 40}
	if got := p.SQLClauseMySQL(); got != "LIMIT 40, 20" {
		t.Errorf("SQLClauseMySQL: expected %q, got %q", "LIMIT 40, 20", got)
	}
}

func TestSQLClauseFirstPage(t *testing.T) {
	p := Pagination{Page: 1, PerPage: 10, Offset: 0}
	if got := p.SQLClause(); got != "LIMIT 10 OFFSET 0" {
		t.Errorf("SQLClause first page: expected %q, got %q", "LIMIT 10 OFFSET 0", got)
	}
}

// --- Invalid input ---

func TestPaginateInvalidPerPage(t *testing.T) {
	r := newRequest("/items?page=1&per_page=abc")
	_, err := Paginate(r)
	if err == nil {
		t.Error("expected error for non-integer per_page")
	}
}

func TestPaginateInvalidPage(t *testing.T) {
	r := newRequest("/items?page=abc")
	_, err := Paginate(r)
	if err == nil {
		t.Error("expected error for non-integer page")
	}
}
