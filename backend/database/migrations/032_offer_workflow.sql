-- 032_offer_workflow.sql
-- Phase 7A: Offer for selected Candidate × Requirement.
CREATE TABLE IF NOT EXISTS recruitment_offers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    selection_id INTEGER NOT NULL REFERENCES recruitment_selections(id),
    offer_reference VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'offered',
    offered_date TIMESTAMP NOT NULL DEFAULT NOW(),
    expected_joining_date TIMESTAMP,
    compensation TEXT,
    terms TEXT,
    notes TEXT,
    decision_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT recruitment_offer_status_valid CHECK (status IN ('draft','offered','accepted','declined','expired','withdrawn'))
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_recruitment_offers_tenant_pair
    ON recruitment_offers(tenant_id, candidate_id, requirement_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_offers_tenant_status
    ON recruitment_offers(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_recruitment_offers_tenant_selection
    ON recruitment_offers(tenant_id, selection_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_offers_candidate_requirement
    ON recruitment_offers(tenant_id, candidate_id, requirement_id);
