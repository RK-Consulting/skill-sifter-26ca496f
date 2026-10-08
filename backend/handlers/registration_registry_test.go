package handlers

import (
	"database/sql"
	"strings"
	"testing"
)

func openRegistryTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if handlerControlDB == nil {
		t.Fatal("control test database is not initialized")
	}
	return handlerControlDB
}

func TestPermanentRegistrationRegistrySchema(t *testing.T) {
	testDB := openRegistryTestDB(t)
	var exists bool
	if err := testDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'platform_registration_registry'
		)
	`).Scan(&exists); err != nil {
		t.Fatalf("check registration registry table: %v", err)
	}
	if !exists {
		t.Fatal("platform_registration_registry table does not exist")
	}

	var columns int
	if err := testDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'platform_registration_registry'
		  AND column_name IN ('email_id', 'first_registered', 'last_tenant_id')
	`).Scan(&columns); err != nil {
		t.Fatalf("check registration registry columns: %v", err)
	}
	if columns != 3 {
		t.Fatalf("registration registry columns = %d, want 3", columns)
	}
}

func TestPermanentRegistrationRegistryRejectsDuplicateEmail(t *testing.T) {
	testDB := openRegistryTestDB(t)
	defer testDB.Close()
	const email = "registry-test@example.com"

	_, _ = testDB.Exec(`DELETE FROM platform_registration_registry WHERE email_id = $1`, email)

	if _, err := testDB.Exec(`
		INSERT INTO platform_registration_registry(email_id, first_registered, last_tenant_id)
		VALUES($1, NOW(), $2)
	`, email, "tenant_registry_test"); err != nil {
		t.Fatalf("insert registration identity: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testDB.Exec(`DELETE FROM platform_registration_registry WHERE email_id = $1`, email)
	})

	var registered bool
	if err := testDB.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM platform_registration_registry
			WHERE email_id = $1
		)
	`, email).Scan(&registered); err != nil {
		t.Fatalf("check registration identity: %v", err)
	}
	if !registered {
		t.Fatal("registered email was not retained")
	}

	_, err := testDB.Exec(`
		INSERT INTO platform_registration_registry(email_id, first_registered, last_tenant_id)
		VALUES($1, NOW(), $2)
	`, email, "tenant_registry_test_2")
	if err == nil {
		t.Fatal("duplicate registration email was accepted")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("duplicate email failed for unexpected reason: %v", err)
	}
}
