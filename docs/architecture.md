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

## 12. Use Case & Requirements Baseline — Candidate Resume Backfill & Recruiter Workflow

**Status: requirements baseline for the AI resume-extraction feature (§11, Decisions 5-7), derived from a validated use case walkthrough with the actual end user's real workflow (R K Consulting). Precedes any architecture or design work for this feature — captured here first, per SE process discipline (Use Case → Requirements → Architecture → Design), before any interface or system-boundary decisions are made.**

### 12.1 Use Case: Manual Recruitment Screening and Data Entry

- **Actor**: Recruiter / Talent Acquisition Specialist (R K Consulting)
- **Goal**: Identify qualified candidates from a batch of received applications and log their professional profiles into the tracking repository for hiring manager review.
- **Trigger**: A new job requisition closes, or a batch of unorganized application files accumulates in the local sourcing folder.
- **Preconditions**: A local folder exists containing unprocessed resume files (PDF, DOCX); the tracking system is accessible; the job description and hiring criteria are defined.

**Main flow, as currently performed manually (pre-automation baseline):**
1. Open the folder containing new resumes.
2. Open a resume file, scan it.
3. Assess the candidate against job requirements.
4. Extract name, email, phone, primary skills, most recent job title.
5. Manually log the extracted details into the tracking system.
6. Move the processed file into a "Processed" subfolder (this is the actor's mechanism for avoiding duplicate work across sessions — not incidental).
7. Repeat for every file in the folder.

**Postcondition / done-state**: the input folder is empty, all candidate profiles are visible in the tracking system, and the recruiter can begin scheduling interviews without needing to reopen raw resume files. (Note: the bar is "the database is trustworthy enough to never need to return to source," not merely "extraction happened.")

### 12.2 Validated Flows (target automated behavior)

**Flow A — One-time backfill.** Approximately 3,000 existing resumes (mixed PDF/DOCX) in a local folder are scanned and loaded into the database as a one-time batch operation, not per-candidate interaction.

**Flow B — Ongoing recruiter workflow, database-first.** After Flow A, the recruiter works against the database, not raw files:
1. A job description arrives; the recruiter searches/matches against candidate database records, not resume files.
2. On a match, the recruiter notes phone/email and calls the candidate.
3. **If not interested** — the recruiter marks a status against that candidate record (open/extensible set: interested, not interested, wait period, on notice period, etc.).
4. **If interested** — the recruiter requests and receives an updated resume.
5. The recruiter goes to the candidate's record (keyed by `candidate_id`), marks the old resume for deletion, and uploads/replaces it with the new one. **This entire flow happens within the existing Candidates section of the application** — no separate upload area.

Clarified rule: "latest resume wins" — no version history is expected or wanted; the old resume is discarded once a newer one is confirmed, not archived.

**Flow C — Brand new candidate**, not sourced from backfill or an existing record: uses the existing **Add Candidate** button/form. Already implemented, no gap identified.

### 12.3 Requirements Baseline

Each requirement is traced to its source flow/step. This is a **draft baseline pending review and correction** — not yet finalized.

**Functional — Backfill (traces to Flow A)**

| ID | Requirement | Trace |
|---|---|---|
| REQ-F-01 | The system shall accept a local folder of resume files (PDF, DOCX) as input for one-time batch processing. | Flow A |
| REQ-F-02 | The system shall extract, per resume: candidate name, email, phone number, and skills. | Flow A |
| REQ-F-03 | The system shall create one candidate record per successfully processed resume, tagged to the correct tenant (`company_name`). | Flow A |
| REQ-F-04 | The system shall retain a reference to the source resume file for each candidate record created via backfill. | Flow A → supports Flow B Step 6 |

**Functional — Recruiter Workflow (traces to Flow B)**

| ID | Requirement | Trace |
|---|---|---|
| REQ-F-05 | The system shall allow the recruiter to search/match candidates against a job description using database records only, without requiring access to source resume files. | Flow B, Steps 1-2 |
| REQ-F-06 | The system shall allow the recruiter to view a matched candidate's phone number and email. | Flow B, Step 3 |
| REQ-F-07 | The system shall provide a status field on each candidate record, supporting an open/extensible set of values (e.g. interested, not interested, wait period, on notice period). | Flow B, Step 4 |
| REQ-F-08 | The system shall allow the recruiter to update a candidate's status independent of any other field change. | Flow B, Step 4 |
| REQ-F-09 | The system shall allow the recruiter to replace a candidate's resume file, keyed by `candidate_id`, removing the prior resume. | Flow B, Step 6 |
| REQ-F-10 | The system shall NOT retain the prior resume file after a successful replacement (latest-resume-only model, no version history). | Flow B, Step 6 |
| REQ-F-11 | Resume replacement and status updates shall occur within the existing Candidates section of the application. | Flow B, Step 6 (explicit) |

**Functional — New Candidate (traces to Flow C)**

| ID | Requirement | Trace | Status |
|---|---|---|---|
| REQ-F-12 | The system shall allow a recruiter to manually add a new candidate not sourced from backfill or resume upload. | Flow C | **Already satisfied by existing Add Candidate feature — no new work required** |

**Data Requirements (schema gaps identified during use-case walkthrough, not present in current `candidates` table per §6/§9.5)**

| ID | Requirement | Trace |
|---|---|---|
| REQ-D-01 | The `candidates` data model shall include a `status` field, distinct from any existing field. | Supports REQ-F-07 |
| REQ-D-02 | The `candidates` data model shall include a reference to a stored resume file, not merely extracted text fields. | Supports REQ-F-04, REQ-F-09 |

**Constraints / Explicitly Out of Scope**

| ID | Statement |
|---|---|
| CON-01 | No resume version history is required — a conscious exclusion, not an oversight (see REQ-F-10). |
| CON-02 | The AstraMind↔SkillSifter HTTPS API and auth work (§11, Decision 5-6) is not required to satisfy this use case. Flow A (backfill) can run as a local one-time job against the resume folder directly; the HTTPS API remains scoped to a separate, future live/ongoing ingestion capability. |

### 12.4 Open Items — Resolved

All items originally flagged as open have been resolved through further use-case clarification:

| Item | Resolution |
|---|---|
| Unreadable/unusual files during backfill (corrupted, scanned-image-only, non-English, unusual formatting) | Logged to a report list (filename + reason) at the end of the backfill run. Not inserted into the database. |
| `position` vs. "most recent job title" | Same field, not two. No schema split needed. |
| Duplicate/update detection (is this an update to an existing candidate, or a new one?) | Match on **name AND email** together, not either alone. |
| Exception handling for "unqualified" candidates during backfill | **Not applicable to Flow A.** "Qualified" is only meaningful relative to a specific job description, and Flow A (backfill) has no JD attached — it is a blind bulk import. Every candidate that is *readable* (see above) is stored, regardless of fit for any particular role. Qualification is evaluated dynamically in Flow B, at search time, against whichever JD the recruiter is currently working — it is a property of a search, not a property of a stored candidate record. |

### 12.5 Data Structure Design

Traces to REQ-D-01 and REQ-D-02.

**Status field (REQ-D-01, REQ-F-07).** REQ-F-07 calls for an open/extensible set of values, not a fixed list. A plain free-text column would satisfy "extensible" but permits inconsistent values over time (`"Not Interested"` vs `"not interested"` vs `"NI"`, silently fragmenting later reporting). Following the same reference-table pattern already established by the existing `roles` table:

```sql
CREATE TABLE candidate_statuses (
    id SERIAL PRIMARY KEY,
    label VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO candidate_statuses (label) VALUES
    ('Interested'),
    ('Not Interested'),
    ('Wait 2 Months'),
    ('On Notice Period');

ALTER TABLE candidates ADD COLUMN status_id INTEGER REFERENCES candidate_statuses(id);
```

Extensible by inserting new rows, no migration required for new status values; consistent values, no free-text drift. Nullable — a freshly backfilled or newly added candidate has no status until the recruiter engages with them (Flow B).

**Resume file reference (REQ-D-02, supports REQ-F-04/09/10).** Resume files reside in exactly two possible locations, never on SkillSifter's own infrastructure: the recruiter's local drive, or Google Drive. The schema stores only a reference to the file's location, never the file itself — no server-side storage, no disk footprint on the droplet.

```sql
ALTER TABLE candidates ADD COLUMN resumesource VARCHAR(20);        -- 'local' or 'gdrive'
ALTER TABLE candidates ADD COLUMN resumereference VARCHAR(500);    -- Google Drive file ID (local: not stored here — see below)
ALTER TABLE candidates ADD COLUMN resumefilename VARCHAR(255);     -- original filename, for display
ALTER TABLE candidates ADD COLUMN resumeupdatedat TIMESTAMP;
```

Since CON-01 explicitly excludes resume version history (latest-resume-only model), this is intentionally a 1:1 relationship on `candidates` directly — not a separate history table, which would over-build against an explicitly excluded requirement.

**Local folder access — frontend-only, not schema.** Local file access from a web application is governed by the browser's File System Access API. A folder handle, once granted via a native Browse dialog, can be persisted client-side (in the browser's IndexedDB) and remembered across sessions — but the underlying read/write *permission* is re-confirmed on each new session as a browser security measure, not assumed indefinitely. Design decision: **on each login, if a remembered folder handle exists, prompt the recruiter with a Yes/No confirmation to re-grant access**; "No" allows Browsing to a different folder instead. This is entirely frontend application state — the local folder path is never sent to or stored by the backend. Only the Google Drive reference is a database concern, since Drive is itself a cloud service the backend can address directly (`resumesource = 'gdrive'`, `resumereference` = Drive file ID).

