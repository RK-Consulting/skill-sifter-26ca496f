package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/RK-Consulting/skill-sifter/auth"
	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// setupAssignmentHandlerTestDB stands up the tables needed to exercise the
// assignment HTTP handlers. Tenant literals are namespaced ("ah_tenant_a"/
// "ah_tenant_b", both id AND display name — companies.name is UNIQUE) to
// avoid the cross-package/cross-file test-pollution issue found and fixed
// in the domain/assignment package's own tests (checkpoint 2).
func setupAssignmentHandlerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvOr2("TEST_DB_HOST", "localhost")
	port := getenvOr2("TEST_DB_PORT", "5432")
	user := getenvOr2("TEST_DB_USER", "postgres")
	password := getenvOr2("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr2("TEST_DB_NAME", "skillsifter_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping assignment handler test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Skipf("skipping assignment handler test: test DB not reachable (%v)", err)
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY, username VARCHAR(100) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL, role VARCHAR(100) NOT NULL,
			tenant_id VARCHAR(255) REFERENCES companies(id), company_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS candidates (
			id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL,
			phone VARCHAR(20), position VARCHAR(100), location VARCHAR(50), experience VARCHAR(100),
			currentctc VARCHAR(100), expectedctc VARCHAR(100), noticeperiod VARCHAR(100),
			jlptlanguage VARCHAR(100), skills VARCHAR(100),
			tenant_id VARCHAR(255) REFERENCES companies(id), company_name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			CONSTRAINT ah_candidates_status_valid CHECK (status IN ('active','inactive','blacklisted','archived'))
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) REFERENCES companies(id), name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'prospect'
		)`,
		`CREATE TABLE IF NOT EXISTS requirements (
			id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) REFERENCES companies(id),
			client_id INTEGER REFERENCES clients(id), title VARCHAR(255) NOT NULL,
			location VARCHAR(100), work_arrangement VARCHAR(50), description TEXT,
			required_skills TEXT, experience_required VARCHAR(100), compensation VARCHAR(255),
			headcount INTEGER NOT NULL DEFAULT 1, language_requirement VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'draft'
		)`,
		`CREATE TABLE IF NOT EXISTS recruitment_assignments (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
			candidate_id INTEGER NOT NULL REFERENCES candidates(id),
			requirement_id INTEGER NOT NULL REFERENCES requirements(id),
			status VARCHAR(50) NOT NULL DEFAULT 'draft',
			created_by_user_id INTEGER NOT NULL REFERENCES users(id),
			owner_user_id INTEGER NOT NULL REFERENCES users(id),
			candidate_snapshot JSONB,
			requirement_snapshot JSONB,
			snapshot_created_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT ah_recruitment_assignments_status_valid CHECK (
				status IN ('draft','screening','submitted','interviewing','offered','joined','rejected','withdrawn')
			),
			CONSTRAINT ah_recruitment_assignments_candidate_requirement_unique UNIQUE (candidate_id, requirement_id)
		)`,
	}
	for _, s := range statements {
		if _, err := testDB.Exec(s); err != nil {
			t.Fatalf("assignment handler test schema setup failed: %v\nstatement: %s", err, s)
		}
	}

	for _, tbl := range []string{"recruitment_assignments", "requirements", "clients", "candidates", "users"} {
		testDB.Exec("DELETE FROM " + tbl + " WHERE tenant_id IN ('ah_tenant_a', 'ah_tenant_b')")
	}
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('ah_tenant_a', 'Assignment Handler Tenant A') ON CONFLICT (id) DO NOTHING`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('ah_tenant_b', 'Assignment Handler Tenant B') ON CONFLICT (id) DO NOTHING`)

	return testDB
}

func getenvOr2(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type ahFixtures struct {
	userID        int
	candidateID   int
	requirementID int
}

func seedAHFixtures(t *testing.T, testDB *sql.DB, tenantID string) ahFixtures {
	t.Helper()
	var f ahFixtures
	testDB.QueryRow(`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ($1, $2, 'x', 'manager', $3, $3) RETURNING id`,
		"user_"+tenantID, tenantID+"user@test.com", tenantID).Scan(&f.userID)
	testDB.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ('Test Candidate', $1, $2, $2, 'active') RETURNING id`,
		tenantID+"cand@test.com", tenantID).Scan(&f.candidateID)
	var clientID int
	testDB.QueryRow(`INSERT INTO clients (name, tenant_id) VALUES ('Test Client', $1) RETURNING id`, tenantID).Scan(&clientID)
	testDB.QueryRow(`INSERT INTO requirements (client_id, title, tenant_id) VALUES ($1, 'Test Req', $2) RETURNING id`, clientID, tenantID).Scan(&f.requirementID)
	return f
}

