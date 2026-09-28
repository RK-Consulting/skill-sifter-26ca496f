-- 037_subscription_lifecycle.sql
-- Phase 9E: subscription plans, checkout handoff and provider event idempotency.
-- SkillSifter stores only subscription state and opaque provider references.
-- Payment instruments, recurring charge execution and financial records remain
-- with the external provider.

CREATE TABLE IF NOT EXISTS platform_plans (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'INR',
    billing_period VARCHAR(20) NOT NULL,
    billing_interval INTEGER NOT NULL DEFAULT 1 CHECK (billing_interval > 0),
    user_limit INTEGER NOT NULL DEFAULT 1 CHECK (user_limit > 0),
    total_count INTEGER NOT NULL DEFAULT 12 CHECK (total_count > 0),
    provider VARCHAR(100) NOT NULL DEFAULT 'razorpay',
    provider_plan_ref VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform_subscription_checkouts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    plan_code VARCHAR(100) NOT NULL REFERENCES platform_plans(code),
    provider VARCHAR(100) NOT NULL,
    provider_subscription_ref VARCHAR(255),
    checkout_url TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_subscription_checkouts_status_chk
        CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED'))
);

CREATE INDEX IF NOT EXISTS idx_platform_subscription_checkouts_tenant
    ON platform_subscription_checkouts(tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS platform_subscription_events (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(100) NOT NULL,
    provider_event_ref VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    provider_subscription_ref VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_event_ref)
);

CREATE INDEX IF NOT EXISTS idx_platform_subscription_events_subscription
    ON platform_subscription_events(provider, provider_subscription_ref);