**Requirements baseline status: complete.** No open items remain pending as of this section.

## 13. RBAC, Licensing & Payment Infrastructure

**Status: design-level decisions for a future commercial/multi-tenant licensing capability. Not in scope for the current v0.2.0 milestone — captured here to preserve the reasoning, following the same discipline as §11/§12.**

### 13.1 Current RBAC state (as verified against the actual codebase, not assumed)

The `roles` table defines four roles (admin, manager, recruiter, team_leader) each with a `permissions` JSONB array — but this data is **not currently enforced anywhere in the code**. `RoleMiddleware` only checks the role *name* string against a per-route allow-list; the `permissions` column is read nowhere. Only two route groups are gated at all (`/api/admin/*`, `/api/manager/*`), and `/api/manager/*` has zero routes registered under it. Every other resource (candidates, jobs, interviews, daily_jobs, business_dev) is accessible to any authenticated user regardless of role. No resource-ownership concept exists (e.g. no `created_by_user_id` on `jobs`).

### 13.2 Commercial licensing model

- **Single-user license**: one seat, role = `admin` (already a strict superset of all other role permissions — no new role needed).
- **Multi-seat license**: `admin`, `manager`, `recruiter`, `team_leader` as distinct seats. **Every company, at any seat count, must always include at least one admin or manager.**
- This floor is satisfied structurally, not by an active check: **Admin is permanently non-deletable, by any role, under any circumstance** (see 13.3), which guarantees a company can never reach zero admins.

