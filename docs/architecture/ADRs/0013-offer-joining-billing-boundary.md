# ADR 0013 — Offer, Joining, and Billing Boundary

**Status:** Accepted  
**Scope:** Core SkillSifter recruitment workflow  
**Runtime context:** Candidate × Requirement

## Decision

Keep the downstream workflow deliberately minimal:

```
Selection = selected
      ↓
Offer exists
      ↓
Offer.accepted = true
      ↓
Joining
      ↓
Joining.joined = true
      ↓
Billing eligible
```

All downstream records remain scoped to Candidate × Requirement. Candidate master state is never changed by Offer, Joining, or Billing.

### Offer

An Offer record means **offer made**.

Fields:
- tenant_id
- candidate_id
- requirement_id
- selection_id
- accepted (boolean)
- audit timestamps

Rules:
1. Offer requires Selection = `selected` for the same Candidate × Requirement.
2. One Offer is permitted per Candidate × Requirement.
3. `accepted = false` or `true`.
4. No Offer status state machine.

### Joining

Joining has only two business facts:

- joining_date
- joined (boolean)

Rules:
1. Joining requires an accepted Offer for the same Candidate × Requirement.
2. One Joining is permitted per Candidate × Requirement.
3. `joined = true` requires joining_date.
4. No Joining status state machine.

### Billing

Billing becomes eligible only when:

```
Offer.accepted = true
        ↓
Joining.joined = true
        ↓
Billing record
```

A Billing record means a recruitment-firm billing event has been raised for the joined Candidate × Requirement.

Authoritative fields:
- tenant_id
- candidate_id
- requirement_id
- client_id
- joining_id
- billing_date
- amount
- currency
- optional invoice_reference
- audit timestamps

Rules:
1. Billing requires a joined Joining record for the same Candidate × Requirement.
2. One Billing record is permitted per Candidate × Requirement.
3. Billing date is the Joining date.
4. Billing amount is supplied as a commercial billing value; Requirement budget and candidate compensation are not used automatically.
5. Billing does not track payment reconciliation or accounting state.

Billing remains recruitment-firm commercial functionality, not a general accounting, payroll, or HRMS system.

## Candidate boundary

```
Candidate A
  ├── Requirement 101 → Offer accepted → Joined → Billing
  └── Requirement 205 → Interviewing
```

No global Candidate field represents offer, acceptance, joining, or billing.

## Non-goals

This boundary does not introduce:
- employee records
- payroll
- attendance
- leave
- employee benefits
- internal approval hierarchies
- accounting ledgers
- GST/tax engines
- payment reconciliation
- Recruitment Assignment runtime dependency
