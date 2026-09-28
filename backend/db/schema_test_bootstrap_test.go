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
	host := getenvDefault("TEST_DB_HOST", "localhost")
	port := getenvDefault("TEST_DB_PORT", "5432")
	user := getenvDefault("TEST_DB_USER", "postgres")
	password := getenvDefault("TEST_DB_PASSWORD", "postgres")

	adminDSN := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=postgres sslmode=disable"
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return err
	}

	var exists bool
	if err := adminDB.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = adminDB.Exec("CREATE DATABASE \"" + dbName + "\"")
	return err
}
