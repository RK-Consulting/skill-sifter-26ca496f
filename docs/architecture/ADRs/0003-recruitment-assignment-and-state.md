# ADR 0003: Recruitment Assignment and Transaction State

**Status:** Superseded by ADR 0012 and the Candidate × Requirement recruitment model.

This ADR documented an earlier Recruitment Assignment domain. The Assignment entity is no longer part of the current runtime recruitment architecture.

Historical migration files and historical data are retained for upgrade compatibility. They must not be used as dependencies by new workflow code.

The current model is:

```text
Client
  ↓
Requirement
  ↓
Candidate × Requirement
  ├── Screening history
  ├── Submission history
  ├── Interview history
  └── Selection decision
```

Candidate master state is independent from any individual requirement outcome. Screening capacity is persisted on Candidate; interview concurrency is scoped to Candidate × Requirement; Selection is unique to Candidate × Requirement.

See ADR 0012 for the authoritative architecture.
