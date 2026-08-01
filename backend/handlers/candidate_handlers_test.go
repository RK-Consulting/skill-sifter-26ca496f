package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// setupTestDB connects to a real Postgres instance for integration testing.
// Skips (not fails) if no test database is reachable, so `go test ./...`
// still works in environments without a DB configured — but when a DB IS
// available (CI, or run manually), this exercises the exact code path that
// broke silently in production: does the SQL actually match the schema?
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "skillsifter_test"
	}

	connStr := "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("skipping integration test: could not open test DB connection: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Skipf("skipping integration test: test DB not reachable (%v). "+
			"Set TEST_DB_HOST/PORT/USER/PASSWORD/NAME, or run against a local Postgres "+
			"with a 'skillsifter_test' database to enable this test.", err)
	}

	testDB.Exec(`
		CREATE TABLE IF NOT EXISTS candidates (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			position VARCHAR(100),
			location VARCHAR(50),
			experience VARCHAR(100),
			currentctc VARCHAR(100),
			expectedctc VARCHAR(100),
			noticeperiod VARCHAR(100),
			jlptlanguage VARCHAR(100),
			skills VARCHAR(100),
			jobdescription VARCHAR(500),
			company_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	// Clean slate for this test run
	testDB.Exec(`DELETE FROM candidates WHERE company_name = 'test_company'`)

	return testDB
}

func withAuthContext(req *http.Request, companyName string) *http.Request {
	ctx := context.WithValue(req.Context(), "companyName", companyName)
	return req.WithContext(ctx)
}

// TestAddCandidateAndGetCandidates is the test that would have caught the
// real production bug: AddCandidate/GetCandidates referenced columns
// (status, source, date_applied, resume_url, cover_letter) that did not
// exist in the actual schema. This test exercises the real SQL against a
// real database, not just Go's type system, so a column mismatch here fails
// loudly instead of silently reaching production.
func TestAddCandidateAndGetCandidates(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()
	db.DB = testDB // substitute the package-level connection used by the real handlers

	candidate := models.Candidate{
		Name:         "Test Candidate",
		Email:        "test.candidate@example.com",
		Phone:        "9999999999",
		Position:     "Software Engineer",
		Location:     "Bangalore",
		Experience:   "5 years",
		CurrentCTC:   "10 LPA",
		ExpectedCTC:  "15 LPA",
		NoticePeriod: "30 days",
		JLPTLanguage: "N/A",
		Skills:       "Go, PostgreSQL, React",
	}
	body, _ := json.Marshal(candidate)

	req := httptest.NewRequest("POST", "/api/candidates", bytes.NewReader(body))
	req = withAuthContext(req, "test_company")
	rec := httptest.NewRecorder()

	AddCandidate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("AddCandidate returned status %d, want %d. Body: %s",
			rec.Code, http.StatusCreated, rec.Body.String())
	}

	// Now verify GetCandidates can read it back without erroring
	getReq := httptest.NewRequest("GET", "/api/candidates", nil)
	getReq = withAuthContext(getReq, "test_company")
	getRec := httptest.NewRecorder()

	GetCandidates(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GetCandidates returned status %d, want %d. Body: %s",
			getRec.Code, http.StatusOK, getRec.Body.String())
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal GetCandidates response: %v", err)
	}
	if !resp.Success {
		t.Errorf("GetCandidates response Success = false, message: %s", resp.Message)
	}
}

// TestDeleteCandidateNonexistentReturnsNotFound checks basic not-found
// handling doesn't regress.
func TestDeleteCandidateNonexistentReturnsNotFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()
	db.DB = testDB // substitute the package-level connection used by the real handlers

	req := httptest.NewRequest("DELETE", "/api/candidates/999999", nil)
	req = withAuthContext(req, "test_company")
	req = mux.SetURLVars(req, map[string]string{"id": "999999"})
	rec := httptest.NewRecorder()

	DeleteCandidate(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d for deleting a nonexistent candidate", rec.Code, http.StatusNotFound)
	}
}
