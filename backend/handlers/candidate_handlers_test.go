package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// setupTestDB uses the authoritative tenant-plane baseline prepared by TestMain.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return handlerTenantDB
}

func withAuthContext(req *http.Request, companyName string) *http.Request {
	// Existing tests identify their tenant by a human-readable name; this
	// helper treats that same string as the tenant_id too (both companies
	// tables above use it as their id), so existing callers keep working
	// unchanged while still exercising the real tenant_id-scoped code path.
	ctx := context.WithValue(req.Context(), "companyName", companyName)
	ctx = context.WithValue(ctx, "tenantID", companyName)
	ctx = db.WithTenantDB(ctx, handlerTenantDB)
	return req.WithContext(ctx)
}

// TestAddCandidateAndGetCandidates is the test that would have caught the
// real production bug: AddCandidate/GetCandidates referenced columns
// (status, source, date_applied, resume_url, cover_letter) that did not
// exist in the actual schema. This test exercises the real SQL against a
// real database, not just Go's type system, so a column mismatch here fails
// loudly instead of silently reaching production.
func TestAddCandidateAndGetCandidates(t *testing.T) {
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
		Status:       "active",
		LanguageExpertise: []models.CandidateLanguageExpertise{
			{
				Language:             "Japanese",
				ProficiencyFramework: "JLPT",
				ProficiencyLevel:     "N2",
			},
		},
		TechnicalExpertise: []models.CandidateExpertise{
			{
				Skill:            "Go",
				Category:         "Programming",
				ProficiencyLevel: "Expert",
			},
			{
				Skill:            "PostgreSQL",
				Category:         "Database",
				ProficiencyLevel: "Advanced",
			},
		},
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

	req := httptest.NewRequest("DELETE", "/api/candidates/999999", nil)
	req = withAuthContext(req, "test_company")
	req = mux.SetURLVars(req, map[string]string{"id": "999999"})
	rec := httptest.NewRecorder()

	DeleteCandidate(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d for deleting a nonexistent candidate", rec.Code, http.StatusNotFound)
	}
}
