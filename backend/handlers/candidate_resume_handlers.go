package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

// candidateResumeResponse is the wire shape for a candidate's resume,
// deliberately small — the Candidates page only needs enough to show
// "a resume is on file" and let the user open/replace it, not the full
// AI-pipeline fields (parser_model, parse_error, extracted_text) that
// ResumeRecord in the bulk Resume AI upload exposes.
type candidateResumeResponse struct {
	ID         int    `json:"id"`
	FileName   string `json:"fileName"`
	UploadedAt string `json:"uploadedAt"`
}

// UploadCandidateResume uploads one resume file and associates it with a
// specific, already-known candidate (the {id} in the URL). This is
// deliberately separate from UploadResumes (resume_ai_handlers.go), which
// bulk-uploads a folder of resumes and matches/creates candidates by
// parsing extracted email/name — a fuzzy match that would risk attaching
// a resume to the wrong candidate here, where the candidate is already
// known with certainty from the URL. No AI parsing runs on this path;
// parsing_status is set directly to 'completed'.
func UploadCandidateResume(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	company, _ := r.Context().Value("companyName").(string)
	userID, _ := r.Context().Value("userID").(int)

	// Confirm the candidate actually belongs to this tenant before
	// accepting a file for it — the same tenant-membership check every
	// other candidate-scoped write in this codebase performs.
	var exists bool
	if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM candidates WHERE id = $1 AND tenant_id = $2)`, candidateID, tenantID).Scan(&exists); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error verifying candidate")
		return
	}
	if !exists {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, "Upload must be multipart/form-data and is limited to 10MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No resume file supplied. Use the file field.")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 10<<20+1))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading uploaded file")
		return
	}
	if len(data) > 10<<20 {
		respondWithError(w, http.StatusBadRequest, "File exceeds 10MB limit")
		return
	}

	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	root := filepath.Join(resumeStoragePath(), safeResumeName(company))
	if err := os.MkdirAll(root, 0750); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create resume storage")
		return
	}

	safeName := safeResumeName(header.Filename)
	path := filepath.Join(root, fmt.Sprintf("candidate%d_%s_%s", candidateID, hashHex[:12], safeName))

	if err := os.WriteFile(path, data, 0640); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving resume file")
		return
	}

	var resumeID int
	err = db.DB.QueryRow(`
		INSERT INTO resumes (
			company_name, candidate_id, file_name, file_path, file_hash,
			mime_type, parsing_status, uploaded_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'completed', $7)
		RETURNING id`,
		company, candidateID, safeName, path, hashHex, header.Header.Get("Content-Type"), userID,
	).Scan(&resumeID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recording resume")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Resume uploaded successfully",
		Data:    candidateResumeResponse{ID: resumeID, FileName: safeName},
	})
}

// GetCandidateResume returns the most recently uploaded resume for a
// candidate, if any, so the Candidates page can show "a resume is on
// file" and let the user re-download it. Returns 404 if none exists yet
// — this is a normal state, not an error, for a candidate who hasn't had
// a resume uploaded.
func GetCandidateResume(w http.ResponseWriter, r *http.Request) {
	candidateID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	var resp candidateResumeResponse
	var uploadedAt time.Time

	err = db.DB.QueryRow(`
		SELECT r.id, r.file_name, r.uploaded_at
		FROM resumes r
		JOIN candidates c ON c.id = r.candidate_id
		WHERE r.candidate_id = $1 AND c.tenant_id = $2
		ORDER BY r.uploaded_at DESC
		LIMIT 1`,
		candidateID, tenantID,
	).Scan(&resp.ID, &resp.FileName, &uploadedAt)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "No resume on file for this candidate")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching resume")
		return
	}

	resp.UploadedAt = uploadedAt.Format(time.RFC3339)

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Resume retrieved successfully",
		Data:    resp,
	})
}
