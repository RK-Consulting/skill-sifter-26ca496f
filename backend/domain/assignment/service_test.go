package assignment

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

// setupAssignmentTestDB always connects to the database prepared by the
// package TestMain. Keeping the database name in package state avoids relying
// on environment mutation across the test process and prevents this suite
// from accidentally reconnecting to the shared handler/db test database.
func setupAssignmentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + assignmentTestDatabaseName + " search_path=public sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("could not open assignment test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		t.Fatalf("assignment test DB not reachable: %v", err)
	}
	return testDB
}