### 13.3 Delete/role hierarchy

| Role | Can delete | Can be deleted by |
|---|---|---|
| Admin | Manager, Recruiter, Team Leader | No one |
| Manager | Recruiter, Team Leader | Admin only |
| Recruiter, Team Leader | Nothing | Admin or Manager |

**Gap against current code**: `/api/manager/*` has no delete routes implemented at all today — this capability needs to be built, not just re-gated.

### 13.4 Job ownership, derived from RBAC permissions (not a separate assumption)

The `manage_jobs` permission belongs to `manager` (and `admin`, via `["all"]`); `recruiter`/`team_leader` hold only `view_jobs`. This means **jobs are a manager-owned artifact by design** — directly resolving who Flow E's (§13.6) per-user auto-matching settings and cleanup should be scoped to.

```sql
ALTER TABLE jobs ADD COLUMN created_by_user_id INTEGER REFERENCES users(id);
```

**Gap against current code**: `/api/jobs` POST is not currently restricted to `manager`/`admin` — this needs to be enforced to match the RBAC design's own intent, not treated as optional.

### 13.5 License/subscription state

```sql
ALTER TABLE companies ADD COLUMN seat_limit INTEGER NOT NULL DEFAULT 1;
ALTER TABLE companies ADD COLUMN license_expires_at TIMESTAMP;
ALTER TABLE companies ADD COLUMN license_active BOOLEAN NOT NULL DEFAULT TRUE;
```

State machine: T-7 days before `license_expires_at` → warning popup (informational only). At `license_expires_at`, if unrenewed → `license_active` set to `FALSE`, all active sessions for that company terminated, 30-day retrieval window begins (retrieval requires payment — see 13.7). At T+30 days unretrieved → permanent deletion of the company's data.

**Decision: the boolean is stored and explicitly transitioned by a scheduled process, not computed live on each check.** A purely computed value (`expires_at > NOW()`) has no natural point to hang the warning/deletion *events* off of — those require an actual state transition, not a continuously re-derived fact. A lightweight sanity check against `expires_at` at login (log-only, not blocking) is recommended as a defensive layer in case the scheduled process ever fails to run.

**Explicit, deliberate exception to the "nothing runs automatically" principle (§13.8)**: license expiry warning, lock, and deletion are time-triggered and must fire whether or not any user takes an action — this is the one case in the whole design where a scheduled background process is required, not optional.

### 13.6 Auto-update settings (resume scan, JD matching) — per-user, not global

Both are token-costing AI operations, default OFF, user-controlled via Settings:
- `auto_resume_update_enabled` — scoped to whichever recruiter/manager/admin enables it (governs Flow A incremental folder scan)
- `auto_jd_matching_enabled` — scoped via `jobs.created_by_user_id` (§13.4), since only managers/admins own jobs

```sql
CREATE TABLE user_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    auto_resume_update_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_jd_matching_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_cleanup_enabled BOOLEAN NOT NULL DEFAULT FALSE
);
```

