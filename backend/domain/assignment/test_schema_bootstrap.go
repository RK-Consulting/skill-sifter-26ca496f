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

func TestMain(m *testing.M) {
	testDB, err := openAssignmentIntegrationDB()
	if err != nil {
		os.Exit(m.Run())
	}
	defer testDB.Close()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(wd, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			fmt.Fprintln(os.Stderr, "could not locate backend go.mod")
			os.Exit(1)
		}
		wd = parent
	}
	if err := os.Chdir(wd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	db.DB = testDB
	if err := db.InitializeSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment integration schema setup failed: %v
", err)
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
