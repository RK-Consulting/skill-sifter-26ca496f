package assignment

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
)

// TestMain runs assignment integration tests against a clean database schema
// produced exclusively by the application's migration runner. The test DB is
// disposable; resetting it here prevents older hand-built fixture schemas
// from masking missing/obsolete columns and constraints.
func TestMain(m *testing.M) {
	testDB, err := openAssignmentIntegrationDB()
	if err != nil {
		// Preserve the repository's convention: integration tests skip when
		// PostgreSQL is unavailable rather than making unit-only runs fail.
		os.Exit(m.Run())
	}
	defer testDB.Close()

	if err := chdirToBackendRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap: %v\n", err)
		os.Exit(1)
	}

	if _, err := testDB.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test database reset failed: %v\n", err)
		os.Exit(1)
	}

	db.DB = testDB
	if err := db.InitializeSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func openAssignmentIntegrationDB() (*sql.DB, error) {
	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_assignment_test")

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	testDB, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		return nil, err
	}
	return testDB, nil
}

func chdirToBackendRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return os.Chdir(wd)
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return fmt.Errorf("could not locate backend go.mod from %q", wd)
		}
		wd = parent
	}
}
