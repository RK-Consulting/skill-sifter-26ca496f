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

func TestRoutesDoNotExposeLegacyCompanyUsersAlias(t *testing.T) {
	router := mux.NewRouter()
	setupProtectedRoutes(router)

	found := false
	err := router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err == nil && path == "/api/company-users" {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking routes: %v", err)
	}
	if found {
		t.Fatal("legacy /api/company-users route must not be registered")
	}
}
