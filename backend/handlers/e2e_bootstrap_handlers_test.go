package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResetE2ESmokeTenantDataOptions(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/e2e/reset", nil)
	rec := httptest.NewRecorder()

	ResetE2ESmokeTenantData(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected OPTIONS /api/e2e/reset to return 204, got %d", rec.Code)
	}
}

func TestResetE2ESmokeTenantDataRejectsNonSmokeTenant(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/e2e/reset", nil)
	req = req.WithContext(context.WithValue(req.Context(), "tenantID", "tenant_other"))
	req = req.WithContext(context.WithValue(req.Context(), "role", "admin"))
	rec := httptest.NewRecorder()

	ResetE2ESmokeTenantData(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected non-smoke tenant reset to return 403, got %d", rec.Code)
	}
}
