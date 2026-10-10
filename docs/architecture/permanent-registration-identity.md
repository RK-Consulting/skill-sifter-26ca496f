# Registration Identity, Tenant Lifecycle, and Production Validation Design

**Status:** Implemented and frozen for v1.0.0  
**Scope:** SaaS registration, permanent registration identity, trial lifecycle, tenant provisioning, login recovery, and paid subscription activation  
**Related schema:** `backend/database/control-plane/001_baseline.sql`  
**Last updated:** 2026-10-07

## 1. Purpose

This document defines the SkillSifter SaaS registration identity and tenant lifecycle.

The key architectural rule is:

> **Tenant data is disposable; registration identity is permanent.**

A trial tenant may expire and its tenant database may be physically deleted. The email address that successfully completed registration remains permanently registered with SkillSifter and cannot be used to create another SkillSifter account.

This separates the lifetime of the customer identity from the lifetime of disposable tenant data.

## 2. Canonical identities

SkillSifter has two different identities with different lifetimes.

### 2.1 Permanent registration identity

The permanent registration identity is the normalized registration email.

It is stored in the control-plane table:

`platform_registration_registry`

Current schema:

```sql
CREATE TABLE IF NOT EXISTS platform_registration_registry (
    email_id VARCHAR(320) PRIMARY KEY,
    first_registered TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_tenant_id VARCHAR(255)
);
```

The table intentionally contains only:

- `email_id`
- `first_registered`
- `last_tenant_id`

It does **not** store:

- password
- phone number
- payment information
- subscription state
- tenant operational data

### 2.2 Tenant identity

`tenant_id` remains the canonical SaaS customer identity for the active tenant.

`company_name` is descriptive information and is not the permanent customer identity.

The tenant and its database are disposable. The registration email is not.

## 3. Registration lifecycle

The registration flow is:

```text
User enters registration details
          ↓
Normalize email
          ↓
Check permanent registration registry
          ↓
Check unexpired pending registration
          ↓
Create pending registration
          ↓
Send email OTP
          ↓
Verify OTP
          ↓
Create platform tenant
          ↓
Create tenant-bound user
          ↓
Create trial subscription
          ↓
Create platform user account
          ↓
Consume verification
          ↓
Persist permanent registration identity
          ↓
Commit control-plane transaction
          ↓
Provision tenant database
          ↓
Tenant READY
```

### 3.1 Ordering constraint

The control-plane tenant must be created before inserting any tenant-bound user.

Required order:

```text
platform_tenants
      ↓
users
      ↓
platform_subscriptions
      ↓
platform_user_accounts
```

This ordering is required because `users.tenant_id` references `platform_tenants.tenant_id`.

This specifically prevents the earlier registration failure where the user insert could fail because its tenant did not yet exist.

## 4. Permanent email uniqueness

Registration checks the permanent registry before creating a new pending registration.

Conceptually:

```sql
SELECT EXISTS(
    SELECT 1
    FROM platform_registration_registry
    WHERE email_id = $1
)
OR EXISTS(
    SELECT 1
    FROM platform_pending_registrations
    WHERE email = $1
      AND email_verified_at IS NULL
      AND expires_at > NOW()
);
```

Therefore:

- an already registered email cannot start another registration;
- an unexpired pending registration cannot be duplicated;
- a different company cannot reuse an already registered email;
- a different username cannot bypass the email identity rule;
- tenant deletion does not release the email address.

Migration 043 protects uniqueness while a registration is pending or active. Migration 044 provides the permanent registration identity.

## 5. What happens after trial deletion

Trial data follows the disposable tenant lifecycle.

```text
TRIAL
  ↓
trial expires
  ↓
tenant becomes EXPIRED
  ↓
data deletion period
  ↓
tenant database deleted
  ↓
platform tenant deleted
  ↓
tenant-owned control records cascade/delete
```

The permanent registry is **not** deleted.

The resulting state is:

```text
Disposable tenant data       → deleted
Tenant database              → deleted
Tenant/user/subscription     → deleted
Permanent registration email → retained
```

The same email therefore remains unavailable for a new SkillSifter registration.

## 6. Provisioning failure and recovery

Tenant provisioning happens after the control-plane transaction has committed.

If provisioning fails:

