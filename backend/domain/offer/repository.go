package offer

import (
"database/sql"
"errors"
 )

var ErrNotFound = errors.New("offer not found")

type Repository interface {
Create(*Offer) error
GetByPair(tenantID string, candidateID, requirementID int) (*Offer, error)
Update(*Offer) error
}

type PostgresRepository struct { db *sql.DB }

func NewPostgresRepository(dbConn *sql.DB) *PostgresRepository { return &PostgresRepository{db: dbConn} }

const offerSelect = `id, tenant_id, candidate_id, requirement_id, selection_id, offer_reference, status, offered_date, expected_joining_date, compensation, terms, notes, decision_at, created_at, last_modified`

func scanOffer(row *sql.Row) (*Offer, error) {
o := &Offer{}
var ref, compensation, terms, notes sql.NullString
if err := row.Scan(&o.ID,&o.TenantID,&o.CandidateID,&o.RequirementID,&o.SelectionID,&ref,&o.Status,&o.OfferedDate,&o.ExpectedJoiningDate,&compensation,&terms,&notes,&o.DecisionAt,&o.CreatedAt,&o.LastModified); err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }; return nil, err }
o.OfferReference=ref.String; o.Compensation=compensation.String; o.Terms=terms.String; o.Notes=notes.String
return o,nil
}

func (r *PostgresRepository) Create(o *Offer) error {
return r.db.QueryRow(`INSERT INTO recruitment_offers (tenant_id,candidate_id,requirement_id,selection_id,offer_reference,status,offered_date,expected_joining_date,compensation,terms,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id,created_at,last_modified`,o.TenantID,o.CandidateID,o.RequirementID,o.SelectionID,nullable(o.OfferReference),o.Status,o.OfferedDate,o.ExpectedJoiningDate,nullable(o.Compensation),nullable(o.Terms),nullable(o.Notes)).Scan(&o.ID,&o.CreatedAt,&o.LastModified)
}

func (r *PostgresRepository) GetByPair(tenantID string,candidateID,requirementID int)(*Offer,error){ return scanOffer(r.db.QueryRow(`SELECT `+offerSelect+` FROM recruitment_offers WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3`,tenantID,candidateID,requirementID)) }

func (r *PostgresRepository) Update(o *Offer) error {
return r.db.QueryRow(`UPDATE recruitment_offers SET status=$1, expected_joining_date=$2, compensation=$3, terms=$4, notes=$5, decision_at=$6, last_modified=CURRENT_TIMESTAMP WHERE id=$7 AND tenant_id=$8 RETURNING last_modified`,o.Status,o.ExpectedJoiningDate,nullable(o.Compensation),nullable(o.Terms),nullable(o.Notes),o.DecisionAt,o.ID,o.TenantID).Scan(&o.LastModified)
}

func nullable(v string) interface{} { if v=="" { return nil }; return v }
