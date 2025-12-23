package repository

import (
	"math"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`      // 1-indexed page number
	PageSize int `json:"page_size"` // Number of items per page
}

// PaginationResult represents paginated results
type PaginationResult struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() {
	p.Page = max(p.Page, 1)
	p.PageSize = max(p.PageSize, 1)
	p.PageSize = min(p.PageSize, 100)
}

// Offset calculates the database offset
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the limit for the query
func (p *PaginationParams) Limit() int {
	return p.PageSize
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(params PaginationParams, totalItems int64) PaginationResult {
	params.Validate()

	totalPages := int(math.Ceil(float64(totalItems) / float64(params.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}
