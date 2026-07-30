# SkillSifter — Architecture Document

## 1. Overview

SkillSifter is a **multi-tenant Applicant Tracking System (ATS)** for recruitment/staffing firms. It is a classic three-tier web application:

1. **Presentation tier** — a React/TypeScript single-page application (SPA), served as static assets via Nginx.
2. **Application tier** — a Go REST API that owns all business logic, authentication, and authorization.
3. **Data tier** — a single shared PostgreSQL database, with tenant isolation enforced at the row level via a `company_name` column.

There is no separate services layer, message queue, or cache — the system is a monolithic API backing a monolithic SPA. This is appropriate for its current scope (a handful of resource types, moderate traffic) and keeps operational complexity low.

## 2. System Context

```
Recruiter / Admin / Manager (browser)
        │  HTTPS
        ▼
   Nginx (static SPA + gzip + SPA fallback routing)
        │  HTTPS (VITE_API_URL, e.g. api.skillsifter.in)
        ▼
   Go REST API (gorilla/mux)
        │  database/sql + lib/pq
        ▼
   PostgreSQL (single shared instance, tenant-scoped rows)
```

External actors are recruiting-company staff (`admin`, `manager`, `recruiter`, `team_leader` roles) interacting entirely through the browser. There are no third-party integrations, webhooks, or external APIs in the current codebase.

## 3. Component Architecture

### 3.1 Frontend (`/src`)

- **Framework**: React 18 + TypeScript, built with Vite.
- **Routing**: `react-router-dom`, with a `ProtectedRoute` wrapper component that gates all authenticated pages behind a client-side check for a `token` + `user` pair in `localStorage`. Unauthenticated access to protected routes redirects to `/login`.
- **State/data fetching**: TanStack Query (`@tanstack/react-query`) wraps the whole app in `App.tsx`, providing caching, retry (1 retry, 1s delay), and a 5-minute stale time for all queries.
- **API client**: `src/services/api.ts` wraps Axios with:
  - A request interceptor that attaches `Authorization: Bearer <token>` from `localStorage` and normalizes request URLs to always include an `/api/` prefix.
  - A response interceptor that logs errors and detects `401` responses (currently logs only — does not auto-redirect to login).
  - Domain-specific service objects: `authService`, `candidateService`, and others per resource (jobs, interviews, daily jobs, business dev, reports).
- **UI layer**: shadcn/ui components built on Radix UI primitives, styled with Tailwind CSS. `src/components/ui` holds ~49 generic UI primitives (buttons, dialogs, forms, etc.); `src/components/dashboard` and `src/components/layout` hold app-specific composition; `src/components/ui-custom` holds bespoke components.
- **Pages** (`src/pages`) map roughly 1:1 to the domain resources: `Candidates`, `AddCandidate`, `Jobs`, `AddJob`, `DailyJobs`, `AddDailyJob`, `BusinessDev`, `AddBusinessDev`, `Interviews`, `InterviewDetails`, `ScheduleInterview`, `Reports`, plus `Login`, `Register`, `Index` (dashboard), and `NotFound`.

### 3.2 Backend (`/backend`)

Written in Go 1.21, structured as a small set of packages rather than a framework-driven layout:

| Package | Responsibility |
|---|---|
| `main` (`main.go`) | Wires up the router, middleware chain, CORS policy, and starts the HTTP server. |
| `auth` | JWT issuance/validation (`golang-jwt/jwt/v5`), `AuthMiddleware` (validates bearer tokens), `RoleMiddleware` (role allow-listing). |
| `db` | Database connection setup (`InitDB`), schema bootstrap (`InitializeSchema`), and environment variable helpers. |
| `handlers` | One file per resource (`candidate_handlers.go`, `job_handlers.go`, `interview_handlers.go`, `daily_job_handlers.go`, `business_dev_handlers.go`, `auth_handlers.go`, `report_handler.go`) plus `common.go` for shared JSON response helpers. |
| `models` | Plain Go structs for every resource, plus response envelopes (`ApiResponse`, `TokenResponse`, report DTOs). |

**Routing** (`main.go`):
- Public routes: root/health/ping endpoints, `/auth/register`, `/auth/login` — each duplicated at both `/x` and `/api/x` for client compatibility.
- Protected routes: mounted under `/api` with `AuthMiddleware` applied to the whole subrouter, and a second, **duplicate, non-`/api`-prefixed** route tree also protected by `AuthMiddleware`. This duplication (routes registered twice, once with `/api` and once without) is a pragmatic way to tolerate frontend requests that may or may not include the prefix, at the cost of maintaining two copies of the route table.
- `setupResourceRoutes` is a small helper that registers the standard 5-route REST CRUD pattern (`GET list`, `POST create`, `GET/{id}`, `PUT/{id}`, `DELETE/{id}`) for each resource, avoiding repetition across candidates/jobs/daily-jobs/interviews/business-dev.
- Admin-only routes (`/api/admin/users`) are additionally wrapped in `RoleMiddleware("admin")`.
- A catch-all `OPTIONS` handler returns `204` for any unmatched preflight request.

