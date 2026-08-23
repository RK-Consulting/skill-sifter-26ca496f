# Codex Engineering Rules

**Status:** Active from v0.4.0

Codex is the implementation agent for SkillSifter. The product owner approves product, architecture, data-model, release, and major technical decisions. Codex may suggest improvements but must not silently decide them.

## Required operating rules

1. Implement only the approved GitHub issue or explicitly approved instruction.
2. Prefer the smallest safe change that meets the acceptance criteria.
3. Preserve the approved architecture, release boundary, domain names, public contracts, and existing behaviour unless the issue explicitly changes them.
4. Use the existing stack and avoid significant dependencies without an approved rationale.
5. Protect authentication, authorization, tenant isolation, and sensitive candidate data in every change.
6. Use deterministic, reviewable migrations for database changes; consider upgrade compatibility, recovery, and tests.
7. Add and run tests appropriate to the change. Do not weaken tests or CI to obtain a pass.
8. Update directly relevant documentation and keep changes focused.
9. Work on a feature/fix/docs branch and use pull requests; do not commit directly to `main` unless explicitly instructed.

## Stop and report

Stop implementation and request direction for an architecture, database-model, API-contract, security, tenant-isolation, dependency, migration, breaking-change, data-loss, release-boundary, or materially ambiguous requirement conflict.

The report must identify the affected components, why implementation cannot safely continue, and viable options. A recommendation is not an authorization.
