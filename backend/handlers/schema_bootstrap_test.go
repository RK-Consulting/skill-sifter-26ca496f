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

const handlerTestDBName = "skillsifter_handler_test"

// TestMain gives the handler integration suite its own PostgreSQL database.
// This prevents package-level integration tests from racing with the db and
// assignment packages when `go test ./...` runs packages concurrently.
func TestMain(m *testing.M) {
	handlerDBName := getenvOr("SKILLSIFTER_HANDLER_TEST_DB", handlerTestDBName)
	if err := os.Setenv("TEST_DB_NAME", handlerDBName); err != nil {
		fmt.Fprintf(os.Stderr, "handler test database environment setup failed: %v\n", err)
		os.Exit(1)
	}

	testDB, err := openHandlerTestDB(handlerDBName)
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

func openHandlerTestDB(dbname string) (*sql.DB, error) {
	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")

	adminConn := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=postgres sslmode=disable"
	adminDB, err := sql.Open("postgres", adminConn)
	if err != nil {
		return nil, err
	}
	defer adminDB.Close()
	if err := adminDB.Ping(); err != nil {
		return nil, err
	}

	var exists bool
	if err := adminDB.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, dbname).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		if _, err := adminDB.Exec(`CREATE DATABASE "` + dbname + `"`); err != nil {
			return nil, err
		}
	}

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
