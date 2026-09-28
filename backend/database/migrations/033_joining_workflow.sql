-- 033_joining_workflow.sql
-- One Joining record per Candidate × Requirement.
-- joining_date is the recruitment joining date; joined records whether it happened.
CREATE TABLE IF NOT EXISTS recruitment_joinings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    offer_id INTEGER NOT NULL REFERENCES recruitment_offers(id),
    joining_date DATE,
    joined BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT recruitment_joining_date_required
        CHECK (NOT joined OR joining_date IS NOT NULL)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_recruitment_joinings_tenant_pair
    ON recruitment_joinings(tenant_id, candidate_id, requirement_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_joinings_tenant_offer
    ON recruitment_joinings(tenant_id, offer_id);
CREATE INDEX IF NOT EXISTS idx_recruitment_joinings_candidate_requirement
    ON recruitment_joinings(tenant_id, candidate_id, requirement_id);
