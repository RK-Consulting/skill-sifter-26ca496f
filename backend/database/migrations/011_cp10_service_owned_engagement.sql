-- 011_cp10_service_owned_engagement.sql
-- CP10 follow-up: candidate engagement ownership is implemented by the
-- application service so assignment creation and candidate claiming occur
-- in one explicit transaction. Migration 010 remains immutable.

ALTER TABLE candidates
    ADD COLUMN IF NOT EXISTS active_recruitment_engagements BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_candidates_active_recruitment_engagements
    ON candidates(tenant_id, active_recruitment_engagements);

-- Remove the legacy trigger implementation from migration 010. Keeping both
-- trigger ownership and service ownership would make the service claim the
-- candidate first and then cause the trigger to reject the assignment.
DROP TRIGGER IF EXISTS trg_claim_candidate_recruitment_engagement
    ON recruitment_assignments;

DROP TRIGGER IF EXISTS trg_release_candidate_recruitment_engagement
    ON recruitment_assignments;

DROP FUNCTION IF EXISTS claim_candidate_recruitment_engagement();
DROP FUNCTION IF EXISTS release_candidate_recruitment_engagement();
