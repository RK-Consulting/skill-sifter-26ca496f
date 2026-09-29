# SkillSifter — Phase 9 Security, UAT & Go-Live Gate

**Status:** Implementation complete; production UAT gate remains operational  
**Date:** 2026-09-29

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

Before production launch, verify with a real deployment:

- [ ] New tenant registration
- [ ] Initial Admin login
- [ ] Admin creates Manager / Recruiter / Team Leader
- [ ] Seat limit blocks the next user
- [ ] Admin account cannot be edited/deleted
- [ ] Each role sees only its permitted operations
- [ ] Cross-tenant resource access returns denial/404
- [ ] Candidate create/read/update/delete isolation
- [ ] Requirement and Client isolation
- [ ] Recruitment lifecycle works end-to-end
- [ ] Joined candidate appears in Billing
- [ ] Billing cannot be created before Joined
- [ ] Tenant database is provisioned successfully
- [ ] Tenant data is written to the tenant database
- [ ] Tenant DB routing survives application restart
- [ ] Account page shows plan, status and seats
- [ ] Checkout handoff reaches the external provider
- [ ] Provider webhook is accepted only with valid signature
- [ ] Activation/renewal/failure/cancellation/expiry state transitions are reflected in access
- [ ] Suspended/expired tenant cannot access protected APIs
- [ ] Admin can retry provisioning if provisioning fails
- [ ] Backup/restore procedure verified
- [ ] Production secrets/configuration verified
- [ ] Database migrations verified
- [ ] HTTPS/reverse proxy verified
- [ ] Application logs and health endpoint verified

## 5. Go-live boundary

SkillSifter is ready to enter deployment/UAT once the deployment-specific checklist above is verified.

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

The next work after UAT is deployment hardening and production rollout, not another architecture redesign.
