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

func TestProductionCORSRejectsDevelopmentAndPreviewOrigins(t *testing.T) {
	t.Setenv("SKILLSIFTER_ENV", "production")
	handler := setupCORS().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, origin := range []string{"http://localhost:5173", "https://pr-123.skill-sifter-26ca496f.pages.dev"} {
		req := httptest.NewRequest(http.MethodGet, "/api/health-check", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("production CORS allowed %q with Access-Control-Allow-Origin %q", origin, got)
		}
	}
}

func TestProductionCORSAllowsCanonicalSite(t *testing.T) {
	t.Setenv("SKILLSIFTER_ENV", "production")
	handler := setupCORS().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health-check", nil)
	req.Header.Set("Origin", "https://skillsifter.in")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://skillsifter.in" {
		t.Fatalf("canonical origin allowed header = %q, want %q", got, "https://skillsifter.in")
	}
}
