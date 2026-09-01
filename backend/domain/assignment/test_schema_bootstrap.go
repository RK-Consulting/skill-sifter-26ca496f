package assignment

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
)

const assignmentTestSchema = "assignment_test"

func getenvOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type fixtures struct {
	userID        int
	candidateID   int
	requirementID int
}

var seedFixturesCounter int64

func seedFixtures(t *testing.T, testDB *sql.DB, tenantID string) fixtures {
	t.Helper()
	n := atomic.AddInt64(&seedFixturesCounter, 1)
	unique := fmt.Sprintf("%s_%d", tenantID, n)
	var f fixtures
	if err := testDB.QueryRow(`INSERT INTO users (username,email,password,role,tenant_id,company_name) VALUES ($1,$2,'x','recruiter',$3,$3) RETURNING id`, "user_"+unique, unique+"user@test.com", tenantID).Scan(&f.userID); err != nil {
		t.Fatalf("seedFixtures: could not insert user: %v", err)
	}
	if err := testDB.QueryRow(`INSERT INTO candidates (name,email,tenant_id,company_name,status) VALUES ('Test Candidate',$1,$2,$2,'active') RETURNING id`, unique+"cand@test.com", tenantID).Scan(&f.candidateID); err != nil {
		t.Fatalf("seedFixtures: could not insert candidate: %v", err)
	}
	var clientID int
	if err := testDB.QueryRow(`INSERT INTO clients (name,tenant_id) VALUES ('Test Client',$1) RETURNING id`, tenantID).Scan(&clientID); err != nil {
		t.Fatalf("seedFixtures: could not insert client: %v", err)
	}
	if err := testDB.QueryRow(`INSERT INTO requirements (client_id,title,tenant_id) VALUES ($1,'Test Req',$2) RETURNING id`, clientID, tenantID).Scan(&f.requirementID); err != nil {
		t.Fatalf("seedFixtures: could not insert requirement: %v", err)
	}
	return f
}

func TestMain(m *testing.M) {
	testDB, err := openAssignmentIntegrationDB()
	if err != nil {
		os.Exit(m.Run())
	}
	defer testDB.Close()
	if err := chdirToBackendRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap: %v\n", err)
		os.Exit(1)
	}
	if _, err := testDB.Exec(`DROP SCHEMA assignment_test CASCADE; CREATE SCHEMA assignment_test;`); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test database reset failed: %v\n", err)
		os.Exit(1)
	}
	db.DB = testDB
	if err := db.InitializeSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "assignment test schema bootstrap failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func openAssignmentIntegrationDB() (*sql.DB, error) {
	host := getenvOr("TEST_DB_HOST", "localhost")
	port := getenvOr("TEST_DB_PORT", "5432")
	user := getenvOr("TEST_DB_USER", "postgres")
	password := getenvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getenvOr("TEST_DB_NAME", "skillsifter_assignment_test")
	baseConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	adminDB, err := sql.Open("postgres", baseConn)
	if err != nil {
		return nil, err
	}
	if err := adminDB.Ping(); err != nil {
		adminDB.Close()
		return nil, err
	}
	if _, err := adminDB.Exec(`CREATE SCHEMA IF NOT EXISTS assignment_test`); err != nil {
		adminDB.Close()
		return nil, err
	}
	adminDB.Close()

	conn := baseConn + " search_path=" + assignmentTestSchema
	testDB, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err := testDB.Ping(); err != nil {
		testDB.Close()
		return nil, err
	}
	return testDB, nil
}

func chdirToBackendRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return os.Chdir(wd)
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return fmt.Errorf("could not locate backend go.mod from %q", wd)
		}
		wd = parent
	}
}
