# Known Issues and Baseline Reconciliation

This file records facts requiring an explicit issue or decision; it is not a substitute for GitHub Issues.

## Release and documentation reconciliation — resolved baseline

- The product owner approved v0.3.0 as the current repository baseline on 2026-08-23. README, changelog, and release notes now use that designation.
- The repository does not currently contain a GitHub Actions workflow, although architecture material refers to GitHub Actions as CI/CD.
- Historical documentation is stale in places: the dashboard's recent activity is now API-backed in code, while other dashboard visuals remain hardcoded; report and feature documentation should be re-verified against running software.

The GitHub milestone and issues must be published after repository authentication is restored.

## Verified technical follow-ups

- Runtime schema initialization uses the authoritative numbered schema definitions under `backend/database/migrations/`. Historical migration files are retained as schema history and must not be rewritten.
- Legacy Jobs handlers were retired as part of the Requirements migration; this historical follow-up is closed.
- Tenant filtering remains handler/query based and should be hardened through an approved design.
- Dashboard recruitment pipeline and charts use hardcoded data and must not be represented as real reporting.
- Requirements are the authoritative recruitment-demand model; Daily Jobs is a separate operational domain.
