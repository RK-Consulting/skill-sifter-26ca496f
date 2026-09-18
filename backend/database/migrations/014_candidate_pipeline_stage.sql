-- The Candidates page "Actions" menu (Screening / Interview / Rejected) was
-- writing to the `status` column, which is reserved for ADR 0004's
-- eligibility gate (active/inactive/blacklisted/archived, enforced by
-- candidates_status_valid) and validated in Go via isValidCandidateStatus.
-- Screening/interview/rejected are not eligibility values — they are
-- recruitment pipeline stages, a different concept — so every such update
-- was rejected by the backend's own validation. This adds a dedicated
-- column for pipeline stage, independent of the eligibility status.
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS pipeline_stage VARCHAR(50) NOT NULL DEFAULT 'new';
