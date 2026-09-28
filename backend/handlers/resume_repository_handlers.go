package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

func resumeID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	return id, err == nil && id > 0
}

func GetResumeDetail(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value("tenantID").(string)
	if tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	id, ok := resumeID(r)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid resume ID")
		return
	}

	var idv int
	var name, mime, status, parseError, parserModel, filePath string
	var candidateID sql.NullInt64
	var uploadedAt time.Time
	var parsedAt sql.NullTime
	err := db.RequestDB(r).QueryRow("SELECT id,file_name,COALESCE(mime_type,''),parsing_status,COALESCE(parse_error,''),COALESCE(parser_model,''),candidate_id,uploaded_at,parsed_at,file_path FROM resumes WHERE id=$1 AND tenant_id=$2", id, tenantID).
		Scan(&idv, &name, &mime, &status, &parseError, &parserModel, &candidateID, &uploadedAt, &parsedAt, &filePath)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Resume not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve resume")
		return
	}

	data := map[string]interface{}{"id": idv, "fileName": name, "mimeType": mime, "status": status, "hasFile": storedResumeFile(filePath), "uploadedAt": uploadedAt}
	if parseError != "" {
		data["error"] = parseError
	}
	if parserModel != "" {
		data["parserModel"] = parserModel
	}
	if candidateID.Valid {
		data["candidateId"] = int(candidateID.Int64)
	}
	if parsedAt.Valid {
		data["parsedAt"] = parsedAt.Time
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Resume retrieved", Data: data})
}

func storedResumeFile(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DownloadResume(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value("tenantID").(string)
	if tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	id, ok := resumeID(r)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid resume ID")
		return
	}

	var path, name, mime string
	err := db.RequestDB(r).QueryRow("SELECT file_path,file_name,COALESCE(mime_type,'') FROM resumes WHERE id=$1 AND tenant_id=$2", id, tenantID).Scan(&path, &name, &mime)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Resume not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve resume")
		return
	}
	if !storedResumeFile(path) {
		respondWithError(w, http.StatusNotFound, "Resume file is unavailable")
		return
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", safeResumeName(name)))
	http.ServeFile(w, r, path)
}

func RetryResume(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value("tenantID").(string)
	if tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	id, ok := resumeID(r)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid resume ID")
		return
	}

	var path, name, text string
	var candidateID sql.NullInt64
	err := db.RequestDB(r).QueryRow("SELECT file_path,file_name,COALESCE(extracted_text,''),candidate_id FROM resumes WHERE id=$1 AND tenant_id=$2", id, tenantID).Scan(&path, &name, &text, &candidateID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Resume not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve resume")
		return
	}
	if !storedResumeFile(path) {
		respondWithError(w, http.StatusNotFound, "Resume file is unavailable")
		return
	}

	if strings.TrimSpace(text) == "" {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read stored resume")
			return
		}
		text = extractResumeText(data, name)
	}
	if strings.TrimSpace(text) == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "No extractable text found. Scanned/image PDFs need OCR.")
		return
	}

	_, _ = db.RequestDB(r).Exec("UPDATE resumes SET parsing_status='processing',parse_error=NULL,parser_model=$1 WHERE id=$2 AND tenant_id=$3", ollamaModel(), id, tenantID)
	ai, parseErr := callOllama(text)
	if parseErr != "" {
		_, _ = db.RequestDB(r).Exec("UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2 AND tenant_id=$3", parseErr, id, tenantID)
		respondWithError(w, http.StatusUnprocessableEntity, parseErr)
		return
	}

	var candidate *models.Candidate
	if candidateID.Valid {
		candidate, err = resumeExistingCandidate(r, int(candidateID.Int64), tenantID, ai)
	} else {
		company, _ := r.Context().Value("companyName").(string)
		candidate, err = upsertResumeCandidate(db.RequestDB(r), company, tenantID, ai)
	}
	if err != nil {
		_, _ = db.RequestDB(r).Exec("UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2 AND tenant_id=$3", err.Error(), id, tenantID)
		respondWithError(w, http.StatusInternalServerError, "Failed to persist extracted candidate intelligence")
		return
	}
	if candidate != nil {
		if err = persistResumeIntelligence(db.RequestDB(r), id, candidate.ID, tenantID, ai); err != nil {
			_, _ = db.RequestDB(r).Exec("UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2 AND tenant_id=$3", err.Error(), id, tenantID)
			respondWithError(w, http.StatusInternalServerError, "Failed to persist extracted candidate intelligence")
			return
		}
		_, _ = db.RequestDB(r).Exec("UPDATE resumes SET candidate_id=$1,parsing_status='completed',parsed_at=NOW(),parse_error=NULL WHERE id=$2 AND tenant_id=$3", candidate.ID, id, tenantID)
	} else {
		_, _ = db.RequestDB(r).Exec("UPDATE resumes SET parsing_status='completed',parsed_at=NOW(),parse_error=NULL WHERE id=$1 AND tenant_id=$2", id, tenantID)
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Resume reprocessed successfully", Data: map[string]interface{}{"resumeId": id, "status": "completed"}})
}

func resumeExistingCandidate(r *http.Request, candidateID int, tenantID string, ai resumeAIResult) (*models.Candidate, error) {
	var c models.Candidate
	err := db.RequestDB(r).QueryRow("SELECT id,name,email,phone,position,location,experience,currentctc,expectedctc,noticeperiod,jobdescription,status,created_at,tenant_id,company_name FROM candidates WHERE id=$1 AND tenant_id=$2", candidateID, tenantID).
		Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Position, &c.Location, &c.Experience, &c.CurrentCTC, &c.ExpectedCTC, &c.NoticePeriod, &c.JobDescription, &c.Status, &c.CreatedAt, &c.TenantID, &c.CompanyName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("candidate not found")
	}
	if err != nil {
		return nil, err
	}
	_, err = db.RequestDB(r).Exec("UPDATE candidates SET name=COALESCE(NULLIF($1,''),name),email=COALESCE(NULLIF($2,''),email),phone=COALESCE(NULLIF($3,''),phone) WHERE id=$4 AND tenant_id=$5", ai.Name, ai.Email, ai.Phone, candidateID, tenantID)
	if err != nil {
		return nil, err
	}
	c.Name = firstNonEmpty(ai.Name, c.Name)
	c.Email = firstNonEmpty(ai.Email, c.Email)
	c.Phone = firstNonEmpty(ai.Phone, c.Phone)
	return &c, nil
}
