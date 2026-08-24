package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// setupClientRequirementTestDB stands up (or reuses) the tables needed to
// exercise the Client/Requirement domain, matching the migration's real
// schema. Skips (does not fail) if no test database is reachable.
func setupClientRequirementTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDB := setupIsolationTestDB(t) // reuses companies/tenant_a/tenant_b setup from tenant_isolation_test.go

	statements := []string{
		`CREATE TABLE IF NOT EXISTS clients (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'prospect',
			contact_email VARCHAR(255),
			contact_phone VARCHAR(50),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT clients_status_valid CHECK (status IN ('prospect', 'active', 'inactive'))
		)`,
		`CREATE TABLE IF NOT EXISTS requirements (
			id SERIAL PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
			client_id INTEGER NOT NULL REFERENCES clients(id),
			title VARCHAR(255) NOT NULL,
			department VARCHAR(100),
			location VARCHAR(100),
			work_arrangement VARCHAR(50),
			status VARCHAR(50) NOT NULL DEFAULT 'draft',
			opened_date TIMESTAMP,
			description TEXT,
			required_skills TEXT,
			experience_required VARCHAR(100),
			compensation VARCHAR(255),
			headcount INTEGER NOT NULL DEFAULT 1,
			language_requirement VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT requirements_status_valid CHECK (status IN ('draft', 'open', 'on_hold', 'filled', 'cancelled')),
			CONSTRAINT requirements_headcount_positive CHECK (headcount > 0)
		)`,
	}
	for _, s := range statements {
		if _, err := testDB.Exec(s); err != nil {
			t.Fatalf("client/requirement test schema setup failed: %v\nstatement: %s", err, s)
		}
	}

	testDB.Exec(`DELETE FROM requirements WHERE tenant_id IN ('tenant_a', 'tenant_b')`)
	testDB.Exec(`DELETE FROM clients WHERE tenant_id IN ('tenant_a', 'tenant_b')`)

	return testDB
}

// --- Client domain tests ---

func TestClient_LifecycleAndValidation(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	t.Run("AddClient defaults status to prospect", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Acme Corp"})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/clients", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddClient(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"prospect"`)) {
			t.Errorf("expected default status prospect, got: %s", rec.Body.String())
		}
	})

	t.Run("AddClient rejects invalid status", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Bad Status Co", "status": "on_fire"})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/clients", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddClient(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", rec.Code)
		}
	})

	t.Run("AddClient rejects empty name", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": ""})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/clients", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddClient(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("client-supplied tenantId in body is ignored", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Injected Co", "tenantId": "tenant_b"})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/clients", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddClient(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}
		var tenantID string
		testDB.QueryRow(`SELECT tenant_id FROM clients WHERE name = 'Injected Co'`).Scan(&tenantID)
		if tenantID != "tenant_a" {
			t.Errorf("client tenant_id = %q, want %q — client-supplied tenantId overrode authenticated tenant", tenantID, "tenant_a")
		}
	})

	t.Run("UpdateClient can transition status prospect -> active", func(t *testing.T) {
		var id int
		testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('Transition Co', 'prospect', 'tenant_a') RETURNING id`).Scan(&id)

		body, _ := json.Marshal(map[string]string{"name": "Transition Co", "status": "active"})
		req := isoCtx(httptest.NewRequest("PUT", "/api/v1/clients/x", bytes.NewReader(body)), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
		rec := httptest.NewRecorder()
		UpdateClient(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200. Body: %s", rec.Code, rec.Body.String())
		}

		var status string
		testDB.QueryRow(`SELECT status FROM clients WHERE id = $1`, id).Scan(&status)
		if status != "active" {
			t.Errorf("status = %q, want %q", status, "active")
		}
	})
}

func TestClient_CrossTenantIsolation(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBClientID int
	testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('B Client Co', 'active', 'tenant_b') RETURNING id`).Scan(&tenantBClientID)

	t.Run("cross-tenant read returns 404", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/v1/clients/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBClientID)})
		rec := httptest.NewRecorder()
		GetClientByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("cross-tenant list never includes another tenant's client", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/v1/clients", nil), "tenant_a")
		rec := httptest.NewRecorder()
		GetClients(rec, req)
		if bytes.Contains(rec.Body.Bytes(), []byte("B Client Co")) {
			t.Error("Tenant A's client list leaked Tenant B's client")
		}
	})

	t.Run("cross-tenant update affects zero rows", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Hijacked", "status": "inactive"})
		req := isoCtx(httptest.NewRequest("PUT", "/api/v1/clients/x", bytes.NewReader(body)), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBClientID)})
		rec := httptest.NewRecorder()
		UpdateClient(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var name string
		testDB.QueryRow(`SELECT name FROM clients WHERE id = $1`, tenantBClientID).Scan(&name)
		if name != "B Client Co" {
			t.Errorf("Tenant B's client was modified via cross-tenant update: %q", name)
		}
	})

	t.Run("cross-tenant delete does not remove the row", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("DELETE", "/api/v1/clients/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBClientID)})
		rec := httptest.NewRecorder()
		DeleteClient(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`, tenantBClientID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's client was deleted by a Tenant A request")
		}
	})
}

