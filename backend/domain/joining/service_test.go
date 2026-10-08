package joining

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

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func fixture(t *testing.T, d *sql.DB) (string, int, int, func()) {
	t.Helper()
	tenant := fmt.Sprintf("joining_test_%d", os.Getpid())
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	var candidateID, clientID, requirementID, selectionID int
	must(d.QueryRow(
		"INSERT INTO candidates(name,email,tenant_id) VALUES($1,$2,$3) RETURNING id",
		"Candidate", tenant+"@candidate", tenant,
	).Scan(&candidateID))
	must(d.QueryRow(
		"INSERT INTO clients(name,status,tenant_id) VALUES($1,$2,$3) RETURNING id",
		"Client", "active", tenant,
	).Scan(&clientID))
	must(d.QueryRow(
		"INSERT INTO requirements(client_id,title,status,tenant_id) VALUES($1,$2,$3,$4) RETURNING id",
		clientID, "Requirement", "open", tenant,
	).Scan(&requirementID))
	must(d.QueryRow(
		"INSERT INTO recruitment_selections(tenant_id,candidate_id,requirement_id,decision) VALUES($1,$2,$3,'selected') RETURNING id",
		tenant, candidateID, requirementID,
	).Scan(&selectionID))
	must(d.QueryRow(
		"INSERT INTO recruitment_offers(tenant_id,candidate_id,requirement_id,selection_id,accepted) VALUES($1,$2,$3,$4,true) RETURNING id",
		tenant, candidateID, requirementID, selectionID,
	).Scan(new(int)))

	cleanup := func() {
		d.Exec("DELETE FROM recruitment_joinings WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_offers WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM recruitment_selections WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM requirements WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM clients WHERE tenant_id=$1", tenant)
		d.Exec("DELETE FROM candidates WHERE tenant_id=$1", tenant)
	}
	return tenant, candidateID, requirementID, cleanup
}

func TestService_CreateJoiningRequiresAcceptedOffer(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()

	if _, err := d.Exec(
		"UPDATE recruitment_offers SET accepted=false WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3",
		tenant, candidateID, requirementID,
	); err != nil {
		t.Fatal(err)
	}

	_, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID: candidateID, RequirementID: requirementID,
	})
	if err != ErrOfferNotFound {
		t.Fatalf("got %v, want ErrOfferNotFound", err)
	}
}

func TestService_CreateJoining(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()

	date := time.Now()
	got, err := NewService(NewPostgresRepository(d), d).Create(tenant, CreateInput{
		CandidateID: candidateID, RequirementID: requirementID, JoiningDate: &date,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 || got.Joined {
		t.Fatalf("unexpected joining: %+v", got)
	}
}

func TestService_JoinedRequiresDate(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()

	s := NewService(NewPostgresRepository(d), d)
	if _, err := s.Create(tenant, CreateInput{CandidateID: candidateID, RequirementID: requirementID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(tenant, candidateID, requirementID, UpdateInput{Joined: true}); err != ErrJoiningDateRequired {
		t.Fatalf("got %v, want ErrJoiningDateRequired", err)
	}

	date := time.Now()
	got, err := s.Update(tenant, candidateID, requirementID, UpdateInput{
		JoiningDate: &date, Joined: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Joined || got.JoiningDate == nil {
		t.Fatalf("unexpected joined record: %+v", got)
	}
}

func TestService_TenantIsolation(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	tenant, candidateID, requirementID, cleanup := fixture(t, d)
	defer cleanup()

	other := tenant + "_other"
	if _, err := NewService(NewPostgresRepository(d), d).Get(other, candidateID, requirementID); err != ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
