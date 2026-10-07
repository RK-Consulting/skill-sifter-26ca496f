package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
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

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenvResumeAITest("TEST_DB_HOST", "localhost"),
		getenvResumeAITest("TEST_DB_PORT", "5432"),
		getenvResumeAITest("TEST_DB_USER", "postgres"),
		getenvResumeAITest("TEST_DB_PASSWORD", "postgres"),
		getenvResumeAITest("TEST_DB_NAME", "skillsifter_test"),
	)

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Resume AI persistence test skipped: could not open test DB: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Skipf("Resume AI persistence test skipped: test DB not reachable: %v", err)
	}

	var exists bool
	if err := testDB.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'candidate_professional_profiles')").Scan(&exists); err != nil {
		testDB.Close()
		t.Fatalf("could not inspect Resume AI schema: %v", err)
	}
	if !exists {
		testDB.Close()
		t.Skip("Resume AI persistence test skipped: RAI-03 schema is not initialized")
	}

	tenantID := fmt.Sprintf("rai03_test_%d", time.Now().UnixNano())
		var candidateID int
	if err := testDB.QueryRow(
		"INSERT INTO candidates (name, email, phone, position, location, experience, currentctc, expectedctc, noticeperiod, jobdescription, status, tenant_id, company_name) VALUES ($1, $2, $3, '', '', '', '', '', '', '', 'active', $4, $5) RETURNING id",
		"RAI-03 Test Candidate", "rai03-"+tenantID+"@example.com", "9000000000", tenantID, tenantID+" Company",
	).Scan(&candidateID); err != nil {
		testDB.Close()
		t.Fatalf("could not create test candidate: %v", err)
	}

	var resumeID int
	if err := testDB.QueryRow(
		"INSERT INTO resumes (tenant_id, candidate_id, file_name, file_path, file_hash, mime_type, extracted_text, parsing_status, parser_model) VALUES ($1, $2, 'rai03-test.txt', '/tmp/rai03-test.txt', $3, 'text/plain', 'test resume', 'processing', 'test-model') RETURNING id",
		tenantID, candidateID, fmt.Sprintf("%064d", candidateID),
	).Scan(&resumeID); err != nil {
		testDB.Close()
		t.Fatalf("could not create test resume: %v", err)
	}

	t.Cleanup(func() {
		_, _ = testDB.Exec("DELETE FROM resumes WHERE id = $1", resumeID)
		_, _ = testDB.Exec("DELETE FROM candidates WHERE id = $1", candidateID)
		_, _ = testDB.Exec("DELETE FROM companies WHERE id = $1", tenantID)
		testDB.Close()
		if db.DB == testDB {
			db.DB = nil
		}
	})

	return resumeAITestFixture{db: testDB, tenantID: tenantID, candidateID: candidateID, resumeID: resumeID}
}

