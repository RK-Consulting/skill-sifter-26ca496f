package handlers

import "testing"

func TestRetiredUserLedgerSchema(t *testing.T) {
	testDB := openRegistryTestDB(t)
	var exists bool
	if err := testDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema='public'
			  AND table_name='platform_retired_users'
		)
	`).Scan(&exists); err != nil {
		t.Fatalf("check retired-user ledger table: %v", err)
	}
	if !exists {
		t.Fatal("platform_retired_users table does not exist")
	}

	var columns int
	if err := testDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema='public'
		  AND table_name='platform_retired_users'
		  AND column_name IN ('tenant_id', 'user_id', 'username', 'email', 'status', 'retired_at')
	`).Scan(&columns); err != nil {
		t.Fatalf("check retired-user ledger columns: %v", err)
	}
	if columns != 6 {
		t.Fatalf("retired-user ledger columns = %d, want 6", columns)
	}
}
