package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
)

type resumeAIResult struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Phone  string   `json:"phone"`
	Skills []string `json:"skills"`
}

type resumeUploadResult struct {
	FileName  string            `json:"fileName"`
	Status    string            `json:"status"`
	ResumeID  int               `json:"resumeId,omitempty"`
	Candidate *models.Candidate `json:"candidate,omitempty"`
	Error     string            `json:"error,omitempty"`
}

var unsafeFileChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func resumeStoragePath() string {
	return db.GetEnv("RESUME_STORAGE_PATH", "./storage/resumes")
}

func safeResumeName(name string) string {
	name = filepath.Base(name)
	name = unsafeFileChars.ReplaceAllString(name, "_")

	if name == "" || name == "." {
		return "resume.bin"
	}

	return name
}

func ollamaBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("OLLAMA_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}

	if os.Getenv("DOCKER_ENV") == "true" || os.Getenv("DOCKER_ENV") == "1" {
		return "http://host.docker.internal:11434"
	}

	return "http://127.0.0.1:11434"
}

func ollamaModel() string {
	return db.GetEnv("OLLAMA_MODEL", "llama3.1:8b")
}

func extractResumeText(data []byte, filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".txt", ".md", ".csv":
		return string(data)

	case ".docx":
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return ""
		}

		for _, f := range zr.File {
			if f.Name != "word/document.xml" {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				return ""
			}

			b, err := io.ReadAll(rc)
			rc.Close()

			if err != nil {
				return ""
			}

			s := strings.ReplaceAll(string(b), "</w:p>", "\n")
			s = strings.ReplaceAll(s, "</w:tr>", "\n")

			return strings.TrimSpace(
				regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, ""),
			)
		}

	case ".pdf":
		s := strings.ReplaceAll(string(data), "\\n", " ")
		s = strings.ReplaceAll(s, "\\r", " ")

		matches := regexp.MustCompile(`\(([^()]{2,300})\)`).
			FindAllStringSubmatch(s, -1)

		parts := make([]string, 0, len(matches))

		for _, m := range matches {
			if len(m) > 1 {
				parts = append(parts, m[1])
			}
		}

		return strings.TrimSpace(strings.Join(parts, " "))
	}

	return ""
}

