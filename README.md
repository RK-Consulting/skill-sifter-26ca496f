# SkillSifter

**Published version: 0.2.0** · [Changelog](CHANGELOG.md)

The approved V1 target architecture, scope, governance rules, and ADR process are documented in [`docs/architecture/v1-baseline.md`](docs/architecture/v1-baseline.md), [`docs/product/v1-scope.md`](docs/product/v1-scope.md), [`CODEX_ENGINEERING_RULES.md`](CODEX_ENGINEERING_RULES.md), and [`docs/architecture/ADRs/`](docs/architecture/ADRs/). The exact current-release designation is tracked for reconciliation in [`ISSUES.md`](ISSUES.md).

SkillSifter is a multi-tenant Applicant Tracking System (ATS) for staffing and recruitment teams. It lets a recruiting company manage candidates, job openings, daily job assignments, interviews, and business-development leads, all scoped per company (tenant) with role-based access control.

The project is a full-stack app: a **React + TypeScript** SPA frontend (in `frontend/`) and a **Go** REST API backend (in `backend/`), backed by **PostgreSQL**.

## Features

Status reflects what's actually verified working as of v0.2.0, not just what's designed — see [`docs/features.md`](docs/features.md) for the full page-by-page audit.

- ✅ **Authentication & Authorization** — JWT-based login/register, role-based access (`admin`, `manager`, `recruiter`, `team_leader`) enforced via middleware, with a delete/edit hierarchy (Admin undeletable, Manager deletable only by Admin)
- ✅ **Multi-tenancy** — every core resource scoped by `company_name`
- ✅ **Candidate management** — add, view, update, delete; fields include position, location, experience, CTC, notice period, skills
- ✅ **Job management** — post and track job openings, scoped to Manager/Admin roles (`manage_jobs` permission)
- ✅ **Daily job tracking** — assign and track daily tasks, with a real assignee dropdown
- ✅ **Business development** — track clients, partners, and contacts
- ✅ **Dashboard: Recent Activity** — real, timestamp-sorted feed across candidates/jobs/business-dev/daily-tasks/interviews
- ⚠️ **Interview scheduling** — display and scheduling not yet verified working, under investigation
- ⚠️ **Reports** — Total Candidates works; source-distribution report currently errors (missing schema column); hiring-trend chart wiring not yet confirmed
- ⚠️ **Dashboard: charts & pipeline** — Hiring Trend chart, Candidate Sources chart, and Recruitment Pipeline widget are not yet wired to real data
- **Dashboard UI** — built with shadcn/ui, Radix primitives, Tailwind CSS, and Recharts

## Tech Stack

**Frontend**
- React 18 + TypeScript
- Vite (build tool/dev server)
- React Router
- TanStack Query (React Query) for data fetching/caching
- Axios for HTTP requests
- shadcn/ui + Radix UI + Tailwind CSS
- React Hook Form + Zod for forms/validation
- Recharts for charts

**Backend**
- Go 1.21
- `gorilla/mux` for routing
- `golang-jwt/jwt` for JWT auth
- `lib/pq` (PostgreSQL driver)
- `rs/cors` for CORS handling
- PostgreSQL database

**Infra**
- Docker & Docker Compose (Postgres + Go API + Nginx-served frontend)
- Nginx (serves the built frontend, handles SPA routing + gzip/caching)

## Project Structure

```
.
├── backend/                  # Go REST API
│   ├── auth/                 # JWT auth & role middleware
│   ├── db/                   # DB connection + env helpers
│   ├── database/
│   │   ├── schema.sql         # Legacy full-schema file (see migrations/ below)
│   │   └── migrations/        # Versioned schema migrations, applied in order by deploy.sh
│   ├── docs/                  # Swagger spec + design doc
│   ├── handlers/               # HTTP handlers (candidates, jobs, interviews, etc.), with tests
│   ├── models/                 # Data models
│   ├── main.go                  # Entry point, routes, server bootstrap
│   └── Dockerfile
├── frontend/                  # React SPA
│   ├── src/
│   │   ├── components/         # UI components (dashboard, layout, shadcn ui)
│   │   ├── hooks/                # Custom hooks
│   │   ├── pages/                # Route-level pages (Candidates, Jobs, Interviews, Reports, ...)
│   │   ├── services/             # API client (Axios) + service wrappers
│   │   └── test/                  # Vitest setup + tests
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── infra/                     # Deployment infrastructure, versioned
│   ├── nginx/                  # Reverse proxy config
│   ├── systemd/                 # Service unit
│   └── scripts/
│       ├── bootstrap.sh          # One-time server setup
│       ├── deploy.sh             # Repeatable deploy — runs the test gate first
│       └── test.sh               # Full backend + frontend test/build gate
├── docs/                       # Architecture decisions, feature audit
├── docker-compose.yml           # Postgres + backend + frontend services (local dev)
├── CHANGELOG.md
└── README.md
```

