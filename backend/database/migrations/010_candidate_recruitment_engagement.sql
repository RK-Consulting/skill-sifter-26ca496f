-- 010_candidate_recruitment_engagement.sql
-- CP10: a candidate may have one active recruitment engagement within the
-- tenant database at a time. This is a simple availability gate, not a
-- cross-tenant coordination mechanism.

ALTER TABLE candidates
    ADD COLUMN IF NOT EXISTS active_recruitment_engagements BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_candidates_active_recruitment_engagements
    ON candidates(tenant_id, active_recruitment_engagements);

-- The recruitment service atomically claims/releases this flag inside the
-- same transaction as assignment processing. No trigger or cross-tenant
-- coordination is required.
