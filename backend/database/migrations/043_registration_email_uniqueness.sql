-- 043_registration_email_uniqueness.sql
-- Registration email addresses must be unique while pending or active.
-- The permanent registration identity is enforced by migration 044.

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_lower
    ON users (LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS uq_pending_registration_email_lower
    ON platform_pending_registrations (LOWER(email))
    WHERE email_verified_at IS NULL;
