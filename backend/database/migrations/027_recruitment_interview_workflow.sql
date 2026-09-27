-- 027_recruitment_interview_workflow.sql
-- Phase 5: agency-first interview workflow.
-- Extends the legacy interviews table without introducing enterprise-HR
-- concepts such as interviewer panels, hiring managers, approvals, or
-- submission dependencies.

ALTER TABLE interviews
    ADD COLUMN IF NOT EXISTS round INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS duration_minutes INTEGER,
    ADD COLUMN IF NOT EXISTS outcome VARCHAR(100),
    ADD COLUMN IF NOT EXISTS candidate_feedback TEXT,
    ADD COLUMN IF NOT EXISTS next_action TEXT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();

ALTER TABLE interviews
    ADD CONSTRAINT interviews_round_positive CHECK (round > 0);

ALTER TABLE interviews
    ADD CONSTRAINT interviews_duration_positive
    CHECK (duration_minutes IS NULL OR duration_minutes > 0);

CREATE INDEX IF NOT EXISTS idx_interviews_tenant_requirement_round
    ON interviews(tenant_id, requirement_id, round, interview_date DESC);

CREATE INDEX IF NOT EXISTS idx_interviews_tenant_candidate
    ON interviews(tenant_id, candidate_id, interview_date DESC);
