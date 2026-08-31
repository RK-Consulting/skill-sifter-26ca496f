package db

import (
	"database/sql"
	"os"
	"path/filepath"
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

	if err := initializeSchemaFromDir(dir); err != nil {
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

	if err := initializeSchemaFromDir(dir); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	writeScratchSchemaDefinition(t, dir, "002_add_column.sql", `ALTER TABLE migration_runner_probe ADD COLUMN extra TEXT;`)

	if err := initializeSchemaFromDir(dir); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count); err != nil {
		t.Fatalf("could not count schema versions: %v", err)
	}
	if count != 2 {
		t.Errorf("schema version count = %d, want 2", count)
	}

	if err := initializeSchemaFromDir(dir); err != nil {
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

	if err := initializeSchemaFromDir(dir); err != nil {
		t.Fatalf("initial schema initialization failed: %v", err)
	}

	if err := os.WriteFile(path, []byte(`CREATE TABLE migration_runner_probe (id SERIAL PRIMARY KEY, changed TEXT);`), 0644); err != nil {
		t.Fatalf("could not modify schema definition: %v", err)
	}

	if err := initializeSchemaFromDir(dir); err == nil {
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

	if err := initializeSchemaFromDir(dir); err == nil {
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
		firstDone <- withSchemaLock(func() error {
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
		secondDone <- withSchemaLock(func() error {
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
