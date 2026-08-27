package assignment

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	_ "github.com/lib/pq"
)

// setupAssignmentTestDB stands up (or reuses) the tables needed to
// exercise the assignment service: companies, users, candidates,
// clients, requirements, recruitment_assignments — matching the real
// migration 008 schema. Skips (does not fail) if no test database is
// reachable, matching this repo's existing integration-test convention.
func setupAssignmentTestDB(t *testing.T) *sql.DB {
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
		t.Skipf("skipping assignment service test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Skipf("skipping assignment service test: test DB not reachable (%v)", err)
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
			CONSTRAINT candidates_status_valid CHECK (status IN ('active','inactive','blacklisted','archived'))
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
			CONSTRAINT recruitment_assignments_status_valid CHECK (
				status IN ('draft','screening','submitted','interviewing','offered','joined','rejected','withdrawn')
			),
			CONSTRAINT recruitment_assignments_candidate_requirement_unique UNIQUE (candidate_id, requirement_id)
		)`,
	}
	for _, s := range statements {
		if _, err := testDB.Exec(s); err != nil {
			t.Fatalf("assignment test schema setup failed: %v\nstatement: %s", err, s)
		}
	}

	// Clean slate for tenant_a / tenant_b.
	testDB.Exec(`DELETE FROM recruitment_assignments WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM requirements WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM clients WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM candidates WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM users WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('asg_tenant_a', 'Assignment Test Tenant A Co') ON CONFLICT (id) DO NOTHING`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('asg_tenant_b', 'Assignment Test Tenant B Co') ON CONFLICT (id) DO NOTHING`)

	return testDB
}

func getenvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// fixtures holds the IDs of a minimal valid tenant_a setup (one user, one
// active candidate, one client+requirement) that tests can build on.
type fixtures struct {
	userID        int
	candidateID   int
	requirementID int
}

var seedFixturesCounter int64

func seedFixtures(t *testing.T, db *sql.DB, tenantID string) fixtures {
	t.Helper()
	// A monotonically increasing suffix so this helper is safe to call
	// multiple times with the same tenantID (e.g. an exhaustive matrix
	// test creating one fresh fixture set per subtest) without colliding
	// on users.email's UNIQUE constraint, which is keyed purely off
	// tenantID otherwise.
	n := atomic.AddInt64(&seedFixturesCounter, 1)
	unique := fmt.Sprintf("%s_%d", tenantID, n)

	var f fixtures
	if err := db.QueryRow(`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ($1, $2, 'x', 'recruiter', $3, $3) RETURNING id`,
		"user_"+unique, unique+"user@test.com", tenantID).Scan(&f.userID); err != nil {
		t.Fatalf("seedFixtures: could not insert user: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ('Test Candidate', $1, $2, $2, 'active') RETURNING id`,
		unique+"cand@test.com", tenantID).Scan(&f.candidateID); err != nil {
		t.Fatalf("seedFixtures: could not insert candidate: %v", err)
	}
	var clientID int
	if err := db.QueryRow(`INSERT INTO clients (name, tenant_id) VALUES ('Test Client', $1) RETURNING id`, tenantID).Scan(&clientID); err != nil {
		t.Fatalf("seedFixtures: could not insert client: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO requirements (client_id, title, tenant_id) VALUES ($1, 'Test Req', $2) RETURNING id`, clientID, tenantID).Scan(&f.requirementID); err != nil {
		t.Fatalf("seedFixtures: could not insert requirement: %v", err)
	}
	return f
}

func TestCreateAssignment_Success(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if a.ID == 0 {
		t.Error("created assignment has no ID")
	}
	if a.Status != StatusDraft {
		t.Errorf("initial status = %q, want %q", a.Status, StatusDraft)
	}
	if a.CreatedByUserID != f.userID {
		t.Errorf("CreatedByUserID = %d, want %d", a.CreatedByUserID, f.userID)
	}
	if a.OwnerUserID != f.userID {
		t.Errorf("OwnerUserID = %d, want %d (should default to actor when not specified)", a.OwnerUserID, f.userID)
	}
}

func TestCreateAssignment_ExplicitOwnerDifferentFromCreator(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	var ownerID int
	db.QueryRow(`INSERT INTO users (username, email, password, role, tenant_id, company_name) VALUES ('owner', 'owner@test.com', 'x', 'manager', 'asg_tenant_a', 'asg_tenant_a') RETURNING id`).Scan(&ownerID)

	svc := NewService(NewPostgresRepository(db), db)
	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID, OwnerUserID: ownerID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if a.CreatedByUserID != f.userID {
		t.Errorf("CreatedByUserID = %d, want %d", a.CreatedByUserID, f.userID)
	}
	if a.OwnerUserID != ownerID {
		t.Errorf("OwnerUserID = %d, want %d", a.OwnerUserID, ownerID)
	}
}

func TestCreateAssignment_RejectsCrossTenantCandidate(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "asg_tenant_a")
	fB := seedFixtures(t, db, "asg_tenant_b")

	svc := NewService(NewPostgresRepository(db), db)
	_, err := svc.CreateAssignment("asg_tenant_a", fA.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fA.requirementID})
	if !errors.Is(err, ErrCandidateNotFound) {
		t.Errorf("error = %v, want ErrCandidateNotFound (Tenant B's candidate must not be usable from Tenant A context)", err)
	}
}

