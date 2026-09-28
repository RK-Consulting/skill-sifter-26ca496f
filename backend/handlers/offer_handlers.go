package handlers

import (
"encoding/json"
"errors"
"net/http"
"strconv"
"time"

"github.com/RK-Consulting/skill-sifter/db"
"github.com/RK-Consulting/skill-sifter/domain/offer"
"github.com/RK-Consulting/skill-sifter/models"
"github.com/gorilla/mux"
"github.com/lib/pq"
)

type offerCreateRequest struct {
OfferReference string `json:"offerReference,omitempty"`
ExpectedJoiningDate *time.Time `json:"expectedJoiningDate,omitempty"`
Compensation string `json:"compensation,omitempty"`
Terms string `json:"terms,omitempty"`
Notes string `json:"notes,omitempty"`
}

type offerUpdateRequest struct {
Status string `json:"status"`
ExpectedJoiningDate *time.Time `json:"expectedJoiningDate,omitempty"`
Compensation string `json:"compensation,omitempty"`
Terms string `json:"terms,omitempty"`
Notes string `json:"notes,omitempty"`
}

func offerService() *offer.Service {
return offer.NewService(offer.NewPostgresRepository(db.DB), db.DB)
}

func parseOfferPair(r *http.Request) (int,int,error) {
candidateID,err:=strconv.Atoi(mux.Vars(r)["candidateId"]); if err!=nil{return 0,0,errors.New("invalid candidate ID")}
requirementID,err:=strconv.Atoi(mux.Vars(r)["requirementId"]); if err!=nil{return 0,0,errors.New("invalid requirement ID")}
return candidateID,requirementID,nil
}

func GetCandidateRequirementOffer(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseOfferPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
tenantID:=r.Context().Value("tenantID").(string)
o,err:=offerService().Get(tenantID,candidateID,requirementID);if errors.Is(err,offer.ErrNotFound){respondWithError(w,http.StatusNotFound,"Offer not found");return};if err!=nil{respondWithError(w,http.StatusInternalServerError,"Error fetching offer");return}
respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"Offer retrieved successfully",Data:o})
}

func CreateCandidateRequirementOffer(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseOfferPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
var req offerCreateRequest;if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{respondWithError(w,http.StatusBadRequest,"Invalid request payload");return};defer r.Body.Close()
tenantID:=r.Context().Value("tenantID").(string)
o,err:=offerService().Create(tenantID,offer.CreateInput{CandidateID:candidateID,RequirementID:requirementID,OfferReference:req.OfferReference,ExpectedJoiningDate:req.ExpectedJoiningDate,Compensation:req.Compensation,Terms:req.Terms,Notes:req.Notes})
switch {case errors.Is(err,offer.ErrCandidateRequirementNotFound):respondWithError(w,http.StatusNotFound,"Candidate or requirement not found");case errors.Is(err,offer.ErrSelectionNotFound):respondWithError(w,http.StatusUnprocessableEntity,"A selected decision is required before creating an offer");case errors.Is(err,offer.ErrOfferExists):respondWithError(w,http.StatusConflict,"Offer already exists for this candidate and requirement");case err!=nil:var pqErr *pq.Error;if errors.As(err,&pqErr)&&pqErr.Code=="23505"{respondWithError(w,http.StatusConflict,"Offer already exists for this candidate and requirement")}else{respondWithError(w,http.StatusInternalServerError,"Error creating offer")};default:respondWithJSON(w,http.StatusCreated,models.ApiResponse{Success:true,Message:"Offer created successfully",Data:o})}
}

func UpdateCandidateRequirementOffer(w http.ResponseWriter,r *http.Request){
candidateID,requirementID,err:=parseOfferPair(r);if err!=nil{respondWithError(w,http.StatusBadRequest,err.Error());return}
var req offerUpdateRequest;if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{respondWithError(w,http.StatusBadRequest,"Invalid request payload");return};defer r.Body.Close()
tenantID:=r.Context().Value("tenantID").(string)
o,err:=offerService().Update(tenantID,candidateID,requirementID,offer.UpdateInput{Status:req.Status,ExpectedJoiningDate:req.ExpectedJoiningDate,Compensation:req.Compensation,Terms:req.Terms,Notes:req.Notes})
switch {case errors.Is(err,offer.ErrNotFound):respondWithError(w,http.StatusNotFound,"Offer not found");case errors.Is(err,offer.ErrInvalidStatus):respondWithError(w,http.StatusBadRequest,"Invalid offer status");case errors.Is(err,offer.ErrInvalidTransition):respondWithError(w,http.StatusUnprocessableEntity,"Invalid offer status transition");case err!=nil:respondWithError(w,http.StatusInternalServerError,"Error updating offer");default:respondWithJSON(w,http.StatusOK,models.ApiResponse{Success:true,Message:"Offer updated successfully",Data:o})}
}
