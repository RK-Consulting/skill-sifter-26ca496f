package assignment

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
)

// seedFixtures creates the minimal tenant-local records shared by the
// assignment integration tests. The companies themselves are provisioned by
// the package TestMain schema bootstrap before these fixtures run.
func seedFixtures(t *testing.T, db *sql.DB, tenantID string) fixtures {
	t.Helper()

	n := atomic.AddInt64(&seedFixturesCounter, 1)
	unique := fmt.Sprintf("%s_%d", tenantID, n)

	var f fixtures
	if err := db.QueryRow(`
		INSERT INTO users (username, email, password, role, tenant_id, company_name)
		VALUES ($1, $2, 'x', 'recruiter', $3, $3)
		RETURNING id`,
		"user_"+unique,
		unique+"user@test.com",
		tenantID,
	).Scan(&f.userID); err != nil {
		t.Fatalf("seedFixtures: could not insert user: %v", err)
	}

	if err := db.QueryRow(`
		INSERT INTO candidates (name, email, tenant_id, company_name, status)
		VALUES ('Test Candidate', $1, $2, $2, 'active')
		RETURNING id`,
		unique+"cand@test.com",
		tenantID,
	).Scan(&f.candidateID); err != nil {
		t.Fatalf("seedFixtures: could not insert candidate: %v", err)
	}

	var clientID int
	if err := db.QueryRow(`
		INSERT INTO clients (name, tenant_id)
		VALUES ('Test Client', $1)
		RETURNING id`, tenantID).Scan(&clientID); err != nil {
		t.Fatalf("seedFixtures: could not insert client: %v", err)
	}

	if err := db.QueryRow(`
		INSERT INTO requirements (client_id, title, tenant_id)
		VALUES ($1, 'Test Req', $2)
		RETURNING id`, clientID, tenantID).Scan(&f.requirementID); err != nil {
		t.Fatalf("seedFixtures: could not insert requirement: %v", err)
	}

	return f
}