func getenvResumeAITest(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestResumeDatePreservesPrecision(t *testing.T) {
	if got, err := resumeDate("2024-06-15"); err != nil || got == nil || got.Format("2006-01-02") != "2024-06-15" {
		t.Fatalf("exact date = %v, err=%v", got, err)
	}
	if got, err := resumeDate("2024-06"); err != nil || got != nil {
		t.Fatalf("month-only date should not become an exact date: got=%v err=%v", got, err)
	}
	if got, err := resumeDate("2024"); err != nil || got != nil {
		t.Fatalf("year-only date should not become an exact date: got=%v err=%v", got, err)
	}
	if _, err := resumeDate("not-a-date"); err == nil {
		t.Fatal("invalid date should be rejected")
	}
	if got := resumeYearFromPartialDate("2024-06", 0); got == nil || *got != 2024 {
		t.Fatalf("month-only year = %v, want 2024", got)
	}
}

func TestPersistResumeIntelligence(t *testing.T) {
	fx := setupResumeAITestFixture(t)
	ai := resumeAIResult{
		CurrentTitle:        "Software Architect",
		ProfessionalSummary: "Structured test profile",
		Skills:              []string{"golang", "Postgres"},
		Languages:           []resumeLanguage{{Name: "English", ProficiencyFramework: "CEFR", ProficiencyLevel: "C1"}},
		EmploymentHistory:   []resumeEmployment{{Employer: "Example Corp", JobTitle: "Architect", StartDate: "2024-06", IsCurrent: true}},
		Education:           []resumeEducation{{Institution: "Example University", Degree: "B.E.", FieldOfStudy: "Electronics", StartDate: "2010", EndYear: 2014}},
		Certifications:      []resumeCertification{{Name: "AWS Certified Developer", Issuer: "AWS", IssueDate: "2024-05-20"}},
		Projects:            []resumeProject{{ProjectName: "Project Atlas", Role: "Lead", Technologies: []string{"golang", "Postgres"}, StartDate: "2023", EndYear: 2024}},
	}

	if err := persistResumeIntelligence(db.DB, fx.resumeID, fx.candidateID, fx.tenantID, ai); err != nil {
		t.Fatalf("persistResumeIntelligence failed: %v", err)
	}

	counts := map[string]int{}
	queries := map[string]string{
		"profile":        "SELECT COUNT(*) FROM candidate_professional_profiles WHERE tenant_id=$1 AND candidate_id=$2",
		"skills":         "SELECT COUNT(*) FROM candidate_expertise WHERE tenant_id=$1 AND candidate_id=$2 AND category='resume_import'",
		"languages":      "SELECT COUNT(*) FROM candidate_language_expertise WHERE tenant_id=$1 AND candidate_id=$2",
		"employment":     "SELECT COUNT(*) FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3",
		"education":      "SELECT COUNT(*) FROM candidate_education WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3",
		"certifications": "SELECT COUNT(*) FROM candidate_certifications WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3",
		"projects":       "SELECT COUNT(*) FROM candidate_projects WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3",
	}
	for name, query := range queries {
		var count int
		var err error
		if name == "skills" || name == "profile" || name == "languages" {
			err = fx.db.QueryRow(query, fx.tenantID, fx.candidateID).Scan(&count)
		} else {
			err = fx.db.QueryRow(query, fx.tenantID, fx.candidateID, fx.resumeID).Scan(&count)
		}
		if err != nil {
			t.Fatalf("%s count query failed: %v", name, err)
		}
		counts[name] = count
	}
	for _, name := range []string{"profile", "languages", "employment", "education", "certifications", "projects"} {
		if counts[name] != 1 {
			t.Errorf("%s count = %d, want 1", name, counts[name])
		}
	}
	if counts["skills"] != 2 {
		t.Errorf("skills count = %d, want 2", counts["skills"])
	}

	var startDate sql.NullTime
	var startYear int
	if err := fx.db.QueryRow("SELECT start_date, start_year FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3", fx.tenantID, fx.candidateID, fx.resumeID).Scan(&startDate, &startYear); err != nil {
		t.Fatal(err)
	}
	if startDate.Valid {
		t.Fatalf("month-only employment date was materialized as %v", startDate.Time)
	}
	if startYear != 2024 {
		t.Fatalf("month-only employment year = %d, want 2024", startYear)
	}
}

func TestPersistResumeIntelligenceProvenance(t *testing.T) {
	fx := setupResumeAITestFixture(t)
	ai := resumeAIResult{
		Skills:            []string{"Go"},
		Languages:         []resumeLanguage{{Name: "English"}},
		EmploymentHistory: []resumeEmployment{{Employer: "Example Corp"}},
		Education:         []resumeEducation{{Institution: "Example University"}},
		Certifications:    []resumeCertification{{Name: "Certification"}},
		Projects:          []resumeProject{{ProjectName: "Project"}},
	}
	if err := persistResumeIntelligence(db.DB, fx.resumeID, fx.candidateID, fx.tenantID, ai); err != nil {
		t.Fatalf("persistResumeIntelligence failed: %v", err)
	}

	queries := map[string]string{
		"profile":       "SELECT source_resume_id FROM candidate_professional_profiles WHERE tenant_id=$1 AND candidate_id=$2",
		"language":      "SELECT source_resume_id FROM candidate_language_expertise WHERE tenant_id=$1 AND candidate_id=$2",
		"employment":    "SELECT source_resume_id FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2",
		"education":     "SELECT source_resume_id FROM candidate_education WHERE tenant_id=$1 AND candidate_id=$2",
		"certification": "SELECT source_resume_id FROM candidate_certifications WHERE tenant_id=$1 AND candidate_id=$2",
		"project":       "SELECT source_resume_id FROM candidate_projects WHERE tenant_id=$1 AND candidate_id=$2",
	}
	for name, query := range queries {
		var sourceResumeID int
		if err := fx.db.QueryRow(query, fx.tenantID, fx.candidateID).Scan(&sourceResumeID); err != nil {
			t.Fatalf("%s provenance query failed: %v", name, err)
		}
		if sourceResumeID != fx.resumeID {
			t.Errorf("%s source_resume_id = %d, want %d", name, sourceResumeID, fx.resumeID)
		}
	}
}

func TestResumeAICandidateAssociationIsTenantScoped(t *testing.T) {
	fx := setupResumeAITestFixture(t)
	otherTenant := fmt.Sprintf("rai03_other_%d", time.Now().UnixNano())
	if _, err := fx.db.Exec("INSERT INTO platform_tenants(tenant_id,company_name,provisioning_status,account_status) VALUES($1,$2,'READY','ACTIVE')", otherTenant, otherTenant+" Company"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = fx.db.Exec("DELETE FROM platform_tenants WHERE tenant_id=$1", otherTenant)
	}()

	var otherCandidateID int
	if err := fx.db.QueryRow(
		"INSERT INTO candidates (name, email, phone, position, location, experience, currentctc, expectedctc, noticeperiod, jobdescription, status, tenant_id, company_name) VALUES ('Other Tenant Candidate', 'shared@example.com', '9111111111', '', '', '', '', '', '', '', 'active', $1, $2) RETURNING id",
		otherTenant, otherTenant+" Company",
	).Scan(&otherCandidateID); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = fx.db.Exec("DELETE FROM candidates WHERE id=$1", otherCandidateID) }()

	if _, err := fx.db.Exec("UPDATE candidates SET email='shared@example.com', phone='9222222222' WHERE id=$1", fx.candidateID); err != nil {
		t.Fatal(err)
	}

	got, err := upsertResumeCandidate(db.DB, "Fixture Company", fx.tenantID, resumeAIResult{
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

	if err := persistResumeIntelligence(db.DB, fx.resumeID, fx.candidateID, fx.tenantID, ai); err == nil {
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
	if err := persistResumeIntelligence(db.DB, fx.resumeID, fx.candidateID, fx.tenantID, first); err != nil {
		t.Fatalf("first persistence failed: %v", err)
	}

	second := resumeAIResult{
		CurrentTitle:      "Updated Title",
		Skills:            []string{"Go", "Postgres"},
		EmploymentHistory: []resumeEmployment{{Employer: "Example Corp", JobTitle: "Principal Architect"}},
		Projects:          []resumeProject{{ProjectName: "Project Atlas v2"}},
	}
	if err := persistResumeIntelligence(db.DB, fx.resumeID, fx.candidateID, fx.tenantID, second); err != nil {
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
