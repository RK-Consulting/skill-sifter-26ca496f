package assignment

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	_ "github.com/lib/pq"
)

// setupAssignmentTestDB stands up (or reuses) the tables needed to
// exercise the assignment service. The shared getenvOr helper lives in
// test_schema_bootstrap.go so the package has a single definition.
func setupAssignmentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_assignment_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping assignment service test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		t.Skipf("skipping assignment service test: test DB not reachable (%v)", err)
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS companies (id VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(100) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL, role VARCHAR(100) NOT NULL, tenant_id VARCHAR(255) REFERENCES companies(id), company_name VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS candidates (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, phone VARCHAR(20), position VARCHAR(100), location VARCHAR(50), experience VARCHAR(100), currentctc VARCHAR(100), expectedctc VARCHAR(100), noticeperiod VARCHAR(100), tenant_id VARCHAR(255) REFERENCES companies(id), company_name VARCHAR(255) NOT NULL, status VARCHAR(50) NOT NULL DEFAULT 'active', active_recruitment_engagements BOOLEAN NOT NULL DEFAULT FALSE, CONSTRAINT candidates_status_valid CHECK (status IN ('active','inactive','blacklisted','archived')))` ,
		`CREATE TABLE IF NOT EXISTS candidate_language_expertise (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id), candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE, language VARCHAR(100) NOT NULL, proficiency_framework VARCHAR(50) NOT NULL, proficiency_level VARCHAR(50) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW(), CONSTRAINT candidate_language_expertise_unique UNIQUE (candidate_id, language, proficiency_framework, proficiency_level))`,
		`CREATE TABLE IF NOT EXISTS candidate_expertise (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id), candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE, skill VARCHAR(100) NOT NULL, category VARCHAR(100) NOT NULL, proficiency_level VARCHAR(50) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW(), CONSTRAINT candidate_expertise_unique UNIQUE (candidate_id, skill, category))`,
		`CREATE TABLE IF NOT EXISTS clients (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) REFERENCES companies(id), name VARCHAR(255) NOT NULL, status VARCHAR(50) NOT NULL DEFAULT 'prospect')`,
		`CREATE TABLE IF NOT EXISTS requirements (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) REFERENCES companies(id), client_id INTEGER REFERENCES clients(id), title VARCHAR(255) NOT NULL, location VARCHAR(100), work_arrangement VARCHAR(50), description TEXT, required_skills TEXT, experience_required VARCHAR(100), compensation VARCHAR(255), headcount INTEGER NOT NULL DEFAULT 1, language_requirement VARCHAR(255), status VARCHAR(50) NOT NULL DEFAULT 'draft')`,
		`CREATE TABLE IF NOT EXISTS recruitment_assignments (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id), candidate_id INTEGER NOT NULL REFERENCES candidates(id), requirement_id INTEGER NOT NULL REFERENCES requirements(id), status VARCHAR(50) NOT NULL DEFAULT 'draft', created_by_user_id INTEGER NOT NULL REFERENCES users(id), owner_user_id INTEGER NOT NULL REFERENCES users(id), candidate_snapshot JSONB, requirement_snapshot JSONB, snapshot_created_at TIMESTAMP, created_at TIMESTAMP NOT NULL DEFAULT NOW(), last_modified TIMESTAMP NOT NULL DEFAULT NOW(), CONSTRAINT recruitment_assignments_status_valid CHECK (status IN ('draft','screening','submitted','interviewing','offered','joined','rejected','withdrawn')), CONSTRAINT recruitment_assignments_candidate_requirement_unique UNIQUE (candidate_id, requirement_id))`,
		`CREATE TABLE IF NOT EXISTS audit_events (id SERIAL PRIMARY KEY, tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id), actor_user_id INTEGER NOT NULL REFERENCES users(id), entity_type VARCHAR(100) NOT NULL, entity_id INTEGER NOT NULL, action VARCHAR(100) NOT NULL, occurred_at TIMESTAMP NOT NULL DEFAULT NOW(), correlation_id VARCHAR(100), metadata JSONB NOT NULL DEFAULT '{}'::JSONB)`,
	}

	for _, statement := range statements {
		if _, err := testDB.Exec(statement); err != nil {
			t.Fatalf("assignment test schema setup failed: %v\nstatement: %s", err, statement)
		}
	}

	testDB.Exec(`DELETE FROM audit_events WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM recruitment_assignments WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM candidate_language_expertise WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM candidate_expertise WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM requirements WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM clients WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM candidates WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`DELETE FROM users WHERE tenant_id IN ('asg_tenant_a', 'asg_tenant_b')`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('asg_tenant_a', 'Assignment Test Tenant A Co') ON CONFLICT (id) DO NOTHING`)
	testDB.Exec(`INSERT INTO companies (id, name) VALUES ('asg_tenant_b', 'Assignment Test Tenant B Co') ON CONFLICT (id) DO NOTHING`)

	return testDB
}

type fixtures struct {
	userID        int
	candidateID   int
	requirementID int
}

var seedFixturesCounter int64
