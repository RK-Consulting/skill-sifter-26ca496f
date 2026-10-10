# AI Resume + Reporting

## Local Ollama

SkillSifter can process resumes with a local Ollama instance. No resume text is sent to a cloud LLM by this feature.

1. Install Ollama locally.
2. Start Ollama.
3. Pull the configured model, for example `ollama pull llama3.1:8b`.
4. Start SkillSifter with `OLLAMA_URL=http://127.0.0.1:11434` for a non-containerized backend.
5. For Docker Compose, the backend defaults to `http://host.docker.internal:11434`.

Optional environment variables:

- `OLLAMA_URL` — Ollama base URL.
- `OLLAMA_MODEL` — extraction model, default `llama3.1:8b`.
- `RESUME_STORAGE_PATH` — local resume storage, default `./storage/resumes` outside Docker.

## Resume ingestion

`POST /api/resume-ai/upload` accepts the `files` multipart field. PDF, DOCX and plain text files are supported. The backend hashes every file, prevents duplicate uploads per tenant, extracts text, sends only the extracted text to Ollama, and stores structured name/email/phone/skills in PostgreSQL.

Each parsed resume is linked to a candidate. Technical skills are persisted in the authoritative `candidate_expertise` table with category `resume_import` and proficiency `unspecified`. The retired `skills` and `candidate_skills` tables are no longer used by Resume AI.

Scanned/image PDFs without an embedded text layer are marked `failed` with a clear OCR-required error instead of creating unreliable candidate data.

## AI search

`GET /api/resume-ai/search?q=Java%20Spring%20AWS` searches candidate name, email, phone, skills and normalized skills. Search activity is recorded in the tenant-scoped `resume_search_logs` table.

## Reporting

Reporting reads tenant-owned recruitment/resume records and the application-owned `audit_events` table. Recruitment mutations write explicit audit events in Go; business workflow meaning is not delegated to database triggers. Legacy Jobs, daily-jobs, business-development, and `activity_logs` trigger infrastructure are not part of the authoritative tenant-plane baseline. Available period reports are:

- daily — last 30 days
- monthly — last 12 months
- quarterly — last 8 quarters
- yearly — last 5 years

The UI presents operational tables instead of relying only on charts. It also includes a detailed activity log.

## Important limitation

The initial PDF extractor intentionally uses the Go standard library only. Text-based PDFs are supported through common PDF text-stream patterns, while scanned/image PDFs are flagged for OCR. A future OCR worker can be added without changing the resume database contract.


## Tenant isolation

Resume records carry the authoritative `tenant_id` from the authenticated request context. Duplicate detection, repository listing and candidate/expertise joins are tenant-scoped. `company_name` is control-plane descriptive data and is not used as the tenant isolation boundary.

The authoritative tenant baseline requires `resumes.tenant_id` and uses tenant/hash uniqueness. The clean baseline is initialized from `backend/database/tenant-plane/001_baseline.sql`; historical migration files are retained only as historical evidence and are not replayed to create new tenant databases.
