-- 033_joining_workflow.sql
-- Phase 7B: Joining for an accepted Candidate × Requirement offer.

CREATE TABLE IF NOT EXISTS recruitment_joinings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    offer_id INTEGER NOT NULL REFERENCES recruitment_offers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expected_joining_date TIMESTAMP,
    actual_joining_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT recruitment_joining_status_valid
        CHECK (status IN ('pending','joined','not_joined','withdrawn')),
    CONSTRAINT recruitment_joining_actual_date_valid
        CHECK ((status = 'joined') = (actual_joining_date IS NOT NULL))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_recruitment_joinings_tenant_pair
    ON recruitment_joinings(tenant_id, candidate_id, requirement_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_joinings_tenant_status
    ON recruitment_joinings(tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_recruitment_joinings_tenant_offer
    ON recruitment_joinings(tenant_id, offer_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_joinings_candidate_requirement
    ON recruitment_joinings(tenant_id, candidate_id, requirement_id);