// --- Requirement domain tests ---

func TestRequirement_LifecycleAndValidation(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var clientID int
	testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('Req Test Client', 'active', 'tenant_a') RETURNING id`).Scan(&clientID)

	t.Run("AddRequirement defaults status to draft and headcount to 1", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"title": "Backend Engineer", "clientId": clientID})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/requirements", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddRequirement(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"draft"`)) {
			t.Errorf("expected default status draft, got: %s", rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"headcount":1`)) {
			t.Errorf("expected default headcount 1, got: %s", rec.Body.String())
		}
	})

	t.Run("AddRequirement rejects invalid status", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"title": "Bad Status Req", "clientId": clientID, "status": "vibing"})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/requirements", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddRequirement(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", rec.Code)
		}
	})

	t.Run("AddRequirement rejects missing clientId", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"title": "No Client Req"})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/requirements", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddRequirement(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("AddRequirement rejects a client_id belonging to another tenant", func(t *testing.T) {
		var otherClientID int
		testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('Tenant B Client', 'active', 'tenant_b') RETURNING id`).Scan(&otherClientID)

		body, _ := json.Marshal(map[string]interface{}{"title": "Cross Tenant Client Req", "clientId": otherClientID})
		req := isoCtx(httptest.NewRequest("POST", "/api/v1/requirements", bytes.NewReader(body)), "tenant_a")
		rec := httptest.NewRecorder()
		AddRequirement(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404 (must not attach a requirement to another tenant's client)", rec.Code)
		}
	})

	t.Run("UpdateRequirement can transition status draft -> open -> filled", func(t *testing.T) {
		var id int
		testDB.QueryRow(`INSERT INTO requirements (client_id, title, status, tenant_id) VALUES ($1, 'Transition Req', 'draft', 'tenant_a') RETURNING id`, clientID).Scan(&id)

		for _, status := range []string{"open", "filled"} {
			body, _ := json.Marshal(map[string]interface{}{"title": "Transition Req", "clientId": clientID, "status": status, "headcount": 1})
			req := isoCtx(httptest.NewRequest("PUT", "/api/v1/requirements/x", bytes.NewReader(body)), "tenant_a")
			req = mux.SetURLVars(req, map[string]string{"id": itoa(id)})
			rec := httptest.NewRecorder()
			UpdateRequirement(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("transition to %q: status = %d, want 200. Body: %s", status, rec.Code, rec.Body.String())
			}
		}

		var finalStatus string
		testDB.QueryRow(`SELECT status FROM requirements WHERE id = $1`, id).Scan(&finalStatus)
		if finalStatus != "filled" {
			t.Errorf("final status = %q, want %q", finalStatus, "filled")
		}
	})
}

func TestRequirement_CrossTenantIsolation(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var tenantBClientID, tenantBReqID int
	testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('B Client', 'active', 'tenant_b') RETURNING id`).Scan(&tenantBClientID)
	testDB.QueryRow(`INSERT INTO requirements (client_id, title, status, tenant_id) VALUES ($1, 'B Requirement', 'open', 'tenant_b') RETURNING id`, tenantBClientID).Scan(&tenantBReqID)

	t.Run("cross-tenant read by known ID returns 404", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/v1/requirements/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBReqID)})
		rec := httptest.NewRecorder()
		GetRequirementByID(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("cross-tenant list never includes another tenant's requirement", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/v1/requirements", nil), "tenant_a")
		rec := httptest.NewRecorder()
		GetRequirements(rec, req)
		if bytes.Contains(rec.Body.Bytes(), []byte("B Requirement")) {
			t.Error("Tenant A's requirement list leaked Tenant B's requirement")
		}
	})

	t.Run("cross-tenant update affects zero rows", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"title": "Hijacked", "clientId": tenantBClientID, "status": "cancelled", "headcount": 1})
		req := isoCtx(httptest.NewRequest("PUT", "/api/v1/requirements/x", bytes.NewReader(body)), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBReqID)})
		rec := httptest.NewRecorder()
		UpdateRequirement(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("Tenant A was able to update Tenant B's requirement")
		}
		var title string
		testDB.QueryRow(`SELECT title FROM requirements WHERE id = $1`, tenantBReqID).Scan(&title)
		if title != "B Requirement" {
			t.Errorf("Tenant B's requirement was modified via cross-tenant update: %q", title)
		}
	})

	t.Run("cross-tenant delete does not remove the row", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("DELETE", "/api/v1/requirements/x", nil), "tenant_a")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBReqID)})
		rec := httptest.NewRecorder()
		DeleteRequirement(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
		var stillExists bool
		testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM requirements WHERE id = $1)`, tenantBReqID).Scan(&stillExists)
		if !stillExists {
			t.Error("Tenant B's requirement was deleted by a Tenant A request")
		}
	})

	t.Run("own-tenant access succeeds", func(t *testing.T) {
		req := isoCtx(httptest.NewRequest("GET", "/api/v1/requirements/x", nil), "tenant_b")
		req = mux.SetURLVars(req, map[string]string{"id": itoa(tenantBReqID)})
		rec := httptest.NewRecorder()
		GetRequirementByID(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 (Tenant B reading its own requirement must succeed). Body: %s", rec.Code, rec.Body.String())
		}
	})
}

