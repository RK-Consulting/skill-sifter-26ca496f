# SkillSifter — Current Implementation Status

**Baseline:** v1.0.0 production release
**Date:** 2026-09-30

## 1. PR / Merge Status

All implementation PRs in the current workstream are complete.

| PR | Scope | Status |
|---|---|---|
| #80 | Phase 9 bridge tenancy/subscription/login architecture | Merged |
| #81 | Control-plane subscription access in login/RBAC | Merged |
| #82 | Recruitment lifecycle API integration | Merged |
| #83 | Superseded duplicate lifecycle UI PR | Closed — not merged |
| #84 | Recruitment lifecycle UI integration | Merged |
| #85 | Interviews aligned with Candidate × Requirement lifecycle | Merged |
| #86 | Billing worklist | Merged |
| #87 | Login, RBAC and subscription access hardening | Merged |

PR #83 must remain closed for historical reference; its work was superseded by #84.

## 2. Implemented Recruitment Product

The operational recruitment lifecycle is implemented through:

```
Requirement
    ↓
Candidate × Requirement
    ↓
Screening
    ↓
Submission
    ↓
Feedback
    ↓
Interview
    ↓
Selection
    ↓
Offer
    ↓
Joining
    ↓
Joined
    ↓
Billing
```

The Candidate × Requirement relationship is the durable recruitment context.

### Offer / Joining boundary

Offer remains intentionally simple:

- offer exists = offer made
- `accepted = true` = offer accepted

Joining remains intentionally simple:

- `joining_date` = scheduled joining date
- `joined = true` = candidate joined

Billing is permitted only after joined.

No HRMS workflow, internal interviewer policy, invoice engine, GST/accounting system, payment ledger, or banking workflow is part of this lifecycle.

## 3. Implemented Billing

PR #86 added:

- tenant-scoped billing worklist
- every joined Candidate × Requirement visible in one worklist
- billed/pending state
- search
- direct billing creation for pending joined records
- lifecycle navigation

SkillSifter billing remains an operational recruitment record. Formal invoicing, accounting, GST/tax, banking and client payment collection remain outside SkillSifter.

## 4. Implemented Phase 9 Platform

### Control plane and access

Implemented:

- `platform_tenants`
- `platform_subscriptions`
- `platform_user_accounts`
- subscription user limits
- tenant/account/subscription/provisioning status
- trusted tenant and role resolution
- protected-request subscription access checks
- public tenant registration with initial Admin
- Admin User Management UI and API
- fixed V1 roles

### V1 privilege enforcement

Implemented across the protected application surface:

- Clients
- Requirements
- Candidates
- Screening
- Submissions
- Feedback
- Interviews
- Selection
- Offers
- Joining
- Billing
- Reports
- Resume/AI
- tenant/user administration

Separate Business Development and Daily Tasks tables/modules were retired from the final tenant schema; partner/contact details are represented on Client records.

No permission builder or arbitrary role model was introduced.

### Tenant isolation

Implemented:

- authenticated tenant ID as the security boundary
- tenant-scoped create/read/update/delete paths
- cross-tenant known-ID negative tests
- client-supplied tenant identity cannot override authenticated tenant identity
- reporting and tenant isolation checks
- tenant DB routing with request-scoped access

### Tenant DB provisioning

Implemented:

- deterministic tenant database identity
- create-or-resume provisioning
- tenant schema migration
- idempotency
- provisioning status
- activation only after successful provisioning
- registration-time provisioning
- administrator retry path

### Tenant DB routing

Implemented:

- authenticated tenant → tenant DB resolution
- request-scoped tenant DB
- tenant-owned handler routing
- control-plane queries remain on the control DB
- request DB lifecycle cleanup

### Subscription lifecycle

Implemented:

- runtime plan catalog
- external Razorpay checkout handoff
- signed webhook processing
- activation
- renewal
- payment failure/suspension
- cancellation
- expiry
- reactivation/resumption
- account/subscription APIs

SkillSifter does not process payment instruments or maintain a financial ledger.

### Account and Subscription UI

Implemented:

- account details
- current plan
- subscription status
- seat usage
- available plans
- checkout handoff
- administrator-only cancellation

## 5. Phase 9 Security / UAT / Go-Live

The Phase 9 implementation and production go-live gate are complete for v1.0.0.

Production verification included the deployment gate, service health, PostgreSQL connectivity, API health and final Playwright production smoke. The final smoke run passed all four tests in 19.1 seconds.

The acceptance checklist is maintained in:

`docs/phase-9-security-uat-go-live.md`

Required production verification includes authentication, fixed-role authorization, cross-tenant isolation, tenant DB provisioning/routing, lifecycle completion, subscription state transitions, webhook signature validation, migration/backup verification, production configuration, HTTPS and health monitoring.

## 6. v1.0.0 Release Evidence

- Production deployment succeeded on the final security-fixed main commit.
- API health returned `{"status":"OK"}`.
- Final production smoke: **4 passed (19.1s)**.
- Frontend release version is `1.0.0` and is displayed in the application footer.
- Release documentation and architecture are frozen with the v1.0.0 release.

## 7. Documentation Work Remaining

Documentation was reconciled during the final code audit. Historical ADRs and archived design notes remain historical records; the final control-plane and tenant-plane baselines are authoritative.

Older documents that describe company-name tenancy or the earlier authentication model are historical and must not be treated as the current runtime contract.

## 8. Explicit Non-Goals

The Phase 9 implementation does not introduce:

- microservices
- custom permission builders
- arbitrary user roles
- HRMS functionality
- financial ledger
- GST/accounting module
- banking/payment-instrument storage
- custom payment gateway
- per-customer servers by default

The recruitment lifecycle remains frozen.

## 9. Recommended Implementation Order

```
Documentation baseline
        ↓
Admin User Management UI
        ↓
V1 Privilege Matrix enforcement
        ↓
Tenant Isolation audit/tests
        ↓
Tenant DB Provisioning
        ↓
Tenant DB Routing
        ↓
Subscription lifecycle/provider
        ↓
Account + Subscription UI
        ↓
Phase 9 security/UAT/go-live
```

The recruitment lifecycle should remain frozen while these platform concerns are completed.