func callOllama(text string) (resumeAIResult, string) {
	prompt := `You are a resume extraction service.

Extract only:
- name
- email
- phone
- technical skills

Return ONLY valid JSON matching:

{
  "name": "",
  "email": "",
  "phone": "",
  "skills": []
}

Skills must be concise normalized technical skill names.

Examples:
["Go", "PostgreSQL", "React", "Docker"]

Never invent values.
Missing fields must be empty.
Do not include language names as technical skills.

Resume text:

` + text

	payload := map[string]interface{}{
		"model":  ollamaModel(),
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"options": map[string]interface{}{
			"temperature": 0,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return resumeAIResult{}, "Could not encode Ollama request: " + err.Error()
	}

	client := &http.Client{Timeout: 90 * time.Second}

	resp, err := client.Post(
		ollamaBaseURL()+"/api/generate",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return resumeAIResult{},
			"Ollama unavailable at " + ollamaBaseURL() + ": " + err.Error()
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		return resumeAIResult{},
			fmt.Sprintf("Ollama returned %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return resumeAIResult{}, "Invalid Ollama response: " + err.Error()
	}

	var parsed resumeAIResult

	if err := json.Unmarshal([]byte(out.Response), &parsed); err != nil {
		return resumeAIResult{}, "Ollama JSON parse failed: " + err.Error()
	}

	return parsed, ""
}

func upsertResumeCandidate(
	company string,
	tenantID string,
	ai resumeAIResult,
) (*models.Candidate, error) {
	if ai.Name == "" && ai.Email == "" && ai.Phone == "" {
		return nil, nil
	}

	var existing models.Candidate

	err := db.DB.QueryRow(`
		SELECT id, name, email, phone, position, location, experience,
		       currentctc, expectedctc, noticeperiod, jobdescription,
		       status, created_at, tenant_id, company_name
		FROM candidates
		WHERE tenant_id = $1
		  AND (
		       (email <> '' AND lower(email) = lower($2))
		    OR (phone <> '' AND phone = $3)
		  )
		ORDER BY id
		LIMIT 1`,
		tenantID,
		ai.Email,
		ai.Phone,
	).Scan(
		&existing.ID,
		&existing.Name,
		&existing.Email,
		&existing.Phone,
		&existing.Position,
		&existing.Location,
		&existing.Experience,
		&existing.CurrentCTC,
		&existing.ExpectedCTC,
		&existing.NoticePeriod,
		&existing.JobDescription,
		&existing.Status,
		&existing.CreatedAt,
		&existing.TenantID,
		&existing.CompanyName,
	)

	if err == nil {
		_, err = db.DB.Exec(`
			UPDATE candidates
			SET name = COALESCE(NULLIF($1, ''), name),
			    email = COALESCE(NULLIF($2, ''), email),
			    phone = COALESCE(NULLIF($3, ''), phone)
			WHERE id = $4 AND tenant_id = $5`,
			ai.Name,
			ai.Email,
			ai.Phone,
			existing.ID,
			tenantID,
		)
		if err != nil {
			return nil, err
		}

		existing.Name = firstNonEmpty(ai.Name, existing.Name)
		existing.Email = firstNonEmpty(ai.Email, existing.Email)
		existing.Phone = firstNonEmpty(ai.Phone, existing.Phone)

		return &existing, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	err = db.DB.QueryRow(`
		INSERT INTO candidates (
			name,
			email,
			phone,
			status,
			tenant_id,
			company_name
		)
		VALUES ($1, $2, $3, 'active', $4, $5)
		RETURNING id, created_at`,
		firstNonEmpty(ai.Name, "Unknown Candidate"),
		ai.Email,
		ai.Phone,
		tenantID,
		company,
	).Scan(
		&existing.ID,
		&existing.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	existing.Name = firstNonEmpty(ai.Name, "Unknown Candidate")
	existing.Email = ai.Email
	existing.Phone = ai.Phone
	existing.Status = "active"
	existing.TenantID = tenantID
	existing.CompanyName = company

	return &existing, nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}

	return b
}

func saveCandidateTechnicalExpertise(
	candidateID int,
	tenantID string,
	skills []string,
) error {
	for _, skill := range skills {
		skill = strings.TrimSpace(skill)

		if skill == "" {
			continue
		}

		_, err := db.DB.Exec(`
			INSERT INTO candidate_expertise (
				tenant_id,
				candidate_id,
				skill,
				category,
				proficiency_level
			)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (candidate_id, skill, category)
			DO UPDATE SET
				proficiency_level = EXCLUDED.proficiency_level,
				updated_at = NOW()`,
			tenantID,
			candidateID,
			skill,
			"resume_import",
			"unspecified",
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func UploadResumes(w http.ResponseWriter, r *http.Request) {
	company, _ := r.Context().Value("companyName").(string)
	tenantID, _ := r.Context().Value("tenantID").(string)
	userID, _ := r.Context().Value("userID").(int)

	if tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	if err := r.ParseMultipartForm(25 << 20); err != nil {
		respondWithError(
			w,
			http.StatusBadRequest,
			"Upload must be multipart/form-data and request is limited to 25MB",
		)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		files = r.MultipartForm.File["resume"]
	}

	if len(files) == 0 {
		respondWithError(
			w,
			http.StatusBadRequest,
			"No resume files supplied. Use the files field.",
		)
		return
	}

	root := filepath.Join(
		resumeStoragePath(),
		safeResumeName(company),
	)

	if err := os.MkdirAll(root, 0750); err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Could not create resume storage",
		)
		return
	}

	results := make([]resumeUploadResult, 0, len(files))

	for _, fh := range files {
		res := resumeUploadResult{
			FileName: safeResumeName(fh.Filename),
			Status:   "failed",
		}

		if fh.Size > 10<<20 {
			res.Error = "File exceeds 10MB limit"
			results = append(results, res)
			continue
		}

		src, err := fh.Open()
		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		data, err := io.ReadAll(
			io.LimitReader(src, 10<<20+1),
		)
		src.Close()

		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		if len(data) > 10<<20 {
			res.Error = "File exceeds 10MB limit"
			results = append(results, res)
			continue
		}

		hash := sha256.Sum256(data)
		hashHex := hex.EncodeToString(hash[:])

		var duplicateID int

		err = db.DB.QueryRow(`
			SELECT id
			FROM resumes
			WHERE company_name = $1
			  AND file_hash = $2`,
			company,
			hashHex,
		).Scan(&duplicateID)

		if err == nil {
			res.Status = "duplicate"
			res.ResumeID = duplicateID
			results = append(results, res)
			continue
		}

		if err != sql.ErrNoRows {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		path := filepath.Join(
			root,
			fmt.Sprintf(
				"%s_%s",
				hashHex[:12],
				safeResumeName(fh.Filename),
			),
		)

		if err := os.WriteFile(path, data, 0640); err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		text := extractResumeText(data, fh.Filename)

		var resumeID int

		err = db.DB.QueryRow(`
			INSERT INTO resumes (
				company_name,
				file_name,
				file_path,
				file_hash,
				mime_type,
				extracted_text,
				parsing_status,
				parser_model,
				uploaded_by
			)
			VALUES (
				$1, $2, $3, $4, $5, $6,
				'processing', $7, $8
			)
			RETURNING id`,
			company,
			fh.Filename,
			path,
			hashHex,
			fh.Header.Get("Content-Type"),
			text,
			ollamaModel(),
			userID,
		).Scan(&resumeID)

		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		res.ResumeID = resumeID

		if strings.TrimSpace(text) == "" {
			_, _ = db.DB.Exec(`
				UPDATE resumes
				SET parsing_status = 'failed',
				    parse_error = $1
				WHERE id = $2`,
				"No extractable text found. Scanned/image PDFs need OCR.",
				resumeID,
			)

			res.Error = "No extractable text found"
			results = append(results, res)
			continue
		}

		ai, parseErr := callOllama(text)

		if parseErr != "" {
			_, _ = db.DB.Exec(`
				UPDATE resumes
				SET parsing_status = 'failed',
				    parse_error = $1
				WHERE id = $2`,
				parseErr,
				resumeID,
			)

			res.Error = parseErr
			results = append(results, res)
			continue
		}

		candidate, err := upsertResumeCandidate(
			company,
			tenantID,
			ai,
		)

		if err != nil {
			_, _ = db.DB.Exec(`
				UPDATE resumes
				SET parsing_status = 'failed',
				    parse_error = $1
				WHERE id = $2`,
				err.Error(),
				resumeID,
			)

			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		if candidate != nil {
			if err := saveCandidateTechnicalExpertise(
				candidate.ID,
				tenantID,
				ai.Skills,
			); err != nil {
				_, _ = db.DB.Exec(`
					UPDATE resumes
					SET parsing_status = 'failed',
					    parse_error = $1
					WHERE id = $2`,
					err.Error(),
					resumeID,
				)

				res.Error = err.Error()
				results = append(results, res)
				continue
			}

			_, _ = db.DB.Exec(`
				UPDATE resumes
				SET candidate_id = $1,
				    parsing_status = 'completed',
				    parsed_at = NOW(),
				    parse_error = NULL
				WHERE id = $2`,
				candidate.ID,
				resumeID,
			)

			res.Candidate = candidate
		} else {
			_, _ = db.DB.Exec(`
				UPDATE resumes
				SET parsing_status = 'completed',
				    parsed_at = NOW(),
				    parse_error = NULL
				WHERE id = $1`,
				resumeID,
			)
		}

		res.Status = "completed"
		results = append(results, res)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Resume processing completed",
		Data:    results,
	})
}

func SearchResumes(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value("tenantID").(string)
	actor, _ := r.Context().Value("userID").(int)

	if tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	if q == "" {
		respondWithError(w, http.StatusBadRequest, "q is required")
		return
	}

	limit := 50

	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	started := time.Now()
	pattern := "%" + strings.ToLower(q) + "%"

	rows, err := db.DB.Query(`
		SELECT
			c.id,
			c.name,
			c.email,
			c.phone,
			COALESCE(
				STRING_AGG(
					DISTINCT ce.skill || ' [' ||
					ce.category || ', ' ||
					ce.proficiency_level || ']',
					', '
				),
				''
			) AS technical_expertise,
			COALESCE(r.file_name, ''),
			COALESCE(r.parsing_status, '')
		FROM candidates c
		LEFT JOIN candidate_expertise ce
			ON ce.candidate_id = c.id
		   AND ce.tenant_id = c.tenant_id
		LEFT JOIN resumes r
			ON r.candidate_id = c.id
		   AND r.company_name = c.company_name
		WHERE c.tenant_id = $1
		  AND (
			   lower(c.name) LIKE $2
			OR lower(c.email) LIKE $2
			OR lower(c.phone) LIKE $2
			OR EXISTS (
				SELECT 1
				FROM candidate_expertise search_ce
				WHERE search_ce.candidate_id = c.id
				  AND search_ce.tenant_id = c.tenant_id
				  AND lower(search_ce.skill) LIKE $2
			)
		  )
		GROUP BY
			c.id,
			c.name,
			c.email,
			c.phone,
			r.file_name,
			r.parsing_status
		ORDER BY c.name
		LIMIT $3`,
		tenantID,
		pattern,
		limit,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to search resumes")
		return
	}
	defer rows.Close()

	type result struct {
		ID                 int    `json:"id"`
		Name               string `json:"name"`
		Email              string `json:"email"`
		Phone              string `json:"phone"`
		TechnicalExpertise string `json:"technicalExpertise"`
		ResumeFile         string `json:"resumeFile"`
		Status             string `json:"resumeStatus"`
	}

	out := make([]result, 0)

	for rows.Next() {
		var x result

		if err := rows.Scan(
			&x.ID,
			&x.Name,
			&x.Email,
			&x.Phone,
			&x.TechnicalExpertise,
			&x.ResumeFile,
			&x.Status,
		); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"Error reading search result",
			)
			return
		}

		out = append(out, x)
	}

	if err := rows.Err(); err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Error reading search results",
		)
		return
	}

	duration := time.Since(started).Milliseconds()

	_, _ = db.DB.Exec(`
		INSERT INTO resume_search_logs (
			company_name,
			actor_user_id,
			query_text,
			resumes_searched,
			results_count,
			duration_ms
		)
		SELECT
			$1,
			$2,
			$3,
			COUNT(*),
			$4,
			$5
		FROM resumes
		WHERE company_name = $1`,
		r.Context().Value("companyName"),
		actor,
		q,
		len(out),
		duration,
	)

	_, _ = db.DB.Exec(`
		INSERT INTO activity_logs (
			company_name,
			actor_user_id,
			action,
			entity_type,
			description,
			metadata
		)
		VALUES (
			$1,
			$2,
			'RESUME_SEARCHED',
			'resume_search',
			$3,
			$4
		)`,
		r.Context().Value("companyName"),
		actor,
		"Resume search: "+q,
		`{"results":`+strconv.Itoa(len(out))+`}`,
	)

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Resume search completed",
		Data:    out,
	})
}

