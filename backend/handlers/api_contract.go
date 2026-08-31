package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	defaultPage  = 1
	defaultLimit = 50
	maxLimit     = 100
)

// Pagination describes the bounded pagination contract for V1 collection APIs.
// Page is 1-based and Limit is capped server-side.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PaginatedResponse is the additive V1 response shape for collection APIs.
// It preserves the existing success/message/data envelope and adds pagination
// metadata without breaking existing response consumers.
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// parsePagination applies the canonical V1 page/limit contract.
// Missing parameters use safe defaults. Invalid or out-of-range values are
// rejected instead of silently changing the caller's requested page.
func parsePagination(r *http.Request) (Pagination, error) {
	page := defaultPage
	limit := defaultLimit

	if raw := r.URL.Query().Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return Pagination{}, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxLimit {
			return Pagination{}, fmt.Errorf("limit must be between 1 and %d", maxLimit)
		}
		limit = parsed
	}

	return Pagination{Page: page, Limit: limit}, nil
}

func paginationOffset(p Pagination) int {
	return (p.Page - 1) * p.Limit
}
