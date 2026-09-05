# SkillSifter

**Open-source, multi-tenant Applicant Tracking System (ATS) for staffing and recruitment teams.**

[![Backend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml)
[![Status](https://img.shields.io/badge/status-active%20development-blue.svg)](#project-status)

SkillSifter is a full-stack recruitment platform designed to help staffing organizations manage candidates, requirements, recruitment assignments, interviews, and related operational workflows in a secure multi-tenant environment.

The project combines a **Go REST API**, **React + TypeScript frontend**, and **PostgreSQL** database, with automated backend and frontend verification through GitHub Actions.

> **Current release:** `v0.5.3`  
> **Release focus:** CP11 Production Quality + CP12 Recruitment Workflow UAT / Go-Live Readiness  
> **Project status:** Active development / pre-1.0

---

## ✨ What is SkillSifter?

SkillSifter is being built as a practical ATS platform for staffing and recruitment organizations.

The platform is centered around the recruitment lifecycle:

- **Candidates** — maintain candidate profiles and recruitment information
- **Candidate expertise** — maintain structured technical and language expertise
- **Requirements** — manage client/job requirements
- **Recruitment assignments** — connect candidates with requirements through controlled assignment workflows
- **Screening and submission** — support recruiter-driven candidate progression
- **Interviews** — manage interview-related recruitment activity
- **Decision and commercial workflow** — support later-stage recruitment progression toward joining
- **Resume intelligence** — ingest, parse, and search resume information
- **Multi-tenancy** — isolate organizational data by tenant
- **Role-based access control** — control access according to user roles and permissions
- **Auditability** — preserve important operational history and assignment activity

The architecture is intentionally modular so individual domains can evolve without turning the application into an unnecessarily tightly coupled monolith.

---

## 🚀 Current Capabilities

### Candidate Management

- Candidate creation and management
- Candidate profile information
- Tenant-scoped candidate access
- Candidate status management
- Structured technical expertise
- Structured language expertise with proficiency levels
- Resume association and metadata
- Recruitment activity tracking

### Recruitment Requirements

- Requirement management
- Client/requirement domain separation
- Requirement status lifecycle
- Structured requirement information
- Tenant-scoped requirement access

### Recruitment Assignment

Recruitment assignments are now treated as a controlled domain workflow rather than a loosely coordinated set of handler operations.

Current capabilities include:

- Candidate-to-requirement assignment
- Controlled assignment lifecycle
- Validated assignment state transitions
- Tenant isolation
- Tenant-aware assignment actor authorization
- Assignment audit records
- Tenant-aware audit enforcement
- Immutable candidate snapshots
- Immutable requirement snapshots
- Transactional state changes
- Regression coverage around state, authorization, snapshots, and audit behavior

### Recruitment Workflow

The current product workflow is organized around the following progression:

```text
Candidate
    ↓
Candidate Expertise
    ↓
Requirement
    ↓
Assignment
    ↓
Screening
    ↓
Submission
    ↓
Interview
    ↓
Decision
    ↓
Commercial / Joining
```

The v0.5.3 release focuses on validating this workflow through automated regression coverage and focused recruiter UAT rather than introducing unnecessary browser-test or infrastructure complexity.

### Resume Intelligence

- Resume ingestion foundation
- Resume metadata
- Resume parsing status
- Extracted technical expertise
- Recruiter-assisted resume search
- Local AI/Ollama integration foundation

### Authentication & Authorization

- JWT-based authentication
- Role-based authorization
- Tenant-aware access control
- Protected API routes
- Administrative user management
- Tenant-aware assignment actor validation

### Audit & Tenant Isolation

Security-sensitive recruitment operations enforce tenant boundaries at the domain level.

This includes:

- Tenant-scoped candidate and requirement access
- Tenant-scoped assignment operations
- Assignment actor validation
- Cross-tenant actor rejection
- Tenant-aware assignment audit enforcement
- Regression coverage for cross-tenant access paths

### CI & Quality Gates

The repository maintains separate CI pipelines for the two major application layers:

- **Backend CI** — formatting, repository/migration structure checks, schema validation, build, `go vet`, tests, and coverage reporting
- **Frontend CI** — dependency installation, lint, tests, and production build

The CI quality gate is intentionally lean and focused on repeatable engineering verification.

---

## 🧱 Architecture

SkillSifter follows a modular full-stack architecture:

```text
┌─────────────────────────────────────────────┐
│                 Frontend                    │
│                                             │
│ React + TypeScript + Vite                   │
│ React Router + TanStack Query               │
│ shadcn/ui + Radix UI + Tailwind CSS        │
└──────────────────────┬──────────────────────┘
                       │
                       │ REST / JSON
                       ▼
┌─────────────────────────────────────────────┐
│                  Backend                    │
│                                             │
│ Go REST API                                 │
│ Authentication / Authorization              │
│ Domain services                             │
│ Recruitment workflows                       │
│ Resume processing                           │
│ Audit / tenant isolation                    │
└──────────────────────┬──────────────────────┘
                       │
                       │ SQL
                       ▼
┌─────────────────────────────────────────────┐
│                PostgreSQL                   │
│                                             │
│ Tenant data                                 │
│ Candidates                                  │
│ Requirements                                │
│ Assignments                                 │
│ Expertise                                   │
│ Resumes                                     │
│ Audit data                                  │
└─────────────────────────────────────────────┘
```

Database schema changes are maintained through versioned migrations under:

```text
backend/database/migrations/
```

The migration set is treated as the authoritative database schema source.

---

## 🛠️ Technology Stack

### Frontend

- React 18
- TypeScript
- Vite
- React Router
- TanStack Query
- Axios
- shadcn/ui
- Radix UI
- Tailwind CSS
- React Hook Form
- Zod
- Recharts
- Vitest

### Backend

- Go
- `gorilla/mux`
- PostgreSQL
- `lib/pq`
- JWT authentication
- CORS
- Standard Go testing tooling

### Infrastructure

- Docker
- Docker Compose
- PostgreSQL
- Nginx
- GitHub Actions
- Linux deployment tooling

### Local AI

- Ollama integration foundation for local resume-processing and recruiter-assisted intelligence workflows

---

## 📁 Repository Structure

```text
.
├── backend/
│   ├── auth/                       # Authentication and authorization
│   ├── db/                         # Database connection and migration support
│   ├── database/
│   │   └── migrations/             # Authoritative versioned database schema
│   ├── domain/
│   │   └── assignment/             # Recruitment assignment domain
│   ├── handlers/                   # HTTP/API handlers
│   ├── models/                     # Application data models
│   ├── docs/                       # Backend API documentation
│   ├── main.go                     # API entry point
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── components/             # Reusable UI components
│   │   ├── hooks/                  # React hooks
│   │   ├── pages/                  # Application pages
│   │   ├── services/               # API clients/services
│   │   └── test/                   # Frontend tests
│   ├── public/
│   ├── package.json
│   ├── package-lock.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── infra/
│   ├── nginx/
│   ├── systemd/
│   └── scripts/
│
├── docs/
│   ├── architecture/
│   ├── product/
│   ├── project/
│   └── release/
│
├── .github/
│   └── workflows/
│       ├── backend-ci.yml
│       └── frontend-ci.yml
│
├── docker-compose.yml
├── CHANGELOG.md
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── RELEASE_NOTES.md
└── README.md
```

---

## ⚡ Getting Started

### Prerequisites

For local development:

- Go
- Node.js
- npm
- PostgreSQL

Docker is recommended for running the complete application stack.

### Option 1 — Docker Compose

```bash
git clone https://github.com/RK-Consulting/skill-sifter-26ca496f.git
cd skill-sifter-26ca496f

docker compose up --build
```

The application stack consists of:

```text
Frontend → Nginx
              ↓
          Go API
              ↓
         PostgreSQL
```

### Option 2 — Run Backend Locally

```bash
cd backend

go mod download
go run main.go
```

The backend initializes the database and applies the versioned migrations during startup.

### Run Frontend Locally

```bash
cd frontend

npm install
npm run dev
```

The development frontend is served by Vite.

---

## 🧪 Testing

### Backend

From `backend/`:

```bash
go build ./...
go vet ./...
go test ./...
```

### Frontend

From `frontend/`:

```bash
npm ci
npm run lint
npm run test
npm run build
```

### Full Local Gate

The repository also provides a combined verification gate:

```bash
bash infra/scripts/test.sh
```

The engineering principle is simple:

> **If it does not pass locally, it should not pass CI.**

The v0.5.3 release continues to favor a small, repeatable quality gate over unnecessary testing infrastructure.

---

## 🔄 Continuous Integration

SkillSifter uses separate GitHub Actions workflows for maintainability.

### Backend CI

```text
Pull Request / Push
        ↓
 Repository checks
        ↓
 Migration/schema validation
        ↓
 Go formatting
        ↓
      Build
        ↓
       Vet
        ↓
      Tests
        ↓
    Coverage
        ↓
       PASS
```

Workflow:

```text
.github/workflows/backend-ci.yml
```

### Frontend CI

```text
Pull Request / Push
        ↓
     npm ci
        ↓
       Lint
        ↓
       Test
        ↓
      Build
        ↓
       PASS
```

Workflow:

```text
.github/workflows/frontend-ci.yml
```

Backend and frontend workflows remain separate so each application layer has an independent and maintainable verification pipeline.

---

## 🔐 Security & Multi-Tenancy

Multi-tenancy is a core architectural boundary of SkillSifter.

Security-sensitive recruitment operations are designed to prevent cross-tenant data and actor relationships.

Key controls include:

- Tenant-scoped data access
- Tenant-aware authorization
- Role-based access control
- Assignment actor tenant validation
- Cross-tenant actor rejection
- Tenant-aware audit relationships
- Regression coverage for tenant isolation

The application is pre-1.0 software, so security hardening and validation remain ongoing engineering responsibilities.

---

## 🌿 Development Workflow

The repository follows a feature-branch and pull-request workflow.

```text
Feature Branch
      ↓
Implementation
      ↓
Local Verification
      ↓
Pull Request
      ↓
Backend CI ──────┐
                 ├──→ Quality Gate
Frontend CI ─────┘
      ↓
Review
      ↓
Merge
      ↓
Release Checkpoint
```

The `main` branch is the release integration branch. Development branches may contain work for upcoming checkpoints and should not be assumed to represent the current production baseline.

Governance is intentionally lightweight: automation and process are used where they provide a clear engineering benefit without creating unnecessary overhead.

---

## 📚 Documentation

Project documentation is organized under `docs/` and evolves alongside the implementation.

Useful documentation areas include:

- Architecture
- Product scope
- Architecture Decision Records
- Engineering principles
- Release process
- Versioning policy
- Project backlog
- Release readiness

For API details, see:

```text
backend/docs/swagger.yaml
```

Release-specific documentation is maintained under:

```text
docs/release/
```

---

## 📌 Project Status

SkillSifter is **active development software** and remains **pre-1.0**.

The project has progressed beyond the initial ATS foundation into recruitment workflow hardening and production-readiness validation.

### v0.5.3 Status

The current release combines two checkpoints:

- **CP11 — Production Quality Gate**
- **CP12 — Recruitment Workflow UAT / Go-Live Readiness**

The release confirms the existing CI and regression infrastructure as the quality gate and adds a focused go-live readiness process for the recruiter workflow.

The final application-level gate remains manual recruiter UAT against the running application before production cutover.

### Current Engineering Priorities

- Production hardening
- Recruitment workflow validation
- Tenant isolation and authorization assurance
- Regression safety
- Focused recruiter UAT
- Production deployment readiness
- Evidence-driven reliability and performance improvements

Production observability, performance optimization, and additional infrastructure should be driven by real usage evidence rather than speculative engineering.

---

## 🗺️ Roadmap

SkillSifter is being developed through incremental checkpoints rather than attempting to implement the entire product in a single release.

### Foundation — Completed

- [x] Core ATS foundation
- [x] Authentication and authorization
- [x] Multi-tenant foundation
- [x] Candidate management
- [x] Requirement domain foundation
- [x] Structured candidate expertise foundation
- [x] Recruitment assignment foundation
- [x] Assignment tenant isolation
- [x] Assignment actor authorization
- [x] Assignment audit foundation
- [x] Assignment state-machine foundation
- [x] Candidate/requirement snapshot handling
- [x] Backend CI
- [x] Frontend CI
- [x] Protected `main`
- [x] Resume ingestion/search foundation

### Production Readiness — Current

- [x] CP11 Production Quality Gate
- [x] CP12 Recruitment Workflow UAT / Go-Live Readiness checkpoint
- [x] Backend regression quality gate
- [x] Frontend regression/build quality gate
- [x] Tenant isolation release validation
- [x] Assignment authorization release validation
- [ ] Final production recruiter UAT
- [ ] Production cutover

### Post-Go-Live / V1

- [ ] Production hardening based on real usage
- [ ] Operational observability based on real production needs
- [ ] Reliability improvements based on production evidence
- [ ] Expanded automated workflow coverage where justified
- [ ] V1 product scope completion
- [ ] V1 release preparation

The detailed implementation roadmap is maintained separately from this README.

---

## 📦 Releases

SkillSifter follows:

- **Semantic Versioning**
- **Keep a Changelog**
- Incremental development checkpoints

### Current Release — v0.5.3

**Released:** September 2, 2026

v0.5.3 combines CP11 Production Quality and CP12 Recruitment Workflow UAT / Go-Live Readiness.

The release is intentionally a **lean stabilization release**. It does not introduce a new observability platform, distributed tracing infrastructure, continuous-deployment system, or speculative performance framework.

No `v0.5.2` release was created; the CP11 and CP12 checkpoints were delivered together as v0.5.3.

### Release History

| Release | Focus |
|---|---|
| `v0.5.3` | CP11 Production Quality + CP12 Go-Live Readiness |
| `v0.5.1` | Recruitment workflow stabilization and verification |
| `v0.4.0` | Recruitment Assignment State Machine, tenant isolation, audit integrity, candidate expertise, CI foundations |
| `v0.3.0` | Resume intelligence foundation and V1 architecture/product baseline |
| `v0.2.0` | ATS stabilization, RBAC, testing infrastructure, and critical fixes |
| `v0.1.0` | Initial documented baseline |

See:

- [`CHANGELOG.md`](CHANGELOG.md)
- [`RELEASE_NOTES.md`](RELEASE_NOTES.md)
- [`docs/release/`](docs/release/)

---

## 🎯 Release Philosophy

SkillSifter follows a deliberately incremental engineering approach:

```text
LEAN
  ↓
STABLE
  ↓
LIVE
  ↓
OBSERVE
  ↓
FIX
  ↓
OPTIMIZE
  ↓
EXPAND
```

The project avoids introducing infrastructure or architectural complexity before there is evidence that the complexity is needed.

This principle is particularly important as SkillSifter moves from development checkpoints toward real production usage.

---

## 🤝 Contributing

Contributions, bug reports, documentation improvements, and engineering discussions are welcome.

Before contributing, please read:

- [`CONTRIBUTING.md`](CONTRIBUTING.md)
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md)

For architectural changes, review the existing project documentation and architecture decisions before introducing new patterns or dependencies.

Changes should normally be developed on a feature branch and submitted through a pull request with the applicable CI checks passing.

---

## 📄 License

A project license has not yet been finalized.

Until a license is added to the repository, the source code should **not** be assumed to be available for unrestricted redistribution or commercial use.

---

## ⭐ Project

**SkillSifter**

An open-source foundation for modern staffing and recruitment operations.

Built with:

**Go · React · TypeScript · PostgreSQL · Docker · GitHub Actions**