// ahCtx builds a request context matching what AuthMiddleware would set.
func ahCtx(req *http.Request, tenantID string, userID int, role string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenantID", tenantID)
	ctx = context.WithValue(ctx, "companyName", tenantID)
	ctx = context.WithValue(ctx, "userID", userID)
	ctx = context.WithValue(ctx, "role", role)
	return req.WithContext(ctx)
}

// --- Success paths ---

func TestAssignmentHandlers_CreateGetListUpdateDelete_Success(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")

	// Create
	body, _ := json.Marshal(map[string]int{"candidateId": f.candidateID, "requirementId": f.requirementID})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
	rec := httptest.NewRecorder()
	AddAssignment(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("AddAssignment status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
	}
	var createResp struct {
		Data assignmentResponse `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	if createResp.Data.Status != "draft" {
		t.Errorf("created status = %q, want %q", createResp.Data.Status, "draft")
	}
	if createResp.Data.OwnerUserID != f.userID {
		t.Errorf("owner defaulted to %d, want %d (creator)", createResp.Data.OwnerUserID, f.userID)
	}
	id := createResp.Data.ID

	// Get by ID
	getReq := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments/x", nil), "ah_tenant_a", f.userID, "manager")
	getReq = mux.SetURLVars(getReq, map[string]string{"id": itoa(id)})
	getRec := httptest.NewRecorder()
	GetAssignmentByID(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GetAssignmentByID status = %d, want 200. Body: %s", getRec.Code, getRec.Body.String())
	}

	// List
	listReq := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments", nil), "ah_tenant_a", f.userID, "manager")
	listRec := httptest.NewRecorder()
	GetAssignments(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("GetAssignments status = %d, want 200. Body: %s", listRec.Code, listRec.Body.String())
	}
	var listResp struct {
		Data []assignmentResponse `json:"data"`
	}
	json.Unmarshal(listRec.Body.Bytes(), &listResp)
	if len(listResp.Data) != 1 {
		t.Errorf("list length = %d, want 1", len(listResp.Data))
	}

	// Update (reassign owner)
	var newOwnerID int
	testDB.QueryRow(`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ('newowner', 'newowner@test.com', 'x', 'manager', 'ah_tenant_a', 'ah_tenant_a') RETURNING id`).Scan(&newOwnerID)
	updateBody, _ := json.Marshal(map[string]int{"ownerUserId": newOwnerID})
	updateReq := ahCtx(httptest.NewRequest("PUT", "/api/v1/assignments/x", bytes.NewReader(updateBody)), "ah_tenant_a", f.userID, "manager")
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": itoa(id)})
	updateRec := httptest.NewRecorder()
	UpdateAssignment(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("UpdateAssignment status = %d, want 200. Body: %s", updateRec.Code, updateRec.Body.String())
	}
	var updateResp struct {
		Data assignmentResponse `json:"data"`
	}
	json.Unmarshal(updateRec.Body.Bytes(), &updateResp)
	if updateResp.Data.OwnerUserID != newOwnerID {
		t.Errorf("owner after update = %d, want %d", updateResp.Data.OwnerUserID, newOwnerID)
	}
	if updateResp.Data.Status != "draft" {
		t.Errorf("status changed via UpdateAssignment to %q, want unchanged %q (lifecycle transitions are out of scope this checkpoint)", updateResp.Data.Status, "draft")
	}

	// Delete
	delReq := ahCtx(httptest.NewRequest("DELETE", "/api/v1/assignments/x", nil), "ah_tenant_a", f.userID, "manager")
	delReq = mux.SetURLVars(delReq, map[string]string{"id": itoa(id)})
	delRec := httptest.NewRecorder()
	DeleteAssignment(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("DeleteAssignment status = %d, want 200. Body: %s", delRec.Code, delRec.Body.String())
	}

	var stillExists bool
	testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM recruitment_assignments WHERE id = $1)`, id).Scan(&stillExists)
	if stillExists {
		t.Error("assignment still exists after DeleteAssignment")
	}
}

// --- Unauthenticated access ---

func TestAssignmentHandlers_Unauthenticated(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	os.Setenv("JWT_SECRET", "test-secret-for-unauthenticated-check")

	handlersToCheck := map[string]http.HandlerFunc{
		"GET list":  GetAssignments,
		"POST":      AddAssignment,
		"GET by ID": GetAssignmentByID,
		"PUT":       UpdateAssignment,
		"DELETE":    DeleteAssignment,
	}

	for name, h := range handlersToCheck {
		t.Run(name, func(t *testing.T) {
			protected := auth.AuthMiddleware(h)
			req := httptest.NewRequest("GET", "/api/v1/assignments", nil) // no Authorization header at all
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s without Authorization header: status = %d, want 401", name, rec.Code)
			}
		})
	}
}

