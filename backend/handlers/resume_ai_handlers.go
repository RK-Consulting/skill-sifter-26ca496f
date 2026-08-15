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

type resumeAIResult struct { Name string `json:"name"`; Email string `json:"email"`; Phone string `json:"phone"`; Skills []string `json:"skills"` }
type resumeUploadResult struct { FileName string `json:"fileName"`; Status string `json:"status"`; ResumeID int `json:"resumeId,omitempty"`; Candidate *models.Candidate `json:"candidate,omitempty"`; Error string `json:"error,omitempty"` }
var unsafeFileChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func resumeStoragePath() string { return db.GetEnv("RESUME_STORAGE_PATH", "./storage/resumes") }
func safeResumeName(name string) string { name=filepath.Base(name); name=unsafeFileChars.ReplaceAllString(name,"_"); if name==""||name=="."{return "resume.bin"}; return name }

func extractResumeText(data []byte, filename string) string {
    switch strings.ToLower(filepath.Ext(filename)) {
    case ".txt", ".md", ".csv": return string(data)
    case ".docx":
        zr,err:=zip.NewReader(bytes.NewReader(data),int64(len(data)));if err!=nil{return ""}
        for _,f:=range zr.File{if f.Name!="word/document.xml"{continue};rc,err:=f.Open();if err!=nil{return ""};b,err:=io.ReadAll(rc);rc.Close();if err!=nil{return ""};s:=strings.ReplaceAll(string(b),"</w:p>","\n");s=strings.ReplaceAll(s,"</w:tr>","\n");return strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s,""))}
    case ".pdf":
        s:=strings.ReplaceAll(string(data),"\\n"," ");s=strings.ReplaceAll(s,"\\r"," ");matches:=regexp.MustCompile(`\(([^()]{2,300})\)`).FindAllStringSubmatch(s,-1);parts:=make([]string,0,len(matches));for _,m:=range matches{if len(m)>1{parts=append(parts,m[1])}};return strings.TrimSpace(strings.Join(parts," "))
    }
    return ""
}

func callOllama(text string) (resumeAIResult,string) {
    model:=db.GetEnv("OLLAMA_MODEL","llama3.1:8b")
    prompt:=`You are a resume extraction service. Extract only the requested fields. Return ONLY valid JSON matching {"name":"","email":"","phone":"","skills":[]}. Skills must be concise normalized skill names. Never invent values. Missing fields must be empty. Resume text:\n\n`+text
    payload:=map[string]interface{}{"model":model,"prompt":prompt,"stream":false,"format":"json","options":map[string]interface{}{"temperature":0}}
    body,_:=json.Marshal(payload);client:=&http.Client{Timeout:90*time.Second};resp,err:=client.Post(db.GetEnv("OLLAMA_URL","http://127.0.0.1:11434")+"/api/generate","application/json",bytes.NewReader(body));if err!=nil{return resumeAIResult{},"Ollama unavailable: "+err.Error()};defer resp.Body.Close();if resp.StatusCode>=300{b,_:=io.ReadAll(io.LimitReader(resp.Body,4096));return resumeAIResult{},fmt.Sprintf("Ollama returned %d: %s",resp.StatusCode,string(b))};var out struct{Response string `json:"response"`};if err:=json.NewDecoder(resp.Body).Decode(&out);err!=nil{return resumeAIResult{},"Invalid Ollama response: "+err.Error()};var parsed resumeAIResult;if err:=json.Unmarshal([]byte(out.Response),&parsed);err!=nil{return resumeAIResult{},"Ollama JSON parse failed: "+err.Error()};return parsed,""
}

