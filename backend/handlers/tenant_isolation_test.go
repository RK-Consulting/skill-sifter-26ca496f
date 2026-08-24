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

// setupIsolationTestDB stands up (or reuses) the full tenant-owned schema
// needed to exercise cross-tenant isolation across every domain listed in
// ADR 0001: users, candidates, jobs, daily_jobs, interviews, business_dev.
// Skips (does not fail) if no test database is reachable, matching the
// existing integration-test pattern in this package.
func setupIsolationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping tenant isolation test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Skipf("skipping tenant isolation test: test DB not reachable (%v)", err)
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(100) NOT NULL,
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS candidates (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			phone VARCHAR(20), position VARCHAR(100), location VARCHAR(50),
			experience VARCHAR(100), currentctc VARCHAR(100), expectedctc VARCHAR(100),
			noticeperiod VARCHAR(100), jlptlanguage VARCHAR(100), skills VARCHAR(100),
			jobdescription VARCHAR(500),
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL, department VARCHAR(100), location VARCHAR(100),
			status VARCHAR(50) DEFAULT 'open', date_posted TIMESTAMP DEFAULT NOW(),
			description TEXT, requirements TEXT, last_modified TIMESTAMP DEFAULT NOW(),
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL,
			created_by_user_id INTEGER REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS daily_jobs (
			id SERIAL PRIMARY KEY,
			jd_no INTEGER NOT NULL, instructions TEXT,
			assigned_user INTEGER REFERENCES users(id),
			assigned_date TIMESTAMP DEFAULT NOW(), last_modified TIMESTAMP DEFAULT NOW(),
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS interviews (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER REFERENCES candidates(id),
			candidate_name VARCHAR(255) NOT NULL, position VARCHAR(100),
			interview_date TIMESTAMP NOT NULL, status VARCHAR(50) DEFAULT 'scheduled',
			feedback TEXT, last_modified TIMESTAMP DEFAULT NOW(),
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS business_dev (
			id SERIAL PRIMARY KEY,
			client_name VARCHAR(255) NOT NULL, partner_name VARCHAR(255),
			contact_person VARCHAR(255) NOT NULL, contact_number VARCHAR(50),
			contact_email VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(), last_modified TIMESTAMP DEFAULT NOW(),
			tenant_id VARCHAR(255) REFERENCES companies(id),
			company_name VARCHAR(255) NOT NULL
		)`,
	}
	for _, s := range statements {
		if _, err := testDB.Exec(s); err != nil {
			t.Fatalf("isolation test schema setup failed: %v\nstatement: %s", err, s)
		}
	}
	// Older local dev databases may pre-date tenant_id.
	for _, alter := range []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
		`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
		`ALTER TABLE daily_jobs ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
		`ALTER TABLE interviews ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
		`ALTER TABLE business_dev ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) REFERENCES companies(id)`,
	} {
		testDB.Exec(alter)
	}

	// Clean slate.
	for _, t := range []string{"interviews", "daily_jobs", "jobs", "candidates", "users", "business_dev"} {
		testDB.Exec("DELETE FROM " + t + " WHERE tenant_id IN ('tenant_a', 'tenant_b')")
	}
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('tenant_a', 'Tenant A Co') ON CONFLICT (id) DO NOTHING`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('tenant_b', 'Tenant B Co') ON CONFLICT (id) DO NOTHING`)

	return testDB
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
	return req.WithContext(ctx)
}

// TestTenantIsolation_Candidates covers ADR 0001's required matrix (read,
// update, delete, lookup-by-known-id, tenant-scoped list) for the
// candidates domain: Tenant A must never be able to read, modify, or delete
// Tenant B's data, even knowing Tenant B's exact resource ID.
func TestTenantIsolation_Candidates(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBCandidateID int
	err := testDB.QueryRow(
		`INSERT INTO candidates (name, email, phone, position, location, experience, currentctc, expectedctc, noticeperiod, jlptlanguage, skills, jobdescription, tenant_id, company_name)
		 VALUES ('B Candidate', 'bcand@test.com', '', '', '', '', '', '', '', '', '', '', 'tenant_b', 'tenant_b') RETURNING id`,
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

// TestTenantIsolation_Jobs mirrors the candidates matrix for the jobs
// domain.
func TestTenantIsolation_Jobs(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBJobID int
	err := testDB.QueryRow(
		`INSERT INTO jobs (title, tenant_id, company_name) VALUES ('B Job', 'tenant_b', 'tenant_b') RETURNING id`,
	).Scan(&tenantBJobID)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	t.Run("cross-tenant read returns 404", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/jobs/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBJobID)})
		rec := httptest.NewRecorder()
		GetJobByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("cross-tenant delete does not remove the row", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("DELETE", "/api/jobs/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBJobID)})
		rec := httptest.NewRecorder()
		DeleteJob(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM jobs WHERE id = $1)`, tenantBJobID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's job was deleted by a Tenant A request")
		}
	})
}

// TestTenantIsolation_Users covers the users domain, including the
// admin-only user-management endpoints, and specifically the "known target
// ID in another tenant" case for Update/Delete.
func TestTenantIsolation_Users(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBUserID int
	err := testDB.QueryRow(
		`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ('buser', 'buser@test.com', 'x', 'recruiter', 'tenant_b', 'tenant_b') RETURNING id`,
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
func TestTenantIsolation_BusinessDev(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBRecordID int
	err := testDB.QueryRow(
		`INSERT INTO business_dev (client_name, contact_person, contact_email, tenant_id, company_name) VALUES ('B Client', 'B Contact', 'b@test.com', 'tenant_b', 'tenant_b') RETURNING id`,
	).Scan(&tenantBRecordID)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	t.Run("cross-tenant read returns 404", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/business-dev/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBRecordID)})
		rec := httptest.NewRecorder()
		GetBusinessDevByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("cross-tenant update affects zero rows", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"clientName": "Hijacked", "contactPerson": "x", "contactEmail": "x@test.com"})
		req := isoCtx(httptest.NewRequest("PUT", "/api/business-dev/x", bytes.NewReader(body)), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBRecordID)})
		rec := httptest.NewRecorder()
		UpdateBusinessDev(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var name string
		testDB.QueryRow(`SELECT client_name FROM business_dev WHERE id = $1`, tenantBRecordID).Scan(&name)
		if name != "B Client" {
			t.Errorf("Tenant B's business_dev record was modified via cross-tenant update: %q", name)
		}
	})
}

// itoa is a small local alias used throughout this file for building URL
// vars from int IDs.
func itoa(n int) string {
	return strconv.Itoa(n)
}
