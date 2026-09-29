-- 039_tbd_subscription_plan.sql
-- Phase 9 smoke-test placeholder.
-- Real commercial plan values will be defined before paid subscription go-live.
-- Keep this plan inactive so it cannot be selected for customer checkout.

INSERT INTO platform_plans (
    code,
    name,
    amount_minor,
    currency,
    billing_period,
    billing_interval,
    user_limit,
    total_count,
    provider,
    provider_plan_ref,
    active
)
VALUES (
    'TBD',
    'TBD',
    0,
    'INR',
    'TBD',
    1,
    1,
    1,
    'razorpay',
    'TBD',
    FALSE
)
ON CONFLICT (code) DO UPDATE
SET
    name = EXCLUDED.name,
    amount_minor = EXCLUDED.amount_minor,
    currency = EXCLUDED.currency,
    billing_period = EXCLUDED.billing_period,
    billing_interval = EXCLUDED.billing_interval,
    user_limit = EXCLUDED.user_limit,
    total_count = EXCLUDED.total_count,
    provider = EXCLUDED.provider,
    provider_plan_ref = EXCLUDED.provider_plan_ref,
    active = FALSE,
    updated_at = NOW();