```text
Registration verified
      ↓
Control-plane transaction committed
      ↓
Tenant provisioning attempted
      ↓
Provisioning fails
      ↓
Tenant marked provisioning FAILED
      ↓
Admin can authenticate through provisioning recovery
      ↓
Provision tenant again
      ↓
Tenant READY
```

Normal application login requires a READY tenant.

Provisioning recovery deliberately does not require the tenant to already be READY. This prevents the original failure mode where a provisioning failure also prevented the administrator from reaching the provisioning endpoint needed to repair it.

Provisioning must remain idempotent.

## 7. Email OTP

Trial registration requires email verification.

Expected flow:

```text
Registration
    ↓
Email OTP
    ↓
Correct OTP
    ↓
Verified registration
```

Incorrect or expired OTPs are rejected.

The OTP input must display entered digits visibly to the user.

## 8. Paid subscription and phone verification

Phone verification is required for paid checkout, not for trial registration.

The paid flow is:

```text
Trial account
     ↓
Choose paid plan
     ↓
Verify administrator phone
     ↓
Continue to payment
     ↓
External payment provider
     ↓
Provider confirmation/webhook
     ↓
Activate subscription
```

Without a verified administrator phone number, checkout must remain blocked.

SkillSifter does not store payment-instrument data. Payment-provider-specific financial information remains with the external provider.

## 9. Tenant isolation

`tenant_id` is the canonical SaaS customer context.

The client must not be able to select:

- another tenant ID;
- another tenant database;
- another company;
- another tenant role.

Authentication establishes the trusted tenant context before tenant application access.

Tenant isolation and RBAC remain separate:

```text
Authenticated identity
        │
        ├── tenant_id → which customer/data
        │
        └── role       → which operations
```

## 10. Security and lifecycle invariants

The following rules are architectural invariants:

1. A verified registration email is permanently reserved.
2. Tenant deletion never deletes the permanent registration identity.
3. `tenant_id` is the canonical active SaaS customer identity.
4. Company name is descriptive, not identity.
5. Tenant-bound users are created only after their tenant exists.
6. Provisioning failure must not make recovery impossible.
7. Normal tenant access requires a READY tenant.
8. Provisioning recovery may operate on a non-READY tenant.
9. Trial registration requires email verification.
10. Paid checkout requires administrator phone verification.
11. Payment instruments remain outside SkillSifter.
12. Tenant A must never access Tenant B data.

## 11. Production acceptance test

The minimum production smoke test is:

```text
Fresh email
   ↓
Register
   ↓
Receive OTP
   ↓
Verify
   ↓
Tenant provisioned
   ↓
Login
   ↓
Dashboard
   ↓
Register same email again
   ↓
HTTP 409 / registration rejected
```

The paid-path check is:

```text
Trial admin
   ↓
Attempt checkout without phone verification
   ↓
Blocked
   ↓
Verify phone
   ↓
Checkout allowed
```

Do not deliberately corrupt the production database or provisioning infrastructure merely to reproduce the historical provisioning failure. The recovery path should be verified through existing state/tests and, where necessary, a controlled non-destructive staging test.

## 12. Long-term lifecycle validation

The full trial-retention rule cannot be validated by waiting during every release.

The intended lifecycle is:

```text
REGISTERED EMAIL
      ↓
TRIAL
      ↓
TRIAL EXPIRED
      ↓
DATA DELETION PERIOD
      ↓
TENANT DB DELETED
      ↓
TENANT CONTROL DATA DELETED
      ↓
REGISTRATION EMAIL STILL EXISTS
      ↓
RE-REGISTRATION BLOCKED
```

A scheduled lifecycle/integration test should verify this behavior outside the normal release smoke test.

## 13. Related implementation

- Control-plane baseline: registration email uniqueness while pending/active and permanent registration identity.
- Registration verification handler: creates the tenant before tenant-bound user records.
- Provisioning access path: allows administrator recovery when provisioning is not READY.
- Trial cleanup: deletes disposable tenant data without touching the permanent registration registry.
- Subscription checkout: requires administrator phone verification before payment.

## 14. Design decision

This architecture intentionally avoids introducing another customer entity or retaining disposable tenant data solely to remember a registration.

The permanent registry is a small control-plane identity record.

The design principle is:

> **Remember the registration identity; delete the disposable tenant.**

This keeps the SaaS lifecycle simple while ensuring that a SkillSifter registration email remains globally unique for the lifetime of the platform.
