# SkillSifter v1.0.0 — Release Notes

**Release date:** 2026-09-30  
**Status:** Production release

SkillSifter v1.0.0 is the first production release of the multi-tenant recruitment intelligence platform. It consolidates the V1 recruitment workflow, SaaS account/subscription layer, tenant database isolation, fixed-role authorization, administration, production hardening and go-live verification.

## What is included

### Recruitment operations

- Clients and client contacts
- Requirements
- Candidate management and expertise
- Candidate × Requirement recruitment context
- Screening
- Submission
- Feedback
- Interviews
- Selection
- Offers
- Joining
- Joined-candidate billing worklist
- Daily recruitment tasks
- Recruitment history and auditability
- Operational reporting

### Resume and AI assistance

- Resume ingestion and extraction foundation
- Resume intelligence
- Recruiter-assisted candidate-to-requirement matching
- AI-assisted search/reporting
- Recruiter remains responsible for selection and submission decisions

### SaaS account and subscription

- Tenant registration
- Initial tenant Admin
- Plan catalog
- Checkout handoff
- Provider webhook processing
- Webhook signature validation
- Idempotent subscription events
- Activation
- Renewal
- Payment failure/suspension
- Cancellation
- Expiry
- Reactivation/resumption
- Server-side subscription access enforcement
- Account and Subscription UI
- Seat limits

### Tenant isolation

- Control-plane tenant/account/subscription state
- Dedicated PostgreSQL database per subscribed tenant
- Deterministic provisioning
- Tenant schema initialization
- READY gating
- Authenticated tenant DB routing
- Request-scoped tenant DB access
- Tenant ownership defense-in-depth
- Cross-tenant negative tests
- Tenant isolation independent of frontend visibility

### Authorization and administration

Fixed V1 roles:

- Admin
- Manager
- Recruiter
- Team Leader

Implemented:

- Server-side V1 privilege matrix
- Admin User Management
- User creation/update/permitted deletion
- Tenant-admin protection
- Server-side seat enforcement

## Security and production hardening

- Production JWT configuration fails closed when `JWT_SECRET` is missing.
- Production runtime is explicitly marked as production.
- Tenant DB routing is enforced on the V1 API.
- Deployment tests use persistent production database credentials before service restart.
- Obsolete duplicate deployment migration execution was removed.
- Production schema definition 039 was restored to match the applied production schema.
- Nginx, systemd, database connectivity and API health are checked during deployment.

## Release verification

### Production deployment

Verified on commit `bce3d5dac7a078a9af83c65456a0cc0a17279cf8`:

- Deployment test gate passed.
- Backend built successfully.
- Nginx configuration passed.
- systemd service active.
- PostgreSQL connection successful.
- API health endpoint returned `{"status":"OK"}`.

### Production smoke

Final Playwright production smoke:

```text
Running 4 tests using 1 worker

✓ public login and registration pages load
✓ authenticated application smoke across Phase 9 modules
✓ subscription/account surfaces load without browser errors
✓ mutating production smoke uses the permanent account and clears only its test data

4 passed (19.1s)
```

## Recruitment boundary

The V1 lifecycle is:

```text
Requirement
 → Candidate × Requirement
 → Screening
 → Submission
 → Feedback
 → Interview
 → Selection
 → Offer
 → Joining
 → JOINED
 → BILLING
```

Offer is **made + accepted**.  
Joining is **joining date + joined**.

## Explicit non-goals

v1.0.0 does not introduce:

- HRMS
- microservices
- arbitrary RBAC or permission builders
- accounting or financial ledger
- GST/accounting module
- banking/payment-instrument storage
- custom payment gateway
- autonomous recruiting
- automated external sourcing/scraping
- WhatsApp automation
- candidate/client portals
- complex vector-database architecture
- advanced autonomous analytics

## Architecture status

The V1 architecture is frozen at release.

Post-v1.0.0 work is new product work and should be driven by real production usage, not by reopening the V1 architecture without evidence.
