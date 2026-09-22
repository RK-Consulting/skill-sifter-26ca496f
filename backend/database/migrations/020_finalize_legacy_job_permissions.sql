-- 020_finalize_legacy_job_permissions.sql
-- Finalize removal of legacy Jobs permissions using the corrected JSONB
-- implementation. Migration 016 is historical and must remain immutable.
--
-- This migration is intentionally idempotent. Existing environments where
-- the legacy permissions were already removed are unaffected.

UPDATE roles
SET permissions = (
    SELECT COALESCE(
        jsonb_agg(to_jsonb(value)),
        '[]'::jsonb
    )
    FROM jsonb_array_elements_text(permissions) AS p(value)
    WHERE value NOT IN ('manage_jobs', 'view_jobs')
)
WHERE permissions ?| ARRAY['manage_jobs', 'view_jobs'];
