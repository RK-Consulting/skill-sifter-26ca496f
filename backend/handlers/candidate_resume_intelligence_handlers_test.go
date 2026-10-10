package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCandidateResumeIntelligenceRequiresTenant(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/candidates/1/resume-intelligence", nil)
	rec := httptest.NewRecorder()
	GetCandidateResumeIntelligence(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestGetCandidateResumeIntelligenceRejectsInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/candidates/not-an-id/resume-intelligence", nil)
	req = req.WithContext(context.WithValue(context.Background(), "tenantID", "tenant_a"))
	rec := httptest.NewRecorder()
	GetCandidateResumeIntelligence(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
