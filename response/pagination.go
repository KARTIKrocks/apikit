package response

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PageMeta represents offset-based pagination metadata.
type PageMeta struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"per_page"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// CursorMeta represents cursor-based pagination metadata.
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Total      int    `json:"total,omitempty"`
}

// Paginated writes a paginated response with offset-based metadata.
//
//	response.Paginated(w, users, response.PageMeta{
//	    Page: 1, PerPage: 20, Total: 150, TotalPages: 8,
//	})
func Paginated(w http.ResponseWriter, data any, meta PageMeta) {
	write(w, http.StatusOK, Envelope{
		Success:   true,
		Data:      data,
		Meta:      &meta,
		Timestamp: time.Now().Unix(),
	})
}

// PaginatedWithMessage writes a paginated response with a message.
func PaginatedWithMessage(w http.ResponseWriter, message string, data any, meta PageMeta) {
	write(w, http.StatusOK, Envelope{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      &meta,
		Timestamp: time.Now().Unix(),
	})
}

// CursorPaginated writes a cursor-paginated response.
//
//	response.CursorPaginated(w, events, response.CursorMeta{
//	    NextCursor: "eyJpZCI6MTAwfQ==",
//	    HasMore:    true,
//	    Total:      500,
//	})
func CursorPaginated(w http.ResponseWriter, data any, meta CursorMeta) {
	write(w, http.StatusOK, Envelope{
		Success:   true,
		Data:      data,
		Meta:      &meta,
		Timestamp: time.Now().Unix(),
	})
}

// NewPageMeta creates a PageMeta from page, perPage, and total count.
// TotalPages, HasNext, and HasPrevious are computed automatically.
// If perPage is <= 0, TotalPages defaults to 0.
func NewPageMeta(page, perPage, total int) PageMeta {
	var totalPages int
	if perPage > 0 {
		totalPages = total / perPage
		if total%perPage > 0 {
			totalPages++
		}
	}
	return PageMeta{
		Page:        page,
		PerPage:     perPage,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// SetLinkHeader sets an RFC 5988 Link header for paginated responses.
//
//	response.SetLinkHeader(w, "https://api.example.com/users", 2, 20, 100)
//	// Link: <https://api.example.com/users?page=3&per_page=20>; rel="next",
//	//       <https://api.example.com/users?page=1&per_page=20>; rel="prev",
//	//       <https://api.example.com/users?page=1&per_page=20>; rel="first",
//	//       <https://api.example.com/users?page=5&per_page=20>; rel="last"
func SetLinkHeader(w http.ResponseWriter, baseURL string, page, perPage, total int) {
	var totalPages int
	if perPage > 0 {
		totalPages = total / perPage
		if total%perPage > 0 {
			totalPages++
		}
	}
	if totalPages == 0 {
		return
	}

	var links []string

	buildLink := func(p int, rel string) string {
		u, err := url.Parse(baseURL)
		if err != nil {
			// Fallback: use baseURL as-is.
			return fmt.Sprintf("<%s?page=%d&per_page=%d>; rel=\"%s\"", baseURL, p, perPage, rel)
		}
		q := u.Query()
		q.Set("page", fmt.Sprintf("%d", p))
		q.Set("per_page", fmt.Sprintf("%d", perPage))
		u.RawQuery = q.Encode()
		return fmt.Sprintf("<%s>; rel=\"%s\"", u.String(), rel)
	}

	if page < totalPages {
		links = append(links, buildLink(page+1, "next"))
	}
	if page > 1 {
		links = append(links, buildLink(page-1, "prev"))
	}
	links = append(links, buildLink(1, "first"))
	links = append(links, buildLink(totalPages, "last"))

	w.Header().Set("Link", strings.Join(links, ", "))
}
