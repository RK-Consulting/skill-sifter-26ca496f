-- 043_registration_email_uniqueness.sql
-- A trial email is a unique customer identity at registration time.
-- Enforce uniqueness in the database as well as in the API.

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_lower
    ON users (LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS uq_pending_registration_email_lower
    ON platform_pending_registrations (LOWER(email))
    WHERE email_verified_at IS NULL;
