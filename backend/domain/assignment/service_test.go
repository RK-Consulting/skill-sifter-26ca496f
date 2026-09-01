package assignment

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

// setupAssignmentTestDB keeps the existing assignment integration-test
// database entry point. Schema/bootstrap helpers and shared fixtures live in
// test_schema_bootstrap.go and fixtures_test.go.
func setupAssignmentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_assignment_test")

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping assignment service test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		t.Skipf("skipping assignment service test: test DB not reachable (%v)", err)
	}
	return testDB
}
