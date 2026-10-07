-- 044_persistent_registration_registry.sql
-- Permanent platform-level registration registry.
-- This survives tenant/trial deletion and prevents reuse of an email address.

CREATE TABLE IF NOT EXISTS platform_registration_registry (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    first_registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_tenant_id VARCHAR(255),
    status VARCHAR(32) NOT NULL DEFAULT 'REGISTERED'
        CHECK (status IN ('REGISTERED', 'TRIAL_EXPIRED', 'DELETED')),
    CONSTRAINT uq_platform_registration_registry_email UNIQUE (email)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_registration_registry_email_lower
    ON platform_registration_registry (LOWER(email));

COMMENT ON TABLE platform_registration_registry IS
    'Permanent platform registration history. Email addresses remain registered even after tenant/trial data is deleted.';

COMMENT ON COLUMN platform_registration_registry.email IS
    'Canonical email address retained permanently to prevent reuse for a new registration.';
