package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaginationDefaults(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/clients", nil)
	pagination, err := parsePagination(r)
	if err != nil {
		t.Fatalf("parsePagination() error = %v", err)
	}
	if pagination.Page != 1 || pagination.Limit != 50 {
		t.Fatalf("parsePagination() = %+v, want page=1 limit=50", pagination)
	}
}

func TestParsePaginationAcceptsBoundedValues(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/clients?page=3&limit=25", nil)
	pagination, err := parsePagination(r)
	if err != nil {
		t.Fatalf("parsePagination() error = %v", err)
	}
	if pagination.Page != 3 || pagination.Limit != 25 {
		t.Fatalf("parsePagination() = %+v, want page=3 limit=25", pagination)
	}
}

func TestParsePaginationRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "zero page", url: "/api/v1/clients?page=0"},
		{name: "negative page", url: "/api/v1/clients?page=-1"},
		{name: "non numeric page", url: "/api/v1/clients?page=abc"},
		{name: "zero limit", url: "/api/v1/clients?limit=0"},
		{name: "limit above maximum", url: "/api/v1/clients?limit=101"},
		{name: "non numeric limit", url: "/api/v1/clients?limit=abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			if _, err := parsePagination(r); err == nil {
				t.Fatal("parsePagination() error = nil, want validation error")
			}
		})
	}
}

func TestParsePaginationRejectsOverflowingPage(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/clients?page=9223372036854775807&limit=100", nil)
	if _, err := parsePagination(r); err == nil {
		t.Fatal("parsePagination() error = nil, want overflow validation error")
	}
}

func TestPaginationOffset(t *testing.T) {
	got := paginationOffset(Pagination{Page: 3, Limit: 25})
	if got != 50 {
		t.Fatalf("paginationOffset() = %d, want 50", got)
	}
}