// --- Tenant isolation ---

func TestAssignmentHandlers_CrossTenantIsolation(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	fA := seedAHFixtures(t, testDB, "ah_tenant_a")
	fB := seedAHFixtures(t, testDB, "ah_tenant_b")

	var assignmentBID int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_b', $1, $2, $3, $3) RETURNING id`,
		fB.candidateID, fB.requirementID, fB.userID).Scan(&assignmentBID)

	t.Run("cross-tenant read returns 404", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments/x", nil), "ah_tenant_a", fA.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(assignmentBID)})
		rec := httptest.NewRecorder()
		GetAssignmentByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("cross-tenant list never includes another tenant's assignment", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments", nil), "ah_tenant_a", fA.userID, "manager")
		rec := httptest.NewRecorder()
		GetAssignments(rec, req)
		var resp struct {
			Data []assignmentResponse `json:"data"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		for _, a := range resp.Data {
			if a.ID == assignmentBID {
				t.Error("Tenant A's assignment list leaked Tenant B's assignment")
			}
		}
	})

	t.Run("cross-tenant update (owner reassignment) affects zero rows", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"ownerUserId": fA.userID})
		req := ahCtx(httptest.NewRequest("PUT", "/api/v1/assignments/x", bytes.NewReader(body)), "ah_tenant_a", fA.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(assignmentBID)})
		rec := httptest.NewRecorder()
		UpdateAssignment(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var ownerID int
		testDB.QueryRow(`SELECT owner_user_id FROM recruitment_assignments WHERE id = $1`, assignmentBID).Scan(&ownerID)
		if ownerID != fB.userID {
			t.Errorf("Tenant B's assignment owner changed to %d via cross-tenant update, want unchanged %d", ownerID, fB.userID)
		}
	})

	t.Run("cross-tenant delete does not remove the row", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("DELETE", "/api/v1/assignments/x", nil), "ah_tenant_a", fA.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(assignmentBID)})
		rec := httptest.NewRecorder()
		DeleteAssignment(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM recruitment_assignments WHERE id = $1)`, assignmentBID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's assignment was deleted by a Tenant A request")
		}
	})

	t.Run("cross-tenant candidate rejected on create with 404", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"candidateId": fB.candidateID, "requirementId": fA.requirementID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", fA.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("own-tenant access succeeds", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments/x", nil), "ah_tenant_b", fB.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(assignmentBID)})
		rec := httptest.NewRecorder()
		GetAssignmentByID(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 (Tenant B reading its own assignment must succeed)", rec.Code)
		}
	})
}

// --- Role authorization (route-level: managerOnly wrapping) ---

