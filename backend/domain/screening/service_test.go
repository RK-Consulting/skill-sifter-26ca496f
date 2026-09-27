package screening

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

func openScreeningTestDB(t *testing.T) *sql.DB {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	old, _ := os.Getwd()
	if err := os.Chdir(filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))); err != nil { t.Fatal(err) }
	defer os.Chdir(old)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", envScreen("TEST_DB_HOST","localhost"), envScreen("TEST_DB_PORT","5432"), envScreen("TEST_DB_USER","postgres"), envScreen("TEST_DB_PASSWORD","postgres"), envScreen("TEST_DB_NAME","skillsifter_test"))
	d, err := sql.Open("postgres", dsn)
	if err != nil { t.Skip(err) }
	if err := d.Ping(); err != nil { d.Close(); t.Skip(err) }
	appdb.DB = d
	if err := appdb.InitializeSchema(); err != nil { d.Close(); t.Fatal(err) }
	return d
}

func envScreen(k, fallback string) string { if v := os.Getenv(k); v != "" { return v }; return fallback }

func seedScreening(t *testing.T, d *sql.DB) (string, int, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("screening_test_%d", time.Now().UnixNano())
	must := func(err error) { if err != nil { t.Fatal(err) } }
	_, err := d.Exec("INSERT INTO companies (id,name) VALUES ($1,$2)", tenant, tenant); must(err)
	var userID, candidateID, clientID, requirementID int
	must(d.QueryRow("INSERT INTO users (username,email,password,role,tenant_id,company_name) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id", "screen-user", tenant+"@u", "x", "recruiter", tenant, tenant).Scan(&userID))
	must(d.QueryRow("INSERT INTO candidates (name,email,tenant_id,company_name) VALUES ($1,$2,$3,$4) RETURNING id", "Candidate", tenant+"@c", tenant, tenant).Scan(&candidateID))
	must(d.QueryRow("INSERT INTO clients (name,status,tenant_id) VALUES ($1,$2,$3) RETURNING id", "Client", "active", tenant).Scan(&clientID))
	must(d.QueryRow("INSERT INTO requirements (client_id,title,status,tenant_id) VALUES ($1,$2,$3,$4) RETURNING id", clientID, "Requirement", "open", tenant).Scan(&requirementID))
	cleanup := func() {
		d.Exec("DELETE FROM recruitment_screenings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM users WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM companies WHERE id=$1", tenant)
	}
	return tenant, candidateID, requirementID, userID, cleanup
}

func TestService_CreateAndListCandidateRequirementScreenings(t *testing.T) {
	d := openScreeningTestDB(t); defer d.Close()
	tenant, candidateID, requirementID, userID, cleanup := seedScreening(t, d); defer cleanup()
	s := NewService(NewPostgresRepository(d), d)
	got, err := s.CreateScreening(tenant, CreateInput{CandidateID:candidateID, RequirementID:requirementID, RecruiterUserID:userID, RecruiterAssessment:"Strong fit"})
	if err != nil { t.Fatal(err) }
	if got.ID == 0 || got.CandidateID != candidateID || got.RequirementID != requirementID { t.Fatalf("unexpected screening: %+v", got) }
	rows, err := s.ListByCandidateRequirement(tenant, candidateID, requirementID)
	if err != nil { t.Fatal(err) }
	if len(rows) != 1 || rows[0].RecruiterAssessment != "Strong fit" { t.Fatalf("unexpected rows: %+v", rows) }
}

func TestService_TenantIsolation(t *testing.T) {
	d := openScreeningTestDB(t); defer d.Close()
	tenant, candidateID, requirementID, userID, cleanup := seedScreening(t, d); defer cleanup()
	other := tenant + "_other"
	if _, err := d.Exec("INSERT INTO companies (id,name) VALUES ($1,$2)", other, other); err != nil { t.Fatal(err) }
	defer d.Exec("DELETE FROM companies WHERE id=$1", other)
	s := NewService(NewPostgresRepository(d), d)
	if _, err := s.CreateScreening(tenant, CreateInput{CandidateID:candidateID, RequirementID:requirementID, RecruiterUserID:userID}); err != nil { t.Fatal(err) }
	if _, err := s.ListByCandidateRequirement(other, candidateID, requirementID); err != ErrCandidateNotFound { t.Fatalf("got %v, want ErrCandidateNotFound", err) }
}
