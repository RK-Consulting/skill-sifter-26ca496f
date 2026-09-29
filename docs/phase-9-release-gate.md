# SkillSifter Phase 9 — Security / UAT / Go-Live Release Gate

This issue is the final release gate for SkillSifter Phase 9.

**Scope:** verify the implementation already consolidated into `main`.  
**Rule:** no new architecture, HRMS scope, microservices, accounting/ledger, or new permission framework is added under this gate.

## 1. Security verification

- [ ] Authentication: login/registration/JWT tenant identity verified
- [ ] Tenant identity cannot be overridden by client-supplied `tenant_id`
- [ ] Public registration cannot select an existing tenant or arbitrary role
- [ ] RBAC verified for admin / manager / recruiter / team_leader
- [ ] Tenant A cannot read Tenant B data
- [ ] Tenant A cannot modify/delete Tenant B data
- [ ] Tenant-owned requests use the authenticated tenant database
- [ ] Control-plane authentication/subscription data remains on control DB
- [ ] Expired/cancelled subscription blocks protected access
- [ ] Webhook authentication/signature validation verified
- [ ] Duplicate webhook is idempotent
- [ ] Tenant admin cannot be accidentally deleted/demoted
- [ ] User seat limits enforced server-side

## 2. Negative/security tests

- [ ] Tenant A → Tenant B candidate access: DENY
- [ ] Tenant A → Tenant B requirement access: DENY
- [ ] Tenant A → Tenant B client access: DENY
- [ ] Tenant A → Tenant B billing access: DENY
- [ ] Recruiter → admin-only operation: DENY
- [ ] Team leader → restricted operation: DENY
- [ ] Non-admin → user management: DENY
- [ ] Expired subscription → protected API: DENY
- [ ] Invalid provider webhook signature: DENY
- [ ] Duplicate provider webhook: no duplicate state/effect

## 3. Tenant database

- [ ] Tenant DB provisioning succeeds for a new tenant
- [ ] Tenant schema migrations apply cleanly
- [ ] Tenant seed/admin creation succeeds
- [ ] Provisioning status is enforced
- [ ] Login is blocked until tenant DB is READY
- [ ] Admin provisioning retry works
- [ ] Tenant DB routing verified for tenant-owned domains
- [ ] Cross-tenant negative tests pass

## 4. Admin User Management

- [ ] Admin can list users
- [ ] Admin can create users
- [ ] Admin can update users
- [ ] Admin can delete users where permitted
- [ ] Tenant admin protection verified
- [ ] Seat usage/limit displayed correctly
- [ ] Non-admin cannot access user management

## 5. Subscription lifecycle

- [ ] Plans are available
- [ ] Checkout handoff works
- [ ] Provider webhook is authenticated
- [ ] Checkout/webhook activates subscription
- [ ] Renewal transition verified
- [ ] Payment failure transition verified
- [ ] Cancellation transition verified
- [ ] Expiry transition verified
- [ ] Duplicate webhook is idempotent
- [ ] Account API reads subscription from control DB
- [ ] Subscription state is enforced server-side

## 6. Account & Subscription UI

- [ ] Account details displayed
- [ ] Current plan/status displayed
- [ ] Available plans displayed
- [ ] Checkout action works
- [ ] Admin-only cancellation works
- [ ] No accounting/financial-ledger UI added

## 7. Recruitment lifecycle UAT

- [ ] Requirement
- [ ] Assignment
- [ ] Screening
- [ ] Submission
- [ ] Feedback
- [ ] Interview
- [ ] Selection
- [ ] Offer
- [ ] Joining
- [ ] JOINED
- [ ] Billing

**Offer:** made + accepted only.  
**Joining:** joining date + joined only.

## 8. CI release gate

- [ ] Formatting passes
- [ ] Build passes
- [ ] Vet passes
- [ ] Backend tests pass
- [ ] PostgreSQL-backed tests pass
- [ ] Frontend lint/tests pass
- [ ] Frontend build passes
- [ ] Migration on clean database passes
- [ ] Tenant provisioning/routing tests pass
- [ ] Security negative tests pass

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

- [ ] Production JWT secret configured
- [ ] Production database credentials configured
- [ ] Production provider credentials/webhook secret configured
- [ ] CORS configured for production
- [ ] Production frontend API URL configured
- [ ] HTTPS enabled
- [ ] Secure cookie/session settings verified where applicable
- [ ] Production migrations prepared
- [ ] Database backup confirmed
- [ ] Rollback procedure confirmed
- [ ] No secrets committed to repository

## 10. Production smoke test

- [ ] Register tenant
- [ ] Tenant DB provisioned
- [ ] Admin login
- [ ] Create user
- [ ] User login
- [ ] Create client
- [ ] Create requirement
- [ ] Add candidate
- [ ] Assignment
- [ ] Screening
- [ ] Submission
- [ ] Feedback
- [ ] Interview
- [ ] Offer
- [ ] Joining
- [ ] Mark joined
- [ ] Billing
- [ ] Account/subscription check

## Final release decision

- [ ] CI = PASS
- [ ] Security = PASS
- [ ] Tenant isolation = PASS
- [ ] RBAC = PASS
- [ ] Subscription = PASS
- [ ] Database migration = PASS
- [ ] UAT = PASS
- [ ] Production smoke = PASS

### GO-LIVE

When all required boxes above are checked, Phase 9 is complete and the architecture is frozen.

Post-go-live work is new product work and must not be added to this release gate.