func TestAssignmentRoutes_ManagerOnlyWrapping(t *testing.T) {
	// This directly verifies the same mechanism main.go's route
	// registration uses to wrap mutating assignment handlers
	// (auth.RoleMiddleware("admin","manager"), i.e. what managerOnly()
	// wraps in main.go) — not a new authorization scheme.
	protectedCreate := auth.RoleMiddleware("admin", "manager")(http.HandlerFunc(AddAssignment))

	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader([]byte(`{}`))), "ah_tenant_a", 1, "recruiter")
	rec := httptest.NewRecorder()
	protectedCreate.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("recruiter role on manager/admin-protected AddAssignment: status = %d, want 403", rec.Code)
	}

	req2 := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader([]byte(`{}`))), "ah_tenant_a", 1, "manager")
	rec2 := httptest.NewRecorder()
	protectedCreate.ServeHTTP(rec2, req2)
	if rec2.Code == http.StatusForbidden {
		t.Errorf("manager role on manager/admin-protected AddAssignment was forbidden, want role check to pass (may still fail validation, but not on role)")
	}
}

// --- Nonexistent assignment ---

func TestAssignmentHandlers_NonexistentAssignment(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")

	getReq := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments/x", nil), "ah_tenant_a", f.userID, "manager")
	getReq = mux.SetURLVars(getReq, map[string]string{"id": "999999"})
	getRec := httptest.NewRecorder()
	GetAssignmentByID(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Errorf("GetAssignmentByID nonexistent: status = %d, want 404", getRec.Code)
	}

	delReq := ahCtx(httptest.NewRequest("DELETE", "/api/v1/assignments/x", nil), "ah_tenant_a", f.userID, "manager")
	delReq = mux.SetURLVars(delReq, map[string]string{"id": "999999"})
	delRec := httptest.NewRecorder()
	DeleteAssignment(delRec, delReq)
	if delRec.Code != http.StatusNotFound {
		t.Errorf("DeleteAssignment nonexistent: status = %d, want 404", delRec.Code)
	}

	updateBody, _ := json.Marshal(map[string]int{"ownerUserId": f.userID})
	updateReq := ahCtx(httptest.NewRequest("PUT", "/api/v1/assignments/x", bytes.NewReader(updateBody)), "ah_tenant_a", f.userID, "manager")
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": "999999"})
	updateRec := httptest.NewRecorder()
	UpdateAssignment(updateRec, updateReq)
	if updateRec.Code != http.StatusNotFound {
		t.Errorf("UpdateAssignment nonexistent: status = %d, want 404", updateRec.Code)
	}
}

// --- Malformed requests ---

