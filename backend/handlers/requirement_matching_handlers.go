package handlers

import (
	"net/http"
	"strconv"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/matching"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
)

type requirementMatchCandidate struct {
	ID         int
	Location   string
	Experience string
	Notice     string
	Skills     []string
	Languages  []matching.LanguageEvidence
	Certs      []string
}

func loadRequirementForMatching(tenantID string, id int) (models.Requirement, error) {
	var req models.Requirement
	err := db.DB.QueryRow(`
		SELECT id, client_id, COALESCE(job_id, ''), COALESCE(job_type, ''), title, COALESCE(department, ''),
			COALESCE(experience_required, ''), COALESCE(budget, ''), COALESCE(language_requirement, ''),
			COALESCE(certifications_required, ''), COALESCE(notice_period, ''),
			COALESCE(work_arrangement, ''), COALESCE(mandatory_requirements, ''),
			COALESCE(description, ''), status, COALESCE(location, ''), headcount,
			COALESCE(opened_date, created_at), created_at, last_modified, tenant_id
		FROM requirements WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	).Scan(
		&req.ID, &req.ClientID, &req.JobID, &req.JobType, &req.Title, &req.Department,
		&req.ExperienceRequired, &req.Budget, &req.LanguageRequirements,
		&req.CertificationsRequired, &req.NoticePeriod, &req.WorkArrangement,
		&req.MandatoryRequirements, &req.Description, &req.Status, &req.Location,
		&req.Headcount, &req.OpenedDate, &req.CreatedAt, &req.LastModified, &req.TenantID,
	)
	return req, err
}

func loadCandidateForMatching(tenantID string, candidateID int) (requirementMatchCandidate, error) {
	candidate := requirementMatchCandidate{ID: candidateID, Skills: []string{}, Languages: []matching.LanguageEvidence{}, Certs: []string{}}

	var profileExperience string
	err := db.DB.QueryRow(`
		SELECT COALESCE(c.location, ''), COALESCE(c.experience, ''), COALESCE(c.noticeperiod, ''),
		       COALESCE(p.total_experience, '')
		FROM candidates c
		LEFT JOIN candidate_professional_profiles p
		  ON p.tenant_id = c.tenant_id AND p.candidate_id = c.id
		WHERE c.id = $1 AND c.tenant_id = $2`,
		candidateID, tenantID,
	).Scan(&candidate.Location, &candidate.Experience, &candidate.Notice, &profileExperience)
	if err != nil {
		return candidate, err
	}
	if profileExperience != "" {
		candidate.Experience = profileExperience
	}

	rows, err := db.DB.Query(`SELECT skill FROM candidate_expertise WHERE tenant_id = $1 AND candidate_id = $2 ORDER BY id`, tenantID, candidateID)
	if err != nil {
		return candidate, err
	}
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			rows.Close()
			return candidate, err
		}
		candidate.Skills = append(candidate.Skills, skill)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return candidate, err
	}
	rows.Close()

	var sourceResumeID int
	err = db.DB.QueryRow(`
		SELECT id FROM resumes
		WHERE tenant_id = $1 AND candidate_id = $2
		  AND parsing_status = 'completed' AND parsed_at IS NOT NULL
		ORDER BY parsed_at DESC, id DESC LIMIT 1`,
		tenantID, candidateID,
	).Scan(&sourceResumeID)
	if err == nil {
		rows, err = db.DB.Query(`
			SELECT language, proficiency_level
			FROM candidate_language_expertise
			WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3
			ORDER BY id`, tenantID, candidateID, sourceResumeID)
		if err != nil {
			return candidate, err
		}
		for rows.Next() {
			var language, proficiency string
			if err := rows.Scan(&language, &proficiency); err != nil {
				rows.Close()
				return candidate, err
			}
			candidate.Languages = append(candidate.Languages, matching.LanguageEvidence{Language: language, Proficiency: proficiency})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return candidate, err
		}
		rows.Close()

		rows, err = db.DB.Query(`
			SELECT name FROM candidate_certifications
			WHERE tenant_id = $1 AND candidate_id = $2 AND source_resume_id = $3
			ORDER BY id`, tenantID, candidateID, sourceResumeID)
		if err != nil {
			return candidate, err
		}
		for rows.Next() {
			var cert string
			if err := rows.Scan(&cert); err != nil {
				rows.Close()
				return candidate, err
			}
			candidate.Certs = append(candidate.Certs, cert)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return candidate, err
		}
		rows.Close()
	}
	return candidate, nil
}

func buildRequirementMatch(req models.Requirement, candidate requirementMatchCandidate) matching.MatchResult {
	return matching.Evaluate(
		matching.RequirementEvidence{
			RequiredSkills:         "",
			MandatoryRequirements: req.MandatoryRequirements,
			ExperienceRequired:     req.ExperienceRequired,
			LanguageRequirements:   req.LanguageRequirements,
			CertificationsRequired: req.CertificationsRequired,
			Location:               req.Location,
			NoticePeriod:           req.NoticePeriod,
			WorkArrangement:        req.WorkArrangement,
		},
		matching.CandidateEvidence{
			TechnicalSkills: candidate.Skills,
			Languages:       candidate.Languages,
			Certifications:  candidate.Certs,
			Experience:      candidate.Experience,
			Location:        candidate.Location,
			NoticePeriod:    candidate.Notice,
			WorkArrangement: "",
		},
	)
}

func GetRequirementCandidateMatch(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || requirementID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}
	candidateID, err := strconv.Atoi(mux.Vars(r)["candidateId"])
	if err != nil || candidateID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}
	req, err := loadRequirementForMatching(tenantID, requirementID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}
	candidate, err := loadCandidateForMatching(tenantID, candidateID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Candidate not found")
		return
	}
	result := buildRequirementMatch(req, candidate)
	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirement match evidence retrieved",
		Data:    map[string]interface{}{
			"requirementId": req.ID,
			"jobId":        req.JobID,
			"candidateId":  candidate.ID,
			"match": result,
		},
	})
}

func GetRequirementMatches(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	requirementID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || requirementID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid requirement ID")
		return
	}
	req, err := loadRequirementForMatching(tenantID, requirementID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Requirement not found")
		return
	}

	rows, err := db.DB.Query(`SELECT id FROM candidates WHERE tenant_id = $1 ORDER BY id`, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve candidates")
		return
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var candidateID int
		if err := rows.Scan(&candidateID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan candidate")
			return
		}
		candidate, err := loadCandidateForMatching(tenantID, candidateID)
		if err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"candidateId": candidateID,
			"match": buildRequirementMatch(req, candidate),
		})
	}
	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve candidates")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Requirement match evidence retrieved",
		Data: map[string]interface{}{
			"requirementId": req.ID,
			"jobId":        req.JobID,
			"matches":      results,
		},
	})
}
