package db

import (
	"database/sql"
	"os"
	"testing"
)

// TestApplyMigrations_LegacyV03DatabaseUpgrade simulates the actual upgrade
// path that matters most: an existing v0.3 production database, which has
// never had schema_migrations tracking (it didn't exist until this
// prerequisite), running the new deterministic runner for the first time.
//
// This reproduces the exact production shape as closely as this sandbox
// allows:
//   - Baseline tables (companies, roles, users, candidates, jobs,
//     daily_jobs, interviews) created the way InitializeSchema() creates
//     them on every real deployment, WITH real data in them.
//   - business_dev created via the same CREATE TABLE the pre-existing lazy
//     fallback in handlers/business_dev_handlers.go uses, WITH a real row
//     — this is the table 001_baseline.sql's old "DROP TABLE IF EXISTS
//     business_dev" would have destroyed.
//   - 002_ai_reporting.sql's effects already present, matching the old
//     runner's actual historical behaviour (it always executed that one
//     file, every startup, unconditionally).
//   - No schema_migrations table at all (this is the first time the new
//     runner has ever touched this database).
//
// This is the "representative v0.3 database" upgrade scenario the PR #44
// review requires: it must prove business_dev data survives, and that only
// the genuinely pending migrations (001 for its idempotent non-business_dev
// statements, 003-006, since 002 was already applied historically) execute
// without error or data loss.
func TestApplyMigrations_LegacyV03DatabaseUpgrade(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	// --- Simulate a real, already-running v0.3 production database ---

	mustExec(t, testDB, `DROP TABLE IF EXISTS schema_migrations`)
	mustExec(t, testDB, `DROP TABLE IF EXISTS resume_search_logs, activity_logs, candidate_skills, skills, resumes, interviews, daily_jobs, jobs, candidates, users, roles, companies, business_dev, schema_version CASCADE`)

	// Baseline tables as InitializeSchema() creates them (see
	// backend/db/database.go createTableFromStruct / models.go), with real
	// data — an active tenant with real recruitment data, exactly what
	// must not be lost.
	mustExec(t, testDB, `CREATE TABLE schema_version (version INTEGER PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT NOW())`)
	mustExec(t, testDB, `INSERT INTO schema_version (version) VALUES (1)`)
	mustExec(t, testDB, `CREATE TABLE companies (id VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL DEFAULT NOW())`)
	mustExec(t, testDB, `INSERT INTO companies (id, name) VALUES ('comp_rk_consulting', 'RK Consulting')`)
	mustExec(t, testDB, `CREATE TABLE roles (id SERIAL PRIMARY KEY, name VARCHAR(50) NOT NULL UNIQUE, permissions JSONB NOT NULL DEFAULT '[]'::JSONB, created_at TIMESTAMP NOT NULL DEFAULT NOW())`)
	mustExec(t, testDB, `CREATE TABLE users (id SERIAL PRIMARY KEY, username VARCHAR(100) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL, role VARCHAR(100) NOT NULL, company_name VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT NOW())`)
	mustExec(t, testDB, `INSERT INTO users (username, email, password, role, company_name) VALUES ('legacyadmin', 'legacy@rk.test', 'x', 'admin', 'RK Consulting')`)
	mustExec(t, testDB, `CREATE TABLE candidates (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, phone VARCHAR(20), position VARCHAR(100), location VARCHAR(50), experience VARCHAR(100), currentctc VARCHAR(100), expectedctc VARCHAR(100), noticeperiod VARCHAR(100), jlptlanguage VARCHAR(100), skills VARCHAR(100), jobdescription VARCHAR(500), company_name VARCHAR(255) NOT NULL)`)
	mustExec(t, testDB, `INSERT INTO candidates (name, email, company_name) VALUES ('Legacy Candidate', 'legacycand@test.com', 'RK Consulting')`)
	mustExec(t, testDB, `CREATE TABLE jobs (id SERIAL PRIMARY KEY, title VARCHAR(255) NOT NULL, department VARCHAR(100), location VARCHAR(100), status VARCHAR(50) DEFAULT 'open', date_posted TIMESTAMP DEFAULT NOW(), description TEXT, requirements TEXT, last_modified TIMESTAMP DEFAULT NOW(), company_name VARCHAR(255) NOT NULL)`)
	mustExec(t, testDB, `CREATE TABLE daily_jobs (id SERIAL PRIMARY KEY, jd_no INTEGER NOT NULL, instructions TEXT, assigned_user INTEGER REFERENCES users(id), assigned_date TIMESTAMP DEFAULT NOW(), last_modified TIMESTAMP DEFAULT NOW(), company_name VARCHAR(255) NOT NULL)`)
	mustExec(t, testDB, `CREATE TABLE interviews (id SERIAL PRIMARY KEY, candidate_id INTEGER REFERENCES candidates(id), candidate_name VARCHAR(255) NOT NULL, position VARCHAR(100), interview_date TIMESTAMP NOT NULL, status VARCHAR(50) DEFAULT 'scheduled', feedback TEXT, last_modified TIMESTAMP DEFAULT NOW(), company_name VARCHAR(255) NOT NULL)`)

	// business_dev created via the exact CREATE TABLE the pre-existing lazy
	// fallback in handlers/business_dev_handlers.go uses (not via any
	// migration, since 001_baseline.sql's migration has never actually
	// executed on production). This is the table that must survive.
	mustExec(t, testDB, `CREATE TABLE business_dev (id SERIAL PRIMARY KEY, client_name VARCHAR(255) NOT NULL, partner_name VARCHAR(255), contact_person VARCHAR(255) NOT NULL, contact_number VARCHAR(50), contact_email VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT NOW(), last_modified TIMESTAMP DEFAULT NOW(), company_name VARCHAR(255) NOT NULL)`)
	mustExec(t, testDB, `INSERT INTO business_dev (client_name, partner_name, contact_person, contact_number, contact_email, company_name) VALUES ('Acme Client Corp', 'Referral Partner Inc', 'Jane Contact', '555-0100', 'jane@acmeclient.test', 'RK Consulting')`)

	// 002_ai_reporting.sql's effects, matching the old runner's actual
	// historical behaviour: it unconditionally executed this one file on
	// every startup, so any real production database already has these
	// effects regardless of what schema_migrations (which didn't exist)
	// says.
	mustExec(t, testDB, `ALTER TABLE candidates ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW()`)
	mustExec(t, testDB, `CREATE TABLE IF NOT EXISTS resumes (id SERIAL PRIMARY KEY, company_name VARCHAR(255) NOT NULL, candidate_id INTEGER REFERENCES candidates(id) ON DELETE SET NULL, file_name VARCHAR(500) NOT NULL, file_path TEXT NOT NULL, file_hash VARCHAR(64) NOT NULL, mime_type VARCHAR(120), extracted_text TEXT, parsing_status VARCHAR(30) NOT NULL DEFAULT 'pending', parser_model VARCHAR(120), parse_error TEXT, uploaded_by INTEGER REFERENCES users(id) ON DELETE SET NULL, uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(), parsed_at TIMESTAMP)`)
	mustExec(t, testDB, `CREATE TABLE IF NOT EXISTS activity_logs (id SERIAL PRIMARY KEY, company_name VARCHAR(255) NOT NULL, action VARCHAR(100), entity_type VARCHAR(100), entity_id VARCHAR(100), description TEXT, actor_user_id INTEGER, metadata JSONB, created_at TIMESTAMP NOT NULL DEFAULT NOW())`)
	mustExec(t, testDB, `INSERT INTO schema_version(version) VALUES (2) ON CONFLICT (version) DO NOTHING`)

	// Crucially, no schema_migrations table — the new tracked runner has
	// never touched this database before.

	// --- Now run the real, fixed migration set (001-006) exactly as the
	// deployed binary would on startup ---

	// go test's working directory is this package's own directory
	// (backend/db), not the backend module root migrationsDir() expects at
	// runtime, so the real directory is referenced relative to here rather
	// than via migrationsDir() (which is tuned for how the compiled binary
	// is actually launched in dev/production, not the test runner).
	dir := "../database/migrations"
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("could not locate real migrations directory at %q: %v", dir, err)
	}
	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("upgrade from legacy v0.3 database failed: %v", err)
	}

	// --- The critical assertion: business_dev data survived ---

	var clientName, contactEmail string
	if err := testDB.QueryRow(`SELECT client_name, contact_email FROM business_dev WHERE contact_email = 'jane@acmeclient.test'`).Scan(&clientName, &contactEmail); err != nil {
		t.Fatalf("business_dev data did NOT survive the upgrade (this is the exact data-loss scenario the DROP TABLE bug caused): %v", err)
	}
	if clientName != "Acme Client Corp" {
		t.Errorf("business_dev.client_name = %q, want %q — data was altered during upgrade", clientName, "Acme Client Corp")
	}

	var bdCount int
	testDB.QueryRow(`SELECT COUNT(*) FROM business_dev`).Scan(&bdCount)
	if bdCount != 1 {
		t.Errorf("business_dev row count = %d, want 1 (table must not have been dropped and left empty, or duplicated)", bdCount)
	}

	// --- Migrations that had never actually run before (003, 004) took
	// effect for the first time, as intended ---

	var skillAliasesExists bool
	testDB.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'skill_aliases')`).Scan(&skillAliasesExists)
	if !skillAliasesExists {
		t.Error("skill_aliases table missing after upgrade — migration 003 did not apply")
	}

	var createdByColExists bool
	testDB.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'jobs' AND column_name = 'created_by_user_id')`).Scan(&createdByColExists)
	if !createdByColExists {
		t.Error("jobs.created_by_user_id missing after upgrade — migration 004 did not apply")
	}

	// --- Tenant identity (006) correctly backfilled onto legacy data ---

	var candidateTenantID, businessDevTenantID string
	testDB.QueryRow(`SELECT tenant_id FROM candidates WHERE email = 'legacycand@test.com'`).Scan(&candidateTenantID)
	testDB.QueryRow(`SELECT tenant_id FROM business_dev WHERE contact_email = 'jane@acmeclient.test'`).Scan(&businessDevTenantID)
	if candidateTenantID != "comp_rk_consulting" {
		t.Errorf("legacy candidate tenant_id = %q, want %q", candidateTenantID, "comp_rk_consulting")
	}
	if businessDevTenantID != "comp_rk_consulting" {
		t.Errorf("legacy business_dev tenant_id = %q, want %q", businessDevTenantID, "comp_rk_consulting")
	}

	// --- All 6 migrations recorded ---

	applied, err := appliedVersions()
	if err != nil {
		t.Fatalf("appliedVersions failed: %v", err)
	}
	for v := 1; v <= 6; v++ {
		if !applied[v] {
			t.Errorf("migration %d not recorded as applied after upgrade", v)
		}
	}

	// --- Simulate the next server restart: only pending migrations run,
	// which here means none, and business_dev data is still untouched ---

	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("second run (simulating next restart) failed: %v", err)
	}

	testDB.QueryRow(`SELECT COUNT(*) FROM business_dev`).Scan(&bdCount)
	if bdCount != 1 {
		t.Errorf("business_dev row count after a second runner pass = %d, want 1", bdCount)
	}
}

func mustExec(t *testing.T, testDB *sql.DB, query string) {
	t.Helper()
	if _, err := testDB.Exec(query); err != nil {
		t.Fatalf("setup statement failed: %v\nstatement: %s", err, query)
	}
}
