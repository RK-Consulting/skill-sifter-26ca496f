package handlers

import (
"encoding/json"
"errors"
"net/http"
"strconv"
"time"

"github.com/RK-Consulting/skill-sifter/db"
"github.com/RK-Consulting/skill-sifter/domain/joining"
"github.com/RK-Consulting/skill-sifter/models"
"github.com/gorilla/mux"
"github.com/lib/pq"
)

type joiningCreateRequest struct {
ExpectedJoiningDate *time.Time `json:"expectedJoiningDate,omitempty"`
Notes string `json:"notes,omitempty"`
}

type joiningUpdateRequest struct {
Status string `json:"status"`
ExpectedJoiningDate *time.Time `json:"expectedJoiningDate,omitempty"`
ActualJoiningDate *time.Time `json:"actualJoiningDate,omitempty"`
Notes string `json:"notes,omitempty"`
}

func joiningService() *joining.Service {
return joining.NewService(joining.NewPostgresRepository(db.DB), db.DB)
}

func parseJoiningPair(r *http.Request)(int,int,error){
candidateID,err:=strconv.Atoi(mux.Vars(r)["candidateId"]);if err!=nil{return 0,0,errors.New("invalid candidate ID")}
requirementID,err:=strconv.Atoi(mux.Vars(r)["requirementId"]);if err!=nil{return 0,0,errors.New("invalid requirement ID")}
return candidateID,requirementID,nil
}

func GetCandidateRequirementJoining(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseJoiningPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
tenantID:=r.Context().Value("tenantID").(string)
j,err:=joiningService().Get(tenantID,candidateID,requirementID);if errors.Is(err,joining.ErrNotFound){respondWithError(w,http.StatusNotFound,"Joining record not found");return};if err!=nil{respondWithError(w,http.StatusInternalServerError,"Error fetching joining record");return}
respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"Joining record retrieved successfully",Data:j})
}

func CreateCandidateRequirementJoining(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseJoiningPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
var req joiningCreateRequest;if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{respondWithError(w,http.StatusBadRequest,"Invalid request payload");return};defer r.Body.Close()
tenantID:=r.Context().Value("tenantID").(string)
j,err:=joiningService().Create(tenantID,joining.CreateInput{CandidateID:candidateID,RequirementID:requirementID,ExpectedJoiningDate:req.ExpectedJoiningDate,Notes:req.Notes})
switch {case errors.Is(err,joining.ErrCandidateRequirementNotFound):respondWithError(w,http.StatusNotFound,"Candidate or requirement not found");case errors.Is(err,joining.ErrOfferNotFound):respondWithError(w,http.StatusUnprocessableEntity,"An accepted offer is required before creating a joining record");case errors.Is(err,joining.ErrJoiningExists):respondWithError(w,http.StatusConflict,"Joining record already exists for this candidate and requirement");case err!=nil:var pqErr *pq.Error;if errors.As(err,&pqErr)&&pqErr.Code=="23505"{respondWithError(w,http.StatusConflict,"Joining record already exists for this candidate and requirement")}else{respondWithError(w,http.StatusInternalServerError,"Error creating joining record")};default:respondWithJSON(w,http.StatusCreated,models.ApiResponse{Success:true,Message:"Joining record created successfully",Data:j})}
}

func UpdateCandidateRequirementJoining(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseJoiningPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
var req joiningUpdateRequest;if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{respondWithError(w,http.StatusBadRequest,"Invalid request payload");return};defer r.Body.Close()
tenantID:=r.Context().Value("tenantID").(string)
j,err:=joiningService().Update(tenantID,candidateID,requirementID,joining.UpdateInput{Status:req.Status,ExpectedJoiningDate:req.ExpectedJoiningDate,ActualJoiningDate:req.ActualJoiningDate,Notes:req.Notes})
switch {case errors.Is(err,joining.ErrNotFound):respondWithError(w,http.StatusNotFound,"Joining record not found");case errors.Is(err,joining.ErrInvalidStatus):respondWithError(w,http.StatusBadRequest,"Invalid joining status");case errors.Is(err,joining.ErrInvalidTransition):respondWithError(w,http.StatusUnprocessableEntity,"Invalid joining status transition");case errors.Is(err,joining.ErrActualDateRequired):respondWithError(w,http.StatusUnprocessableEntity,"Actual joining date is required when marking joined");case errors.Is(err,joining.ErrActualDateNotAllowed):respondWithError(w,http.StatusUnprocessableEntity,"Actual joining date is only valid when marking joined");case err!=nil:respondWithError(w,http.StatusInternalServerError,"Error updating joining record");default:respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"Joining record updated successfully",Data:j})}
}
