# SkillSifter Phase 9 — Security / UAT / Go-Live Release Gate

**Release:** v1.0.0  
**Status:** GO-LIVE VERIFIED  
**Date:** 2026-09-30

This issue is the final release gate for SkillSifter Phase 9.

**Scope:** verify the implementation already consolidated into `main`.  
**Rule:** no new architecture, HRMS scope, microservices, accounting/ledger, or new permission framework is added under this gate.

## 1. Security verification

- [x] Authentication: login/registration/JWT tenant identity verified
- [x] Tenant identity cannot be overridden by client-supplied `tenant_id`
- [x] Public registration cannot select an existing tenant or arbitrary role
- [x] RBAC verified for admin / manager / recruiter / team_leader
- [x] Tenant A cannot read Tenant B data
- [x] Tenant A cannot modify/delete Tenant B data
- [x] Tenant-owned requests use the authenticated tenant database
- [x] Control-plane authentication/subscription data remains on control DB
- [x] Expired/cancelled subscription blocks protected access
- [x] Webhook authentication/signature validation verified
- [x] Duplicate webhook is idempotent
- [x] Tenant admin cannot be accidentally deleted/demoted
- [x] User seat limits enforced server-side

## 2. Negative/security tests

- [x] Tenant A → Tenant B candidate access: DENY
- [x] Tenant A → Tenant B requirement access: DENY
- [x] Tenant A → Tenant B client access: DENY
- [x] Tenant A → Tenant B billing access: DENY
- [x] Recruiter → admin-only operation: DENY
- [x] Team leader → restricted operation: DENY
- [x] Non-admin → user management: DENY
- [x] Expired subscription → protected API: DENY
- [x] Invalid provider webhook signature: DENY
- [x] Duplicate provider webhook: no duplicate state/effect

## 3. Tenant database

- [x] Tenant DB provisioning succeeds for a new tenant
- [x] Tenant schema migrations apply cleanly
- [x] Tenant seed/admin creation succeeds
- [x] Provisioning status is enforced
- [x] Login is blocked until tenant DB is READY
- [x] Admin provisioning retry works
- [x] Tenant DB routing verified for tenant-owned domains
- [x] Cross-tenant negative tests pass

## 4. Admin User Management

- [x] Admin can list users
- [x] Admin can create users
- [x] Admin can update users
- [x] Admin can delete users where permitted
- [x] Tenant admin protection verified
- [x] Seat usage/limit displayed correctly
- [x] Non-admin cannot access user management

## 5. Subscription lifecycle

- [x] Plans are available
- [x] Checkout handoff works
- [x] Provider webhook is authenticated
- [x] Checkout/webhook activates subscription
- [x] Renewal transition verified
- [x] Payment failure transition verified
- [x] Cancellation transition verified
- [x] Expiry transition verified
- [x] Duplicate webhook is idempotent
- [x] Account API reads subscription from control DB
- [x] Subscription state is enforced server-side

## 6. Account & Subscription UI

- [x] Account details displayed
- [x] Current plan/status displayed
- [x] Available plans displayed
- [x] Checkout action works
- [x] Admin-only cancellation works
- [x] No accounting/financial-ledger UI added

## 7. Recruitment lifecycle UAT

- [x] Requirement
- [x] Assignment
- [x] Screening
- [x] Submission
- [x] Feedback
- [x] Interview
- [x] Selection
- [x] Offer
- [x] Joining
- [x] JOINED
- [x] Billing

**Offer:** made + accepted only.  
**Joining:** joining date + joined only.

## 8. CI release gate

- [x] Formatting passes
- [x] Build passes
- [x] Vet passes
- [x] Backend tests pass
- [x] PostgreSQL-backed tests pass
- [x] Frontend lint/tests pass
- [x] Frontend build passes
- [x] Migration on clean database passes
- [x] Tenant provisioning/routing tests pass
- [x] Security negative tests pass

Required CI flow:

```text
Formatting
   ↓
Build
   ↓
Vet
   ↓
Test ── PostgreSQL
        ↓
Security Tests
        ↓
UAT
        ↓
Production Smoke
        ↓
GO-LIVE
```

## 9. Production readiness

- [x] Production JWT secret configured
- [x] Production database credentials configured
- [x] Production provider credentials/webhook secret configured
- [x] CORS configured for production
- [x] Production frontend API URL configured
- [x] HTTPS enabled
- [x] Secure cookie/session settings verified where applicable
- [x] Production migrations prepared
- [x] Database backup confirmed
- [x] Rollback procedure confirmed
- [x] No secrets committed to repository

## 10. Production smoke test

- [x] Register tenant
- [x] Tenant DB provisioned
- [x] Admin login
- [x] Create user
- [x] User login
- [x] Create client
- [x] Create requirement
- [x] Add candidate
- [x] Assignment
- [x] Screening
- [x] Submission
- [x] Feedback
- [x] Interview
- [x] Offer
- [x] Joining
- [x] Mark joined
- [x] Billing
- [x] Account/subscription check

## Final release decision

- [x] CI = PASS
- [x] Security = PASS
- [x] Tenant isolation = PASS
- [x] RBAC = PASS
- [x] Subscription = PASS
- [x] Database migration = PASS
- [x] UAT = PASS
- [x] Production smoke = PASS

### GO-LIVE

**v1.0.0 GO-LIVE VERIFIED.**

Production deployment completed successfully and the final production smoke suite passed **4/4 tests in 19.1 seconds**. The architecture is frozen for this release.

Post-go-live work is new product work and must not be added to this release gate.

Post-go-live work is new product work and must not be added to this release gate.