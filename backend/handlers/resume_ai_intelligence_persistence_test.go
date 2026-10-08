package handlers

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type resumeAITestFixture struct {
	db          *sql.DB
	tenantID    string
	candidateID int
	resumeID    int
}

func setupResumeAITestFixture(t *testing.T) resumeAITestFixture {
	t.Helper()
	testDB := handlerTenantDB
	if testDB == nil {
		t.Fatal("tenant test database is not initialized")
	}

	tenantID := fmt.Sprintf("rai03_test_%d", time.Now().UnixNano())
	var candidateID int
	if err := testDB.QueryRow(
		"INSERT INTO candidates (name, email, tenant_id) VALUES ('Other Tenant Candidate', 'shared@example.com', $1) RETURNING id"
		otherTenant,
	).Scan(&otherCandidateID); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = fx.db.Exec("DELETE FROM candidates WHERE id=$1", otherCandidateID) }()

	if _, err := fx.db.Exec("UPDATE candidates SET email='shared@example.com', phone='9222222222' WHERE id=$1", fx.candidateID); err != nil {
		t.Fatal(err)
	}

	got, err := upsertResumeCandidate(fx.db, fx.tenantID, resumeAIResult{
		Name:  "Tenant A Candidate",
		Email: "shared@example.com",
		Phone: "9222222222",
	})
	if err != nil {
		t.Fatalf("upsertResumeCandidate failed: %v", err)
	}
	if got == nil || got.ID != fx.candidateID {
		t.Fatalf("associated candidate = %+v, want tenant A candidate %d", got, fx.candidateID)
	}
	if got.ID == otherCandidateID {
		t.Fatal("candidate association crossed tenant boundary")
	}
}

func TestPersistResumeIntelligencePartialFailureRollsBack(t *testing.T) {
	fx := setupResumeAITestFixture(t)
	ai := resumeAIResult{
		CurrentTitle:      "Should Roll Back",
		Skills:            []string{"Go"},
		EmploymentHistory: []resumeEmployment{{Employer: "Example Corp"}},
		Projects:          []resumeProject{{ProjectName: "Broken Project", StartDate: "not-a-date"}},
	}

	if err := persistResumeIntelligence(fx.db, fx.resumeID, fx.candidateID, fx.tenantID, ai); err == nil {
		t.Fatal("expected invalid project date to fail persistence")
	}

	var profileCount, skillCount, employmentCount, projectCount int
	checks := []struct {
		query string
		dest  *int
	}{
		{"SELECT COUNT(*) FROM candidate_professional_profiles WHERE tenant_id=$1 AND candidate_id=$2", &profileCount},
		{"SELECT COUNT(*) FROM candidate_expertise WHERE tenant_id=$1 AND candidate_id=$2 AND category='resume_import'", &skillCount},
		{"SELECT COUNT(*) FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3", &employmentCount},
		{"SELECT COUNT(*) FROM candidate_projects WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3", &projectCount},
	}
	for i, check := range checks {
		var err error
		if i < 2 {
			err = fx.db.QueryRow(check.query, fx.tenantID, fx.candidateID).Scan(check.dest)
		} else {
			err = fx.db.QueryRow(check.query, fx.tenantID, fx.candidateID, fx.resumeID).Scan(check.dest)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if profileCount != 0 || skillCount != 0 || employmentCount != 0 || projectCount != 0 {
		t.Fatalf("partial persistence remained after failure: profile=%d skills=%d employment=%d projects=%d", profileCount, skillCount, employmentCount, projectCount)
	}
}

func TestPersistResumeIntelligenceReprocessingIsSourceScoped(t *testing.T) {
	fx := setupResumeAITestFixture(t)

	first := resumeAIResult{
		CurrentTitle:      "First Title",
		Skills:            []string{"Go"},
		EmploymentHistory: []resumeEmployment{{Employer: "Example Corp", JobTitle: "Architect"}},
		Projects:          []resumeProject{{ProjectName: "Project Atlas"}},
	}
	if err := persistResumeIntelligence(fx.db, fx.resumeID, fx.candidateID, fx.tenantID, first); err != nil {
		t.Fatalf("first persistence failed: %v", err)
	}

	second := resumeAIResult{
		CurrentTitle:      "Updated Title",
		Skills:            []string{"Go", "Postgres"},
		EmploymentHistory: []resumeEmployment{{Employer: "Example Corp", JobTitle: "Principal Architect"}},
		Projects:          []resumeProject{{ProjectName: "Project Atlas v2"}},
	}
	if err := persistResumeIntelligence(fx.db, fx.resumeID, fx.candidateID, fx.tenantID, second); err != nil {
		t.Fatalf("second persistence failed: %v", err)
	}

	var employmentCount, projectCount, skillCount int
	if err := fx.db.QueryRow("SELECT COUNT(*) FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3", fx.tenantID, fx.candidateID, fx.resumeID).Scan(&employmentCount); err != nil {
		t.Fatal(err)
	}
	if err := fx.db.QueryRow("SELECT COUNT(*) FROM candidate_projects WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3", fx.tenantID, fx.candidateID, fx.resumeID).Scan(&projectCount); err != nil {
		t.Fatal(err)
	}
	if err := fx.db.QueryRow("SELECT COUNT(*) FROM candidate_expertise WHERE tenant_id=$1 AND candidate_id=$2 AND category='resume_import'", fx.tenantID, fx.candidateID).Scan(&skillCount); err != nil {
		t.Fatal(err)
	}

	if employmentCount != 1 || projectCount != 1 || skillCount != 2 {
		t.Fatalf("reprocessing counts = employment:%d projects:%d skills:%d, want 1,1,2", employmentCount, projectCount, skillCount)
	}

	var title string
	if err := fx.db.QueryRow("SELECT current_title FROM candidate_professional_profiles WHERE tenant_id=$1 AND candidate_id=$2", fx.tenantID, fx.candidateID).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "Updated Title" {
		t.Fatalf("current title = %q, want Updated Title", title)
	}
}

// RAI-03 persistence coverage validates transactional candidate intelligence boundaries.
// Backend CI verifies formatting and the complete persistence test suite.