func TestAssignmentHandlers_MalformedRequests(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")

	t.Run("invalid JSON body on create", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader([]byte("not json"))), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("missing candidateId on create", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"requirementId": f.requirementID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("missing requirementId on create", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"candidateId": f.candidateID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("non-numeric ID in path", func(t *testing.T) {
		req := ahCtx(httptest.NewRequest("GET", "/api/v1/assignments/x", nil), "ah_tenant_a", f.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": "not-a-number"})
		rec := httptest.NewRecorder()
		GetAssignmentByID(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("missing ownerUserId on update", func(t *testing.T) {
		var id int
		testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $3) RETURNING id`,
			f.candidateID, f.requirementID, f.userID).Scan(&id)

		body, _ := json.Marshal(map[string]int{})
		req := ahCtx(httptest.NewRequest("PUT", "/api/v1/assignments/x", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
		rec := httptest.NewRecorder()
		UpdateAssignment(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
}

// --- Lifecycle transition (checkpoint 4) ---

func TestTransitionAssignment_Success(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")
	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $3) RETURNING id`,
		f.candidateID, f.requirementID, f.userID).Scan(&id)

	body, _ := json.Marshal(map[string]string{"status": "screening"})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
	rec := httptest.NewRecorder()
	TransitionAssignment(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data assignmentResponse `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Status != "screening" {
		t.Errorf("status = %q, want %q", resp.Data.Status, "screening")
	}

	var persisted string
	testDB.QueryRow(`SELECT status FROM recruitment_assignments WHERE id = $1`, id).Scan(&persisted)
	if persisted != "screening" {
		t.Errorf("persisted status = %q, want %q", persisted, "screening")
	}
}

func TestTransitionAssignment_IllegalTransitionReturns409(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")
	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $3) RETURNING id`,
		f.candidateID, f.requirementID, f.userID).Scan(&id)

	// draft -> offered skips screening/submitted/interviewing entirely.
	body, _ := json.Marshal(map[string]string{"status": "offered"})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
	rec := httptest.NewRecorder()
	TransitionAssignment(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409. Body: %s", rec.Code, rec.Body.String())
	}

	var persisted string
	testDB.QueryRow(`SELECT status FROM recruitment_assignments WHERE id = $1`, id).Scan(&persisted)
	if persisted != "draft" {
		t.Errorf("persisted status = %q after rejected transition, want unchanged %q", persisted, "draft")
	}
}

func TestTransitionAssignment_TerminalStateProtection(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")

	for _, terminal := range []string{"joined", "rejected", "withdrawn"} {
		t.Run(terminal, func(t *testing.T) {
			var candID int
			testDB.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ($1, $2, 'ah_tenant_a', 'ah_tenant_a', 'active') RETURNING id`,
				"Cand "+terminal, terminal+"@test.com").Scan(&candID)
			var id int
			testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $4, $4) RETURNING id`,
				candID, f.requirementID, terminal, f.userID).Scan(&id)

			body, _ := json.Marshal(map[string]string{"status": "screening"})
			req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
			req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
			rec := httptest.NewRecorder()
			TransitionAssignment(rec, req)

			if rec.Code != http.StatusConflict {
				t.Errorf("status = %d, want 409 (terminal state %q must reject further transitions)", rec.Code, terminal)
			}

			var persisted string
			testDB.QueryRow(`SELECT status FROM recruitment_assignments WHERE id = $1`, id).Scan(&persisted)
			if persisted != terminal {
				t.Errorf("persisted status = %q, want unchanged terminal %q", persisted, terminal)
			}
		})
	}
}

func TestTransitionAssignment_UnrecognizedStatusReturns400(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")
	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $3) RETURNING id`,
		f.candidateID, f.requirementID, f.userID).Scan(&id)

	for _, status := range []string{"", "cancelled", "APPROVED", "Screening"} {
		t.Run(status, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"status": status})
			req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
			req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
			rec := httptest.NewRecorder()
			TransitionAssignment(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status=%q: response code = %d, want 400", status, rec.Code)
			}
		})
	}
}

func TestTransitionAssignment_CrossTenantReturns404(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	fA := seedAHFixtures(t, testDB, "ah_tenant_a")
	fB := seedAHFixtures(t, testDB, "ah_tenant_b")
	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_b', $1, $2, $3, $3) RETURNING id`,
		fB.candidateID, fB.requirementID, fB.userID).Scan(&id)

	body, _ := json.Marshal(map[string]string{"status": "screening"})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", fA.userID, "manager")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
	rec := httptest.NewRecorder()
	TransitionAssignment(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}

	var persisted string
	testDB.QueryRow(`SELECT status FROM recruitment_assignments WHERE id = $1`, id).Scan(&persisted)
	if persisted != "draft" {
		t.Errorf("Tenant B's assignment status = %q after cross-tenant transition attempt, want unchanged %q", persisted, "draft")
	}
}

func TestTransitionAssignment_RoleAuthorization(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")
	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, $3, $3) RETURNING id`,
		f.candidateID, f.requirementID, f.userID).Scan(&id)

	protected := auth.RoleMiddleware("admin", "manager")(http.HandlerFunc(TransitionAssignment))

	body, _ := json.Marshal(map[string]string{"status": "screening"})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "recruiter")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("recruiter role: status = %d, want 403", rec.Code)
	}
}

func TestTransitionAssignment_Unauthenticated(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	protected := auth.AuthMiddleware(http.HandlerFunc(TransitionAssignment))
	req := httptest.NewRequest("POST", "/api/v1/assignments/1/transition", bytes.NewReader([]byte(`{"status":"screening"}`)))
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// --- Snapshot capture at submission (checkpoint 5) ---

func TestTransitionAssignment_SnapshotCapturedAtSubmission(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")
	testDB.Exec(`UPDATE candidates SET name = 'HTTP Snapshot Candidate' WHERE id = $1`, f.candidateID)
	testDB.Exec(`UPDATE requirements SET title = 'HTTP Snapshot Requirement' WHERE id = $1`, f.requirementID)

	var id int
	testDB.QueryRow(`INSERT INTO recruitment_assignments (tenant_id, candidate_id, requirement_id, status, created_by_user_id, owner_user_id) VALUES ('ah_tenant_a', $1, $2, 'screening', $3, $3) RETURNING id`,
		f.candidateID, f.requirementID, f.userID).Scan(&id)

	body, _ := json.Marshal(map[string]string{"status": "submitted"})
	req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments/x/transition", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
	rec := httptest.NewRecorder()
	TransitionAssignment(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. Body: %s", rec.Code, rec.Body.String())
	}

	var candidateSnapshot, requirementSnapshot []byte
	var snapshotCreatedAt sql.NullTime
	testDB.QueryRow(`SELECT candidate_snapshot, requirement_snapshot, snapshot_created_at FROM recruitment_assignments WHERE id = $1`, id).
		Scan(&candidateSnapshot, &requirementSnapshot, &snapshotCreatedAt)

	if !snapshotCreatedAt.Valid {
		t.Fatal("snapshot_created_at is NULL after submission via HTTP")
	}
	if !bytes.Contains(candidateSnapshot, []byte("HTTP Snapshot Candidate")) {
		t.Errorf("candidate_snapshot does not contain expected name: %s", candidateSnapshot)
	}
	if !bytes.Contains(requirementSnapshot, []byte("HTTP Snapshot Requirement")) {
		t.Errorf("requirement_snapshot does not contain expected title: %s", requirementSnapshot)
	}
}

// --- Service/repository error propagation ---

func TestAssignmentHandlers_ServiceErrorPropagation(t *testing.T) {
	testDB := setupAssignmentHandlerTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	f := seedAHFixtures(t, testDB, "ah_tenant_a")

	t.Run("ineligible candidate status propagates as 422", func(t *testing.T) {
		var inactiveCandID int
		testDB.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ('Inactive Cand', 'inactive@test.com', 'ah_tenant_a', 'ah_tenant_a', 'inactive') RETURNING id`).Scan(&inactiveCandID)

		body, _ := json.Marshal(map[string]int{"candidateId": inactiveCandID, "requirementId": f.requirementID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("duplicate candidate/requirement pair propagates as 409", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"candidateId": f.candidateID, "requirementId": f.requirementID})
		req1 := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec1 := httptest.NewRecorder()
		AddAssignment(rec1, req1)
		if rec1.Code != http.StatusCreated {
			t.Fatalf("first create status = %d, want 201. Body: %s", rec1.Code, rec1.Body.String())
		}

		req2 := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec2 := httptest.NewRecorder()
		AddAssignment(rec2, req2)
		if rec2.Code != http.StatusConflict {
			t.Errorf("duplicate create status = %d, want 409. Body: %s", rec2.Code, rec2.Body.String())
		}
	})

	t.Run("nonexistent candidate propagates as 404", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"candidateId": 9999999, "requirementId": f.requirementID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("owner user not in tenant propagates as 404", func(t *testing.T) {
		fB := seedAHFixtures(t, testDB, "ah_tenant_b")
		var freshCandID int
		testDB.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ('Fresh Cand', 'fresh@test.com', 'ah_tenant_a', 'ah_tenant_a', 'active') RETURNING id`).Scan(&freshCandID)

		body, _ := json.Marshal(map[string]int{"candidateId": freshCandID, "requirementId": f.requirementID, "ownerUserId": fB.userID})
		req := ahCtx(httptest.NewRequest("POST", "/api/v1/assignments", bytes.NewReader(body)), "ah_tenant_a", f.userID, "manager")
		rec := httptest.NewRecorder()
		AddAssignment(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
