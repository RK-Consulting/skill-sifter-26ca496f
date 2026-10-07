# SkillSifter — Final Control-Plane Architecture

## 1. Purpose
The SkillSifter control plane is the authoritative platform-level database.

Its responsibility is limited to:
1. platform registration identity
2. tenant identity
3. tenant database routing
4. platform account access/routing
5. subscription and plan state
6. checkout state
7. payment-provider webhook idempotency
8. registration and verification workflow state
9. tenant provisioning/deprovisioning state

It is NOT the recruitment database.
The control plane must not contain candidates, requirements, assignments, screenings, submissions, interviews, selections, offers, joining records, billing records, or recruitment audit history.

## 2. Final control-plane tables

### platform_tenants
The single SaaS customer root.
Fields: tenant_id, company_name, account_status, provisioning_status, tenant_database, created_at, updated_at, trial_started_at, trial_expires_at, data_deletion_at.
Allowed account_status: ACTIVE, SUSPENDED, EXPIRED, TERMINATED.
Allowed provisioning_status: PENDING, PROVISIONING, READY, FAILED.
Status values are physical data protection only; lifecycle transitions belong to Go.

### platform_plans
Commercial plan catalogue. Fields: code, name, amount_minor, currency, billing_period, billing_interval, user_limit, total_count, provider, provider_plan_ref, active, created_at, updated_at.
SkillSifter stores plan metadata required to operate subscriptions, but never payment instruments, card data, bank data, or financial transaction details.

### platform_subscriptions
Authoritative platform subscription state. Fields: id, tenant_id, plan_code, status, starts_at, ends_at, provider, provider_subscription_ref, user_limit, created_at, updated_at.
Allowed status: TRIAL, ACTIVE, PAST_DUE, SUSPENDED, CANCELLED, EXPIRED.
Go owns lifecycle transitions. PostgreSQL protects physical validity and uniqueness.

### platform_subscription_checkouts
Tracks checkout handoff, not payment processing. Fields: id, tenant_id, plan_code, provider, provider_subscription_ref, checkout_url, status, created_at, updated_at.
Allowed status: PENDING, COMPLETED, FAILED, CANCELLED.
The payment provider remains the system of record for payment execution.

### platform_subscription_events
Permanent provider-webhook idempotency record. Fields: id, provider, provider_event_ref, tenant_id, provider_subscription_ref, event_type, received_at.
Unique key: (provider, provider_event_ref).
The database performs the atomic claim; Go interprets new versus duplicate.

### platform_registration_registry
Permanent registration identity. Fields: email_id, first_registered, last_tenant_id.
This survives trial deletion and tenant deletion. It answers: has this email ever completed SkillSifter registration?
Registration must use an atomic claim; pre-check followed by INSERT is insufficient.

### platform_pending_registrations
Temporary signup state. Fields: id, username, email, company_name, password_hash, plan_code, created_at, expires_at, email_verified_at.
This exists only while registration is incomplete.

### platform_verification_codes
Temporary verification state. Fields: id, purpose, registration_id, user_id, destination, code_hash, attempts, expires_at, consumed_at, created_at.
Purposes: EMAIL_SIGNUP and PHONE_SUBSCRIPTION.
Consumption must be atomic and Go interprets affected-row count.

### platform_user_accounts
Final platform account routing/access record. Fields: user_id, tenant_id, email, role, created_at, updated_at.
Final architecture: this table must not have a foreign key into a tenant database.
The current FK to users is a migration-era compatibility relationship and must disappear after physical separation.
The control plane knows email/user identity -> tenant_id -> tenant_database. The tenant database owns the detailed user record.

## 3. Explicitly excluded
Tenant databases alone own: users as tenant-domain records, candidates, candidate expertise, clients, requirements, assignments, screening, submissions, feedback, interviews, selections, offers, joining, billing, recruitment audit/history, recruitment activity, AI candidate intelligence, resume data, and operational recruitment reports.

## 4. No companies table
companies is not a final domain entity. The final customer identity is platform_tenants.tenant_id. Company name is descriptive tenant data, never a tenant identity key.

## 5. Foreign-key policy
Control-plane FKs are allowed only between control-plane entities. Never create control-plane FKs into tenant databases. Cross-database integrity is an application responsibility.

## 6. Cascade policy
Do not use database cascades as the tenant lifecycle engine.
Tenant deletion is an explicit Go operation: validate termination -> disable access -> stop provisioning -> delete/retire tenant database -> remove control-plane operational records -> retain permanent registration identity -> record final outcome.

## 7. Concurrency rules
Every control-plane race must have an atomic database operation.
Registration email claim; pending registration; verification; phone OTP; subscription transition; checkout/provider reference; webhook idempotency; provisioning; and deletion/recovery must all return explicit PostgreSQL results to Go.

Universal pattern: Go requests transition -> PostgreSQL atomic operation -> success or conflict/no-op -> Go interprets the result.

## 8. Constraints versus domain rules
PostgreSQL protects primary keys, uniqueness, non-null physical requirements, basic numeric/date validity, provider-event uniqueness, and registration identity uniqueness.
Go owns lifecycle transitions, authorization, subscription eligibility, trial rules, provisioning, account activation/suspension, OTP workflow, payment-provider interpretation, tenant deletion, and recruitment workflow.

## 9. Triggers
Final control plane: no generic business-logic triggers.
Do not use triggers for audit generation, tenant routing, subscription transitions, registration workflow, recruitment workflow, activity logging, or authorization.

## 10. Audit
Important control-plane operations should perform the business state change and audit event in the same logical transaction. Audit events are generated by Go, not generic PostgreSQL row-level triggers.

## 11. Migration architecture
The final deployment has two independent schema roots:
database/control-plane/
database/tenant-plane/

Each has independent schema versioning.
There is no numeric cutoff separating control and tenant migrations, no runtime DROP of historical objects, no dynamic legacy cleanup, and no shared migration history.
A new tenant receives the current tenant schema baseline. The control database receives the current control-plane baseline.

## 12. Final invariant
> The control plane knows who the customer is, where the customer's database is, whether the customer is allowed to use the platform, and what platform subscription state applies. It does not know the customer's recruitment data.

## 13. Acceptance criteria
- platform_tenants is the sole tenant root
- companies is absent from the final control schema
- no control table FK references tenant-domain tables
- subscription lifecycle is application-owned
- provider events are idempotent
- registration identity is permanent and atomically claimed
- verification consumption is atomic
- provisioning transitions are atomic
- tenant deletion is application-orchestrated
- no generic activity trigger remains
- no business workflow trigger remains
- no recruitment tables exist in the control database
- control and tenant migrations are independently versioned
- tenant database routing is explicit
- concurrency conflicts are returned to Go and handled explicitly

## 14. Decision
This is the control-plane architecture to freeze.
Do not add another abstraction layer unless a real production requirement demonstrates that one is necessary.
The next engineering step is to construct the final control-plane schema baseline, then independently construct the final tenant-plane schema baseline, and audit the Go code against those two contracts.