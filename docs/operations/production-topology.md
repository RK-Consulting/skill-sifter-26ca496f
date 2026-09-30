# Production Topology, Configuration Recovery & Operational Boundaries

**Status:** Future operational reference  
**Applies to:** SkillSifter production  
**Release baseline:** v1.0.0  
**Date:** 2026-09-30

## 1. Purpose

SkillSifter is intentionally deployed as independently replaceable infrastructure boundaries. The production system must not become dependent on one hosting provider, one server, or one bundled deployment model.

The current production arrangement is:

```
                         INTERNET
                            |
                     DNS / Nameserver
                            |
             +--------------+--------------+
             |                             |
             v                             v
      Cloudflare Pages              DigitalOcean
      Frontend hosting              Production server
      React application                   |
                                          |
                                  +-------+-------+
                                  |               |
                                Nginx          Backend
                                  |               |
                                  +-------+-------+
                                          |
                                      PostgreSQL
```

GitHub is the source repository for application code and reproducible infrastructure definitions.

## 2. Production boundaries

| Boundary | Current location | Role | Replacement principle |
|---|---|---|---|
| DNS / nameserver | Current DNS provider | Domain resolution | Can move independently |
| Frontend | Cloudflare Pages | React SPA delivery | Can move independently |
| Source repository | GitHub private repository | Source/config/documentation | Remains provider-independent |
| Backend | DigitalOcean | Go API | Can move to another server/provider |
| Nginx | DigitalOcean | HTTPS/API reverse proxy | Configuration must remain reproducible |
| PostgreSQL | DigitalOcean, outside application container | Control DB + tenant DBs | Can move to another server/provider |
| Production secrets | Server + protected recovery storage | Runtime credentials | Never committed in plaintext |

The fact that Nginx, backend and PostgreSQL currently share one DigitalOcean server is a **cost/deployment choice**, not an application architecture dependency.

PostgreSQL is deliberately kept outside the backend container so it can later be moved to separate infrastructure without redesigning SkillSifter.

## 3. Recovery model

Production recovery has three different classes of assets.

### 3.1 Reproducible from GitHub

These belong in the private repository:

- application source
- frontend source
- database migrations
- Dockerfiles
- Docker Compose
- Nginx configuration
- systemd configuration
- deployment/bootstrap scripts
- CI workflows
- API and architecture documentation
- release metadata

### 3.2 Sensitive operational configuration

These must **not** be committed in plaintext:

- `backend/.env`
- DB passwords
- `JWT_SECRET`
- payment/provider secrets
- webhook secrets
- Cloudflare/API credentials
- TLS private keys
- other production credentials

The future recovery model is to keep an **encrypted production configuration backup** in the private repository, with its decryption/recovery key stored separately.

GitHub private-repository access is protected by account authentication and Google Authenticator 2FA. This provides an additional access-control layer, but repository privacy does not justify storing plaintext production secrets.

### 3.3 Production data

Database schema is reproducible from migrations. **Production data is not.**

Backups must therefore cover:

- control-plane PostgreSQL data
- all tenant PostgreSQL databases
- any other persistent production data required by the application

Database backup/recovery is a separate responsibility from Git source recovery.

## 4. Independent migration principle

A provider or infrastructure component may be replaced without changing the SkillSifter application architecture.

For example:

```
Current:
Cloudflare -> DigitalOcean Nginx -> Backend -> PostgreSQL

Possible future:
Cloudflare/new frontend host -> Provider A -> Backend -> Provider B PostgreSQL
```

The application should continue to depend on interfaces/configuration, not on the physical location of a provider.

## 5. Nginx configuration is source-controlled

The repository copy of the production Nginx configuration is the intended source of truth:

```
infra/nginx/api.skillsifter.in.conf
```

The live configuration is installed under:

```
/etc/nginx/sites-available/api.skillsifter.in
```

### Critical operational rule

**Do not make an emergency change to the live Nginx configuration and leave it there as an undocumented permanent change.**

This is particularly important for:

- CORS corrections
- proxy headers
- allowed origins
- TLS settings
- request/body limits
- timeout changes
- security headers
- API routing

