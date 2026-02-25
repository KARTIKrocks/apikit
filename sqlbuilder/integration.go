package sqlbuilder

import (
	"github.com/KARTIKrocks/apikit/request"
)

// ApplyPagination applies request.Pagination to a SelectBuilder.
// This is a standalone function alternative to the method on SelectBuilder.
func ApplyPagination(sb *SelectBuilder, p request.Pagination) *SelectBuilder {
	return sb.ApplyPagination(p)
}

// ApplySort applies request.SortField slices to a SelectBuilder.
// This is a standalone function alternative to the method on SelectBuilder.
func ApplySort(sb *SelectBuilder, sorts []request.SortField, allowedColumns map[string]string) *SelectBuilder {
	return sb.ApplySort(sorts, allowedColumns)
}

// ApplyFilters applies request.Filter slices to a SelectBuilder.
// This is a standalone function alternative to the method on SelectBuilder.
func ApplyFilters(sb *SelectBuilder, filters []request.Filter, allowedColumns map[string]string) *SelectBuilder {
	return sb.ApplyFilters(filters, allowedColumns)
}

