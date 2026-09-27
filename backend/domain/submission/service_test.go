package submission

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	appdb "github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
)

func openSubmissionTestDB(t *testing.T) *sql.DB {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	old, _ := os.Getwd()
	if err := os.Chdir(filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", envSub("TEST_DB_HOST", "localhost"), envSub("TEST_DB_PORT", "5432"), envSub("TEST_DB_USER", "postgres"), envSub("TEST_DB_PASSWORD", "postgres"), envSub("TEST_DB_NAME", "skillsifter_test"))
	d, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skip(err)
	}
	if err := d.Ping(); err != nil {
		d.Close()
		t.Skip(err)
	}
	appdb.DB = d
	if err := appdb.InitializeSchema(); err != nil {
		d.Close()
		t.Fatal(err)
	}
	return d
}

func envSub(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func seedSubmission(t *testing.T, d *sql.DB) (string, int, int, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("submission_test_%d", time.Now().UnixNano())
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := d.Exec("INSERT INTO companies (id,name) VALUES ($1,$2)", tenant, tenant)
	must(err)
	var userID, candidateID, clientID, requirementID int
	must(d.QueryRow("INSERT INTO users (username,email,password,role,tenant_id,company_name) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id", "sub-user", tenant+"@u", "x", "recruiter", tenant, tenant).Scan(&userID))
	must(d.QueryRow("INSERT INTO candidates (name,email,tenant_id,company_name) VALUES ($1,$2,$3,$4) RETURNING id", "Candidate", tenant+"@c", tenant, tenant).Scan(&candidateID))
	must(d.QueryRow("INSERT INTO clients (name,status,tenant_id) VALUES ($1,$2,$3) RETURNING id", "Client", "active", tenant).Scan(&clientID))
	must(d.QueryRow("INSERT INTO requirements (client_id,title,status,tenant_id) VALUES ($1,$2,$3,$4) RETURNING id", clientID, "Requirement", "open", tenant).Scan(&requirementID))
	cleanup := func() {
		d.Exec("DELETE FROM recruitment_submission_feedback WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_submissions WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_screenings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM users WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM companies WHERE id=$1", tenant)
	}
	return tenant, candidateID, requirementID, userID, clientID, cleanup
}

func seedCompletedScreening(t *testing.T, d *sql.DB, tenant string, candidateID, requirementID, userID int) {
	t.Helper()
	if _, err := d.Exec("INSERT INTO recruitment_screenings (tenant_id,candidate_id,requirement_id,recruiter_user_id,status) VALUES ($1,$2,$3,$4,'completed')", tenant, candidateID, requirementID, userID); err != nil {
		t.Fatal(err)
	}
}

func TestService_SubmitRequiresCompletedScreening(t *testing.T) {
	d := openSubmissionTestDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, userID, clientID, cleanup := seedSubmission(t, d)
	defer cleanup()
	s := NewService(NewPostgresRepository(d), d)
	_, err := s.Submit(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID, SubmittedByUserID: userID, RecipientType: RecipientClient, RecipientClientID: &clientID})
	if err == nil || err.Error() != "a completed screening is required before submission" {
		t.Fatalf("got %v, want completed-screening gate", err)
	}
}

func TestService_SubmitAndRejectDuplicate(t *testing.T) {
	d := openSubmissionTestDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, userID, clientID, cleanup := seedSubmission(t, d)
	defer cleanup()
	seedCompletedScreening(t, d, tenant, candidateID, requirementID, userID)
	s := NewService(NewPostgresRepository(d), d)
	got, err := s.Submit(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID, SubmittedByUserID: userID, RecipientType: RecipientClient, RecipientClientID: &clientID, RecipientName: "Client", RecipientEmail: "client@test"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 || len(got.CandidateSnapshot) == 0 || len(got.RequirementSnapshot) == 0 {
		t.Fatalf("snapshots missing: %+v", got)
	}
	_, err = s.Submit(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID, SubmittedByUserID: userID, RecipientType: RecipientClient, RecipientClientID: &clientID})
	if err != ErrAlreadySubmitted {
		t.Fatalf("got %v, want ErrAlreadySubmitted", err)
	}
}

func TestService_SubmitRejectsCrossTenantCandidate(t *testing.T) {
	d := openSubmissionTestDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, userID, clientID, cleanup := seedSubmission(t, d)
	defer cleanup()
	other := tenant + "_other"
	if _, err := d.Exec("INSERT INTO companies (id,name) VALUES ($1,$2)", other, other); err != nil {
		t.Fatal(err)
	}
	defer d.Exec("DELETE FROM companies WHERE id=$1", other)
	s := NewService(NewPostgresRepository(d), d)
	_, err := s.Submit(other, CreateInput{CandidateID: candidateID, RequirementID: requirementID, SubmittedByUserID: userID, RecipientType: RecipientClient, RecipientClientID: &clientID})
	if err != ErrCandidateNotFound {
		t.Fatalf("got %v, want ErrCandidateNotFound", err)
	}
}
