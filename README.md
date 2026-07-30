# SkillSifter

SkillSifter is a multi-tenant Applicant Tracking System (ATS) for staffing and recruitment teams. It lets a recruiting company manage candidates, job openings, daily job assignments, interviews, and business-development leads, all scoped per company (tenant) with role-based access control.

The project is a full-stack app: a **React + TypeScript** SPA frontend and a **Go** REST API backend, backed by **PostgreSQL**.

## Features

- **Authentication & Authorization** — JWT-based login/register, with role-based access (`admin`, `manager`, `recruiter`, `team_leader`) enforced via middleware.
- **Multi-tenancy** — Every core resource (candidates, jobs, interviews, daily jobs, business dev) is scoped by `company_name`, so multiple recruiting firms can use the same deployment in isolation.
- **Candidate management** — Add, view, update, and delete candidates with fields like position, location, experience, CTC (current/expected), notice period, language, and skills.
- **Job management** — Post and track job openings (title, department, location, status, description, requirements).
- **Daily job tracking** — Assign daily job descriptions/instructions to users.
- **Interview scheduling** — Schedule interviews against candidates, track status and feedback.
- **Business development** — Track clients, partners, and contacts for BD pipeline.
- **Reports** — Aggregated hiring reports (interviews per month) and source reports.
- **Dashboard UI** — Built with shadcn/ui, Radix primitives, Tailwind CSS, and Recharts for data visualization.

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
│   ├── database/             # SQL schema (schema.sql)
│   ├── docs/                 # Swagger spec + design doc
│   ├── handlers/             # HTTP handlers (candidates, jobs, interviews, etc.)
│   ├── models/                # Data models
│   ├── main.go                # Entry point, routes, server bootstrap
│   └── Dockerfile
├── src/                      # React frontend
│   ├── components/           # UI components (dashboard, layout, shadcn ui)
│   ├── hooks/                 # Custom hooks
│   ├── pages/                 # Route-level pages (Candidates, Jobs, Interviews, Reports, ...)
│   ├── services/              # API client (Axios) + service wrappers
│   └── App.tsx                 # Routes & app shell
├── docker-compose.yml         # Postgres + backend + frontend services
├── Dockerfile                 # Frontend build (Vite → Nginx)
├── nginx.conf
└── package.json
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
npm install
# Create a .env file with:
# VITE_API_URL=http://localhost:8080
npm run dev
```
The frontend will be available at http://localhost:5173.

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
