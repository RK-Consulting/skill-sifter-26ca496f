-- 044_persistent_registration_registry.sql
-- Permanent platform-level registration identity.
-- This survives tenant/trial deletion and prevents reuse of an email address.

CREATE TABLE IF NOT EXISTS platform_registration_registry (
    email_id VARCHAR(320) PRIMARY KEY,
    first_registered TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_tenant_id VARCHAR(255)
);

COMMENT ON TABLE platform_registration_registry IS
    'Permanent registration identity. Email addresses remain registered after tenant/trial data is deleted.';

COMMENT ON COLUMN platform_registration_registry.email_id IS
    'Canonical email address. This is the permanent SkillSifter registration identity.';

COMMENT ON COLUMN platform_registration_registry.first_registered IS
    'Timestamp when this email first completed registration.';

COMMENT ON COLUMN platform_registration_registry.last_tenant_id IS
    'Most recent tenant associated with this registration. Historical reference only.';
