package submission

import (
	"database/sql"
	"fmt"
	appdb "github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func submissionDB(t *testing.T) *sql.DB {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	old, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	d, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", envS("TEST_DB_HOST", "localhost"), envS("TEST_DB_PORT", "5432"), envS("TEST_DB_USER", "postgres"), envS("TEST_DB_PASSWORD", "postgres"), envS("TEST_DB_NAME", "skillsifter_test")))
	if err != nil {
		t.Skip(err)
	}
	if err = d.Ping(); err != nil {
		d.Close()
		t.Skip(err)
	}
	if err = appdb.InitializeSchema(); err != nil {
		d.Close()
		t.Fatal(err)
	}
	return d
}
func envS(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
func subFixture(t *testing.T, d *sql.DB) (string, int, int, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("submission_test_%d", time.Now().UnixNano())
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
		var uid, cid, rid, client int
	must(d.QueryRow("INSERT INTO users (username,email,password,role,tenant_id) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id", "sub-user", tenant+"@u", "x", "recruiter", tenant).Scan(&uid))
	must(d.QueryRow("INSERT INTO candidates (name,email,tenant_id,company_name) VALUES ($1,$2,$3,$4) RETURNING id", "Candidate", tenant+"@c", tenant).Scan(&cid))
	must(d.QueryRow("INSERT INTO clients (name,status,tenant_id) VALUES ($1,$2,$3) RETURNING id", "Client", "active", tenant).Scan(&client))
	must(d.QueryRow("INSERT INTO requirements (client_id,title,status,tenant_id) VALUES ($1,$2,$3,$4) RETURNING id", client, "Requirement", "open", tenant).Scan(&rid))
	clean := func() {
		d.Exec("DELETE FROM recruitment_submission_feedback WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_submissions WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_screenings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM users WHERE tenant_id=$1", tenant)
			}
	return tenant, cid, rid, uid, client, clean
}
func completedScreening(t *testing.T, d *sql.DB, tenant string, cid, rid, uid int) {
	t.Helper()
	_, e := d.Exec("INSERT INTO recruitment_screenings (tenant_id,candidate_id,requirement_id,recruiter_user_id,status) VALUES ($1,$2,$3,$4,'completed')", tenant, cid, rid, uid)
	if e != nil {
		t.Fatal(e)
	}
}

func TestService_SubmitRequiresCompletedScreening(t *testing.T) {
	d := submissionDB(t)
	defer d.Close()
	tenant, cid, rid, uid, client, clean := subFixture(t, d)
	defer clean()
	s := NewService(NewPostgresRepository(d), d)
	_, err := s.Submit(tenant, CreateInput{CandidateID: cid, RequirementID: rid, SubmittedByUserID: uid, RecipientType: RecipientClient, RecipientClientID: &client})
	if err == nil || err.Error() != "a completed screening is required before submission" {
		t.Fatalf("got %v", err)
	}
}
func TestService_SubmitAndRejectDuplicate(t *testing.T) {
	d := submissionDB(t)
	defer d.Close()
	tenant, cid, rid, uid, client, clean := subFixture(t, d)
	defer clean()
	completedScreening(t, d, tenant, cid, rid, uid)
	s := NewService(NewPostgresRepository(d), d)
	got, err := s.Submit(tenant, CreateInput{CandidateID: cid, RequirementID: rid, SubmittedByUserID: uid, RecipientType: RecipientClient, RecipientClientID: &client, RecipientName: "Client", RecipientEmail: "client@test"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 || len(got.CandidateSnapshot) == 0 || len(got.RequirementSnapshot) == 0 {
		t.Fatalf("snapshots missing: %+v", got)
	}
	_, err = s.Submit(tenant, CreateInput{CandidateID: cid, RequirementID: rid, SubmittedByUserID: uid, RecipientType: RecipientClient, RecipientClientID: &client})
	if err != ErrAlreadySubmitted {
		t.Fatalf("got %v, want ErrAlreadySubmitted", err)
	}
}
func TestService_SubmitRejectsCrossTenantCandidate(t *testing.T) {
	d := submissionDB(t)
	defer d.Close()
	tenant, cid, rid, uid, client, clean := subFixture(t, d)
	defer clean()
	other := tenant + "_other"
			s := NewService(NewPostgresRepository(d), d)
	_, err := s.Submit(other, CreateInput{CandidateID: cid, RequirementID: rid, SubmittedByUserID: uid, RecipientType: RecipientClient, RecipientClientID: &client})
	if err != ErrCandidateNotFound {
		t.Fatalf("got %v, want ErrCandidateNotFound", err)
	}
}
