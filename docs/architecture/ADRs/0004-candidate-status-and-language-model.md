# ADR 0004: Candidate Status and Generic Language Expertise

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** V04-05, Issue #35

## Context

The current candidate model has a `jlptlanguage` field. V1 must support generic language expertise and distinguish candidate relationship status from assignment-specific status.

Candidate state and Recruitment Assignment state are independent concerns. A candidate may participate in multiple recruitment assignments while retaining one candidate-level status.

## Decision

### 1. Candidate status

Candidate status is controlled and independent of Recruitment Assignment lifecycle.

The initial candidate status vocabulary is:

| Status | Meaning |
|---|---|
| `active` | Candidate is available for recruitment activity. |
| `inactive` | Candidate exists but is temporarily unavailable. |
| `blacklisted` | Candidate must not normally be submitted to new requirements. |
| `archived` | Historical candidate record retained; no new recruitment activity. |

Normal transitions are:

```text
active -> inactive
active -> blacklisted
active -> archived
inactive -> active
inactive -> archived
blacklisted -> active
blacklisted -> archived
```

`archived` is terminal for normal workflow. Reopening an archived candidate requires an explicit administrative/product decision rather than an ordinary status update.

Candidate status must never be overwritten by an assignment lifecycle transition. Existing assignments remain historically valid when candidate status changes.

### 2. Candidate eligibility

A candidate in `inactive`, `blacklisted`, or `archived` status is not eligible for a new recruitment submission by default. Existing assignment records remain available for historical and operational purposes.

Any exception workflow for blacklisted or otherwise unavailable candidates requires an explicit authorization decision and is outside Issue #35.

### 3. Language expertise

Language expertise is a candidate attribute, not an assignment lifecycle state.

The candidate model should support generic language expertise rather than treating Japanese/JLPT as the universal language model. Existing `jlptlanguage` data is retained for backward compatibility until a dedicated generic-language migration is approved.

The future generic representation is conceptually:

```text
candidate
   |
   +--> language expertise
          +--> language
          +--> proficiency level
```

The platform must not infer a common proficiency scale across languages where the underlying source data does not support it. A future normalized language taxonomy may define language-specific proficiency frameworks and mappings as a separate migration/release decision.

### 4. Data migration and release scope

Issue #35 does not perform a destructive or speculative migration of `jlptlanguage` into a generic language taxonomy.

Existing language data remains available. Generic language/proficiency normalization is additive and requires an explicit migration design before altering or removing the legacy representation.

### 5. Separation from assignment state

The following are intentionally separate:

```text
Candidate status
    !=
Recruitment Assignment status
```

For example, an `active` candidate may simultaneously have:

```text
Assignment A -> interviewing
Assignment B -> submitted
Assignment C -> rejected
```

Conversely, making a candidate inactive does not rewrite or invalidate the historical state of existing assignments.

## Consequences

### Positive

- Candidate availability is modeled independently from recruitment transactions.
- Multiple concurrent recruitment assignments are supported without corrupting candidate state.
- Existing JLPT-oriented data can remain compatible while the platform evolves toward generic language expertise.
- Assignment lifecycle remains owned by ADR 0003.

### Trade-offs

- Candidate status requires explicit transition validation.
- A generic language taxonomy remains a future data-modeling effort.
- Legacy `jlptlanguage` cannot be safely discarded until a migration and compatibility strategy is approved.

## Implementation Boundary

This ADR does not authorize deletion of `jlptlanguage`, speculative conversion of historical language values, or an exception workflow for blacklisted candidates. Those require separate decisions.