func upsertResumeCandidate(company string,ai resumeAIResult)(*models.Candidate,error){
    if ai.Name==""&&ai.Email==""&&ai.Phone==""{return nil,nil};var existing models.Candidate
    err:=db.DB.QueryRow(`SELECT id,name,email,phone,position,location,experience,currentctc,expectedctc,noticeperiod,jlptlanguage,skills,jobdescription,created_at,company_name FROM candidates WHERE company_name=$1 AND ((email<>'' AND lower(email)=lower($2)) OR (phone<>'' AND phone=$3)) ORDER BY id LIMIT 1`,company,ai.Email,ai.Phone).Scan(&existing.ID,&existing.Name,&existing.Email,&existing.Phone,&existing.Position,&existing.Location,&existing.Experience,&existing.CurrentCTC,&existing.ExpectedCTC,&existing.NoticePeriod,&existing.JLPTLanguage,&existing.Skills,&existing.JobDescription,&existing.CreatedAt,&existing.CompanyName)
    skills:=strings.Join(ai.Skills,", ")
    if err==nil{if skills==""{skills=existing.Skills};_,err=db.DB.Exec(`UPDATE candidates SET name=COALESCE(NULLIF($1,''),name),email=COALESCE(NULLIF($2,''),email),phone=COALESCE(NULLIF($3,''),phone),skills=$4 WHERE id=$5 AND company_name=$6`,ai.Name,ai.Email,ai.Phone,skills,existing.ID,company);if err!=nil{return nil,err};existing.Name=firstNonEmpty(ai.Name,existing.Name);existing.Email=firstNonEmpty(ai.Email,existing.Email);existing.Phone=firstNonEmpty(ai.Phone,existing.Phone);existing.Skills=skills;return &existing,nil}
    if err!=sql.ErrNoRows{return nil,err};err=db.DB.QueryRow(`INSERT INTO candidates(name,email,phone,skills,company_name) VALUES($1,$2,$3,$4,$5) RETURNING id,created_at`,firstNonEmpty(ai.Name,"Unknown Candidate"),ai.Email,ai.Phone,skills,company).Scan(&existing.ID,&existing.CreatedAt);if err!=nil{return nil,err};existing.Name=firstNonEmpty(ai.Name,"Unknown Candidate");existing.Email=ai.Email;existing.Phone=ai.Phone;existing.Skills=skills;existing.CompanyName=company;return &existing,nil
}
func firstNonEmpty(a,b string)string{if strings.TrimSpace(a)!=""{return a};return b}
func saveCandidateSkills(candidateID int,skills []string)error{for _,skill:=range skills{skill=strings.TrimSpace(skill);if skill==""{continue};normalized:=strings.ToLower(skill);var id int;if err:=db.DB.QueryRow(`INSERT INTO skills(name,normalized_name) VALUES($1,$2) ON CONFLICT(normalized_name) DO UPDATE SET name=EXCLUDED.name RETURNING id`,skill,normalized).Scan(&id);err!=nil{return err};if _,err:=db.DB.Exec(`INSERT INTO candidate_skills(candidate_id,skill_id,confidence,source) VALUES($1,$2,1.0,'resume_ai') ON CONFLICT(candidate_id,skill_id) DO UPDATE SET confidence=EXCLUDED.confidence,source=EXCLUDED.source`,candidateID,id);err!=nil{return err}};return nil}

