# ADR 0009: Docker Backend Architecture and Dev-Branch Deployment

**Status:** Proposed
**Decider:** Product owner
**Related branch:** `dev`

## Context

SkillSifter's frontend already runs on Cloudflare Pages, which builds and
deploys automatically from the repository. The backend has, until now, been
deployed manually to a single DigitalOcean droplet via
`infra/scripts/deploy.sh`: a native Go binary managed by systemd, reading
configuration from a local `.env` file and applying migrations directly
against PostgreSQL.

During recent independent development, a second, undocumented deployment
path appeared on the same droplet: a Dockerized backend and a
containerized PostgreSQL instance, started by hand and never reconciled
with the native path's configuration or credentials. Nginx was pointed at
the new container without the underlying inconsistency being resolved,
leaving two live backend processes on the same host with different
 database credentials — one correct, one stale. This produced a
hard-to-diagnose production authentication failure with no corresponding
code change.

This ADR defines the target backend architecture going forward and the
deployment discipline needed to avoid a repeat of that failure mode. This
is a forward-looking architectural decision for a revamped codebase, not a
migration or reconciliation of prior implementations.

## Decision

### 1. Containerize the backend application only

The Go API backend will run as a Docker container. The container image is
the deployable unit: it is built once (in CI) and run unchanged across
environments. No environment-specific logic is baked into the image;
environment differences are expressed entirely through environment
variables supplied at run time.

### 2. PostgreSQL never runs inside Docker

PostgreSQL is explicitly and permanently excluded from the container
architecture. It runs as an independent, natively-managed service (systemd)
on its own host.

This is a deliberate data-safety boundary, not a temporary simplification:

- The database's lifecycle (backup, restore, upgrade, restart) must never
  be coupled to the application container's lifecycle. A bad image, a
  failed deploy, or a container crash must never be able to affect the
  database process.
- The application connects to PostgreSQL purely over the network, using
  `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` supplied as
  container environment variables. The application container has no
  awareness of whether PostgreSQL is local or remote.
- This makes the eventual move of PostgreSQL to a dedicated server (the
  planned path once the business scales) a configuration change
  (`DB_HOST`), not an architectural change. No application code or
  container image changes when that move happens.
- No containerized PostgreSQL (e.g. `postgres:*` images) is to be run in
  any environment, including local development, to keep the connectivity
  model consistent across dev, staging, and production and to prevent a
  repeat of the credential/host drift that caused the recent incident.

### 3. Target system architecture

```
GitHub (single source of truth; CI builds and pushes the image; CD deploys it)
  │
  ├── Frontend → Cloudflare Pages (independent build/deploy, unaffected by backend changes)
  │
  └── Backend → DigitalOcean droplet
         ├── Docker container: Go API only (stateless, rebuildable, portable)
         └── PostgreSQL: native systemd service, independent of the container
                (today: same droplet as the API container
                 future: its own dedicated server — DB_HOST change only)
```

### 4. Container-to-database connectivity

The container reaches PostgreSQL over the network using the `DB_HOST`
value, not a Docker-internal DNS alias. `--network host` is used for the
backend container so that `DB_HOST=localhost` works identically to the
previous native deployment, while remaining a pure configuration value that
changes to a real hostname or IP when PostgreSQL moves to its own server.
The application must not assume PostgreSQL is reachable via a
Docker-Compose service name; no docker-compose-managed database service is
part of this architecture.

### 5. Branch and deployment discipline

- All Docker/CD implementation work happens on `dev`. Production traffic
  continues to be served by the existing, proven deployment until `dev` is
demonstrated stable.
- `dev` deploys automatically on push, via CI building and pushing the
  container image and CD pulling and restarting it on the target
environment. No manual `docker run`, `docker compose up`, or hand-edited
  compose files on the server are permitted as a deployment mechanism —
  this is exactly the practice that caused the recent incident and must
  not recur.
- `dev`'s deployment target is isolated from the current production
  service (distinct container name and port at minimum) so that testing
the new architecture carries zero risk to what is currently live.
- Promotion from `dev` to `main` — and therefore to production — is a
  deliberate, explicit action taken only once the containerized backend
  has been verified stable, not an automatic consequence of CI passing on
  `dev`.
- Any server-side configuration required for the containerized backend
  (nginx routing, systemd unit if used to supervise the container,
  environment files) is committed to the repository and synced from git on
  deploy, matching the existing `deploy.sh` discipline of never leaving
  live server configuration to drift from what is tracked in version
  control.

## Consequences

- The database is protected from application-layer failures by
  construction, not by operational discipline alone.
- Moving PostgreSQL to a dedicated server later requires no application or
  image changes — only a `DB_HOST` update and, if desired, a
  network/firewall configuration step.
- The frontend, backend, and database become three independently
  deployable units, each with its own lifecycle and blast radius,
  consistent with the target SaaS-scale architecture.
- An explicit, git-driven CD mechanism for `dev` must be designed and
  implemented before this ADR can be considered fulfilled; this ADR
  establishes the target architecture and constraints but does not itself
  specify the CD tool or exact automation mechanism, which is
  implementation work to follow.
- No existing production behavior changes as a result of this ADR alone;
  it governs new work on `dev` only, until promotion.

## Open Implementation Questions

The following are intentionally left for implementation-time decision,
within the constraints above:

- Exact CD mechanism for `dev` (e.g. GHCR image push plus an SSH-triggered
  pull/restart action, versus a self-hosted runner on the droplet).
- Whether `dev` deploys to the same droplet as production (isolated by
  container name/port) or a separate environment.
- Systemd-vs-pure-Docker supervision of the backend container (i.e.
  whether a systemd unit wraps `docker run`, as the prior native service
  did, or the container is managed directly by the CD mechanism).
