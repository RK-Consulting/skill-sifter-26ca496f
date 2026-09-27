-- 026_recruitment_submission_feedback.sql
-- Phase 4: structured client / hiring-manager feedback against candidate submissions.
CREATE TABLE IF NOT EXISTS recruitment_submission_feedback (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    submission_id INTEGER NOT NULL REFERENCES recruitment_submissions(id) ON DELETE CASCADE,
    feedback_by_user_id INTEGER NOT NULL REFERENCES users(id),
    outcome VARCHAR(50) NOT NULL,
    reason_code VARCHAR(100),
    comments TEXT,
    next_action TEXT,
    feedback_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT recruitment_submission_feedback_outcome_valid
        CHECK (outcome IN ('shortlist', 'hold', 'reject'))
);

CREATE INDEX IF NOT EXISTS idx_recruitment_submission_feedback_tenant
    ON recruitment_submission_feedback(tenant_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_submission_feedback_submission
    ON recruitment_submission_feedback(submission_id, feedback_at DESC);

CREATE INDEX IF NOT EXISTS idx_recruitment_submission_feedback_actor
    ON recruitment_submission_feedback(feedback_by_user_id);