The current deployment process copies the repository Nginx configuration to the live server. Therefore, a manual live-only correction can be overwritten by the next deployment.

### Emergency Nginx change procedure

If a production change must be made directly on the server:

1. Make the minimum required live change.
2. Test it with `nginx -t`.
3. Reload Nginx.
4. Immediately reproduce the final configuration in:
   `infra/nginx/api.skillsifter.in.conf`.
5. Commit the repository change.
6. Treat the Git version as the permanent configuration.
7. On the next deployment, verify that the live file matches the repository version.

The live server is not a second configuration repository.

## 6. Configuration drift

Configuration drift means:

> The live production configuration differs from the configuration stored in Git.

Nginx drift is especially dangerous because a deployment can legitimately overwrite a manual production correction.

Therefore, before declaring the release operationally frozen, Nginx drift should be checked and reconciled.

The same principle applies to:

- systemd unit configuration
- deployment scripts
- application configuration templates
- other files copied from Git to the production server

## 7. Emergency Nginx change checklist

Use this when a production issue requires a direct live Nginx correction, especially for CORS, proxy headers, TLS, routing, request limits or security headers.

1. **Make the smallest necessary live change.**
2. Run:
   ```
   nginx -t
   ```
3. If the configuration test passes, reload Nginx:
   ```
   systemctl reload nginx
   ```
4. Verify the affected production behaviour.
5. Copy the final working configuration back into:
   ```
   infra/nginx/api.skillsifter.in.conf
   ```
6. Review the Git diff carefully. Remove any temporary/debug-only change.
7. Commit and push the reconciled configuration to GitHub.
8. The next deployment should then pass the Nginx drift check.
9. If the drift check still reports a difference, **do not bypass it**. Compare the live and repository files and reconcile them deliberately.

### CORS-specific reminder

CORS corrections are particularly easy to make directly on the server and forget. Treat every production CORS correction as a source-code/infrastructure change:

```
Live CORS fix
    ↓
Test with nginx -t
    ↓
Reload Nginx
    ↓
Verify browser/API behaviour
    ↓
Copy final config into Git
    ↓
Commit
    ↓
Next deployment verifies no drift
```

Never solve a recurring CORS problem by permanently editing only `/etc/nginx/` on the server.

### Before declaring the incident closed

Confirm:

- the live Nginx file matches the repository file;
- `nginx -t` passes;
- the affected frontend/API flow works;
- the Git change is committed and pushed;
- no temporary credentials, debugging headers or unrelated changes were left in the configuration.

## 8. Docker and production distinction

`docker-compose.yml` provides a self-contained development/reproducibility environment.

It must not be interpreted as the production topology.

Production deliberately keeps PostgreSQL outside the application container.

This separation is intentional and must be preserved unless a future architecture decision changes it.

## 8. Recovery sequence

A future replacement DigitalOcean/application server should be recoverable approximately as follows:

```
New server
   |
   v
Install OS/runtime dependencies
   |
   v
Clone private GitHub repository
   |
   +--> restore encrypted production configuration
   |
   +--> install Nginx configuration
   |
   +--> install systemd configuration
   |
   +--> build backend
   |
   +--> restore PostgreSQL / connect to external PostgreSQL
   |
   +--> run authoritative schema/migration checks
   |
   +--> configure TLS
   |
   +--> health check
   |
   v
Production service
```

Actual production recovery should be tested periodically rather than assumed.

## 9. Security and recovery credentials

Keep the following recovery assets outside the Git repository plaintext:

- GitHub recovery codes
- Google Authenticator recovery/backup capability
- encryption/decryption key for encrypted production configuration
- database backup credentials
- provider recovery credentials
- TLS certificate/private-key recovery material where applicable

Do not store all recovery mechanisms on the same device or in the same location.

## 10. V1 operational principle

SkillSifter v1.0.0 intentionally favors:

- simple infrastructure
- low recurring cost
- clear physical and logical boundaries
- reproducible configuration
- portable PostgreSQL
- source-controlled infrastructure
- minimal operational tooling

It does not require an enterprise secrets platform, Kubernetes, service mesh, or microservice infrastructure.

The goal is **replaceability without architectural redesign**.
