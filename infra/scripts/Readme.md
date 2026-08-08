# SkillSifter — Backup & Provisioning

Server backup and fresh-server provisioning scripts, scoped exclusively to
SkillSifter. These live in SkillSifter's own repo and do not reference or
depend on any other application that may share the same physical server.

## ⚠️ Assumptions to verify

These scripts were written from external observation (Nginx config,
`systemctl`, `ps`) — not from inspecting SkillSifter's own source or deploy
tooling directly. Before relying on them, confirm:

- **`.env` location** — assumed `/var/www/skillsifter/backend/.env`. If
  SkillSifter actually configures itself via the systemd unit's
  `Environment=` directives instead, the `.env` backup step will just find
  nothing (harmless), but the real config lives in the unit file, which
  *is* captured.
- **Redis usage** — not confirmed whether SkillSifter uses Redis at all. If
  it does, no separate handling is needed here since Redis is a single
  shared instance on this server; AlphaForge's backup already snapshots it.
- **Existing deploy pipeline** — the Nginx config references
  `infra/scripts/deploy.sh` inside SkillSifter's own repo. That script is
  the authoritative source for how this app is actually meant to be built
  and deployed. Treat `provision.sh` here as a disaster-recovery stopgap
  (restores the last-known compiled binary), not a replacement for that
  proper build pipeline.

## `backup.sh`

Run on the production server (needs `sudo`). Captures:
- `skillsifter` database (Postgres) — custom + plain SQL format
- SkillSifter's Nginx site config (`api.skillsifter.in`)
- Its SSL certificate
- Compiled binary, `.env` (if present), systemd unit file
- Shared server-level facts (UFW rules, cron, installed packages)

```bash
sudo bash backup.sh
```

Output: `/home/harisha/backups/skillsifter_backup_<timestamp>.tar.gz`

**⚠️ Contains live secrets.** Download off the server; don't leave copies
lying around longer than needed.

## `provision.sh`

Sets up a **brand-new** server and restores SkillSifter from a `backup.sh`
archive as a disaster-recovery stopgap: installs Nginx, PostgreSQL,
Certbot, UFW; restores the database; restores the last-known binary +
systemd service.

```bash
sudo bash provision.sh /path/to/skillsifter_backup_<timestamp>.tar.gz
```

For a genuine, deliberate migration (not disaster recovery), prefer
rebuilding SkillSifter fresh from source via its own repo's build/deploy
process rather than relying on the restored binary long-term.

**Deliberately manual, not automated:**
- DNS cutover and propagation wait
- SSL certificate re-issuance
- Setting a *new* DB password (the script forces this — never reuses old,
  potentially-exposed credentials)

## Status

- Both scripts are syntax-checked but have not yet been run against a real
  production restore, and the assumptions above have not been verified
  against SkillSifter's actual source/config. Verify the `.env` path and
  Redis usage before treating this as a trustworthy disaster-recovery plan.