-- 010_candidate_recruitment_engagement.sql
-- CP10: a candidate may have one active recruitment engagement within the
-- tenant database at a time. This is a simple availability gate, not a
-- cross-tenant coordination mechanism.

ALTER TABLE candidates
    ADD COLUMN IF NOT EXISTS active_recruitment_engagements BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_candidates_active_recruitment_engagements
    ON candidates(tenant_id, active_recruitment_engagements);

-- Atomically acquire the candidate when a recruitment assignment is first
-- created. The UPDATE lock on the candidate row makes two simultaneous
-- attempts deterministic: exactly one INSERT can acquire the candidate.
CREATE OR REPLACE FUNCTION claim_candidate_recruitment_engagement()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE candidates
       SET active_recruitment_engagements = TRUE
     WHERE id = NEW.candidate_id
       AND tenant_id = NEW.tenant_id
       AND active_recruitment_engagements = FALSE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'candidate % already has an active recruitment engagement', NEW.candidate_id
            USING ERRCODE = '23514',
                  CONSTRAINT = 'candidate_active_recruitment_engagement';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_claim_candidate_recruitment_engagement
    ON recruitment_assignments;

CREATE TRIGGER trg_claim_candidate_recruitment_engagement
BEFORE INSERT ON recruitment_assignments
FOR EACH ROW
EXECUTE FUNCTION claim_candidate_recruitment_engagement();

-- Terminal assignment outcomes release the candidate for a future
-- recruitment engagement. Joined, rejected and withdrawn are terminal
-- outcomes in the assignment lifecycle.
CREATE OR REPLACE FUNCTION release_candidate_recruitment_engagement()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('joined', 'rejected', 'withdrawn')
       AND OLD.status IS DISTINCT FROM NEW.status THEN
        UPDATE candidates
           SET active_recruitment_engagements = FALSE
         WHERE id = NEW.candidate_id
           AND tenant_id = NEW.tenant_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_release_candidate_recruitment_engagement
    ON recruitment_assignments;

CREATE TRIGGER trg_release_candidate_recruitment_engagement
AFTER UPDATE OF status ON recruitment_assignments
FOR EACH ROW
EXECUTE FUNCTION release_candidate_recruitment_engagement();
