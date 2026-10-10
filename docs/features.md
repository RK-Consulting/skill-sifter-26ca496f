# SkillSifter — Feature Specification & Implementation Status

Status key:
- ✅ Implemented and released in v1.0.0.
- ⛔ Out of scope — deliberately excluded.

## 1. Product boundary

SkillSifter is a recruitment intelligence platform, not an HRMS.

Workflow: Requirement → Candidate × Requirement → Screening → Submission → Feedback → Interview → Selection → Offer → Joining → JOINED → Billing.

Out of scope: HRMS employee lifecycle, interviewer panels, hiring-manager hierarchy, client interview panels, internal approvals, internal evaluation forms, payroll/attendance/leave, accounting ledger, microservices and configurable permission-builder frameworks.

## 2. Recruitment features

| Feature | Status |
|---|---|
| Candidates | ✅ |
| Candidate expertise | ✅ |
| Clients | ✅ |
| Requirements | ✅ |
| Candidate × Requirement assignment context | ✅ |
| Screening | ✅ |
| Submission | ✅ |
| Feedback | ✅ |
| Interviews | ✅ |
| Selection | ✅ |
| Offer | ✅ |
| Joining | ✅ |
| Billing | ✅ |
| Recruitment history/audit | ✅ |
| Resume intelligence foundation | ✅ |

Separate Daily Jobs and Business Development tables/modules are not part of the final tenant schema. Partner/contact details belong to Client records.

## 3. Requirements

Requirement fields: Client, Job Type, Job Title, Department, Experience Required, Budget, Language Requirements, Certifications Required, Notice Period, Mode of Work, Mandatory Requirements, Job Description, Status, Job Location and Number of Open Positions.

## 4. Offer and Joining

Offer = made + accepted.

Joining = joining date + joined.

## 5. Authentication and RBAC

Roles: Admin, Manager, Recruiter, Team Leader.

Server-side capabilities include JWT authentication, protected routes, role authorization, tenant-aware identity, admin user management, seat-limit enforcement and admin account protection.

## 6. Multi-tenancy

Control plane owns authentication identity, tenant registry/provisioning state and subscription/account state.

Tenant DB owns recruitment and operational data.

Phase 9 includes tenant DB provisioning, migrations, routing, READY gating, tenant ownership checks and cross-tenant negative tests.

## 7. Admin User Management

Tenant administrators can list, create, update and delete permitted users, while seeing seat usage/limits. Tenant-admin protection and server-side authorization are enforced.

## 8. Subscription lifecycle

Plans → Checkout → Provider Webhook → Activation → Renewal → Payment Failure / Cancellation / Expiry.

Implemented: plans, checkout handoff, webhook processing, activation, renewal, payment failure, cancellation, expiry, idempotency, server-side enforcement, Account & Subscription UI.

No accounting ledger is part of SkillSifter.

## 9. Security

Phase 9 includes authenticated tenant identity, server-side RBAC, tenant DB routing, ownership checks, cross-tenant rejection, protected admin operations, subscription enforcement and webhook validation.

## 10. CI and quality

Backend: Formatting → Build → Vet → Test with PostgreSQL.

Frontend: Install → Lint → Test → Build.

Release: CI → Security → UAT → Production Smoke → GO-LIVE.

## 11. Docker

No Phase 9 Docker architecture expansion is required. PostgreSQL remains independent infrastructure; backend and frontend remain containerized; Docker Compose remains the local development/verification stack.

## 12. Release verification

Implementation is consolidated on main and released as v1.0.0.

- 🧪 CI evidence on release commit
- 🧪 Security negative-test execution
- 🧪 Production UAT
- 🧪 Production smoke test

Only after these gates pass should the project be declared GO-LIVE.

## 13. Post-go-live

Freeze the architecture after Phase 9. New functionality becomes post-go-live product work and should be driven by real usage evidence.

See docs/phase-9-release-gate.md and GitHub Issue #99.