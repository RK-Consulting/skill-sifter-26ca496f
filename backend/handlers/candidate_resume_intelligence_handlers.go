package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"

)

func GetCandidateResumeIntelligence(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" { respondWithError(w, http.StatusUnauthorized, "Tenant context missing"); return }
	candidateID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || candidateID <= 0 { respondWithError(w, http.StatusBadRequest, "Invalid candidate ID"); return }

	var resumeID int
	var fileName string
	var parsedAt time.Time
	var parserModel sql.NullString
	err = db.DB.QueryRow("SELECT id, file_name, parsed_at, parser_model FROM resumes WHERE id=(SELECT id FROM resumes WHERE tenant_id=$1 AND candidate_id=$2 AND parsing_status='completed' AND parsed_at IS NOT NULL ORDER BY parsed_at DESC, id DESC LIMIT 1)", tenantID, candidateID).Scan(&resumeID,&fileName,&parsedAt,&parserModel)
	if err == sql.ErrNoRows {
		respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"No processed Resume AI intelligence found",Data:map[string]interface{}{"sourceResume":nil,"professionalProfile":nil,"languages":[]interface{}{},"technicalExpertise":[]interface{}{},"employmentHistory":[]interface{}{},"education":[]interface{}{},"certifications":[]interface{}{},"projects":[]interface{}{}}})
		return
	}
	if err != nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve current resume"); return }
	sourceResume:=map[string]interface{}{"id":resumeID,"fileName":fileName,"parsedAt":parsedAt}
	if parserModel.Valid { sourceResume["parserModel"]=parserModel.String }

	profile, err := queryCandidateResumeRows("SELECT id, source_resume_id, current_title, professional_summary, location, total_experience, relevant_experience FROM candidate_professional_profiles WHERE tenant_id=$1 AND candidate_id=$2",tenantID,candidateID,nil,[]string{"id","sourceResumeId","currentTitle","professionalSummary","location","totalExperience","relevantExperience"})
	if err != nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve professional profile"); return }
	var professional interface{}
	if len(profile)>0 { professional=profile[0] }

	languages,err:=queryCandidateResumeRows("SELECT id, language, proficiency_framework, proficiency_level, source_resume_id FROM candidate_language_expertise WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3 ORDER BY id",tenantID,candidateID,resumeID,[]string{"id","language","proficiencyFramework","proficiencyLevel","sourceResumeId"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve languages"); return }
	technical,err:=queryCandidateResumeRows("SELECT id, skill, category, proficiency_level FROM candidate_expertise WHERE tenant_id=$1 AND candidate_id=$2 ORDER BY id",tenantID,candidateID,nil,[]string{"id","skill","category","proficiencyLevel"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve technical expertise"); return }
	employment,err:=queryCandidateResumeRows("SELECT id, employer, job_title, start_date, end_date, start_year, end_year, is_current, description, sort_order, source_resume_id FROM candidate_employment_history WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3 ORDER BY sort_order,id",tenantID,candidateID,resumeID,[]string{"id","employer","jobTitle","startDate","endDate","startYear","endYear","isCurrent","description","sortOrder","sourceResumeId"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve employment history"); return }
	education,err:=queryCandidateResumeRows("SELECT id, institution, degree, field_of_study, start_date, end_date, start_year, end_year, description, sort_order, source_resume_id FROM candidate_education WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3 ORDER BY sort_order,id",tenantID,candidateID,resumeID,[]string{"id","institution","degree","fieldOfStudy","startDate","endDate","startYear","endYear","description","sortOrder","sourceResumeId"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve education"); return }
	certifications,err:=queryCandidateResumeRows("SELECT id, name, issuer, issue_date, expiry_date, issue_year, expiry_year, credential_reference, source_resume_id FROM candidate_certifications WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3 ORDER BY id",tenantID,candidateID,resumeID,[]string{"id","name","issuer","issueDate","expiryDate","issueYear","expiryYear","credentialReference","sourceResumeId"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve certifications"); return }
	projects,err:=queryCandidateResumeRows("SELECT id, project_name, description, role, technologies, start_date, end_date, start_year, end_year, sort_order, source_resume_id FROM candidate_projects WHERE tenant_id=$1 AND candidate_id=$2 AND source_resume_id=$3 ORDER BY sort_order,id",tenantID,candidateID,resumeID,[]string{"id","projectName","description","role","technologies","startDate","endDate","startYear","endYear","sortOrder","sourceResumeId"})
	if err!=nil { respondWithError(w,http.StatusInternalServerError,"Failed to retrieve projects"); return }
	respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"Candidate Resume AI intelligence retrieved",Data:map[string]interface{}{"sourceResume":sourceResume,"professionalProfile":professional,"languages":languages,"technicalExpertise":technical,"employmentHistory":employment,"education":education,"certifications":certifications,"projects":projects}})
}

func queryCandidateResumeRows(query, tenantID string, candidateID, sourceResumeID interface{}, keys []string) ([]map[string]interface{}, error) {
	args:=[]interface{}{tenantID,candidateID}; if sourceResumeID!=nil { args=append(args,sourceResumeID) }
	rows,err:=db.DB.Query(query,args...); if err!=nil{return nil,err}; defer rows.Close()
	result:=make([]map[string]interface{},0)
	for rows.Next(){ values:=make([]interface{},len(keys)); dest:=make([]interface{},len(keys)); for i:=range values {dest[i]=&values[i]}; if err:=rows.Scan(dest...);err!=nil{return nil,err}; item:=make(map[string]interface{},len(keys)); for i,key:=range keys {item[key]=values[i]}; result=append(result,item) }
	return result,rows.Err()
}
