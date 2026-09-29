package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const schemaTestDBName = "skillsifter_schema_test"

func TestMain(m *testing.M) {
	dbName := getenvDefault("SKILLSIFTER_SCHEMA_TEST_DB", schemaTestDBName)
	host := getenvDefault("TEST_DB_HOST", getenvDefault("DB_HOST", "localhost"))
	port := getenvDefault("TEST_DB_PORT", getenvDefault("DB_PORT", "5432"))
	user := getenvDefault("TEST_DB_USER", getenvDefault("DB_USER", "postgres"))
	password := getenvDefault("TEST_DB_PASSWORD", getenvDefault("DB_PASSWORD", "postgres"))

	testDSN := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbName + " sslmode=disable"
	testDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema test database bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "schema test database bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.Setenv("TEST_DB_NAME", dbName); err != nil {
		fmt.Fprintf(os.Stderr, "schema test database environment setup failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
