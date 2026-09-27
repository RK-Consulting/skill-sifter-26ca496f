-- 028_recruitment_selection.sql
-- Phase 6: agency-first selection decision.
CREATE TABLE IF NOT EXISTS recruitment_selections (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    assignment_id INTEGER NOT NULL REFERENCES recruitment_assignments(id),
    decision VARCHAR(20) NOT NULL,
    decision_notes TEXT,
    next_action TEXT,
    decided_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT recruitment_selection_decision_valid CHECK (decision IN ('selected', 'rejected')),
    CONSTRAINT recruitment_selection_assignment_unique UNIQUE (assignment_id)
);

CREATE INDEX IF NOT EXISTS idx_recruitment_selections_tenant_decided
    ON recruitment_selections(tenant_id, decided_at DESC);

CREATE INDEX IF NOT EXISTS idx_recruitment_selections_tenant_assignment
    ON recruitment_selections(tenant_id, assignment_id);