func ListResumes(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value("tenantID").(string)
	companyName, _ := r.Context().Value("companyName").(string)

	if tenantID == "" || companyName == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			r.id,
			r.file_name,
			r.parsing_status,
			COALESCE(r.parse_error, ''),
			r.uploaded_at,
			r.parsed_at,
			COALESCE(c.name, ''),
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			COALESCE(
				(
					SELECT STRING_AGG(
						ce.skill || ' [' ||
						ce.category || ', ' ||
						ce.proficiency_level || ']',
						', '
						ORDER BY ce.id
					)
					FROM candidate_expertise ce
					WHERE ce.candidate_id = c.id
					  AND ce.tenant_id = $1
				),
				''
			)
		FROM resumes r
		LEFT JOIN candidates c
			ON c.id = r.candidate_id
		   AND c.tenant_id = $1
		   AND c.company_name = $2
		WHERE r.company_name = $2
		ORDER BY r.uploaded_at DESC`,
		tenantID,
		companyName,
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to list resumes",
		)
		return
	}
	defer rows.Close()

	type item struct {
		ID                 int        `json:"id"`
		FileName           string     `json:"fileName"`
		Status             string     `json:"status"`
		Error              string     `json:"error,omitempty"`
		UploadedAt         time.Time  `json:"uploadedAt"`
		ParsedAt           *time.Time `json:"parsedAt,omitempty"`
		Name               string     `json:"name"`
		Email              string     `json:"email"`
		Phone              string     `json:"phone"`
		TechnicalExpertise string     `json:"technicalExpertise"`
	}

	out := make([]item, 0)

	for rows.Next() {
		var x item

		if err := rows.Scan(
			&x.ID,
			&x.FileName,
			&x.Status,
			&x.Error,
			&x.UploadedAt,
			&x.ParsedAt,
			&x.Name,
			&x.Email,
			&x.Phone,
			&x.TechnicalExpertise,
		); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"Error reading resumes",
			)
			return
		}

		out = append(out, x)
	}

	if err := rows.Err(); err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Error reading resumes",
		)
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Resumes retrieved",
		Data:    out,
	})
}

func GetResumeHealth(w http.ResponseWriter, r *http.Request) {
	url := ollamaBaseURL()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url + "/api/tags")
	if err != nil {
		respondWithJSON(w, http.StatusOK, models.ApiResponse{
			Success: true,
			Message: "Ollama is not reachable",
			Data: map[string]interface{}{
				"available": false,
				"url":       url,
				"model":     ollamaModel(),
				"error":     err.Error(),
			},
		})
		return
	}

	defer resp.Body.Close()

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Ollama health checked",
		Data: map[string]interface{}{
			"available": resp.StatusCode < 300,
			"url":       url,
			"model":     ollamaModel(),
		},
	})
}