**Middleware chain** (outermost to innermost): logging → CORS → auth (bearer token parsing/validation) → role check (where applicable) → handler.

### 3.3 Data tier

Single PostgreSQL database (`backend/database/schema.sql`), initialized automatically on backend startup via `db.InitializeSchema()` (idempotent `CREATE TABLE IF NOT EXISTS` statements) and also mountable directly into the Postgres container as an init script in Docker Compose.

Core tables: `companies`, `roles`, `users`, `candidates`, `jobs`, `daily_jobs`, `interviews`, `business_dev`. See §5 for the tenancy model and §6 for the full entity diagram.

## 4. Authentication & Authorization

- **Mechanism**: Stateless JWT, HMAC-SHA256 signed (`HS256`).
- **Token contents**: `userId`, `email`, `role`, `companyName`, plus standard `exp` claim. Tokens expire 24 hours after issuance.
- **Password storage**: bcrypt (`golang.org/x/crypto/bcrypt`, default cost) — the plaintext password is never persisted.
- **Registration flow** (`RegisterUser` in `auth_handlers.go`):
  1. Runs inside a single DB transaction.
  2. If the submitted `companyName` doesn't already exist in the `companies` table, it is created on the fly (ID derived deterministically from the lowercased, underscore-joined name), and the registering user becomes that company's first user.
  3. The first user for a new company is auto-assigned the `admin` role (unless a role was explicitly supplied); subsequent users default to `recruiter` unless a role is given.
  4. Role is validated against a fixed allow-list (`admin`, `manager`, `recruiter`, `team_leader`).
  5. On success, a JWT is generated and returned alongside the created user (password omitted).
- **Login flow** (`LoginUser`): looks up the user by email, verifies the bcrypt hash, issues a new JWT.
- **Authorization**:
  - `AuthMiddleware` extracts and validates the bearer token, then injects `userID`, `email`, `role`, and `companyName` into the request context for downstream handlers.
  - `RoleMiddleware(allowedRoles...)` reads `role` from context and rejects (`403`) if it isn't in the allow-list. Used today only to gate `/api/admin/*`.
  - **Tenant scoping is not enforced by middleware** — each handler is individually responsible for filtering queries by the `companyName` pulled from the JWT-derived context (e.g. `report_handler.go` does `WHERE company_name = $1`). This is a convention rather than a structural guarantee; a handler that forgets the filter would leak cross-tenant data.

## 5. Multi-Tenancy Model

SkillSifter uses **shared-database, shared-schema, discriminator-column** multi-tenancy:

- Every business table (`candidates`, `jobs`, `daily_jobs`, `interviews`, `business_dev`, `users`) carries a `company_name VARCHAR(255) NOT NULL` column.
- Tenants are identified by company **name**, not a surrogate numeric ID (the `companies.id` field exists but is a slug-like string, e.g. `comp_acme_corp`, generated at registration time — it is not used as a foreign key by the other tables, which instead reference `company_name` directly).
- Indexes exist on `company_name` for every tenant-scoped table (`idx_candidates_company`, `idx_jobs_company`, etc.) to keep per-tenant queries efficient.
- Roles (`admin`, `manager`, `recruiter`, `team_leader`) are global definitions with a fixed permission list stored as JSONB — they are not currently tenant-customizable.

**Trade-off worth flagging**: using the human-readable company name (rather than an immutable ID) as the tenancy key means a company rename would require a coordinated update across every table, and it also makes company names effectively globally unique identifiers rather than just display labels.

## 6. Data Model (Entity Overview)

- **companies** (`id` PK, `name` unique) — one row per tenant.
- **roles** (`id` PK, `name` unique, `permissions` JSONB) — seeded with `admin`, `manager`, `recruiter`, `team_leader`.
- **users** (`id` PK, `email` unique, `password` hash, `role`, `company_name`) — role is stored as a string on the user row rather than a FK to `roles.id`.
- **candidates** (`id` PK, contact/skill fields, `company_name`).
- **jobs** (`id` PK, title/department/location/status, `company_name`).
- **daily_jobs** (`id` PK, `assigned_user` → `users.id`, `company_name`).
- **interviews** (`id` PK, `candidate_id` → `candidates.id`, status/feedback, `company_name`).
- **business_dev** (`id` PK, client/partner/contact fields, `company_name`).

