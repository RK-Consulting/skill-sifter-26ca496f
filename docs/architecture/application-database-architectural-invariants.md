# SkillSifter — Application/Database Architectural Invariants

## Purpose

This document is the governing engineering contract for the greenfield SkillSifter platform reset.

The project is not live. Existing manually created database data may be discarded. Historical migrations are implementation history, not a compatibility contract.

The objective is a durable system boundary in which Go owns domain meaning and PostgreSQL provides reliable persistence and atomic concurrency primitives without silently becoming a second domain-logic engine.

## 1. Domain authority

The Go application is the authoritative owner of:
- business rules
- workflow transitions
- authorization decisions
- lifecycle decisions
- validation meaning
- conflict/error interpretation
- tenant/account/subscription state transitions
- orchestration across control-plane and tenant-plane data

A PostgreSQL feature must not silently implement business behavior that the application cannot observe.

## 2. PostgreSQL responsibility

PostgreSQL is responsible for:
- durable persistence
- transactions
- atomic writes
- isolation
- concurrency arbitration
- efficient indexing/query execution

Database constraints may protect physical data integrity where appropriate, but a consequential database decision must be observable by the application.

## 3. Database decision -> application decision loop

Whenever PostgreSQL arbitrates a race or conflict, the result must return explicitly to Go.

Conceptually:

    Go command
       |
       v
    PostgreSQL atomic operation
       |
       v
    explicit result
       |
       v
    Go interprets result
       |
       +---- continue -> next operation
       |
       +---- stop -> domain result/error -> client

Examples:
- INSERT ... ON CONFLICT DO NOTHING RETURNING ...
- UPDATE ... WHERE state = expected_state RETURNING ...
- row-count == 0 interpreted by Go as a state/concurrency conflict
- PostgreSQL unique/exclusion/index arbitration translated into a domain error where required

Generic PostgreSQL error text must not become the public application contract.

## 4. No hidden domain logic in PostgreSQL

The greenfield schema must not rely on:
- business triggers
- hidden state-transition triggers
- compatibility triggers
- legacy Jobs behavior
- automatic cross-domain mutation
- database-side workflow engines
- implicit cascade behavior as the application's lifecycle decision

If a state changes because of a business rule, Go must perform or explicitly interpret that transition.

## 5. Foreign-key policy

Foreign keys are not a substitute for application domain logic.

For the greenfield reset:
- remove legacy cross-boundary FKs
- remove companies-based relationships
- remove historical compatibility FKs
- do not make the database silently enforce tenant/domain relationships that are already owned by application logic
- retain a relationship in PostgreSQL only when it is deliberately classified as structural persistence integrity rather than hidden domain behavior
- every retained relationship must be documented and tested

Tenant isolation is primarily an application/routing invariant because tenant databases are physically separated.

## 6. Tenant identity

tenant_id is the single canonical SkillSifter customer identity.
- platform_tenants.tenant_id is the control-plane customer identity.
- Company name is descriptive data, not identity.
- companies is not part of the final model.
- Tenant data is disposable.
- Registration identity is not disposable.

## 7. Permanent registration identity

A successfully verified registration permanently reserves its canonical email address in the control plane.

The permanent registry:
- survives trial expiry
- survives tenant database deletion
- survives tenant/control-plane customer cleanup
- contains only the minimum registration identity information
- does not become a user/account/subscription store

Application code owns the interpretation of registration conflicts.

## 8. Control plane vs tenant plane

The control plane owns:
- tenant identity
- permanent registration identity
- platform authentication/account index
- subscription state
- plan definitions
- payment-provider opaque references
- provisioning state
- trial lifecycle metadata

The tenant database owns recruitment-domain data.

There must be no accidental dependence on the old shared companies model.

## 9. Tenant lifecycle

    registration
       -> verification
       -> tenant creation
       -> provisioning
       -> READY
       -> trial/subscription lifecycle
       -> expiry/cancellation
       -> data deletion

The application owns this lifecycle. Tenant deletion must not depend on database cascades to decide what the product means by deletion.

## 10. Migration architecture

This is a greenfield reset.

The final implementation must not carry forward:
- migrations 001–045 as a historical execution chain
- compatibility bridges
- backfills
- legacy Jobs migrations
- companies cleanup migrations
- temporary smoke-account seed migrations
- obsolete subscription compatibility migrations

The schema should be rebuilt from clean final baselines with an explicit control-plane/tenant-plane migration boundary.

The migration runner must not contain historical version assumptions such as 'versions <= 34 are tenant schema'.

## 11. Seed data

Schema definition and environment/bootstrap data are separate concerns.

Production customer records, smoke accounts, temporary test tenants, and manually fed data must never be embedded in the authoritative production schema baseline.

Commercial plan definitions may be provisioned explicitly as controlled bootstrap/configuration data.

## 12. State and concurrency

Concurrency safety is required.

The preferred pattern is:
1. Go determines the intended state transition.
2. PostgreSQL performs an atomic operation.
3. PostgreSQL returns success/conflict/current result.
4. Go interprets that result.
5. Go either continues the workflow or returns the domain outcome.

This preserves both correctness under concurrency and application ownership of meaning.

## 13. Observability and diagnosability

Every consequential transition must be diagnosable from application-level logs/results.

A production operator should be able to determine:
- what operation was attempted
- which tenant was involved
- what PostgreSQL atomically decided
- how Go interpreted that result
- what next state was selected
- whether the operation completed or stopped

No important state transition should exist only inside PostgreSQL.

## 14. Testability

Tests must verify the application/database boundary, including:
- concurrent conflicting operations
- duplicate registration attempts
- trial expiry and deletion
- provisioning failure/recovery
- subscription lifecycle transitions
- tenant isolation
- database arbitration results returned to Go
- absence of hidden triggers and legacy schema objects
- clean installation from empty control and tenant databases

## 15. Security

Security-sensitive decisions remain in application code:
- authentication
- authorization
- tenant routing
- subscription access
- provisioning access
- OTP verification
- payment-provider webhook verification

PostgreSQL may protect storage primitives, but security policy must remain observable and testable in Go.

## 16. Greenfield acceptance rule

A schema/code change is not accepted merely because tests pass.

It must also satisfy:

> Can a future engineer understand the complete domain behavior by reading the Go application and the final schema, without having to reconstruct historical migration intent or discover hidden PostgreSQL business behavior?

If not, the design is not finished.

## 17. Engineering principle

**PostgreSQL may arbitrate concurrency, but Go owns the meaning of the result.**

**PostgreSQL stores state; Go decides what that state means.**

**No silent second brain.**