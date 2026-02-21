package response

import (
	"net/http"
	"time"
)

// PageMeta represents offset-based pagination metadata.
type PageMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
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
// TotalPages is computed automatically. If perPage is <= 0, TotalPages defaults to 0.
func NewPageMeta(page, perPage, total int) PageMeta {
	var totalPages int
	if perPage > 0 {
		totalPages = total / perPage
		if total%perPage > 0 {
			totalPages++
		}
	}
	return PageMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
