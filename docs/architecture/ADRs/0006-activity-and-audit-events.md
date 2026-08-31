# ADR 0006: Activity and Audit Events

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** V04-08, Issue #35

## Context

Operational reporting and historical traceability require authoritative events while protecting candidate data and avoiding duplicate event logic.

The existing `activity_logs` trigger mechanism records complete row JSON for several legacy domains and is scoped by `company_name`. It predates the tenant-isolation architecture and may contain sensitive candidate information. New transactional domains must not blindly extend that mechanism.

## Decision

### 1. Audit events are a first-class logical event model

New audit coverage uses an append-only audit event model with the following conceptual fields:

```text
id
tenant_id
actor_user_id
entity_type
entity_id
action
occurred_at
correlation_id
metadata
```

The exact physical table/API implementation is part of the relevant implementation issue, but every new event must carry authoritative tenant identity and actor information.

### 2. Tenant ownership

Every new audit event is tenant-scoped.

`tenant_id` is derived from authenticated request context and is never accepted as an authorization value from request payloads or query parameters.

Audit queries must be tenant-scoped and must fail closed across tenants.

### 3. No full-row snapshots by default

New audit events must not automatically serialize entire entity rows with `to_jsonb(NEW)`.

Event metadata contains only information required to understand the business event.

For example:

```json
{
  "from": "submitted",
  "to": "interviewing"
}
```

is preferred over embedding an entire candidate or requirement record.

Candidate PII, contact details, resumes, and other sensitive data must not be copied into audit metadata unless explicitly required by an approved event definition and protected by the applicable redaction rules.

### 4. Recruitment Assignment event coverage

Recruitment Assignment lifecycle events defined by ADR 0003 are auditable, including:

```text
assignment.created
assignment.screening_started
assignment.submitted
assignment.interviewing_started
assignment.offered
assignment.joined
assignment.rejected
assignment.withdrawn
assignment.owner_changed
assignment.snapshot_created
```

The implementation may add additional events when required by an approved workflow, but must not silently broaden audit scope to unrelated domains.

### 5. Audit events are immutable

Audit events are append-only historical records.

Normal application operations must not update or delete existing audit events.

Corrections, if ever required, must be represented by a new event rather than rewriting historical records.

### 6. Existing `activity_logs` remains legacy

The existing `activity_logs` mechanism is not replaced by Issue #35 and is not retroactively redefined by this ADR.

Its existing company-name scoping and full-row metadata behavior remain legacy behavior until a separate migration/modernization decision is approved.

New audit event implementation must not depend on extending that legacy mechanism merely for convenience.

### 7. Retention

Audit events are retained for the lifetime of the associated business transaction unless a later platform-level compliance or retention policy explicitly requires archival or deletion.

Retention policy is a platform concern and must not be hard-coded into the Recruitment Assignment domain.

### 8. Access control

Audit events are not exposed through unrestricted generic endpoints. Access is role-controlled and tenant-scoped.

## Consequences

### Positive

- Provides an authoritative event model for new transactional workflows.
- Prevents accidental propagation of candidate PII through full-row audit serialization.
- Makes tenant ownership explicit.
- Preserves audit immutability.
- Allows legacy activity reporting to coexist with the new audit model without forcing an unrelated migration into Issue #35.

### Trade-offs

- Two event mechanisms exist temporarily: legacy `activity_logs` and the new audit-event model.
- A future migration/consolidation effort will be required if the platform wants one unified event store.
- Event definitions must be deliberately maintained as workflows evolve.

## Implementation Boundary

This ADR does not authorize migration of legacy `activity_logs`, changes to existing trigger behavior, or automatic backfill of historical activity records. Those require a separate architectural and migration decision.
