package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// setupSchemaTestDB connects to a real Postgres instance and points the
// package-level DB at it. Tests always begin with a clean schema-version table
// and scratch probe table.
func setupSchemaTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvDefault("TEST_DB_HOST", "localhost")
	port := getenvDefault("TEST_DB_PORT", "5432")
	user := getenvDefault("TEST_DB_USER", "postgres")
	password := getenvDefault("TEST_DB_PASSWORD", "postgres")
	dbname := getenvDefault("TEST_DB_NAME", "skillsifter_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping schema initialization test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		t.Skipf("skipping schema initialization test: test DB not reachable (%v)", err)
	}

	if _, err := testDB.Exec(`DROP TABLE IF EXISTS schema_versions`); err != nil {
		t.Fatalf("could not reset schema_versions: %v", err)
	}
	if _, err := testDB.Exec(`DROP TABLE IF EXISTS migration_runner_probe`); err != nil {
		t.Fatalf("could not reset migration_runner_probe: %v", err)
	}

	return testDB
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func writeScratchSchemaDefinition(t *testing.T, dir, filename, sqlText string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(sqlText), 0644); err != nil {
		t.Fatalf("could not write scratch schema definition %s: %v", filename, err)
	}
}

func TestInitializeSchema_FreshInstall(t *testing.T) {
	testDB := setupSchemaTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchSchemaDefinition(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, note TEXT);`)
	writeScratchSchemaDefinition(t, dir, "002_seed_probe.sql", `INSERT INTO migration_runner_probe (note) VALUES ('seeded by 002');`)

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err != nil {
		t.Fatalf("initializeSchemaFromDir failed on fresh install: %v", err)
	}

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM migration_runner_probe`).Scan(&count); err != nil {
		t.Fatalf("could not query probe table: %v", err)
	}
	if count != 1 {
		t.Errorf("probe row count = %d, want 1", count)
	}

	var checksumCount int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_versions WHERE checksum IS NOT NULL`).Scan(&checksumCount); err != nil {
		t.Fatalf("could not query schema checksums: %v", err)
	}
	if checksumCount != 2 {
		t.Errorf("checksum row count = %d, want 2", checksumCount)
	}
}

func TestInitializeSchema_OnlyRunsPending(t *testing.T) {
	testDB := setupSchemaTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchSchemaDefinition(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, note TEXT);`)

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	writeScratchSchemaDefinition(t, dir, "002_add_column.sql", `ALTER TABLE migration_runner_probe ADD COLUMN extra TEXT;`)

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count); err != nil {
		t.Fatalf("could not count schema versions: %v", err)
	}
	if count != 2 {
		t.Errorf("schema version count = %d, want 2", count)
	}

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err != nil {
		t.Fatalf("no-op restart failed: %v", err)
	}
}