Relationships are intentionally loose: most foreign keys (e.g. `interviews.candidate_id`) are nullable-by-schema references without `ON DELETE` behavior specified, and cross-resource joins (e.g. candidate ↔ interview by name rather than strictly by ID in some UI flows) are handled in application code rather than enforced by the schema.

## 7. Deployment Architecture

Three containers, orchestrated via `docker-compose.yml`:

| Service | Image / Build | Notes |
|---|---|---|
| `postgres` | `postgres:15-alpine` | Persists to a named volume (`pgdata`); mounts `schema.sql` as a Docker init script; has a healthcheck (`pg_isready`) gating backend startup. |
| `backend` | Built from `backend/Dockerfile` (`golang:1.21-alpine`) | Compiles `main.go` to a binary and runs it directly; waits for Postgres to be healthy. |
| `frontend` | Built from root `Dockerfile` (multi-stage: `node:18-alpine` build → `nginx:alpine` serve) | Vite build output served by Nginx; Nginx config adds gzip and a SPA `try_files` fallback so client-side routing works on refresh. |

Backend configuration is entirely environment-variable driven (`PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`), with sane local defaults baked into `db/env.go`. The frontend's API base URL is baked in at build time via `VITE_API_URL`.

CORS is currently a hardcoded allow-list in `main.go` (`skillsifter.in` domains + common localhost dev ports), rather than environment-configurable — changing deployment domains requires a code change and rebuild.

## 8. Cross-Cutting Concerns

- **Logging**: a single `loggingMiddleware` logs `remote_addr method path` for every request to stdout. No structured logging, request IDs, or log levels.
- **Error handling**: handlers use a shared `respondWithJSON`/`respondWithError` pair (`handlers/common.go`) to return a consistent `{success, message, data?}` JSON envelope with an appropriate HTTP status code.
- **Validation**: minimal and ad hoc — mostly presence checks on required fields in handlers (e.g. registration requires `email`, `password`, `username`, `companyName`). No schema-based request validation library is used.
- **API documentation**: a Swagger/OpenAPI spec exists at `backend/docs/swagger.yaml`, plus some Swagger-style doc comments on models (e.g. `HiringReportResponse`), though it's unclear from the code alone whether the spec is auto-generated or hand-maintained and kept in sync.

## 9. Known Architectural Risks & Technical Debt

These are worth addressing before treating the system as production-hardened:

1. **Hardcoded JWT signing key** (`auth.go`: `skill_sifter_secret_key`) — should move to an environment variable / secrets manager. As-is, anyone with source access can forge valid tokens for any user/role/company.
2. **Tenant isolation is convention-based, not structural** — nothing prevents a future handler from omitting the `company_name` filter and leaking cross-tenant data. A shared query-builder/middleware that injects the tenant filter automatically would close this gap.
3. **Duplicated route tables** (`/api/...` and bare `/...` registered separately in `main.go`) — doubles the surface area for routing bugs; a single canonical prefix with a redirect or rewrite would be simpler to maintain.
4. **Hardcoded CORS origins** — deployment-specific domains are baked into the binary rather than configured via environment.
5. **No tenant → role FK integrity** — `users.role` is a free-standing string, not a foreign key to `roles.id`, so the two can drift out of sync.
6. **A compiled binary (`backend/skillsifter`) is checked into version control** — build artifacts shouldn't be committed; this bloats the repo and can go stale relative to source.
7. **No automated tests** were found in the repository — no `_test.go` files in the backend, no frontend test setup (e.g. Vitest/Jest) evident in `package.json`.
8. **No rate limiting / brute-force protection** on `/auth/login`.

## 10. Suggested Evolution Paths

Not implemented today, but natural next steps if the system grows:

- Extract tenant-scoping into a middleware or repository-layer helper so it's enforced once rather than per-handler.
- Move secrets (JWT key, DB credentials) to environment variables / a secrets manager, and out of source.
- Add integration tests around the auth and multi-tenancy boundaries specifically, since those are the highest-risk areas.
- Introduce structured logging and request tracing if the system will run at meaningful scale or need multi-instance debugging.
- Consider a proper migrations tool (e.g. `golang-migrate`) in place of the current idempotent-`CREATE TABLE`-on-boot approach, to support safer schema evolution over time.

## 11. Design Decisions Log — Future AI Integration (AstraMind)

**Status: vision-level, not in scope for v0.2.0 or any near-term release. Captured here so the reasoning isn't lost, not as a build spec.**

### Context

