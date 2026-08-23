# Migration numbering note (Issue #33 prerequisite)

`002_skill_aliases.sql` was renumbered to `003_skill_aliases.sql`
(and `003_job_ownership.sql` / `004_candidates_created_at.sql` shifted to
`004_` / `005_` accordingly) to resolve the duplicate `002_*` prefix flagged
in ADR 0007 ("Existing duplicate `002_*` files must be reconciled during
implementation; the runner must not silently choose between them").

File contents are unchanged — only filenames were renumbered. This was a
prerequisite for `backend/db/migrations.go` to run every file in this
directory in a single deterministic order (previously the runner only ever
executed `002_ai_reporting.sql`, hardcoded, and never ran the other files at
all).

No new domain logic, checksum verification, or migration tooling was added
as part of this reconciliation — see the v0.5 Issue #33 PR description for
the full scope boundary.
