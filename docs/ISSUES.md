# SkillSifter — Issue Tracking Log

Persistent, cross-release tracking of known issues and limitations. The
CHANGELOG references this file rather than duplicating it — this is the
single source of truth for what's open, what's closed, and when.

| ID      | Description                                                                                          | Status  | Opened | Closed  |
|---------|-------------------------------------------------------------------------------------------------------|---------|--------|---------|
| ISS-001 | `GET /api/reports/sources` returns `500` — queries a `candidates.source` column that doesn't exist    | Open    | v0.2.0 | —       |
| ISS-002 | Interviews display/scheduling not verified working                                                    | Closed  | v0.2.0 | v0.3.0  |
| ISS-003 | CORS origins hardcoded in source, not environment-configurable                                        | Open    | v0.2.0 | —       |
| ISS-004 | No automated CI (GitHub Actions) — `test.sh` gate exists but is run manually, not on push              | Open    | v0.2.0 | —       |
| ISS-005 | `backup.sh`/`provision.sh` are syntax-checked but not yet run against a real production restore        | Open    | v0.3.0 | —       |
| ISS-006 | 512MB droplet running tight on memory (swap in active use) — not yet a hard failure, worth monitoring  | Open    | v0.3.0 | —       |

## Notes

**ISS-002** — closed by 7816709 (Interviews reschedule now updates the
existing record instead of creating a duplicate) and 54755ad (Interviews
scheduling `candidateId` bug fixed). Verified against production traffic
during the v0.3.0 Docker cutover.

**ISS-003** — 54755ad added the Cloudflare Pages root domain to allowed
CORS origins, which is real progress, but the underlying issue (origins
are hardcoded in source rather than driven by an environment variable) is
not fixed. Left open rather than closed to avoid overstating the fix —
revisit if this needs to be genuinely env-configurable before onboarding
external SaaS customers.

**ISS-006** — `alphaforge-backend` (unrelated app sharing this droplet)
and Docker daemon overhead are the larger contributors, not SkillSifter's
own container (which runs ~2-10MiB RSS). Candidate fixes if this becomes a
real problem: bump droplet to 1GB, or reclaim ~50MiB from an idle PM2
daemon and `multipathd` (see chat history from the v0.3.0 Docker migration
for details) — not done as part of this release since it's unrelated to
SkillSifter itself.