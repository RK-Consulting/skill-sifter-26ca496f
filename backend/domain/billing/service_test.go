package billing

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

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	old, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env("TEST_DB_HOST", "localhost"), env("TEST_DB_PORT", "5432"),
		env("TEST_DB_USER", "postgres"), env("TEST_DB_PASSWORD", "postgres"),
		env("TEST_DB_NAME", "skillsifter_test"),
	)
	d, err := sql.Open("postgres", dsn)
	if err != nil { t.Skip(err) }
	if err := d.Ping(); err != nil { d.Close(); t.Skip(err) }
	appdb.DB = d
	if err := appdb.InitializeSchema(); err != nil { d.Close(); t.Fatal(err) }
	return d
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" { return v }
	return fallback
}

func fixture(t *testing.T, d *sql.DB) (string, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("billing_test_%d", os.Getpid())
	must := func(err error) { if err != nil { t.Fatal(err) } }
	exec := func(q string, args ...interface{}) { _, err := d.Exec(q, args...); must(err) }
	exec("INSERT INTO companies(id,name) VALUES($1,$2)", tenant, tenant)

	var candidateID, clientID, requirementID, selectionID int
	must(d.QueryRow("INSERT INTO candidates(name,email,tenant_id,company_name) VALUES($1,$2,$3,$4) RETURNING id",
		"Candidate", tenant+"@candidate", tenant, tenant).Scan(&candidateID))
	must(d.QueryRow("INSERT INTO clients(name,status,tenant_id) VALUES($1,$2,$3) RETURNING id",
		"Client", "active", tenant).Scan(&clientID))
	must(d.QueryRow("INSERT INTO requirements(client_id,title,status,tenant_id) VALUES($1,$2,$3,$4) RETURNING id",
		clientID, "Requirement", "open", tenant).Scan(&requirementID))
	must(d.QueryRow("INSERT INTO recruitment_selections(tenant_id,candidate_id,requirement_id,decision) VALUES($1,$2,$3,'selected') RETURNING id",
		tenant, candidateID, requirementID).Scan(&selectionID))
	var offerID int
	must(d.QueryRow("INSERT INTO recruitment_offers(tenant_id,candidate_id,requirement_id,selection_id,accepted) VALUES($1,$2,$3,$4,true) RETURNING id",
		tenant, candidateID, requirementID, selectionID).Scan(&offerID))
	_ = offerID
	date := time.Now()
	mustExec := exec
	mustExec("INSERT INTO recruitment_joinings(tenant_id,candidate_id,requirement_id,offer_id,joining_date,joined) VALUES($1,$2,$3,$4,$5,true)",
		tenant, candidateID, requirementID, offerID, date)

	cleanup := func() {
		d.Exec("DELETE FROM recruitment_billings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_joinings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_offers WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_selections WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM companies WHERE id=$1", tenant)
	}
	return tenant, candidateID, requirementID, cleanup
}

func TestService_CreateBillingRequiresJoined(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()
	if _, err := d.Exec("UPDATE recruitment_joinings SET joined=false, joining_date=NULL WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3",
		tenant, candidateID, requirementID); err != nil { t.Fatal(err) }
	_, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID: candidateID, RequirementID: requirementID, Amount: "50000.00", Currency: "INR",
	})
	if err != ErrJoiningNotFound { t.Fatalf("got %v, want ErrJoiningNotFound", err) }
}

func TestService_CreateBilling(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()
	got, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID: candidateID, RequirementID: requirementID, Amount: "50000.00", Currency: "inr",
		InvoiceReference: "INV-001",
	})
	if err != nil { t.Fatal(err) }
	if got.ID == 0 || got.BillingDate.IsZero() || got.Currency != "INR" || got.Amount != "50000.00" {
		t.Fatalf("unexpected billing: %+v", got)
	}
}

func TestService_DuplicateBilling(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()
	s := NewService(NewPostgresRepository(d), d)
	if _, err := s.Create(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID, Amount: "50000.00", Currency: "INR"}); err != nil { t.Fatal(err) }
	if _, err := s.Create(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID, Amount: "60000.00", Currency: "INR"}); err != ErrBillingExists {
		t.Fatalf("got %v, want ErrBillingExists", err)
	}
}

func TestService_TenantIsolation(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()
	other := tenant + "_other"
	if _, err := d.Exec("INSERT INTO companies(id,name) VALUES($1,$2)", other, other); err != nil { t.Fatal(err) }
	defer d.Exec("DELETE FROM companies WHERE id=$1", other)
	if _, err := NewService(NewPostgresRepository(d), d).Get(other, candidateID, requirementID); err != ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
