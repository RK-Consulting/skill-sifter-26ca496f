package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestPublicRoutesDoNotExposeE2EBootstrap(t *testing.T) {
	router := mux.NewRouter()
	setupPublicRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/api/e2e/bootstrap", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/e2e/bootstrap status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPublicRoutesDoNotExposeLegacyCompanyUsersAlias(t *testing.T) {
	router := mux.NewRouter()
	setupProtectedRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/company-users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// The route must be absent, not merely protected by a frontend menu.
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/company-users status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