## API Overview

The backend exposes a JSON REST API (mounted both at root and under `/api` for compatibility).

**Public routes**
- `GET /`, `GET /api` — API welcome message
- `GET /health-check`, `GET /ping` — health checks
- `POST /auth/register` — register a new user
- `POST /auth/login` — log in, returns a JWT

**Protected routes** (require `Authorization: Bearer <token>`)
- `GET/POST /api/candidates`, `GET/PUT/DELETE /api/candidates/{id}`
- `GET/POST /api/jobs`, `GET/PUT/DELETE /api/jobs/{id}`
- `GET/POST /api/daily-jobs`, `GET/PUT/DELETE /api/daily-jobs/{id}`
- `GET/POST /api/interviews`, `GET/PUT/DELETE /api/interviews/{id}`
- `GET/POST /api/business-dev`, `GET/PUT/DELETE /api/business-dev/{id}`
- `GET /api/reports/hiring`, `GET /api/reports/sources`

**Admin-only routes** (role: `admin`)
- `GET/POST /api/admin/users`, `PUT/DELETE /api/admin/users/{id}`

A Swagger spec is available at `backend/docs/swagger.yaml`.

## Data Model

Core tables (see `backend/database/schema.sql`):
- `companies` — tenants
- `roles` — predefined roles with permission sets
- `users` — belongs to a company, has a role
- `candidates`, `jobs`, `daily_jobs`, `interviews`, `business_dev` — all scoped by `company_name`

## Getting Started

### Option 1: Docker Compose (recommended)

This spins up Postgres, the Go API, and the Nginx-served frontend together.

```sh
git clone <this-repo-url>
cd skill-sifter-26ca496f
docker compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Postgres: localhost:5432 (db: `skillsifter`, user/pass: `postgres`/`postgres`)

The Postgres container automatically runs `backend/database/schema.sql` on first boot.

> Note: the `docker-compose.yml` frontend build sets `VITE_API_URL=https://api.skillsifter.in` by default — override this env var for local development if you want the containerized frontend to talk to your local backend instead.

### Option 2: Run locally without Docker

**Prerequisites:** Node.js & npm, Go 1.21+, PostgreSQL.

**Backend**
```sh
cd backend
# Set env vars (or rely on defaults in db/env.go), e.g.:
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=skillsifter
export PORT=8080

go mod download
go run main.go
```
The backend calls `db.InitDB()` and `db.InitializeSchema()` on startup, so it will create/verify the schema automatically.

**Frontend**
```sh
cd frontend
npm install
# Create a .env file with:
# VITE_API_URL=http://localhost:8080
npm run dev
```
The frontend will be available at http://localhost:5173.

## Testing

Run the full backend + frontend gate (gofmt, vet, build, test on the backend; lint, test, build on the frontend) in one command:
```sh
bash infra/scripts/test.sh
```
This is the same gate `deploy.sh` runs automatically before touching the live service on every deploy.

## Configuration

The backend reads configuration from environment variables (with fallbacks defined in `backend/db/env.go`):

| Variable      | Purpose                        | Default (compose) |
|---------------|---------------------------------|--------------------|
| `PORT`        | API server port                | `8080`             |
| `DB_HOST`     | Postgres host                  | `postgres`         |
| `DB_PORT`     | Postgres port                  | `5432`             |
| `DB_USER`     | Postgres user                  | `postgres`         |
| `DB_PASSWORD` | Postgres password              | `postgres`         |
| `DB_NAME`     | Postgres database name         | `skillsifter`      |

The frontend reads `VITE_API_URL` at build time to know where the API lives.

## Security Notes

A few things worth flagging before using this in production:

- **JWT secret is hardcoded** in `backend/auth/auth.go` (`skill_sifter_secret_key`). This should be moved to an environment variable before any real deployment.
- **CORS origins are hardcoded** in `backend/main.go` (`skillsifter.in` domains + localhost). Update as needed for your deployment.
- A compiled binary (`backend/skillsifter`) appears to be checked into the repo — consider adding it to `.gitignore`.

## License

No license file is currently present in the repository. Add one (e.g. MIT, Apache-2.0) if you intend to open-source this project.
