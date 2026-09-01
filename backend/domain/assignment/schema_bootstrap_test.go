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

// TestMain ensures assignment integration tests run against the same
// migration-defined schema as the application. Individual tests may still
// create/clean their fixture rows, but they must not silently run against a
// stale hand-built table definition.
func TestMain(m *testing.M) {
	testDB, err := openAssignmentTestDB()
	if err != nil {
		// Individual integration tests already skip when PostgreSQL is not
		// available. Preserve that behavior instead of making the whole
		// package unusable for unit-only runs.
		os.Exit(m.Run())
	}
	defer testDB.Close()

	if err := chdirToBackendRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap: %v\n", err)
		os.Exit(1)
	}

	db.DB = testDB
	if err := db.InitializeSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func openAssignmentTestDB() (*sql.DB, error) {
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
