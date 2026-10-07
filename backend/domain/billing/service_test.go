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

func billingTestDB(t *testing.T) *sql.DB {
	t.Helper()

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env("TEST_DB_HOST", "localhost"),
		env("TEST_DB_PORT", "5432"),
		env("TEST_DB_USER", "postgres"),
		env("TEST_DB_PASSWORD", "postgres"),
		env("TEST_DB_NAME", "skillsifter_test"),
	)

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

func billingFixture(t *testing.T, d *sql.DB) (string, int, int, func()) {
	t.Helper()

	tenant := fmt.Sprintf("billing_test_%d", os.Getpid())
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := d.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}

	
	var candidateID, clientID, requirementID, selectionID int
	if err := d.QueryRow(
		"INSERT INTO candidates(name,email,tenant_id) VALUES($1,$2,$3,$4) RETURNING id",
		"Candidate", tenant+"@candidate", tenant, tenant,
	).Scan(&candidateID); err != nil {
		t.Fatal(err)
	}

	if err := d.QueryRow(
		"INSERT INTO clients(name,status,tenant_id) VALUES($1,$2,$3) RETURNING id",
		"Client", "active", tenant,
	).Scan(&clientID); err != nil {
		t.Fatal(err)
	}

	if err := d.QueryRow(
		"INSERT INTO requirements(client_id,title,status,tenant_id) VALUES($1,$2,$3,$4) RETURNING id",
		clientID, "Requirement", "open", tenant,
	).Scan(&requirementID); err != nil {
		t.Fatal(err)
	}

	if err := d.QueryRow(
		"INSERT INTO recruitment_selections(tenant_id,candidate_id,requirement_id,decision) VALUES($1,$2,$3,'selected') RETURNING id",
		tenant, candidateID, requirementID,
	).Scan(&selectionID); err != nil {
		t.Fatal(err)
	}

	var offerID int
	if err := d.QueryRow(
		"INSERT INTO recruitment_offers(tenant_id,candidate_id,requirement_id,selection_id,accepted) VALUES($1,$2,$3,$4,true) RETURNING id",
		tenant, candidateID, requirementID, selectionID,
	).Scan(&offerID); err != nil {
		t.Fatal(err)
	}

	date := time.Now()
	exec(
		"INSERT INTO recruitment_joinings(tenant_id,candidate_id,requirement_id,offer_id,joining_date,joined) VALUES($1,$2,$3,$4,$5,true)",
		tenant, candidateID, requirementID, offerID, date,
	)

	cleanup := func() {
		d.Exec("DELETE FROM recruitment_billings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_joinings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_offers WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_selections WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
			}

	return tenant, candidateID, requirementID, cleanup
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestService_CreateBillingRequiresJoined(t *testing.T) {
	d := billingTestDB(t)
	defer d.Close()

	tenant, candidateID, requirementID, cleanup := billingFixture(t, d)
	defer cleanup()

	if _, err := d.Exec(
		"UPDATE recruitment_joinings SET joined=false, joining_date=NULL WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3",
		tenant, candidateID, requirementID,
	); err != nil {
		t.Fatal(err)
	}

	_, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID:   candidateID,
		RequirementID: requirementID,
		Amount:        "50000.00",
		Currency:      "INR",
	})
	if err != ErrJoiningNotFound {
		t.Fatalf("got %v, want ErrJoiningNotFound", err)
	}
}

func TestService_CreateBilling(t *testing.T) {
	d := billingTestDB(t)
	defer d.Close()

	tenant, candidateID, requirementID, cleanup := billingFixture(t, d)
	defer cleanup()

	got, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID:      candidateID,
		RequirementID:    requirementID,
		Amount:           "50000.00",
		Currency:         "inr",
		InvoiceReference: "INV-001",
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.ID == 0 || got.BillingDate.IsZero() || got.Currency != "INR" || got.Amount != "50000.00" {
		t.Fatalf("unexpected billing: %+v", got)
	}
}

func TestService_DuplicateBilling(t *testing.T) {
	d := billingTestDB(t)
	defer d.Close()

	tenant, candidateID, requirementID, cleanup := billingFixture(t, d)
	defer cleanup()

	service := NewService(NewPostgresRepository(d), d)
	input := CreateInput{
		CandidateID:   candidateID,
		RequirementID: requirementID,
		Amount:        "50000.00",
		Currency:      "INR",
	}

	if _, err := service.Create(tenant, input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(tenant, input); err != ErrBillingExists {
		t.Fatalf("got %v, want ErrBillingExists", err)
	}
}

func TestService_TenantIsolation(t *testing.T) {
	d := billingTestDB(t)
	defer d.Close()

	tenant, candidateID, requirementID, cleanup := billingFixture(t, d)
	defer cleanup()

	other := tenant + "_other"
		if _, err := NewService(NewPostgresRepository(d), d).Get(other, candidateID, requirementID); err != ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
