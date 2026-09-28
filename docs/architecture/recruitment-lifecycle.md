# SkillSifter Recruitment Lifecycle

**Status:** Frozen architectural/product baseline  
**Related ADRs:** ADR 0010, ADR 0011, ADR 0012, ADR 0013  
**Runtime model:** Candidate × Requirement

## Product definition

SkillSifter is a **recruitment intelligence platform** for recruitment/staffing firms and their client-driven hiring requirements. It is not an internal corporate HRMS.

## Canonical lifecycle

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
Offer accepted
     ↓
Joining date
     ↓
Joined
     ↓
Billing
```

The Candidate × Requirement pair is the authoritative runtime recruitment context. Historical Recruitment Assignment data may remain in the database for upgrade compatibility, but Assignment is not a runtime dependency.

## Selection

Selection is a Candidate × Requirement decision:
- `selected`
- `rejected`

Selection requires a completed Interview for the same pair and does not mutate Candidate master state.

## Offer

The Offer record itself means **offer made**.

```
Offer exists
    ↓
accepted = false / true
```

An Offer can be created only when Selection = `selected` for the same Candidate × Requirement.

There is no Offer status state machine.

## Joining

Joining has only two business facts:

```
joining_date = date
joined       = false / true
```

Joining requires an accepted Offer for the same Candidate × Requirement.

If `joined = true`, joining_date is required.

There is no scheduled/no-show/cancelled state machine.

## Billing

Billing is eligible only after:

```
Offer.accepted = true
        ↓
Joining.joined = true
        ↓
Billing
```

Billing date = joining_date.

Billing belongs to the Candidate × Requirement / Client recruitment transaction, never to Candidate globally.

## Explicitly out of core scope

- Employee HRMS
- Attendance
- Leave
- Payroll
- Employee benefits
- Performance management
- Appraisals
- Corporate HR policy engines
- Internal approval hierarchies
- Client interview panels
- Internal hiring-manager workflows
- Enterprise interview evaluation forms
- HRMS-specific AI
- General accounting ledger
- GST/tax engine
- Payment reconciliation
