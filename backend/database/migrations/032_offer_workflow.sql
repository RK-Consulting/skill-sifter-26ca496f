-- 032_offer_workflow.sql
-- One Offer record per selected Candidate × Requirement.
-- The Offer record means "offer made"; accepted is the only downstream decision.
CREATE TABLE IF NOT EXISTS recruitment_offers (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    selection_id INTEGER NOT NULL REFERENCES recruitment_selections(id),
    accepted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_recruitment_offers_tenant_pair
    ON recruitment_offers(tenant_id, candidate_id, requirement_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_offers_tenant_selection
    ON recruitment_offers(tenant_id, selection_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_offers_candidate_requirement
    ON recruitment_offers(tenant_id, candidate_id, requirement_id);