func UploadResumes(w http.ResponseWriter,r *http.Request){
    company:=r.Context().Value("companyName").(string);userID:=r.Context().Value("userID").(int);if err:=r.ParseMultipartForm(25<<20);err!=nil{respondWithError(w,400,"Upload must be multipart/form-data and request is limited to 25MB");return};files:=r.MultipartForm.File["files"];if len(files)==0{files=r.MultipartForm.File["resume"]};if len(files)==0{respondWithError(w,400,"No resume files supplied. Use the files field.");return};root:=filepath.Join(resumeStoragePath(),safeResumeName(company));if err:=os.MkdirAll(root,0750);err!=nil{respondWithError(w,500,"Could not create resume storage");return};results:=make([]resumeUploadResult,0,len(files))
    for _,fh:=range files{res:=resumeUploadResult{FileName:safeResumeName(fh.Filename),Status:"failed"};if fh.Size>10<<20{res.Error="File exceeds 10MB limit";results=append(results,res);continue};src,err:=fh.Open();if err!=nil{res.Error=err.Error();results=append(results,res);continue};data,err:=io.ReadAll(io.LimitReader(src,10<<20+1));src.Close();if err!=nil{res.Error=err.Error();results=append(results,res);continue};if len(data)>10<<20{res.Error="File exceeds 10MB limit";results=append(results,res);continue};hash:=sha256.Sum256(data);hashHex:=hex.EncodeToString(hash[:]);var duplicateID int;if err=db.DB.QueryRow(`SELECT id FROM resumes WHERE company_name=$1 AND file_hash=$2`,company,hashHex).Scan(&duplicateID);err==nil{res.Status="duplicate";res.ResumeID=duplicateID;results=append(results,res);continue}else if err!=sql.ErrNoRows{res.Error=err.Error();results=append(results,res);continue};path:=filepath.Join(root,fmt.Sprintf("%s_%s",hashHex[:12],safeResumeName(fh.Filename)));if err:=os.WriteFile(path,data,0640);err!=nil{res.Error=err.Error();results=append(results,res);continue};text:=extractResumeText(data,fh.Filename);var resumeID int;err=db.DB.QueryRow(`INSERT INTO resumes(company_name,file_name,file_path,file_hash,mime_type,extracted_text,parsing_status,parser_model,uploaded_by) VALUES($1,$2,$3,$4,$5,$6,'processing',$7,$8) RETURNING id`,company,fh.Filename,path,hashHex,fh.Header.Get("Content-Type"),text,db.GetEnv("OLLAMA_MODEL","llama3.1:8b"),userID).Scan(&resumeID);if err!=nil{res.Error=err.Error();results=append(results,res);continue};res.ResumeID=resumeID;if strings.TrimSpace(text)==""{_,_=db.DB.Exec(`UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2`,"No extractable text found. Scanned/image PDFs need OCR.",resumeID);res.Error="No extractable text found";results=append(results,res);continue};ai,parseErr:=callOllama(text);if parseErr!=""{_,_=db.DB.Exec(`UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2`,parseErr,resumeID);res.Error=parseErr;results=append(results,res);continue};candidate,err:=upsertResumeCandidate(company,ai);if err!=nil{_,_=db.DB.Exec(`UPDATE resumes SET parsing_status='failed',parse_error=$1 WHERE id=$2`,err.Error(),resumeID);res.Error=err.Error();results=append(results,res);continue};if candidate!=nil{_=saveCandidateSkills(candidate.ID,ai.Skills);_,_=db.DB.Exec(`UPDATE resumes SET candidate_id=$1,parsing_status='completed',parsed_at=NOW(),parse_error=NULL WHERE id=$2`,candidate.ID,resumeID);res.Candidate=candidate}else{_,_=db.DB.Exec(`UPDATE resumes SET parsing_status='completed',parsed_at=NOW(),parse_error=NULL WHERE id=$1`,resumeID)};res.Status="completed";results=append(results,res)}
    respondWithJSON(w,200,models.ApiResponse{Success:true,Message:"Resume processing completed",Data:results})
}