SkillSifter's product vision extends beyond CRUD candidate/job tracking into AI-assisted recruitment: searching unstructured resume files (`.docx`/`.pdf`) by skill set, extracting structured candidate data (name, phone, email) from them, digging information out of connected sources like Google Drive, and answering natural-language questions over both document content and SkillSifter's own database. This would be a headline differentiating feature (USP), targeted at a **future release (tentatively v1.5+)**, not the current hardening milestone.

The AI capability is being built as a **separate project, AstraMind** (github.com/harishnagaraju/astramind) — an existing, mature, modular Go RAG engine with document ingestion, chunking, semantic search, retrieval-augmented Q&A, and multi-provider LLM support (OpenAI-compatible APIs, local Ollama). AstraMind predates and is independent of SkillSifter.

### Decision 1 — AstraMind is a platform; SkillSifter is one of its consumers, not a fork of it

AstraMind is intended to serve as generic, reusable AI infrastructure for multiple future applications (SkillSifter today; potentially unrelated verticals, e.g. a law-firm document tool, later). It runs as its **own separately deployed service**, callable over HTTPS, rather than being merged into SkillSifter's Go backend or any other consumer's codebase. This keeps AstraMind independently deployable, scalable, and — eventually — separately resourced (AI inference workloads have very different hardware needs than a CRUD API).

### Decision 2 — AstraMind stays domain-agnostic ("Option B" boundary)

AstraMind's core must never contain domain-specific concepts borrowed from any one consumer (no notion of "candidate," "resume," or "company" inside AstraMind itself). It exposes only generic primitives — ingest a document, embed, semantic search, ask a question of a knowledge base. All domain interpretation (e.g. recognizing a chunk as a candidate's resume, extracting name/phone/email) happens in the **calling application's** own code, using AstraMind's primitives as building blocks.

Rejected alternative: a true in-process "plugin" system where SkillSifter-specific code loads and runs inside AstraMind's own process. Rejected because it would make AstraMind's "stable platform" claim false in practice — any domain-specific branch in AstraMind's core is domain-specific code with extra steps, and would need rework for every new consumer rather than staying stable.

### Decision 3 — Two-level tenancy is a real risk to design around, not an afterthought

Once multiple applications call a shared AstraMind, and each application (like SkillSifter) has its own internal multi-tenancy (multiple companies), there are two nested scoping boundaries to get right:
1. **Application-level**: AstraMind must know *which calling application* it's serving (e.g. API-key-per-application), but should never need to know about that application's internal tenants.
2. **Tenant-level**: SkillSifter must enforce *which company* a request is scoped to **before** calling AstraMind — AstraMind should only ever receive an already-scoped payload, never raw access to "all of SkillSifter's data" with a company filter it's trusted to apply correctly itself.

Getting this backwards (trusting AstraMind to understand and enforce company-level scoping) would reproduce the same class of cross-tenant leak risk already flagged for SkillSifter's own database layer (§9.2) — except spanning two systems, which is harder to catch in review or testing.

### Decision 4 — Ship RAG + prompting first; defer fine-tuned adapters (LoRA) to a later phase

The long-term shape for per-application specialization is a shared base LLM with small, swappable **LoRA adapters** per domain (e.g. a "resume extraction" adapter for SkillSifter) — this is an established production pattern (similar to how vLLM/S-LoRA serve many adapters off one base model), not a novel idea, and it's compatible with the Decision 2 boundary since an adapter is a trained-weights artifact, not application code living inside AstraMind.

However, fine-tuning (even efficient LoRA fine-tuning) requires (a) a real labeled training corpus, which can only come from observing real usage and real mistakes in production, and (b) GPU compute budget, which isn't justified pre-revenue. Decision: **launch the AI feature using RAG + prompting against AstraMind's existing capabilities first**, let real usage generate the training data and reveal actual failure modes, and treat LoRA fine-tuning as a deliberate v2+ optimization once both real usage data and GPU budget exist — not a prerequisite to shipping.

### Decision 5 — v1 AI scope: two use cases, classified by what actually needs AI

A broad review of SkillSifter's feature set shows most of it is structured, database-oriented CRUD (Candidates, Jobs, Daily Tasks, Business Dev, Interviews) — plain SQL and filter forms are the right tool there, not AI. Only two genuine AI use cases were identified:

**Use Case 1 — Resume Extraction.** Input: unstructured files (PDF/.docx), from a local folder initially, Google Drive later. Output: a structured candidate record (Name, Email, Phone, Skills) written to the `candidates` table. Nature: unstructured → structured. This is AstraMind's ingestion/RAG territory. Note: the candidate's primary key (auto-generated ID) is **not** an AI concern — plain database auto-increment/UUID logic, generated at insert time by ordinary application code, not by AstraMind.

