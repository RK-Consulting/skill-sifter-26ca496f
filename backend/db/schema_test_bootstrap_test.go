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
	if err := ensureSchemaTestDatabase(dbName); err != nil {
		fmt.Fprintf(os.Stderr, "schema test database bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("TEST_DB_NAME", dbName); err != nil {
		fmt.Fprintf(os.Stderr, "schema test database environment setup failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func ensureSchemaTestDatabase(dbName string) error {
	host := getenvDefault("TEST_DB_HOST", getenvDefault("DB_HOST", "localhost"))
	port := getenvDefault("TEST_DB_PORT", getenvDefault("DB_PORT", "5432"))
	user := getenvDefault("TEST_DB_USER", getenvDefault("DB_USER", "postgres"))
	password := getenvDefault("TEST_DB_PASSWORD", getenvDefault("DB_PASSWORD", "postgres"))

	conn := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbName + " sslmode=disable"
	testDB, err := sql.Open("postgres", conn)
	if err != nil {
		return err
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		return fmt.Errorf("database %q is not available: %w", dbName, err)
	}
	return nil
}
