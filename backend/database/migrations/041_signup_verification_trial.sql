-- 041_signup_verification_trial.sql
-- Email verification creates the 2-day trial. Phone verification gates paid checkout.

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(30), ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP, ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMP;
ALTER TABLE platform_tenants ADD COLUMN IF NOT EXISTS trial_started_at TIMESTAMP, ADD COLUMN IF NOT EXISTS trial_expires_at TIMESTAMP, ADD COLUMN IF NOT EXISTS data_deletion_at TIMESTAMP;

CREATE TABLE IF NOT EXISTS platform_pending_registrations (
 id BIGSERIAL PRIMARY KEY,
 username VARCHAR(255) NOT NULL,
 email VARCHAR(255) NOT NULL,
 company_name VARCHAR(255) NOT NULL,
 password_hash VARCHAR(255) NOT NULL,
 plan_code VARCHAR(100) NOT NULL REFERENCES platform_plans(code),
 created_at TIMESTAMP NOT NULL DEFAULT NOW(),
 expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),
 email_verified_at TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_pending_registrations_email ON platform_pending_registrations(LOWER(email)) WHERE email_verified_at IS NULL;

CREATE TABLE IF NOT EXISTS platform_verification_codes (
 id BIGSERIAL PRIMARY KEY,
 purpose VARCHAR(30) NOT NULL,
 registration_id BIGINT REFERENCES platform_pending_registrations(id) ON DELETE CASCADE,
 user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
 destination VARCHAR(255) NOT NULL,
 code_hash VARCHAR(128) NOT NULL,
 attempts INTEGER NOT NULL DEFAULT 0,
 expires_at TIMESTAMP NOT NULL,
 consumed_at TIMESTAMP,
 created_at TIMESTAMP NOT NULL DEFAULT NOW(),
 CONSTRAINT platform_verification_codes_purpose_chk CHECK (purpose IN ('EMAIL_SIGNUP','PHONE_SUBSCRIPTION'))
);
CREATE INDEX IF NOT EXISTS idx_platform_verification_codes_lookup ON platform_verification_codes(purpose,destination,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_platform_tenants_trial_cleanup ON platform_tenants(data_deletion_at) WHERE data_deletion_at IS NOT NULL;
