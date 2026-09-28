-- 034_billing_workflow.sql
-- One Billing record per joined Candidate × Requirement.
-- Billing is a recruitment-firm commercial record, not an accounting ledger.
CREATE TABLE IF NOT EXISTS recruitment_billings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    candidate_id INTEGER NOT NULL REFERENCES candidates(id),
    requirement_id INTEGER NOT NULL REFERENCES requirements(id),
    client_id INTEGER NOT NULL REFERENCES clients(id),
    joining_id INTEGER NOT NULL REFERENCES recruitment_joinings(id),
    billing_date DATE NOT NULL,
    amount NUMERIC(14,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    invoice_reference VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_recruitment_billings_tenant_pair
    ON recruitment_billings(tenant_id, candidate_id, requirement_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_billings_tenant_client
    ON recruitment_billings(tenant_id, client_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_billings_tenant_joining
    ON recruitment_billings(tenant_id, joining_id);

CREATE INDEX IF NOT EXISTS idx_recruitment_billings_candidate_requirement
    ON recruitment_billings(tenant_id, candidate_id, requirement_id);
