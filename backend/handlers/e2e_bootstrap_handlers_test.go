package handlers

import (
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
