package response

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// --- NewPageMeta ---

func TestNewPageMeta(t *testing.T) {
	m := NewPageMeta(2, 20, 100)
	if m.Page != 2 {
		t.Errorf("Page: expected 2, got %d", m.Page)
	}
	if m.PerPage != 20 {
		t.Errorf("PerPage: expected 20, got %d", m.PerPage)
	}
	if m.Total != 100 {
		t.Errorf("Total: expected 100, got %d", m.Total)
	}
	if m.TotalPages != 5 {
		t.Errorf("TotalPages: expected 5, got %d", m.TotalPages)
	}
	if !m.HasNext {
		t.Error("expected HasNext=true for page 2 of 5")
	}
	if !m.HasPrevious {
		t.Error("expected HasPrevious=true for page 2")
	}
}

func TestNewPageMetaFirstPage(t *testing.T) {
	m := NewPageMeta(1, 10, 50)
	if m.HasPrevious {
		t.Error("expected HasPrevious=false on first page")
	}
	if !m.HasNext {
		t.Error("expected HasNext=true on first page with more pages")
	}
}

func TestNewPageMetaLastPage(t *testing.T) {
	m := NewPageMeta(5, 10, 50)
	if m.HasNext {
		t.Error("expected HasNext=false on last page")
	}
	if !m.HasPrevious {
		t.Error("expected HasPrevious=true on last page")
	}
}

func TestNewPageMetaSinglePage(t *testing.T) {
	m := NewPageMeta(1, 10, 5)
	if m.TotalPages != 1 {
		t.Errorf("TotalPages: expected 1, got %d", m.TotalPages)
	}
	if m.HasNext {
		t.Error("expected HasNext=false for single page")
	}
	if m.HasPrevious {
		t.Error("expected HasPrevious=false for single page")
	}
}

func TestNewPageMetaZeroPerPage(t *testing.T) {
	m := NewPageMeta(1, 0, 100)
	if m.TotalPages != 0 {
		t.Errorf("TotalPages: expected 0 for zero per_page, got %d", m.TotalPages)
	}
}

func TestNewPageMetaRoundsUp(t *testing.T) {
	m := NewPageMeta(1, 10, 51)
	if m.TotalPages != 6 {
		t.Errorf("TotalPages: expected 6 for 51 items with per_page=10, got %d", m.TotalPages)
	}
}

// --- SetLinkHeader ---

func TestSetLinkHeaderMiddlePage(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users", 2, 20, 100)

	link := w.Header().Get("Link")
	if link == "" {
		t.Fatal("expected Link header to be set")
	}
	if !strings.Contains(link, `rel="next"`) {
		t.Error("expected next rel in Link header")
	}
	if !strings.Contains(link, `rel="prev"`) {
		t.Error("expected prev rel in Link header")
	}
	if !strings.Contains(link, `rel="first"`) {
		t.Error("expected first rel in Link header")
	}
	if !strings.Contains(link, `rel="last"`) {
		t.Error("expected last rel in Link header")
	}
	if !strings.Contains(link, "page=3") {
		t.Error("expected next page=3 in Link header")
	}
	if !strings.Contains(link, "page=1") {
		t.Error("expected prev/first page=1 in Link header")
	}
	if !strings.Contains(link, "page=5") {
		t.Error("expected last page=5 in Link header")
	}
}

func TestSetLinkHeaderFirstPage(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users", 1, 20, 100)

	link := w.Header().Get("Link")
	if strings.Contains(link, `rel="prev"`) {
		t.Error("first page should not have prev link")
	}
	if !strings.Contains(link, `rel="next"`) {
		t.Error("first page should have next link")
	}
}

func TestSetLinkHeaderLastPage(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users", 5, 20, 100)

	link := w.Header().Get("Link")
	if strings.Contains(link, `rel="next"`) {
		t.Error("last page should not have next link")
	}
	if !strings.Contains(link, `rel="prev"`) {
		t.Error("last page should have prev link")
	}
}

func TestSetLinkHeaderSinglePage(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users", 1, 20, 10)

	link := w.Header().Get("Link")
	if strings.Contains(link, `rel="next"`) {
		t.Error("single page should not have next link")
	}
	if strings.Contains(link, `rel="prev"`) {
		t.Error("single page should not have prev link")
	}
	if !strings.Contains(link, `rel="first"`) {
		t.Error("single page should have first link")
	}
	if !strings.Contains(link, `rel="last"`) {
		t.Error("single page should have last link")
	}
}

func TestSetLinkHeaderZeroTotal(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users", 1, 20, 0)

	link := w.Header().Get("Link")
	if link != "" {
		t.Error("expected no Link header for zero total")
	}
}

func TestSetLinkHeaderPreservesExistingQueryParams(t *testing.T) {
	w := httptest.NewRecorder()
	SetLinkHeader(w, "https://api.example.com/users?status=active", 2, 10, 50)

	link := w.Header().Get("Link")
	if !strings.Contains(link, "status=active") {
		t.Error("expected existing query params to be preserved")
	}
}
