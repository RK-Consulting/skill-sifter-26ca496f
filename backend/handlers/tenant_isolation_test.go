package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// setupIsolationTestDB returns the clean tenant-plane baseline prepared by TestMain.
// Control-plane state lives in handlerControlDB; tenant handlers receive the
// tenant database through request context.
func setupIsolationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if handlerTenantDB == nil || handlerControlDB == nil {
		t.Fatal("handler test databases are not initialized")
	}

	cleanupTables := []string{
		"recruitment_submission_feedback", "recruitment_submissions",
		"recruitment_screenings", "recruitment_selections",
		"recruitment_offers", "recruitment_joinings", "recruitment_billings",
		"interviews", "requirements", "clients",
		"candidate_language_expertise", "candidate_expertise",
		"resumes", "candidates", "users", "audit_events",
	}
	for _, table := range cleanupTables {
		if _, err := handlerTenantDB.Exec("DELETE FROM " + table + " WHERE tenant_id IN ('tenant_a','tenant_b')"); err != nil {
			t.Fatalf("tenant baseline cleanup for %s failed: %v", table, err)
		}
	}
	return handlerTenantDB
}

func getenvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// isoCtx builds a request context matching what AuthMiddleware would set:
// tenantID is the authoritative scoping value; companyName/role/userID are
// carried for display/compatibility.
func isoCtx(req *http.Request, tenantID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenantID", tenantID)
	ctx = context.WithValue(ctx, "companyName", tenantID)
	ctx = context.WithValue(ctx, "role", "admin")
	ctx = context.WithValue(ctx, "userID", 1)
	ctx = db.WithTenantDB(ctx, handlerTenantDB)
	return req.WithContext(ctx)
}

// TestTenantIsolation_Candidates covers ADR 0001's required matrix (read,
// update, delete, lookup-by-known-id, tenant-scoped list) for the
// candidates domain: Tenant A must never be able to read, modify, or delete
// Tenant B's data, even knowing Tenant B's exact resource ID.
func TestTenantIsolation_Candidates(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()

	var tenantBCandidateID int
	err := testDB.QueryRow(
		`INSERT INTO candidates (name, email, phone, position, location, experience, currentctc, expectedctc, noticeperiod, jobdescription, tenant_id, company_name)
		 VALUES ('B Candidate', 'bcand@test.com', '', '', '', '', '', '', '', '', 'tenant_b', 'tenant_b') RETURNING id`,
	).Scan(&tenantBCandidateID)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	t.Run("cross-tenant read by known ID returns 404, not the record", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/candidates/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBCandidateID)})
		rec := httptest.NewRecorder()
		GetCandidateByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404 (must not disclose Tenant B's resource to Tenant A)", rec.Code)
		}
	})

	t.Run("cross-tenant list never includes another tenant's rows", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/candidates", nil), "tenant_a")
		rec := httptest.NewRecorder()
		GetCandidates(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("B Candidate")) {
			t.Error("Tenant A's candidate list leaked Tenant B's candidate")
		}
	})

	t.Run("cross-tenant update affects zero rows and returns 404", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Hijacked", "email": "hijacked@test.com"})
		req := isoCtx(httptest.NewRequest("PUT", "/api/candidates/x", bytes.NewReader(body)), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBCandidateID)})
		rec := httptest.NewRecorder()
		UpdateCandidate(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("Tenant A was able to update Tenant B's candidate")
		}

		var name string
		testDB.QueryRow(`SELECT name FROM candidates WHERE id = $1`, tenantBCandidateID).Scan(&name)
		if name != "B Candidate" {
			t.Errorf("Tenant B's candidate name changed to %q via a cross-tenant update", name)
		}
	})

	t.Run("cross-tenant delete affects zero rows and returns 404", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("DELETE", "/api/candidates/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBCandidateID)})
		rec := httptest.NewRecorder()
		DeleteCandidate(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}

		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM candidates WHERE id = $1)`, tenantBCandidateID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's candidate was deleted by a Tenant A request")
		}
	})

	t.Run("own-tenant read succeeds", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/candidates/x", nil), "tenant_b")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBCandidateID)})
		rec := httptest.NewRecorder()
		GetCandidateByID(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 (Tenant B reading its own candidate must succeed). Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("AddCandidate ignores a client-supplied tenant_id/company_name and uses the authenticated tenant", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "New Candidate", "email": "new@test.com",
			"tenantId": "tenant_b", "companyName": "tenant_b", // attempted override
		})
		req := isoCtx(httptest.NewRequest("POST", "/api/candidates", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddCandidate(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}

		var tenantID string
		err := testDB.QueryRow(`SELECT tenant_id FROM candidates WHERE email = 'new@test.com'`).Scan(&tenantID)
		if err != nil {
			t.Fatalf("could not read back created candidate: %v", err)
		}
		if tenantID != "tenant_a" {
			t.Errorf("candidate tenant_id = %q, want %q — client-supplied tenantId in the request body overrode the authenticated tenant", tenantID, "tenant_a")
		}
	})
}

// TestTenantIsolation_Users covers the users domain, including the
// admin-only user-management endpoints, and specifically the "known target
// ID in another tenant" case for Update/Delete.
func TestTenantIsolation_Users(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()

	var tenantBUserID int
	err := testDB.QueryRow(
		`INSERT INTO users (username, email, password, role, tenant_id) VALUES ('buser', 'buser@test.com', 'x', 'recruiter', 'tenant_b') RETURNING id`,
	).Scan(&tenantBUserID)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	t.Run("GetUsers for tenant_a never returns tenant_b's user", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/users", nil), "tenant_a")
		rec := httptest.NewRecorder()
		GetUsers(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("buser")) {
			t.Error("Tenant A's user list leaked Tenant B's user")
		}
	})

	t.Run("DeleteUser with Tenant B's known user ID from Tenant A context is denied", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("DELETE", "/api/users/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBUserID)})
		rec := httptest.NewRecorder()
		DeleteUser(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, tenantBUserID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's user was deleted via a Tenant A admin request")
		}
	})

	t.Run("CreateUser ignores a client-supplied tenantId/companyName", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "newuser", "email": "newuser@test.com", "password": "x", "role": "recruiter",
			"tenantId": "tenant_b", "companyName": "tenant_b", // attempted override
		})
		req := isoCtx(httptest.NewRequest("POST", "/api/users", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		CreateUser(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}

		var tenantID string
		testDB.QueryRow(`SELECT tenant_id FROM users WHERE email = 'newuser@test.com'`).Scan(&tenantID)
		if tenantID != "tenant_a" {
			t.Errorf("created user tenant_id = %q, want %q — client-supplied value overrode authenticated tenant", tenantID, "tenant_a")
		}
	})
}

// TestTenantIsolation_BusinessDev covers the business_dev domain.