// TestGetByID_HandlesNullOptionalFields guards against the class of bug
// found while writing these tests: scanning a nullable DB column straight
// into a Go string/time.Time field errors on NULL. A client/requirement
// created with only its required fields set (every optional field NULL)
// must still be readable.
func TestGetByID_HandlesNullOptionalFields(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var clientID int
	testDB.QueryRow(`INSERT INTO clients (name, tenant_id) VALUES ('Minimal Client', 'tenant_a') RETURNING id`).Scan(&clientID)

	clientReq := isoCtx(httptest.NewRequest("GET", "/api/v1/clients/x", nil), "tenant_a")
	clientReq = mux.SetURLVars(clientReq, map[string]string{"id": itoa(clientID)})
	clientRec := httptest.NewRecorder()
	GetClientByID(clientRec, clientReq)
	if clientRec.Code != http.StatusOK {
		t.Errorf("GetClientByID with all-NULL optional fields: status = %d, want 200. Body: %s", clientRec.Code, clientRec.Body.String())
	}

	var reqID int
	testDB.QueryRow(`INSERT INTO requirements (client_id, title, tenant_id) VALUES ($1, 'Minimal Requirement', 'tenant_a') RETURNING id`, clientID).Scan(&reqID)

	reqReq := isoCtx(httptest.NewRequest("GET", "/api/v1/requirements/x", nil), "tenant_a")
	reqReq = mux.SetURLVars(reqReq, map[string]string{"id": itoa(reqID)})
	reqRec := httptest.NewRecorder()
	GetRequirementByID(reqRec, reqReq)
	if reqRec.Code != http.StatusOK {
		t.Errorf("GetRequirementByID with all-NULL optional fields: status = %d, want 200. Body: %s", reqRec.Code, reqRec.Body.String())
	}

	listReq := isoCtx(httptest.NewRequest("GET", "/api/v1/requirements", nil), "tenant_a")
	listRec := httptest.NewRecorder()
	GetRequirements(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Errorf("GetRequirements with all-NULL optional fields: status = %d, want 200. Body: %s", listRec.Code, listRec.Body.String())
	}
}

// TestDeleteClient_WithRequirements_IsRejected verifies a client with
// existing requirements cannot be silently deleted, cascading data loss
// into its requirements.
func TestDeleteClient_WithRequirements_IsRejected(t *testing.T) {
	testDB := setupClientRequirementTestDB(t)
	defer testDB.Close()
	db.DB = testDB

	var clientID int
	testDB.QueryRow(`INSERT INTO clients (name, status, tenant_id) VALUES ('Client With Reqs', 'active', 'tenant_a') RETURNING id`).Scan(&clientID)
	testDB.Exec(`INSERT INTO requirements (client_id, title, status, tenant_id) VALUES ($1, 'Dependent Req', 'open', 'tenant_a')`, clientID)

	req := isoCtx(httptest.NewRequest("DELETE", "/api/v1/clients/x", nil), "tenant_a")
	req = mux.SetURLVars(req, map[string]string{"id": itoa(clientID)})
	rec := httptest.NewRecorder()
	DeleteClient(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409 (client has existing requirements)", rec.Code)
	}

	var stillExists bool
	testDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`, clientID).Scan(&stillExists)
	if !stillExists {
		t.Error("client was deleted despite having a dependent requirement")
	}
}
