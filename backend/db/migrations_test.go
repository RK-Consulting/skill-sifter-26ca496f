package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// setupMigrationsTestDB connects to a real Postgres instance and points the
// package-level DB at it, mirroring the pattern already used in
// handlers/candidate_handlers_test.go. Skips (does not fail) if no test
// database is reachable.
func setupMigrationsTestDB(t *testing.T) *sql.DB {
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
		t.Skipf("skipping migration runner test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		t.Skipf("skipping migration runner test: test DB not reachable (%v)", err)
	}

	// Fully reset tracking + any tables the scratch migrations below create,
	// so each test starts from a clean slate regardless of execution order.
	if _, err := testDB.Exec(`DROP TABLE IF EXISTS schema_migrations`); err != nil {
		t.Fatalf("could not reset schema_migrations: %v", err)
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

// writeScratchMigration writes a single migration file into a temp dir for
// use by these tests, never touching the real migrations directory.
func writeScratchMigration(t *testing.T, dir, filename, sql string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(sql), 0644); err != nil {
		t.Fatalf("could not write scratch migration %s: %v", filename, err)
	}
}

// TestApplyMigrations_FreshInstall verifies that on a database with no
// migration history, every migration file is applied in ascending numeric
// order, recorded, and assigned a checksum.
func TestApplyMigrations_FreshInstall(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchMigration(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, note TEXT);`)
	writeScratchMigration(t, dir, "002_seed_probe.sql", `INSERT INTO migration_runner_probe (note) VALUES ('seeded by 002');`)

	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("applyMigrationsFromDir failed on fresh install: %v", err)
	}

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM migration_runner_probe`).Scan(&count); err != nil {
		t.Fatalf("could not query probe table: %v", err)
	}
	if count != 1 {
		t.Errorf("probe row count = %d, want 1 (migration 002 should have run)", count)
	}

	var checksumCount int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE checksum IS NOT NULL`).Scan(&checksumCount); err != nil {
		t.Fatalf("could not query migration checksums: %v", err)
	}
	if checksumCount != 2 {
		t.Errorf("checksum row count = %d, want 2", checksumCount)
	}

	applied, err := appliedVersions()
	if err != nil {
		t.Fatalf("appliedVersions failed: %v", err)
	}
	if !applied[1] || !applied[2] {
		t.Errorf("appliedVersions = %v, want both 1 and 2 recorded", applied)
	}
}

// TestApplyMigrations_UpgradeOnlyRunsPending verifies that re-running the
// migrator after some migrations are already applied only executes the
// pending ones and verifies the checksum of the existing migration.
func TestApplyMigrations_UpgradeOnlyRunsPending(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchMigration(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, note TEXT);`)

	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	writeScratchMigration(t, dir, "002_add_column.sql", `ALTER TABLE migration_runner_probe ADD COLUMN extra TEXT;`)

	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("upgrade run (only 002 pending) failed: %v", err)
	}

	applied, err := appliedVersions()
	if err != nil {
		t.Fatalf("appliedVersions failed: %v", err)
	}
	if len(applied) != 2 || !applied[1] || !applied[2] {
		t.Errorf("appliedVersions = %v, want exactly {1,2}", applied)
	}

	// Restart again with no new files: must be a no-op, not an error.
	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("no-op restart failed: %v", err)
	}
}

// TestApplyMigrations_ChecksumMismatchFails verifies that an applied
// migration cannot be changed silently after it has been recorded.
func TestApplyMigrations_ChecksumMismatchFails(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	path := filepath.Join(dir, "001_create_probe.sql")
	writeScratchMigration(t, dir, "001_create_probe.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY);`)

	if err := applyMigrationsFromDir(dir); err != nil {
		t.Fatalf("initial migration run failed: %v", err)
	}

	if err := os.WriteFile(path, []byte(`CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, changed TEXT);`), 0644); err != nil {
		t.Fatalf("could not modify migration: %v", err)
	}

	if err := applyMigrationsFromDir(dir); err == nil {
		t.Fatal("migration runner accepted modified applied migration")
	}
}

// TestDiscoverMigrations_DuplicateSequenceFails verifies that two files
// claiming the same numeric prefix produce a deterministic error rather
// than an arbitrary pick.
func TestDiscoverMigrations_DuplicateSequenceFails(t *testing.T) {
	dir := t.TempDir()
	writeScratchMigration(t, dir, "003_first.sql", `SELECT 1;`)
	writeScratchMigration(t, dir, "003_second.sql", `SELECT 2;`)

	_, err := discoverMigrations(dir)
	if err == nil {
		t.Fatal("discoverMigrations did not fail on duplicate sequence prefix")
	}
}

// TestApplyMigrations_FailureStopsAndDoesNotRecord verifies that a failing
// migration is not recorded as applied, and that a later migration (which
// would otherwise succeed) never runs.
func TestApplyMigrations_FailureStopsAndDoesNotRecord(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	dir := t.TempDir()
	writeScratchMigration(t, dir, "001_broken.sql", `SELECT * FROM this_table_does_not_exist;`)
	writeScratchMigration(t, dir, "002_would_succeed.sql", `CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY);`)

	err := applyMigrationsFromDir(dir)
	if err == nil {
		t.Fatal("applyMigrationsFromDir succeeded despite a failing migration")
	}

	applied, aerr := appliedVersions()
	if aerr != nil {
		t.Fatalf("appliedVersions failed: %v", aerr)
	}
	if applied[1] {
		t.Error("failed migration 001 was recorded as applied")
	}
	if applied[2] {
		t.Error("migration 002 ran despite migration 001 failing first")
	}

	var exists bool
	if err := DB.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'migration_runner_probe')`).Scan(&exists); err != nil {
		t.Fatalf("could not check probe table existence: %v", err)
	}
	if exists {
		t.Error("migration_runner_probe table exists, meaning 002 ran despite 001 failing")
	}
}

// TestMigrationLockSerializesRunners verifies that two migration runners
// cannot enter the protected section at the same time.
func TestMigrationLockSerializesRunners(t *testing.T) {
	testDB := setupMigrationsTestDB(t)
	defer testDB.Close()
	DB = testDB

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondEntered := make(chan struct{})
	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)

	go func() {
		firstDone <- withMigrationLock(func() error {
			close(firstEntered)
			<-releaseFirst
			return nil
		})
	}()

	select {
	case <-firstEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("first migration runner did not acquire lock")
	}

	go func() {
		secondDone <- withMigrationLock(func() error {
			close(secondEntered)
			return nil
		})
	}()

	select {
	case <-secondEntered:
		t.Fatal("second migration runner entered while first still held lock")
	case <-time.After(200 * time.Millisecond):
	}

	close(releaseFirst)

	select {
	case err := <-firstDone:
		if err != nil {
			t.Fatalf("first migration runner failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first migration runner did not finish")
	}

	select {
	case <-secondEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("second migration runner did not acquire lock after first released it")
	}

	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second migration runner failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second migration runner did not finish")
	}
}