func TestCreateAssignment_RejectsCrossTenantRequirement(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "asg_tenant_a")
	fB := seedFixtures(t, db, "asg_tenant_b")

	svc := NewService(NewPostgresRepository(db), db)
	_, err := svc.CreateAssignment("asg_tenant_a", fA.userID, CreateInput{CandidateID: fA.candidateID, RequirementID: fB.requirementID})
	if !errors.Is(err, ErrRequirementNotFound) {
		t.Errorf("error = %v, want ErrRequirementNotFound (Tenant B's requirement must not be usable from Tenant A context)", err)
	}
}

func TestCreateAssignment_RejectsCrossTenantOwner(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "asg_tenant_a")
	fB := seedFixtures(t, db, "asg_tenant_b")

	svc := NewService(NewPostgresRepository(db), db)
	_, err := svc.CreateAssignment("asg_tenant_a", fA.userID, CreateInput{CandidateID: fA.candidateID, RequirementID: fA.requirementID, OwnerUserID: fB.userID})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("error = %v, want ErrUserNotFound (Tenant B's user must not be assignable as owner from Tenant A context)", err)
	}
}

func TestCreateAssignment_RejectsIneligibleCandidateStatuses(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	for _, status := range []string{"inactive", "blacklisted", "archived"} {
		t.Run(status, func(t *testing.T) {
			var candID int
			db.QueryRow(`INSERT INTO candidates (name, email, tenant_id, company_name, status) VALUES ($1, $2, 'asg_tenant_a', 'asg_tenant_a', $3) RETURNING id`,
				"Candidate "+status, status+"@test.com", status).Scan(&candID)

			_, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: candID, RequirementID: f.requirementID})
			if !errors.Is(err, ErrCandidateNotEligible) {
				t.Errorf("status=%q: error = %v, want ErrCandidateNotEligible", status, err)
			}
		})
	}
}

func TestCreateAssignment_RejectsDuplicateCandidateRequirementPair(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	svc := NewService(NewPostgresRepository(db), db)

	_, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("first CreateAssignment failed: %v", err)
	}

	_, err = svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if !errors.Is(err, ErrDuplicateAssignment) {
		t.Errorf("second CreateAssignment error = %v, want ErrDuplicateAssignment", err)
	}
}

func TestGetAssignment_CrossTenantReturnsNotFound(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)
	created, err := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID})
	if err != nil {
		t.Fatalf("seed CreateAssignment failed: %v", err)
	}

	_, err = svc.GetAssignment("asg_tenant_a", created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound (Tenant A must not read Tenant B's assignment by known ID)", err)
	}
}

