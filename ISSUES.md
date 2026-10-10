# Known Issues and Baseline Reconciliation

This file records facts requiring an explicit issue or decision; it is not a substitute for GitHub Issues.

## Release and documentation reconciliation — resolved baseline

- The product owner approved v0.3.0 as the current repository baseline on 2026-08-23. README, changelog, and release notes now use that designation.
- The repository does not currently contain a GitHub Actions workflow, although architecture material refers to GitHub Actions as CI/CD.

The GitHub milestone and issues must be published after repository authentication is restored.

## Verified technical follow-ups

- Runtime schema initialization uses the final independently versioned control-plane and tenant-plane schema baselines:

- `backend/database/control-plane/001_baseline.sql`
- `backend/database/tenant-plane/001_baseline.sql`

- The historical `backend/database/migrations/` chain was retired for v1.0.0 and remains available through Git history only.

- Legacy Jobs handlers were retired as part of the Requirements migration; this historical follow-up is closed.
- Dashboard charts with hardcoded sample values were removed; remaining pipeline/activity views use API-backed data.
- Requirements are the authoritative recruitment-demand model; retired Daily Jobs and Business Development runtime surfaces were removed from the final V1 code path.
