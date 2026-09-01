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

// TestMain prepares the handler integration database from the application's
// authoritative schema definitions. The assignment package uses a separate
// database, so handler helpers that connect to TEST_DB_NAME can safely use the
// public schema without racing another package's schema reset.
func TestMain(m *testing.M) {
	testDB, err := openHandlerTestDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "handler test database bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := chdirHandlerTestToBackendRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "handler test schema bootstrap: %v\n", err)
		os.Exit(1)
	}

	if _, err := testDB.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		fmt.Fprintf(os.Stderr, "handler test database reset failed: %v\n", err)
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

	conn := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"
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
			return fmt.Errorf("could not locate handler backend go.mod from %q", wd)
		}
		wd = parent
	}
}