A third, independent toggle — `auto_cleanup_enabled` — governs the token-free `job_candidate_matches` expiry cleanup (§12/§13.4). Although this operation costs no tokens, the same "don't run anything without a reason" principle applies: even free background work is wasted during genuinely idle periods (holidays, slow stretches), so this remains user-controlled rather than unconditionally scheduled.

### 13.7 Payment infrastructure — do not build custom

**Decision: do not hand-build payment processing, subscription state machines, or invoicing logic from scratch.** This class of problem carries asymmetric risk — failure modes (tax non-compliance, silent over/undercharging) are invisible until they surface as legal or financial liabilities, not caught by ordinary code review. This is explicitly the same category of mistake as building a custom RTOS when the customer only wanted the Bluetooth stack on top of it — solving an already-solved problem nobody is asking for, at the cost of the actual differentiated product.

**Selected approach: self-hosted, license-fee-free open source over hosted vendor SaaS**, driven by the constraint that this product must compete in a tough pricing market — any vendor cost structured as a percentage of revenue (Kinde's 0.7%, Stripe's ~3.6%+30¢) scales against SkillSifter forever as it grows, while self-hosted infrastructure is a bounded, mostly-fixed cost.

- **RBAC**: extend the existing `roles`/`RoleMiddleware` foundation directly (§13.1 gaps) — **selected approach, in progress**. Casdoor (Apache 2.0, Go-based, self-hosted, Casbin-backed) remains a documented fallback if a more complete IAM feature set is needed later, but is not currently being adopted.
- **Billing/subscription/invoicing — selected: Zoho Books/Billing**, not self-hosted Lago. Reasoning: GST compliance is native to the product, not something to configure correctly ourselves — recurring invoices are generated as full GST-compliant tax invoices with automatic CGST/SGST vs IGST breakup, removing the domain-knowledge risk flagged below as otherwise unavoidable. Zoho is India-headquartered (Chennai); paying in INR to a domestic supplier is a domestic procurement (CGST/SGST) rather than an import of services, avoiding the added compliance layer (equalisation levy, RBI remittance forms) that comes with a foreign vendor. Zoho Billing already integrates UPI directly, pairing cleanly with the UPI-as-primary-rail decision below. A genuinely free tier exists (Zoho Books, businesses under ₹25 lakh annual revenue, no time limit) covering R K Consulting's current stage; paid tiers beyond that are flat fees, not revenue-scaling percentages, consistent with the pricing-competitiveness goal.

  Trade-off acknowledged: this is a hosted vendor, not self-hosted/owned source — a departure from the "borrow the code, run it ourselves" pattern used elsewhere (AstraMind, Casdoor). Accepted specifically because GST/tax correctness carries asymmetric risk (see below) that outweighs the self-hosting preference in this one case.
- **Payment rail — UPI as primary, not a card gateway**: bank-to-bank UPI carries **government-mandated zero MDR** (RBI/NPCI policy, in place since 2020) — no percentage transaction cut, a materially better cost structure than any card-based gateway (Stripe, or card-based Razorpay) for a price-competitive product. A small MDR (under 0.5%) for *large* merchants (~₹1-1.5 crore+ annual turnover) was under government discussion as of March 2026 but not enacted as of the most recent budget cycle, and would not apply to SkillSifter at its current or near-term scale regardless. A gateway (Razorpay or similar) would still typically sit alongside UPI for settlement/reconciliation and to support non-UPI paying customers, but the underlying transaction cost drops close to zero.
- **GST/tax compliance**: with Zoho selected specifically for its native GST handling, this risk is substantially mitigated rather than left as an open domain-knowledge burden — though a short accountant/CA review of the actual configuration (correct GSTIN setup, invoice series, applicable rates) remains sensible before real invoicing begins, as final validation rather than as the primary compliance mechanism.

**Infrastructure constraint, verified, not assumed — now narrowed in scope**: the earlier concern that Lago's self-hosted stack (Postgres + Redis + API + frontend, ~1-2GB+ RAM minimum) would not fit the current 512MB droplet no longer applies to billing, since Zoho is hosted by Zoho, not self-hosted. It remains a relevant constraint if Casdoor or any other self-hosted component is adopted later — the droplet's headroom should be reassessed against whatever is actually chosen to self-host, not assumed sufficient.

### 13.8 Explicit operating principle carried through this entire section

**Nothing runs automatically unless the user explicitly asks for it, with one deliberate, named exception**: license expiry warning/lock/deletion (§13.5), which must be time-triggered regardless of user action. Every other automation discussed (resume folder update, JD auto-matching, match-data cleanup) is opt-in via per-user Settings, default OFF, precisely because these operations carry either a token cost or a resource cost that should not be incurred without a clear reason to.