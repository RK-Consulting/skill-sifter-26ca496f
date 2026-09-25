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

`POST /api/resume-ai/upload` accepts the `files` multipart field. PDF, DOCX and plain text files are supported. The backend hashes every file, prevents duplicate uploads per company, extracts text, sends only the extracted text to Ollama, and stores structured name/email/phone/skills in PostgreSQL.

Each parsed resume is linked to a candidate. Technical skills are persisted in the authoritative `candidate_expertise` table with category `resume_import` and proficiency `unspecified`. The retired `skills` and `candidate_skills` tables are no longer used by Resume AI.

Scanned/image PDFs without an embedded text layer are marked `failed` with a clear OCR-required error instead of creating unreliable candidate data.

## AI search

`GET /api/resume-ai/search?q=Java%20Spring%20AWS` searches candidate name, email, phone, skills and normalized skills. Search activity is recorded in `resume_search_logs` and `activity_logs`.

## Reporting

The reporting subsystem uses `activity_logs`, populated by database triggers for candidate, daily-task, interview, business-development and resume changes. Legacy Jobs activity behavior is removed by a forward migration because Jobs were replaced by Requirements. Available period reports are:

- daily — last 30 days
- monthly — last 12 months
- quarterly — last 8 quarters
- yearly — last 5 years

The UI presents operational tables instead of relying only on charts. It also includes a detailed activity log.

## Important limitation

The initial PDF extractor intentionally uses the Go standard library only. Text-based PDFs are supported through common PDF text-stream patterns, while scanned/image PDFs are flagged for OCR. A future OCR worker can be added without changing the resume database contract.


## Tenant isolation

Resume records carry the authoritative `tenant_id` from the authenticated request context. Duplicate detection, repository listing and candidate/expertise joins are tenant-scoped. `company_name` remains compatibility/display data and is not the tenant isolation boundary.

Migration `017_resume_ai_tenant_isolation.sql` backfills existing resume rows, enforces `tenant_id`, replaces the legacy company/hash uniqueness boundary with tenant/hash uniqueness, and removes the obsolete Jobs activity trigger without modifying historical migrations.