func TestListAssignments_TenantScoped(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	fA := seedFixtures(t, db, "asg_tenant_a")
	fB := seedFixtures(t, db, "asg_tenant_b")
	svc := NewService(NewPostgresRepository(db), db)

	if _, err := svc.CreateAssignment("asg_tenant_a", fA.userID, CreateInput{CandidateID: fA.candidateID, RequirementID: fA.requirementID}); err != nil {
		t.Fatalf("tenant_a seed failed: %v", err)
	}
	if _, err := svc.CreateAssignment("asg_tenant_b", fB.userID, CreateInput{CandidateID: fB.candidateID, RequirementID: fB.requirementID}); err != nil {
		t.Fatalf("tenant_b seed failed: %v", err)
	}

	listA, err := svc.ListAssignments("asg_tenant_a")
	if err != nil {
		t.Fatalf("ListAssignments(tenant_a) failed: %v", err)
	}
	if len(listA) != 1 {
		t.Fatalf("tenant_a list length = %d, want 1", len(listA))
	}
	if listA[0].TenantID != "asg_tenant_a" {
		t.Errorf("leaked non-tenant_a assignment: %+v", listA[0])
	}
}

func TestUpdate_PersistsTransitionAndCrossTenantReturnsNotFound(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	repo := NewPostgresRepository(db)
	svc := NewService(repo, db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}

	if err := a.TransitionTo(StatusScreening); err != nil {
		t.Fatalf("TransitionTo failed: %v", err)
	}
	if err := repo.Update(a); err != nil {
		t.Fatalf("repo.Update failed: %v", err)
	}

	reloaded, err := svc.GetAssignment("asg_tenant_a", a.ID)
	if err != nil {
		t.Fatalf("GetAssignment after update failed: %v", err)
	}
	if reloaded.Status != StatusScreening {
		t.Errorf("persisted status = %q, want %q", reloaded.Status, StatusScreening)
	}

	// Cross-tenant update attempt: a tenant_b-scoped Assignment struct
	// pointed at tenant_a's row must not succeed.
	forged := &Assignment{ID: a.ID, TenantID: "asg_tenant_b", Status: StatusRejected, OwnerUserID: f.userID}
	if err := repo.Update(forged); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-tenant Update error = %v, want ErrNotFound", err)
	}

	// Confirm tenant_a's row was NOT modified by the cross-tenant attempt.
	unchanged, _ := svc.GetAssignment("asg_tenant_a", a.ID)
	if unchanged.Status != StatusScreening {
		t.Errorf("tenant_a's assignment status = %q after cross-tenant update attempt, want unchanged %q", unchanged.Status, StatusScreening)
	}
}

// TestUpdate_NilSnapshotFieldsPersistAsNull guards against a real bug found
// while testing this checkpoint: a nil []byte passed directly to lib/pq for
// a JSONB column does not become SQL NULL (it errors with "invalid input
// syntax for type json"), unlike an untyped nil interface{}. Every
// Update() call before formal submission has nil snapshot fields, so this
// path runs on every ordinary status transition, not just submission.
func TestUpdate_NilSnapshotFieldsPersistAsNull(t *testing.T) {
	db := setupAssignmentTestDB(t)
	defer db.Close()

	f := seedFixtures(t, db, "asg_tenant_a")
	repo := NewPostgresRepository(db)
	svc := NewService(repo, db)

	a, err := svc.CreateAssignment("asg_tenant_a", f.userID, CreateInput{CandidateID: f.candidateID, RequirementID: f.requirementID})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	if a.CandidateSnapshot != nil || a.RequirementSnapshot != nil {
		t.Fatalf("newly created assignment should have nil snapshots, got candidate=%v requirement=%v", a.CandidateSnapshot, a.RequirementSnapshot)
	}

	if err := a.TransitionTo(StatusScreening); err != nil {
		t.Fatalf("TransitionTo failed: %v", err)
	}
	if err := repo.Update(a); err != nil {
		t.Fatalf("repo.Update with nil snapshot fields failed: %v", err)
	}

	reloaded, err := svc.GetAssignment("asg_tenant_a", a.ID)
	if err != nil {
		t.Fatalf("GetAssignment failed: %v", err)
	}
	if reloaded.CandidateSnapshot != nil {
		t.Errorf("reloaded.CandidateSnapshot = %v, want nil", reloaded.CandidateSnapshot)
	}
	if reloaded.SnapshotCreatedAt != nil {
		t.Errorf("reloaded.SnapshotCreatedAt = %v, want nil", reloaded.SnapshotCreatedAt)
	}
}
