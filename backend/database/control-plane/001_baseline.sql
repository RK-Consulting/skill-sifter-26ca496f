-- SkillSifter final control-plane baseline.
-- Platform/SaaS state only. No recruitment-domain tables.
-- No foreign key crosses into a tenant database.

CREATE TABLE platform_tenants (
    tenant_id VARCHAR(255) PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL UNIQUE,
    account_status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    provisioning_status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    tenant_database VARCHAR(255),
    trial_started_at TIMESTAMP,
    trial_expires_at TIMESTAMP,
    data_deletion_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_tenants_account_status_valid
        CHECK (account_status IN ('ACTIVE','SUSPENDED','EXPIRED','TERMINATED')),
    CONSTRAINT platform_tenants_provisioning_status_valid
        CHECK (provisioning_status IN ('PENDING','PROVISIONING','READY','FAILED'))
);

CREATE INDEX idx_platform_tenants_provisioning
    ON platform_tenants(provisioning_status);

CREATE TABLE platform_plans (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    billing_period VARCHAR(20) NOT NULL,
    billing_interval INTEGER NOT NULL DEFAULT 1,
    user_limit INTEGER NOT NULL,
    total_count INTEGER,
    provider VARCHAR(100) NOT NULL,
    provider_plan_ref VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_plans_amount_valid CHECK (amount_minor >= 0),
    CONSTRAINT platform_plans_interval_valid CHECK (billing_interval > 0),
    CONSTRAINT platform_plans_user_limit_valid CHECK (user_limit > 0),
    CONSTRAINT platform_plans_currency_valid CHECK (currency = upper(currency))
);

CREATE UNIQUE INDEX uq_platform_plans_provider_ref
    ON platform_plans(provider, provider_plan_ref)
    WHERE provider_plan_ref IS NOT NULL;

CREATE TABLE platform_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    plan_code VARCHAR(100) NOT NULL REFERENCES platform_plans(code),
    status VARCHAR(30) NOT NULL,
    starts_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMP,
    provider VARCHAR(100),
    provider_subscription_ref VARCHAR(255),
    user_limit INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_subscriptions_status_valid
        CHECK (status IN ('TRIAL','ACTIVE','PAST_DUE','SUSPENDED','CANCELLED','EXPIRED')),
    CONSTRAINT platform_subscriptions_user_limit_valid
        CHECK (user_limit > 0),
    CONSTRAINT platform_subscriptions_dates_valid
        CHECK (ends_at IS NULL OR ends_at >= starts_at)
);

CREATE INDEX idx_platform_subscriptions_tenant_status
    ON platform_subscriptions(tenant_id, status, starts_at DESC);

CREATE UNIQUE INDEX uq_platform_subscriptions_provider_ref
    ON platform_subscriptions(provider, provider_subscription_ref)
    WHERE provider_subscription_ref IS NOT NULL;

CREATE TABLE platform_subscription_checkouts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    plan_code VARCHAR(100) NOT NULL REFERENCES platform_plans(code),
    provider VARCHAR(100) NOT NULL,
    provider_subscription_ref VARCHAR(255),
    checkout_url TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_subscription_checkouts_status_valid
        CHECK (status IN ('PENDING','COMPLETED','FAILED','CANCELLED'))
);

CREATE INDEX idx_platform_subscription_checkouts_tenant
    ON platform_subscription_checkouts(tenant_id, created_at DESC);

CREATE TABLE platform_subscription_events (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(100) NOT NULL,
    provider_event_ref VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    provider_subscription_ref VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_event_ref)
);

CREATE INDEX idx_platform_subscription_events_tenant
    ON platform_subscription_events(tenant_id, received_at DESC);

CREATE TABLE platform_registration_registry (
    email_id VARCHAR(255) PRIMARY KEY,
    first_registered TIMESTAMP NOT NULL DEFAULT NOW(),
    last_tenant_id VARCHAR(255)
);

CREATE TABLE platform_pending_registrations (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    plan_code VARCHAR(100) NOT NULL REFERENCES platform_plans(code),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),
    email_verified_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_platform_pending_registrations_email
    ON platform_pending_registrations(lower(email));

CREATE TABLE platform_verification_codes (
    id BIGSERIAL PRIMARY KEY,
    purpose VARCHAR(30) NOT NULL,
    registration_id BIGINT,
    platform_account_id BIGINT,
    destination VARCHAR(255) NOT NULL,
    code_hash VARCHAR(128) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_verification_codes_purpose_valid
        CHECK (purpose IN ('EMAIL_SIGNUP','PHONE_SUBSCRIPTION')),
    CONSTRAINT platform_verification_codes_attempts_valid
        CHECK (attempts >= 0)
);

CREATE INDEX idx_platform_verification_codes_registration
    ON platform_verification_codes(registration_id, purpose, created_at DESC);

CREATE INDEX idx_platform_verification_codes_account
    ON platform_verification_codes(platform_account_id, purpose, created_at DESC);

CREATE TABLE platform_user_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    user_id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    phone_verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_user_accounts_tenant_user_unique UNIQUE (tenant_id, user_id)
);

CREATE UNIQUE INDEX uq_platform_user_accounts_email
    ON platform_user_accounts(lower(email));

CREATE INDEX idx_platform_user_accounts_tenant
    ON platform_user_accounts(tenant_id);
