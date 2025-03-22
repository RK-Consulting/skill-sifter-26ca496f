
# Skill Sifter Backend

This is the Go backend for the Skill Sifter application.

## Prerequisites

- Go 1.18 or later
- PostgreSQL 12 or later

## Database Setup

1. Create a PostgreSQL database named `skillsifter`
2. Run the schema from `database/schema.sql`

## Configuration

The application uses environment variables for configuration:

- `PORT`: Server port (default: 8080)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: skillsifter)

## Running the Backend

```bash
# Navigate to the backend directory
cd backend

# Download dependencies
go mod download

# Run the server
go run main.go
```

The server will start on the configured port (default 8080).

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login

### Candidates
- `GET /api/candidates` - Get all candidates
- `POST /api/candidates` - Add new candidate
- `GET /api/candidates/{id}` - Get candidate by ID

### Jobs
- `GET /api/jobs` - Get all jobs
- `POST /api/jobs` - Add new job
- `GET /api/jobs/{id}` - Get job by ID

### Daily Jobs
- `GET /api/daily-jobs` - Get all daily job assignments
- `POST /api/daily-jobs` - Add new daily job assignment
- `GET /api/daily-jobs/{id}` - Get daily job assignment by ID

### Interviews
- `GET /api/interviews` - Get all interviews
- `POST /api/interviews` - Schedule new interview
- `GET /api/interviews/{id}` - Get interview by ID
