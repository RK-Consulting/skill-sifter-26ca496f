package screening

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

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	old, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", env("TEST_DB_HOST", "localhost"), env("TEST_DB_PORT", "5432"), env("TEST_DB_USER", "postgres"), env("TEST_DB_PASSWORD", "postgres"), env("TEST_DB_NAME", "skillsifter_test"))
	d, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skip(err)
	}
	if err := d.Ping(); err != nil {
		d.Close()
		t.Skip(err)
	}
	if err := appdb.InitializeTenantSchema(d); err != nil {
		d.Close()
		t.Fatal(err)
	}
	return d
}
func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}

func fixture(t *testing.T, d *sql.DB) (string, int, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("screening_test_%d", time.Now().UnixNano())
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
		var uid, cid, client, rid int
	must(d.QueryRow("INSERT INTO users (username,email,password,role,tenant_id) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id", "screen-user", tenant+"@u", "x", "recruiter", tenant).Scan(&uid))
	must(d.QueryRow("INSERT INTO candidates (name,email,tenant_id,company_name) VALUES ($1,$2,$3,$4) RETURNING id", "Candidate", tenant+"@c", tenant).Scan(&cid))
	must(d.QueryRow("INSERT INTO clients (name,status,tenant_id) VALUES ($1,$2,$3) RETURNING id", "Client", "active", tenant).Scan(&client))
	must(d.QueryRow("INSERT INTO requirements (client_id,title,status,tenant_id) VALUES ($1,$2,$3,$4) RETURNING id", client, "Requirement", "open", tenant).Scan(&rid))
	clean := func() {
		d.Exec("DELETE FROM recruitment_screenings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM users WHERE tenant_id=$1", tenant)
			}
	return tenant, cid, rid, uid, clean
}

func TestService_CreateAndListCandidateRequirementScreenings(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, cid, rid, uid, clean := fixture(t, d)
	defer clean()
	s := NewService(NewPostgresRepository(d), d)
	got, err := s.CreateScreening(tenant, CreateInput{CandidateID: cid, RequirementID: rid, RecruiterUserID: uid, RecruiterAssessment: "Strong fit"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 || got.CandidateID != cid || got.RequirementID != rid {
		t.Fatalf("unexpected record: %+v", got)
	}
	rows, err := s.ListByCandidateRequirement(tenant, cid, rid)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].RecruiterAssessment != "Strong fit" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestService_TenantIsolation(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, cid, rid, uid, clean := fixture(t, d)
	defer clean()
	other := tenant + "_other"
			s := NewService(NewPostgresRepository(d), d)
	if _, err := s.CreateScreening(tenant, CreateInput{CandidateID: cid, RequirementID: rid, RecruiterUserID: uid}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListByCandidateRequirement(other, cid, rid); err != ErrCandidateNotFound {
		t.Fatalf("got %v, want ErrCandidateNotFound", err)
	}
}
