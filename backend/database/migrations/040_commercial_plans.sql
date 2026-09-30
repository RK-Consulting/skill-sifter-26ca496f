-- 040_commercial_plans.sql
-- Launch pricing: two customer-facing plans with monthly/annual billing variants.
-- Razorpay plan references are populated after provider plans are created.

INSERT INTO platform_plans (code,name,amount_minor,currency,billing_period,billing_interval,user_limit,total_count,provider,provider_plan_ref,active)
VALUES
('starter_monthly','Starter',99900,'INR','monthly',1,3,120,'razorpay',NULL,TRUE),
('starter_annual','Starter',999000,'INR','annual',1,3,12,'razorpay',NULL,TRUE),
('professional_monthly','Professional',249900,'INR','monthly',1,10,120,'razorpay',NULL,TRUE),
('professional_annual','Professional',2499000,'INR','annual',1,10,12,'razorpay',NULL,TRUE)
ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name,amount_minor=EXCLUDED.amount_minor,currency=EXCLUDED.currency,billing_period=EXCLUDED.billing_period,billing_interval=EXCLUDED.billing_interval,user_limit=EXCLUDED.user_limit,total_count=EXCLUDED.total_count,provider=EXCLUDED.provider,active=TRUE,updated_at=NOW();

UPDATE platform_plans SET active=FALSE WHERE code='TBD';
