# SkillSifter — Current Implementation Status

**Baseline:** `main` at `ca19e6efac6f98c91f8f9873be20a0c5a03520ac`  
**Date:** 2026-09-28  
**Purpose:** authoritative implementation handoff after completion of the recruitment lifecycle and the first Phase 9 SaaS access work.

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

## 4. Implemented Phase 9 Foundation

### Control plane

Implemented:

- `platform_tenants`
- `platform_subscriptions`
- `platform_user_accounts`
- subscription user limit
- tenant/account status
- subscription status
- provisioning status
- plan code
- external provider reference fields

Existing tenants/users are backfilled into the control-plane bridge.

### Login

Implemented:

- real email/password login
- trusted tenant resolution
- subscription access check on every protected request
- trusted role resolution
- tenant identity in JWT
- current account access endpoint
- frontend display of tenant/company, role, plan and subscription context

### Registration

Implemented:

- public registration creates a new tenant
- first public user is always Admin
- caller cannot choose a role
- caller cannot join an existing tenant through public registration
- initial platform subscription is created as the compatibility/legacy subscription

### Initial RBAC enforcement

Implemented:

- Admin-only tenant user-management API
- fixed V1 roles:
  - admin
  - manager
  - recruiter
  - team_leader
- candidate mutation role restrictions
- trusted role/tenant context from the platform access layer

This is **not yet the complete V1 privilege matrix enforcement**.

## 5. Remaining Implementation

The remaining work is deliberately split into small blocks.

### A. Admin User Management UI

Backend user-management capability exists.

Still required:

- Admin Users screen
- list tenant users
- create additional user
- edit user
- delete user
- assign only Manager / Recruiter / Team Leader
- show subscription seat usage
- prevent Admin promotion
- prevent deletion/editing of the initial Admin
- frontend API integration
- negative authorization tests

No custom permission editor is required.

### B. V1 Privilege Matrix Enforcement

The approved four-role matrix must be enforced across every protected module.

The implementation must cover:

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
- Business Development
- Reports
- Resume/AI operations
- Daily Tasks
- Tenant/user administration
- Tenant configuration

The authorization convention should become consistent domain-action authorization rather than a growing collection of ad-hoc role comparisons.

Tenant isolation remains a separate security invariant.

### C. Tenant Isolation Audit

The application still contains historical shared-schema code.

Remaining work:

- identify all tenant-owned queries
- ensure authoritative tenant ID is used
- remove remaining security dependence on `company_name`
- verify every create/read/update/delete path
- add cross-tenant negative tests
- ensure role checks cannot bypass tenant isolation

### D. Tenant Database Provisioning — Phase 9B

Not yet implemented:

- deterministic tenant database identity
- provisioning abstraction
- database creation
- tenant migration execution
- idempotent provisioning
- provisioning failure handling
- activation only after successful provisioning

### E. Tenant Database Routing — Phase 9C

Not yet implemented:

- authenticated tenant → database resolution
- request-scoped tenant DB context
- replacement of the single global tenant DB connection
- safe connection lifecycle
- routing/isolation tests

This is the largest remaining architectural Phase 9 block.

### F. Subscription Lifecycle / Payment Provider — Phase 9E

Not yet implemented:

- plan catalog/runtime plan management
- external checkout handoff
- provider adapter
- webhook endpoint
- normalized subscription events
- activation
- renewal
- payment failure
- suspension
- cancellation
- expiry
- reactivation

SkillSifter must not become a payment gateway or financial ledger.

### G. Account / Subscription UI — Phase 9F

Not yet implemented:

- company account page
- current plan
- seat usage
- subscription status
- subscription dates
- checkout handoff
- access/suspension state
- administrator-only subscription controls

## 6. Documentation Work Remaining

The following documents must be kept synchronized with implementation:

1. Phase 9 implementation status
2. V1 privilege/action matrix
3. Tenant isolation and routing design
4. Subscription lifecycle and provider boundary
5. Admin user-management workflow
6. API contracts for account/user/subscription operations
7. Go-live/security acceptance checklist
8. Current architecture baseline

Older documents that describe company-name tenancy or the earlier authentication model are historical and must not be treated as the current runtime contract.

## 7. Explicit Non-Goals

The next implementation stages must not introduce:

- microservices
- custom permission builders
- arbitrary user roles
- HRMS functionality
- financial ledger
- GST/accounting module
- banking/payment-instrument storage
- custom payment gateway
- per-customer server by default

## 8. Recommended Implementation Order

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