func TestInitializeSchema_ChecksumMismatchFails(t *testing.T) {
	testDB := setupSchemaTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	path := filepath.Join(dir, "001_create_probe.sql")
	writeScratchSchemaDefinition(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY);`)

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err != nil {
		t.Fatalf("initial schema initialization failed: %v", err)
	}

	if err := os.WriteFile(path, []byte(`CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, changed TEXT);`), 0644); err != nil {
		t.Fatalf("could not modify schema definition: %v", err)
	}

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err == nil {
		t.Fatal("schema initializer accepted modified applied definition")
	}
}

func TestDiscoverSchemaDefinitions_DuplicateSequenceFails(t *testing.T) {
	dir := t.TempDir()
	writeScratchSchemaDefinition(t, dir, "003_first.sql", `SELECT 1;`)
	writeScratchSchemaDefinition(t, dir, "003_second.sql", `SELECT 2;`)

	if _, err := discoverSchemaDefinitions(dir); err == nil {
		t.Fatal("discoverSchemaDefinitions did not fail on duplicate sequence prefix")
	}
}

func TestInitializeSchema_FailureStopsAndDoesNotRecord(t *testing.T) {
	testDB := setupSchemaTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchSchemaDefinition(t, dir, "001_broken.sql", `SELECT * FROM this_table_does_not_exist;`)
	writeScratchSchemaDefinition(t, dir, "002_would_succeed.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY);`)

	if err := initializeSchema(DB, dir, "skill-sifter:test-schema"); err == nil {
		t.Fatal("initializeSchemaFromDir succeeded despite a failing definition")
	}

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count); err != nil {
		t.Fatalf("could not count schema versions: %v", err)
	}
	if count != 0 {
		t.Errorf("schema version count = %d, want 0", count)
	}

	var exists bool
	if err := DB.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'migration_runner_probe')`).Scan(&exists); err != nil {
		t.Fatalf("could not check probe table existence: %v", err)
	}
	if exists {
		t.Error("probe table exists, meaning definition 002 ran despite 001 failing")
	}
}

func TestSchemaLockSerializesInitializers(t *testing.T) {
	testDB := setupSchemaTestDB(t)
	defer testDB.Close()
	DB = testDB

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondEntered := make(chan struct{})
	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)

	go func() {
		firstDone <- withSchemaLock(testDB, "skill-sifter:test-lock", func() error {
			close(firstEntered)
			<-releaseFirst
			return nil
		})
	}()

	select {
	case <-firstEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("first schema initializer did not acquire lock")
	}

	go func() {
		secondDone <- withSchemaLock(testDB, "skill-sifter:test-lock", func() error {
			close(secondEntered)
			return nil
		})
	}()

	select {
	case <-secondEntered:
		t.Fatal("second schema initializer entered while first still held lock")
	case <-time.After(200 * time.Millisecond):
	}

	close(releaseFirst)

	select {
	case err := <-firstDone:
		if err != nil {
			t.Fatalf("first schema initializer failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first schema initializer did not finish")
	}

	select {
	case <-secondEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("second schema initializer did not acquire lock after first released it")
	}

	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second schema initializer failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second schema initializer did not finish")
	}
}

func TestFinalPlaneBaselines_ArePhysicallySeparated(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	host := getenvDefault("TEST_DB_HOST", "localhost")
	port := getenvDefault("TEST_DB_PORT", "5432")
	user := getenvDefault("TEST_DB_USER", "postgres")
	password := getenvDefault("TEST_DB_PASSWORD", "postgres")

	openTestDB := func(name string) *sql.DB {
		t.Helper()
		if err := ensureSchemaTestDatabase(name); err != nil {
			t.Skipf("final-plane baseline test skipped: could not create %s: %v", name, err)
		}
		dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + name + " sslmode=disable"
		d, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Skipf("final-plane baseline test skipped: %v", err)
		}
		if err := d.Ping(); err != nil {
			d.Close()
			t.Skipf("final-plane baseline test skipped: %v", err)
		}
		if _, err := d.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
			d.Close()
			t.Fatalf("reset %s failed: %v", name, err)
		}
		return d
	}

	controlDB := openTestDB(getenvDefault("SKILLSIFTER_CONTROL_SCHEMA_TEST_DB", "skillsifter_control_schema_test"))
	defer controlDB.Close()
	tenantDB := openTestDB(getenvDefault("SKILLSIFTER_TENANT_SCHEMA_TEST_DB", "skillsifter_tenant_schema_test"))
	defer tenantDB.Close()

	DB = controlDB
	if err := InitializeControlSchema(); err != nil {
		t.Fatalf("control-plane baseline initialization failed: %v", err)
	}
	if err := InitializeTenantSchema(tenantDB); err != nil {
		t.Fatalf("tenant-plane baseline initialization failed: %v", err)
	}

	assertTables := func(d *sql.DB, plane string, expected []string, forbidden []string) {
		t.Helper()
		for _, table := range expected {
			var exists bool
			if err := d.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)", table).Scan(&exists); err != nil {
				t.Fatalf("%s table check %s failed: %v", plane, table, err)
			}
			if !exists {
				t.Errorf("%s baseline missing required table %q", plane, table)
			}
		}
		for _, table := range forbidden {
			var exists bool
			if err := d.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)", table).Scan(&exists); err != nil {
				t.Fatalf("%s forbidden-table check %s failed: %v", plane, table, err)
			}
			if exists {
				t.Errorf("%s baseline contains forbidden table %q", plane, table)
			}
		}
	}

	assertTables(controlDB, "control-plane",
		[]string{"schema_versions", "platform_tenants", "platform_plans", "platform_subscriptions", "platform_subscription_checkouts", "platform_subscription_events", "platform_registration_registry", "platform_pending_registrations", "platform_verification_codes", "platform_user_accounts"},
		[]string{"users", "candidates", "clients", "requirements", "interviews", "recruitment_offers", "recruitment_joinings", "recruitment_billings", "companies", "jobs"})

	assertTables(tenantDB, "tenant-plane",
		[]string{"schema_versions", "roles", "users", "clients", "requirements", "candidates", "interviews", "recruitment_screenings", "recruitment_submissions", "recruitment_selections", "recruitment_offers", "recruitment_joinings", "recruitment_billings", "audit_events"},
		[]string{"platform_tenants", "platform_plans", "platform_subscriptions", "platform_user_accounts", "companies", "jobs", "daily_jobs", "business_dev"})

	var crossPlaneFKCount int
	if err := tenantDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name=tc.constraint_name
		 AND ccu.table_schema=tc.table_schema
		WHERE tc.constraint_type='FOREIGN KEY'
		  AND ccu.table_schema='public'
		  AND ccu.table_name LIKE 'platform_%'
	`).Scan(&crossPlaneFKCount); err != nil {
		t.Fatalf("cross-plane FK check failed: %v", err)
	}
	if crossPlaneFKCount != 0 {
		t.Errorf("tenant baseline contains %d cross-plane platform foreign keys", crossPlaneFKCount)
	}
}