**Use Case 2 — Candidate-to-Job Matching / Ranking.** Input: an open job's required skills/description, plus the pool of already-extracted candidate records. Output: candidates ranked by relevance to that job. Nature: structured ↔ structured semantic similarity, reusing the same embeddings generated during Use Case 1 rather than requiring a separate system. This is considered high enough value — arguably the core value proposition of a recruiter's actual job (best-fit matching, not exact keyword matching) — to include in the **same v1 AI milestone** as extraction, not deferred to a later release.

**Decision: v1 AI milestone scope = Use Case 1 (extraction) + Use Case 2 (matching/ranking), built together**, since ranking is a natural, low-incremental-cost extension of the extraction pipeline rather than a separate system to design later.

### Decision 6 — Ideas considered and explicitly deferred or rejected for v1

| Idea | Disposition | Reasoning |
|---|---|---|
| Skill-set search box on Candidates page | **Not an AI feature at MVP** | If skills are stored as structured tags/array from extraction, a plain `WHERE skills @> ARRAY[...]` or `ILIKE` query is fast, cheap, predictable. Semantic search (e.g. "React" matching "frontend developer") is a plausible later upgrade, not a v1 requirement — no evidence yet that plain filtering is insufficient |
| Natural-language DB queries (e.g. "how many candidates from Bangalore applied last month") | **Deferred, not v1** | Genuine AI use case (text-to-SQL) but a power-user convenience feature, ranked below matching/ranking in priority |
| Duplicate candidate detection | **Deferred, and likely non-AI when built** | Simple fuzzy-matching (edit distance on email/phone) gets most of the value without needing AI at all |
| Google Drive ingestion | **Not a new AI capability** | Same extraction pipeline as Use Case 1, just a different file source — a new input adapter, not new AI logic |

### Open question — skill taxonomy normalization (resolved — see Decision 7 below)

Freeform extracted skills (e.g. `"React"`, `"ReactJS"`, `"React.js"` treated as three distinct strings) will quietly degrade both search and matching quality later. Not yet decided: whether extraction should normalize against a fixed skill taxonomy from day one, or whether that's an acceptable v2 refinement. This affects the design of the extraction prompt itself, so it needs to be settled **before** implementation begins, not after.

### Decision 7 — Skill matching resolved at query time, not extraction time (closes the open question above)

**Resolution to the open question above: freeform extraction stands. Matching quality is handled at query time via two independent, complementary layers, rather than by normalizing at extraction.**

1. **Surface-form variants** (`React` / `ReactJS` / `React.js`) — handled via fuzzy/prefix matching (Postgres `pg_trgm` extension, or plain `ILIKE '%term%'` wildcard matching as a simpler starting point) directly against the freeform `skills` array. No lookup table needed — general-purpose, catches capitalization/suffix/pluralization differences automatically.

2. **True abbreviation/full-form pairs** (`ML` / `Machine Learning`, `K8s` / `Kubernetes`, `CI/CD` / `Continuous Integration/Continuous Deployment`) — these don't share enough characters for fuzzy matching to bridge, so they're handled via a small, curated **`skill_aliases` reference table**, expanded at query time before matching against candidates. The tech-abbreviation vocabulary is bounded and slow-growing (roughly 150-300 entries covers the large majority of resume terminology across languages, infra/cloud, AI/ML, databases, practices, and frameworks) — not an open-ended taxonomy problem. The table can be extended over time as new terms enter common usage (e.g. "RAG," "LLM" in recent years), the same way any other schema change ships — via the existing migrations workflow (`backend/database/migrations/`, tracked under issue #11), not a special mechanism.

Neither layer touches the extraction prompt, and neither requires re-processing already-extracted resumes if the alias list grows later — both operate purely at query time. This was chosen over extraction-time normalization because it avoids adding complexity to the extraction prompt (Decision 5/Use Case 1) and keeps the reference vocabulary independently upgradable without touching AstraMind or re-running extraction.

**Explicitly accepted limitation**: this does not build a comprehensive taxonomy or ontology of skills (e.g. no attempt to model that "React" is a subset of "Frontend Development"). It solves lexical variation (same concept, different spelling/abbreviation), not conceptual hierarchy. Revisit only if real usage shows this is an actual gap, not preemptively.

### Explicitly out of scope right now

- No AstraMind integration code, API contracts, or infrastructure should be built as part of v0.2.0.
- Tenant-isolation hardening (§9.2) should still be prioritized for its own sake in the near term — independent of the AI roadmap — but its importance is reinforced by Decision 3 above.