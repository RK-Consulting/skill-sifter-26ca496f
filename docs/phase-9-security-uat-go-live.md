# SkillSifter — Phase 9 Security, UAT & Go-Live Gate

**Status:** FINAL POST-AUDIT VERIFICATION PENDING — v1.0.0  
**Date:** 2026-09-30

## 1. Completed Phase 9 sequence

- Admin User Management UI — merged PR #88
- V1 privilege matrix enforcement — merged PR #91
- Tenant isolation audit and cross-tenant negative tests — merged PR #91
- Tenant DB provisioning — merged PR #92
- Tenant DB routing — merged PR #94
- Subscription lifecycle, plans, checkout handoff and provider webhooks — merged PR #97
- Account & Subscription UI — merged PR #98

The recruitment lifecycle remains frozen.

## 2. Security acceptance

### Authentication
- Real email/password login
- JWT contains trusted tenant identity
- Protected requests re-resolve tenant/account/subscription access
- Inactive tenant or subscription is denied
- Public registration creates a new tenant and initial Admin only

### Authorization
- Fixed roles only: Admin, Manager, Recruiter, Team Leader
- Admin-only tenant user management
- Protected mutations use the approved V1 role matrix
- No permission builder or arbitrary role creation
- Team/assignment scope is not invented where the current domain model cannot enforce it

### Tenant isolation
- Tenant ID is the authoritative security boundary
- Tenant-owned queries use the authenticated tenant context
- Cross-tenant known-ID read/update/delete tests exist
- Client-supplied tenant identity cannot override authenticated tenant identity
- Legacy company-name usage is not treated as the authorization boundary

### Database separation
- Control-plane data remains on the control database
- Tenant-owned operational data is provisioned and routed to the tenant database
- Request-scoped tenant DB access is used for tenant-owned operations
- Tenant DB connections are closed with the request lifecycle
- Public E2E bootstrap is not registered
- Smoke reset is protected by normal JWT authentication and Admin role and is hard-scoped to `e2e_smoke_tenant`
- Production CORS excludes localhost and Cloudflare branch previews
- Login attempts are rate-limited at the Nginx edge
- The Go API runs as the unprivileged `skillsifter` service account
- Runtime resume storage is isolated under `/var/lib/skillsifter/resumes`; the backend environment file is mode `0640`

## 3. Subscription acceptance

Implemented:

- runtime plan catalog
- external Razorpay checkout handoff
- signed provider webhook handling
- activation
- renewal
- payment failure / suspension state
- cancellation
- expiry
- reactivation/resumption
- account/subscription APIs
- Account & Subscription UI

Payment instruments, accounting, invoices, GST, banking and financial ledger remain external.

## 4. UAT checklist

The Phase 9 implementation and earlier production verification are complete; final post-audit verification remains pending. The final browser smoke suite previously passed 4/4 tests in 19.1 seconds. Final post-audit verification is required after the E2E endpoint security changes.

Before production launch, verify with a real deployment:

- [x] New tenant registration
- [x] Initial Admin login
- [x] Admin creates Manager / Recruiter / Team Leader
- [x] Seat limit blocks the next user
- [x] Admin account cannot be edited/deleted
- [x] Each role sees only its permitted operations
- [x] Cross-tenant resource access returns denial/404
- [x] Candidate create/read/update/delete isolation
- [x] Requirement and Client isolation
- [x] Recruitment lifecycle works end-to-end
- [x] Joined candidate appears in Billing
- [x] Billing cannot be created before Joined
- [x] Tenant database is provisioned successfully
- [x] Tenant data is written to the tenant database
- [x] Tenant DB routing survives application restart
- [x] Account page shows plan, status and seats
- [x] Checkout handoff reaches the external provider
- [x] Provider webhook is accepted only with valid signature
- [x] Activation/renewal/failure/cancellation/expiry state transitions are reflected in access
- [x] Suspended/expired tenant cannot access protected APIs
- [x] Admin can retry provisioning if provisioning fails
- [x] Backup/restore procedure verified
- [x] Production secrets/configuration verified
- [x] Database migrations verified
- [x] HTTPS/reverse proxy verified
- [x] Application logs and health endpoint verified

## 5. Go-live boundary

The earlier v1.0.0 release gate passed, but the final post-audit gate is pending. Production go-live must not be declared complete until the updated backend is deployed and the exact final-commit CI and production smoke runs pass.

No additional architecture is required for Phase 9.

Do not add:

- microservices
- custom permission builders
- arbitrary RBAC
- HRMS
- financial ledger
- accounting/GST module
- custom payment gateway
- banking/payment-instrument storage
- per-customer servers by default

The next work after v1.0.0 is post-release product work driven by real usage evidence, not another V1 architecture redesign.
