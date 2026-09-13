-- Add the two fields Business Development tracked that Clients did not:
-- partner_name (an optional referral/introducing partner) and
-- contact_person (the named individual at the client, distinct from a
-- general contact_email/contact_phone). Business Development is being
-- retired as a separate tab — Clients is the single client record going
-- forward, with all fields Business Development had.
ALTER TABLE clients ADD COLUMN IF NOT EXISTS partner_name VARCHAR(255);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS contact_person VARCHAR(255);
