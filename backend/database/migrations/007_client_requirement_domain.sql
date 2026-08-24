-- 007_client_requirement_domain.sql
-- Issue #34 / ADR 0002, Stage 1 of the ADR's staged migration strategy:
-- "Introduce the Client, Client Contact, and Requirement domain model."
--
-- Client Contact is NOT included in this migration. ADR 0002 defines it,
-- but Issue #34's own scope list names only "Client entity and lifecycle"
-- and "Requirement entity and lifecycle" — Client Contact is deferred as a
-- scoping decision, not an oversight (see PR description).
--
-- This migration is purely additive: it creates two new tables and does
-- NOT modify, rename, or drop the existing `jobs` table, its data, or any
-- other existing table. Per ADR 0002: "No destructive replacement or
-- uncontrolled rename is authorized as part of this architecture decision."
--
-- Mapping existing `jobs` rows into `requirements` (ADR 0002's migration
-- strategy step 3) is explicitly NOT performed by this migration — that is
-- a separate, consequential data-migration decision flagged for review
-- rather than decided here (see PR description "Known limitations").

CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'prospect',
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT clients_status_valid CHECK (status IN ('prospect', 'active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_clients_tenant ON clients(tenant_id);

CREATE TABLE IF NOT EXISTS requirements (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES companies(id),
    client_id INTEGER NOT NULL REFERENCES clients(id),
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    location VARCHAR(100),
    work_arrangement VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    opened_date TIMESTAMP,
    description TEXT,
    required_skills TEXT,
    experience_required VARCHAR(100),
    compensation VARCHAR(255),
    headcount INTEGER NOT NULL DEFAULT 1,
    language_requirement VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT requirements_status_valid CHECK (status IN ('draft', 'open', 'on_hold', 'filled', 'cancelled')),
    CONSTRAINT requirements_headcount_positive CHECK (headcount > 0)
);

CREATE INDEX IF NOT EXISTS idx_requirements_tenant ON requirements(tenant_id);
CREATE INDEX IF NOT EXISTS idx_requirements_client ON requirements(client_id);

-- Defensive: a requirement's client must belong to the same tenant as the
-- requirement itself. Postgres doesn't support a cross-table CHECK
-- constraint directly, so this is enforced at the application layer
-- (requirement_handlers.go validates client_id belongs to the
-- authenticated tenant before insert/update) rather than in SQL here.
