# SkillSifter

**Open-source, multi-tenant Applicant Tracking System (ATS) for staffing and recruitment teams.**

[![Backend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/RK-Consulting/skill-sifter-26ca496f/actions/workflows/frontend-ci.yml)
[![Status](https://img.shields.io/badge/status-active%20development-blue.svg)](#project-status)

SkillSifter is a full-stack recruitment platform designed to help staffing organizations manage candidates, requirements, recruitment assignments, interviews, and related operational workflows in a secure multi-tenant environment.

The project combines a **Go REST API**, **React + TypeScript frontend**, and **PostgreSQL** database, with automated backend and frontend CI through GitHub Actions.

> **Current release:** `v0.3.0`  
> **Current development:** Recruitment Assignment domain and state-machine work  
> **Project status:** Active development

---

## ✨ What is SkillSifter?

SkillSifter is being built as a practical ATS platform for staffing and recruitment organizations.

The platform is centered around a few core concepts:

- **Candidates** — maintain candidate profiles and recruitment information
- **Requirements** — manage client/job requirements
- **Recruitment assignments** — connect candidates with requirements through controlled assignment workflows
- **Candidate expertise** — maintain structured technical and language expertise
- **Interviews** — manage interview-related recruitment activity
- **Business development** — maintain recruitment business-development information
- **Multi-tenancy** — isolate organizational data by tenant
- **Role-based access control** — control access according to user roles and permissions
- **Resume processing** — support resume ingestion, parsing and recruiter-assisted search
- **Auditability** — maintain important operational history and assignment activity

The architecture is intentionally modular so individual domains can evolve without turning the application into a tightly coupled monolith.

---

## 🚀 Current Capabilities

### Candidate Management

- Candidate creation and management
- Candidate profile information
- Tenant-scoped candidate access
- Candidate status management
- Structured technical expertise
- Structured language expertise
- Resume association

### Recruitment Requirements

- Requirement management
- Client/requirement domain separation
- Requirement status lifecycle
- Structured requirement information
- Tenant-scoped requirement access

### Recruitment Assignments

The recruitment-assignment domain is currently under active development.

Current work includes:

- Candidate-to-requirement assignment
- Assignment lifecycle
- Assignment state transitions
- Tenant isolation
- Actor validation
- Assignment audit records
- Immutable candidate snapshots
- Immutable requirement snapshots
- Transactional state changes

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

### CI

The repository has separate CI pipelines for the two major application layers:

- **Backend CI** — Go formatting, build, `go vet`, and tests
- **Frontend CI** — dependency installation, ESLint, Vitest tests, and production build

Pull requests targeting `main` are protected by required backend and frontend CI checks.

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
│   └── project/
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

The repository also provides a combined test/build gate:

```bash
bash infra/scripts/test.sh
```

The goal is simple:

> **If it does not pass locally, it should not pass CI.**

---

## 🔄 Continuous Integration

SkillSifter uses separate GitHub Actions workflows for maintainability.

### Backend CI

```text
Pull Request / Push
        ↓
   Go formatting
        ↓
      Build
        ↓
       Vet
        ↓
      Tests
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

Both checks are required for changes targeting `main`.

---

## 🔐 Development & Branch Protection

The `main` branch is protected.

Changes are expected to follow:

```text
Feature Branch
      ↓
Pull Request
      ↓
Backend CI ──────┐
                 ├──→ All checks pass
Frontend CI ─────┘
      ↓
Review
      ↓
Merge
```

Direct unverified changes to `main` are intentionally restricted.

The project keeps governance lightweight: automation is used where it provides a clear engineering benefit without creating unnecessary process overhead.

---

## 📚 Documentation

Additional project documentation is organized under `docs/`.

Useful areas include:

- Architecture
- Product scope
- Architecture decisions
- Engineering principles
- Release process
- Versioning
- Project backlog

The documentation evolves alongside the implementation.

For API details, see:

```text
backend/docs/swagger.yaml
```

---

## 📌 Project Status

SkillSifter is **active development software**.

The project is progressing incrementally toward a production-ready V1 platform.

Current development is focused on strengthening the recruitment domain, particularly:

- Assignment lifecycle
- Assignment state transitions
- Candidate/requirement snapshots
- Tenant isolation
- Auditability
- Structured candidate expertise
- CI enforcement

The repository should therefore be considered **pre-1.0** and subject to API, schema and architectural changes.

---

## 🗺️ Roadmap

The project is being developed in incremental checkpoints rather than attempting to implement the entire product in one release.

Broad priorities include:

- [x] Core ATS foundation
- [x] Authentication and authorization
- [x] Multi-tenant foundation
- [x] Candidate management
- [x] Requirement domain foundation
- [x] Structured candidate expertise foundation
- [x] Recruitment assignment foundation
- [x] Assignment tenant isolation
- [x] Assignment audit foundation
- [x] Backend CI
- [x] Frontend CI
- [x] Protected `main`
- [ ] Complete recruitment assignment state machine
- [ ] Production hardening
- [ ] Expanded automated test coverage
- [ ] V1 release preparation
- [ ] Production release

The detailed implementation roadmap is maintained separately from this README.

---

## 📦 Releases

SkillSifter follows:

- **Semantic Versioning**
- **Keep a Changelog**
- Incremental development checkpoints

See:

- [`CHANGELOG.md`](CHANGELOG.md)
- [`docs/project/`](docs/project/)

### Current Release

**v0.3.0**

The current development branch contains work beyond the last documented release and is intentionally not represented as a new release until the corresponding checkpoint is completed and released.

---

## 🤝 Contributing

Contributions, bug reports and engineering discussions are welcome.

Before contributing, please read:

- [`CONTRIBUTING.md`](CONTRIBUTING.md)
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md)

For architectural changes, review the existing project documentation before introducing new patterns or dependencies.

---

## 📄 License

A project license has not yet been finalized.

Until a license is added to the repository, the source code should not be assumed to be available for unrestricted redistribution or commercial use.

---

## ⭐ Project

**SkillSifter**

An open-source foundation for modern staffing and recruitment operations.

Built with:

**Go · React · TypeScript · PostgreSQL · Docker · GitHub Actions**
