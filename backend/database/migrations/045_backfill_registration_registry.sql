-- 045_backfill_registration_registry.sql
-- Backfill the permanent registration identity from all existing users.
-- Migration 044 established the registry for new registrations, but existing
-- users predate that registry and must also be permanently registered.
--
-- users.email is already protected by a unique constraint/index. Therefore
-- each normalized email maps to at most one existing account.

INSERT INTO platform_registration_registry (email_id, first_registered, last_tenant_id)
SELECT
    LOWER(BTRIM(email)) AS email_id,
    MIN(created_at)::TIMESTAMPTZ AS first_registered,
    MIN(tenant_id) AS last_tenant_id
FROM users
WHERE email IS NOT NULL
  AND BTRIM(email) <> ''
GROUP BY LOWER(BTRIM(email))
ON CONFLICT (email_id) DO NOTHING;

COMMENT ON TABLE platform_registration_registry IS
    'Permanent registration identity. Includes all historical and new registered email addresses. Email addresses remain registered after tenant/trial data is deleted.';
