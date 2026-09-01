package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
)

// TestMain makes the handler integration suite use the authoritative database
// migrations before any package-local fixture setup runs. This prevents
// CREATE TABLE IF NOT EXISTS test fixtures from preserving obsolete schemas.
func TestMain(m *testing.M) {
	testDB, err := openHandlerTestDB()
	if err != nil {
		// The existing integration tests intentionally skip when PostgreSQL is
		// unavailable. Keep that behavior for local unit-only runs.
		os.Exit(m.Run())
	}
	defer testDB.Close()

	if err := chdirHandlerTestToBackendRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "handler test schema bootstrap: %v\n", err)
		os.Exit(1)
	}

	db.DB = testDB
	if err := db.InitializeSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "handler test schema bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func openHandlerTestDB() (*sql.DB, error) {
	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		return nil, err
	}
	return testDB, nil
}

func chdirHandlerTestToBackendRoot() error {
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
