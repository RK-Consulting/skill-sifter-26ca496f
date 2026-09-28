package offer

import (
"database/sql"
"errors"
"fmt"
"strings"
"time"
 )

var (
ErrCandidateRequirementNotFound = errors.New("candidate or requirement not found")
ErrSelectionNotFound = errors.New("selected recruitment decision not found")
ErrOfferExists = errors.New("offer already exists for candidate and requirement")
ErrNotFound = errors.New("offer not found")
ErrInvalidStatus = errors.New("invalid offer status")
ErrInvalidTransition = errors.New("invalid offer status transition")
 )

type Service struct { repo Repository; db *sql.DB }

func NewService(repo Repository, dbConn *sql.DB) *Service { return &Service{repo:repo,db:dbConn} }

func (s *Service) Create(tenantID string,input CreateInput)(*Offer,error){
if input.CandidateID==0 || input.RequirementID==0 { return nil,fmt.Errorf("candidateId and requirementId are required") }
var exists bool
if err:=s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM candidates c JOIN requirements r ON r.tenant_id=c.tenant_id WHERE c.id=$1 AND r.id=$2 AND c.tenant_id=$3 AND r.tenant_id=$3)`,input.CandidateID,input.RequirementID,tenantID).Scan(&exists); err!=nil{return nil,err}
if !exists{return nil,ErrCandidateRequirementNotFound}
var selectionID int
var decision string
if err:=s.db.QueryRow(`SELECT id, decision FROM recruitment_selections WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3`,tenantID,input.CandidateID,input.RequirementID).Scan(&selectionID,&decision); err!=nil { if errors.Is(err,sql.ErrNoRows){return nil,ErrSelectionNotFound}; return nil,err }
if decision!="selected" { return nil,ErrSelectionNotFound }
if _,err:=s.repo.GetByPair(tenantID,input.CandidateID,input.RequirementID); err==nil{return nil,ErrOfferExists}else if !errors.Is(err,ErrNotFound){return nil,err}
o:=&Offer{TenantID:tenantID,CandidateID:input.CandidateID,RequirementID:input.RequirementID,SelectionID:selectionID,OfferReference:strings.TrimSpace(input.OfferReference),Status:StatusOffered,OfferedDate:time.Now(),ExpectedJoiningDate:input.ExpectedJoiningDate,Compensation:strings.TrimSpace(input.Compensation),Terms:strings.TrimSpace(input.Terms),Notes:strings.TrimSpace(input.Notes)}
if err:=s.repo.Create(o);err!=nil{return nil,err}; return o,nil
}

func (s *Service) Get(tenantID string,candidateID,requirementID int)(*Offer,error){return s.repo.GetByPair(tenantID,candidateID,requirementID)}

func (s *Service) Update(tenantID string,candidateID,requirementID int,input UpdateInput)(*Offer,error){
o,err:=s.repo.GetByPair(tenantID,candidateID,requirementID); if err!=nil{return nil,err}
if !validStatus(input.Status){return nil,ErrInvalidStatus}
if !validTransition(o.Status,input.Status){return nil,ErrInvalidTransition}
o.Status=input.Status; o.ExpectedJoiningDate=input.ExpectedJoiningDate; o.Compensation=strings.TrimSpace(input.Compensation); o.Terms=strings.TrimSpace(input.Terms); o.Notes=strings.TrimSpace(input.Notes)
if input.Status==StatusAccepted||input.Status==StatusDeclined||input.Status==StatusExpired||input.Status==StatusWithdrawn {now:=time.Now();o.DecisionAt=&now} else {o.DecisionAt=nil}
if err:=s.repo.Update(o);err!=nil{return nil,err};return o,nil
}

func validStatus(v string)bool{switch v{case StatusDraft,StatusOffered,StatusAccepted,StatusDeclined,StatusExpired,StatusWithdrawn:return true};return false}
func validTransition(from,to string)bool{if from==to{return false};switch from{case StatusDraft:return to==StatusOffered||to==StatusWithdrawn;case StatusOffered:return to==StatusAccepted||to==StatusDeclined||to==StatusExpired||to==StatusWithdrawn;case StatusAccepted:return to==StatusWithdrawn;default:return false}}