func SearchResumes(w http.ResponseWriter,r *http.Request){company:=r.Context().Value("companyName").(string);actor:=r.Context().Value("userID").(int);q:=strings.TrimSpace(r.URL.Query().Get("q"));if q==""{respondWithError(w,400,"q is required");return};limit:=50;if raw:=r.URL.Query().Get("limit");raw!=""{if n,e:=strconv.Atoi(raw);e==nil&&n>0&&n<=200{limit=n}};started:=time.Now();pattern:="%"+strings.ToLower(q)+"%";rows,err:=db.DB.Query(`SELECT c.id,c.name,c.email,c.phone,c.skills,COALESCE(r.file_name,''),COALESCE(r.parsing_status,'') FROM candidates c LEFT JOIN resumes r ON r.candidate_id=c.id AND r.company_name=c.company_name WHERE c.company_name=$1 AND (lower(c.name) LIKE $2 OR lower(c.email) LIKE $2 OR lower(c.phone) LIKE $2 OR lower(c.skills) LIKE $2 OR EXISTS (SELECT 1 FROM candidate_skills cs JOIN skills s ON s.id=cs.skill_id WHERE cs.candidate_id=c.id AND s.normalized_name LIKE $2)) ORDER BY c.name LIMIT $3`,company,pattern,limit);if err!=nil{respondWithError(w,500,"Failed to search resumes");return};defer rows.Close();type result struct{ID int `json:"id"`;Name string `json:"name"`;Email string `json:"email"`;Phone string `json:"phone"`;Skills string `json:"skills"`;ResumeFile string `json:"resumeFile"`;Status string `json:"resumeStatus"`};out:=[]result{};for rows.Next(){var x result;if err:=rows.Scan(&x.ID,&x.Name,&x.Email,&x.Phone,&x.Skills,&x.ResumeFile,&x.Status);err!=nil{respondWithError(w,500,"Error reading search result");return};out=append(out,x)};duration:=time.Since(started).Milliseconds();_,_=db.DB.Exec(`INSERT INTO resume_search_logs(company_name,actor_user_id,query_text,resumes_searched,results_count,duration_ms) SELECT $1,$2,$3,COUNT(*),$4,$5 FROM resumes WHERE company_name=$1`,company,actor,q,len(out),duration);_,_=db.DB.Exec(`INSERT INTO activity_logs(company_name,actor_user_id,action,entity_type,description,metadata) VALUES($1,$2,'RESUME_SEARCHED','resume_search',$3,$4)`,company,actor,"Resume search: "+q,`{"results":`+strconv.Itoa(len(out))+`}`);respondWithJSON(w,200,models.ApiResponse{Success:true,Message:"Resume search completed",Data:out})}

func ListResumes(w http.ResponseWriter,r *http.Request){company:=r.Context().Value("companyName").(string);rows,err:=db.DB.Query(`SELECT r.id,r.file_name,r.parsing_status,COALESCE(r.parse_error,''),r.uploaded_at,r.parsed_at,COALESCE(c.name,''),COALESCE(c.email,''),COALESCE(c.phone,''),COALESCE(c.skills,'') FROM resumes r LEFT JOIN candidates c ON c.id=r.candidate_id WHERE r.company_name=$1 ORDER BY r.uploaded_at DESC`,company);if err!=nil{respondWithError(w,500,"Failed to list resumes");return};defer rows.Close();type item struct{ID int `json:"id"`;FileName string `json:"fileName"`;Status string `json:"status"`;Error string `json:"error,omitempty"`;UploadedAt time.Time `json:"uploadedAt"`;ParsedAt *time.Time `json:"parsedAt,omitempty"`;Name string `json:"name"`;Email string `json:"email"`;Phone string `json:"phone"`;Skills string `json:"skills"`};out:=[]item{};for rows.Next(){var x item;if err:=rows.Scan(&x.ID,&x.FileName,&x.Status,&x.Error,&x.UploadedAt,&x.ParsedAt,&x.Name,&x.Email,&x.Phone,&x.Skills);err!=nil{respondWithError(w,500,"Error reading resumes");return};out=append(out,x)};respondWithJSON(w,200,models.ApiResponse{Success:true,Message:"Resumes retrieved",Data:out})}

func GetResumeHealth(w http.ResponseWriter,r *http.Request){url:=db.GetEnv("OLLAMA_URL","http://127.0.0.1:11434");resp,err:=http.Get(url+"/api/tags");if err!=nil{respondWithJSON(w,200,models.ApiResponse{Success:true,Message:"Ollama is not reachable",Data:map[string]interface{}{"available":false,"url":url,"error":err.Error()}});return};defer resp.Body.Close();respondWithJSON(w,200,models.ApiResponse{Success:true,Message:"Ollama health checked",Data:map[string]interface{}{"available":resp.StatusCode<300,"url":url,"model":db.GetEnv("OLLAMA_MODEL","llama3.1:8b")}})}
