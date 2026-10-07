# SkillSifter — Control-Plane and Tenant-Plane Audit Findings

## Control-plane findings

### CP-01 — Tenant root
`platform_tenants` is the correct SaaS customer root. It should remain in the control plane with `tenant_id` as the canonical identity.

Retain conceptually:
- tenant_id
- company_name as descriptive data
- account_status
- provisioning_status
- tenant_database
- created/updated timestamps
- trial lifecycle timestamps

Review: account/provisioning status CHECKs should remain only if treated as physical value protection; workflow transition rules belong in Go.

### CP-02 — Platform subscription
`platform_subscriptions` is application-owned lifecycle state.

Remove historical compatibility behavior:
- legacy plan
- backfill from companies
- migration-time user-count backfill
- compatibility assumptions.

The application should explicitly update subscription and tenant access state.

### CP-03 — Platform account index
`platform_user_accounts` is a control-plane routing/index record.

The relationship between the platform account and tenant/user is important, but it must not silently determine application behavior through database cascades.

Application code should explicitly manage creation, synchronization, and deletion.

### CP-04 — Registration identity
`platform_registration_registry` is valid final architecture.

Keep it as permanent control-plane state.

Concurrency requirement:
the email claim must be atomic. A pre-check followed by an ordinary INSERT is not sufficient by itself under concurrent requests.

Preferred pattern:
Go attempts an atomic claim; PostgreSQL returns claimed/conflict; Go translates that result to the registration outcome.

### CP-05 — Pending registrations and verification
Pending registration and verification-code tables are operational state.

The current flow is largely correct because the application explicitly orchestrates:
pending -> OTP -> verified -> tenant -> user -> subscription -> platform account -> permanent registry -> provisioning.

However, concurrent verification/replay and duplicate registration races require atomic result handling rather than only pre-checks.

### CP-06 — Subscription provider events
Provider event idempotency is a concurrency requirement, not domain logic owned by PostgreSQL.

PostgreSQL may atomically reject a duplicate provider event.

Go must receive and interpret the duplicate result and decide that the webhook is already processed.

### CP-07 — Tenant deletion
Current cleanup depends on `ON DELETE CASCADE` after deleting `platform_tenants`.

Final design:
Go explicitly executes the control-plane cleanup sequence and treats each database result as part of the operation.

### CP-08 — Migration architecture
The current migration runner is the largest architectural defect in the control plane.

It combines two physically different databases through a historical numeric cutoff.

Final design:
- explicit control-plane schema directory
- explicit tenant schema directory
- independent schema versioning
- no historical numeric cutoff
- no runtime schema cleanup
- no dynamic removal of legacy objects.

## Tenant-plane findings

### TP-01 — Tenant isolation
Because each tenant has its own database, tenant_id is not needed as a database-level cross-tenant security boundary inside that database.

Application routing selects the correct tenant database before domain operations.

Therefore tenant_id fields should be retained only where they are useful domain metadata or operational identity—not merely to support a shared-database foreign-key architecture.

This requires a deliberate field-by-field review before removing them.

### TP-02 — Cross-entity foreign keys
The current tenant schema has relationships such as:
- assignment -> candidate
- assignment -> requirement
- assignment -> user
- screening -> assignment/user
- submission -> assignment/user
- feedback -> submission/user
- selection -> assignment
- offer -> candidate/requirement/selection
- joining -> candidate/requirement/offer
- billing -> candidate/requirement/client/joining.

These relationships currently mix two concerns:
1. record reference/storage
2. domain validity/workflow enforcement.

The second concern belongs in Go.

The first may be represented by IDs without a PostgreSQL FK if the application owns relationship validation.

### TP-03 — Workflow state
Offer, Joining and Billing services already implement workflow meaning in Go.

Therefore the final schema should not duplicate those decisions through DB triggers or workflow constraints.

Example:
`joined=true` requiring a date is currently represented as a database CHECK.

Final audit action: decide whether this is merely physical data validity or a domain rule. Under the agreed principle, the application should validate it and the DB should not silently reject it without an explicit application-visible result.

### TP-04 — Candidate engagement concurrency
The historical trigger implementation is explicitly obsolete.

Final architecture should use an application transaction with an atomic PostgreSQL operation, for example an UPDATE guarded by the expected current value.

Go interprets row-count/result:
- one row affected -> candidate claimed
- zero rows -> candidate was already claimed / state changed concurrently.

This is the model we want.

### TP-05 — Audit events
Audit events are application-generated domain events.

They must be written by Go as part of the transaction that performs the business operation.

No generic row-level PostgreSQL activity trigger should generate them.

### TP-06 — Legacy activity logging
The old `skillsifter_activity_trigger()` captures row JSON and uses company-based tenancy.

This directly violates the final architecture.

Remove it completely from the greenfield baseline.

## Critical race audit

Patterns requiring special attention:

1. registration email claim
2. pending registration creation
3. registration verification replay
4. candidate recruitment engagement claim
5. assignment uniqueness
6. offer creation
7. joining creation
8. billing creation
9. phone OTP consumption
10. payment webhook idempotency
11. subscription lifecycle transitions
12. tenant deletion/provisioning recovery.

For every item, the implementation must answer:

`What happens if two requests arrive at exactly the same time?`

The answer must include an explicit PostgreSQL result returned to Go.

## Current conclusion

The application architecture is already moving in the correct direction in several places. The remaining problem is that the database still carries the accumulated history of earlier architectural decisions.

The clean implementation should therefore be a **schema reconstruction**, not a migration cleanup exercise.

Next implementation artifact: final control-plane schema specification followed by final tenant-plane schema specification.