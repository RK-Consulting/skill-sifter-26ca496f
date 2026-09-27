package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRequirementMatchesRequiresTenant(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/requirements/1/matches", nil)
	rec := httptest.NewRecorder()
	GetRequirementMatches(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestGetRequirementCandidateMatchRejectsInvalidRequirementID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/requirements/not-an-id/matches/1", nil)
	req = req.WithContext(context.WithValue(context.Background(), "tenantID", "tenant_a"))
	rec := httptest.NewRecorder()
	GetRequirementCandidateMatch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetRequirementCandidateMatchRejectsInvalidCandidateID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/requirements/1/matches/not-an-id", nil)
	req = req.WithContext(context.WithValue(context.Background(), "tenantID", "tenant_a"))
	rec := httptest.NewRecorder()
	GetRequirementCandidateMatch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
