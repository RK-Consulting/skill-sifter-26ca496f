# Resume AI Architecture & Requirements

**Status:** Approved design baseline  
**Scope:** Resume ingestion, extraction, candidate intelligence, search, and future requirement matching  
**Repository:** SkillSifter  
**Baseline:** main at 78aac6467faee8e480d9a230f5c4a00488e533b1

## 1. Purpose

Resume AI is a core recruitment-intelligence subsystem of SkillSifter. Its purpose is to transform resume documents into structured candidate intelligence that can be searched, reviewed, and eventually matched against Requirements.

The subsystem must not become an isolated AI demonstration. The target workflow is:

```text
Resume
  -> Document Processing
  -> Structured Candidate Intelligence
  -> Requirement Matching
  -> Explainable Evidence
  -> Recruiter Decision
```

The existing local-Ollama ingestion foundation is retained and evolved rather than replaced.

## 2. Current State

The current implementation provides:

- Bulk folder resume upload.
- PDF, DOCX, TXT and Markdown input at the UI level.
- SHA-256 duplicate detection per company.
- Local resume file storage.
- Text extraction.
- Local Ollama extraction of name, email, phone and technical skills.
- Candidate association by email/phone within tenant scope.
- Technical expertise persistence through `candidate_expertise`.
- Resume parsing status and error persistence.
- Resume repository listing.
- Keyword search across candidate identity and expertise.
- Ollama health endpoint.
- Resume search/activity logging.
- A separate candidate-specific resume upload path.

The current implementation is a foundation, not the final Resume Intelligence experience.

## 3. Architectural Principles

1. **AI extracts; the system structures; rules match; the recruiter decides.**
2. Resume text must remain local to the configured Ollama service unless a future architecture explicitly changes the privacy boundary.
3. Tenant isolation is mandatory for every Resume AI read/write operation.
4. Candidate expertise remains the authoritative technical-skill storage model.
5. Resume AI and candidate-specific resume upload remain separate workflows.
6. AI output is evidence, not unquestioned truth.
7. Requirement matching is a separate responsibility from resume extraction.
8. Matching results must be explainable.
9. Historical database migrations must not be rewritten; corrective changes use new migrations.
10. The first matching implementation should prefer deterministic, inspectable criteria before semantic/embedding-based matching.

## 4. Target Domain Model

### 4.1 Resume

A Resume represents a source document and its processing lifecycle.

Existing core fields include:

- id
- company/tenant association
- candidate association
- file name/path/hash
- MIME type
- extracted text
- parsing status
- parser model
- parse error
- uploaded by
- uploaded timestamp
- parsed timestamp

Future Resume Intelligence may add provenance and extraction metadata without replacing this foundation.

### 4.2 Candidate Intelligence

Candidate intelligence is structured information derived from a resume and/or recruiter input.

Target areas:

- Identity and contact
- Location
- Professional profile
- Employment history
- Total/relevant experience
- Technical expertise
- Languages
- Education
- Certifications
- Projects

The current `candidate_expertise` model remains the authoritative technical-expertise mechanism.

### 4.3 Requirement

Requirement is the recruitment-demand object and is identified by the immutable Job ID.

Relevant matching attributes include:

- Client
- Job ID
- Job Type
- Job Title
- Department
- Experience Required
- Budget
- Language Requirements
- Certifications Required
- Notice Period
- Mode of Work
- Mandatory Requirements
- Job Description
- Status
- Job Location
- Open Positions

Resume AI must consume Requirements; it must not recreate a parallel Jobs model.

## 5. Processing Lifecycle

The logical Resume lifecycle is:

```text
uploaded
  -> queued
  -> processing
  -> extracted
  -> parsed
  -> candidate_linked
  -> completed
```

Failure conditions must be distinguishable, including:

- unsupported format
- corrupt/invalid file
- oversize file
- duplicate
- no extractable text
- OCR required
- text extraction failure
- Ollama unavailable
- model/API error
- invalid AI response
- candidate association failure

The current implementation may continue using its existing persisted states initially; the lifecycle above is the target contract.

## 6. AI Extraction Contract

The current contract extracts:

```json
{
  "name": "",
  "email": "",
  "phone": "",
  "skills": []
}
```

This contract is intentionally small and should remain backward compatible while the subsystem is stabilized.

The target extraction contract will progressively support:

- identity
- contact
- location
- professional summary
- employment history
- total/relevant experience
- technical skills
- languages
- education
- certifications
- projects

AI rules:

- Never invent information.
- Missing information remains empty/unknown.
- Technical skills must be normalized consistently.
- Language names must not be incorrectly stored as technical skills.
- Structured output must be validated before persistence.
- Parser/model identity must be retained for traceability.

## 7. Candidate Association

Bulk Resume AI may identify or create candidates from extracted identity information.

Association must:

1. Stay inside the authenticated tenant.
2. Prefer exact email matching when present.
3. Use phone matching when appropriate.
4. Avoid cross-tenant matches.
5. Avoid overwriting non-empty authoritative candidate fields with empty AI values.
6. Preserve the resume-to-candidate relationship.

Candidate-specific resume upload remains a separate exact-association path because its candidate is already known.

## 8. Expertise Model

`candidate_expertise` is the authoritative technical expertise store.

The historical `skills` and `candidate_skills` model was removed as unused schema. Resume AI must not reintroduce it.

Current Resume AI writes imported technical skills using:

- category = `resume_import`
- proficiency = `unspecified`

Future provenance can add source resume and confidence information when justified by an explicit schema change.

## 9. Search

### Current search

Keyword search covers candidate:

- name
- email
- phone
- technical expertise

### Target search evolution

Search should progress through three layers:

1. Structured/keyword search.
2. Requirement-aware filtering.
3. Optional semantic search.

Semantic/vector infrastructure is not required for the initial implementation.

## 10. Requirement Matching

Matching is a separate subsystem from parsing.

Target flow:

```text
Requirement / Job ID
      |
      v
Matching criteria
      |
      +-- Skills
      +-- Experience
      +-- Languages
      +-- Certifications
      +-- Location
      +-- Notice period
      +-- Work mode
      |
      v
Candidate evidence
      |
      v
Explainable match result
      |
      v
Recruiter review
```

The first matching version should expose evidence such as:

- Matched
- Missing
- Unknown
- Not applicable

A black-box percentage score must not be the primary user experience.

## 11. Explainability

Every future match should be explainable from stored candidate and requirement evidence.

Example:

```text
Skills
  Go             Matched
  PostgreSQL     Matched
  Docker         Matched

Experience
  Required: 5 years
  Candidate: 7 years
  Result: Matched

Language
  English        Matched

Location
  Bangalore      Matched
```

The recruiter must be able to understand why a candidate was surfaced.

## 12. API Direction

Existing endpoints remain compatible during transition:

- POST /resume-ai/upload
- GET /resume-ai/search
- GET /resume-ai/resumes
- GET /resume-ai/health

Future API concepts include:

- GET /resume-ai/resumes/{id}
- GET /resume-ai/resumes/{id}/download
- POST /resume-ai/resumes/{id}/retry
- GET /candidates/{id}/intelligence
- GET /requirements/{jobID}/matches
- GET /requirements/{jobID}/matches/{candidateID}

Exact endpoint naming is subject to the API conventions used elsewhere in SkillSifter.

## 13. UI Direction

The current Resume AI page is retained as the initial ingestion/search surface.

The target experience is:

```text
Resume AI
  |-- Resume Repository
  |-- Upload
  |-- Search
  |-- Resume Detail
  |-- Candidate Intelligence
  |-- Requirement Matching
```

Resume repository should eventually expose:

- Candidate
- Resume
- Processing status
- Upload date
- Expertise
- Error state
- View/download
- Retry where appropriate

Resume detail should expose extracted intelligence and processing metadata.

## 14. Security Requirements

Mandatory controls:

- Tenant-scoped reads and writes.
- Authorization on resume viewing/downloading.
- Safe file names and storage paths.
- File size limits.
- No cross-tenant candidate association.
- Audit of material Resume AI operations.
- No accidental cloud transmission of resume text under the local-Ollama design.

## 15. Reliability Requirements

The system must handle:

- partial failures in multi-file uploads
- Ollama unavailable
- malformed AI output
- unsupported/corrupt documents
- OCR-required documents
- duplicate uploads
- candidate association failures
- database failures

A single bad resume must not invalidate successful processing of unrelated files in the same batch.

## 16. Testing Requirements

### Backend

Test:

- file size limits
- supported formats
- extraction failures
- duplicate detection
- Ollama success/failure/malformed output
- candidate association
- tenant isolation
- expertise persistence
- search behavior
- batch partial failure

### Frontend

Test:

- initial loading
- upload
- upload failures
- duplicate/error rendering
- Ollama status
- search
- empty repository
- result rendering

### End-to-end

The target recruitment workflow is:

```text
Client
  -> Requirement
  -> Resume upload
  -> Resume processing
  -> Candidate creation/link
  -> Expertise
  -> Search/match
  -> Recruiter review
  -> Assignment
  -> Interview
```

## 17. Database Migration Rules

Do not edit historical migrations.

The existing historical AI migration contains legacy Jobs activity-trigger references. Because Jobs have been replaced by Requirements, any remaining obsolete schema behavior must be corrected with a new forward migration.

Documentation must reflect the current `candidate_expertise` model and must not describe the removed `skills`/`candidate_skills` tables as active storage.

## 18. Processing Architecture Evolution

The current upload path processes files synchronously. This is acceptable for the stabilization stage.

For larger batches, the target architecture is:

```text
Upload
  -> create Resume records
  -> queue processing
  -> Resume worker
  -> extraction
  -> AI parsing
  -> candidate association
  -> expertise persistence
```

A background worker should be introduced only after the domain and API contracts are stable.

## 19. Implementation Roadmap

### RAI-01 — Architecture baseline
- This document.
- Correct Resume AI documentation.
- Establish API/domain terminology.

### RAI-02 — Reliability and test foundation
- Dedicated backend Resume AI tests.
- Ollama parsing tests.
- Upload/duplicate/tenant tests.
- Frontend Resume AI tests.
- Preserve existing behavior.

### RAI-03 — Resume intelligence
- Expand structured extraction.
- Add appropriate persistence for experience, education, certifications, languages and projects.
- Preserve provenance.

### RAI-04 — Resume repository
- Resume detail.
- Processing state.
- Error/retry UX.
- Secure view/download.

### RAI-05 — Candidate intelligence
- Integrate Resume intelligence into candidate profile.

### RAI-06 — Requirement matching
- Deterministic matching against Job ID Requirements.
- Explainable evidence.

### RAI-07 — Semantic matching
- Optional embeddings/semantic retrieval.
- Hybrid matching.
- Explainability retained.

## 20. Definition of Done for RAI-02

RAI-02 is complete when:

- Resume AI backend has dedicated automated tests.
- Upload and duplicate behavior is covered.
- Tenant isolation is covered.
- Ollama success/failure/malformed responses are covered.
- Candidate association behavior is covered.
- Search behavior is covered.
- Frontend Resume AI behavior has automated coverage appropriate to the existing frontend test framework.
- Existing full CI remains green.
- No regression is introduced in Candidate, Requirement, Assignment or Interview workflows.
- Documentation accurately describes the current implementation.

## 21. Non-goals for RAI-02

RAI-02 does not include:

- vector database
- semantic matching
- autonomous recruiter decisions
- cloud LLM integration
- complete Resume Detail redesign
- background worker implementation
- large-scale UI redesign

Those belong to later roadmap phases.


## 21. Tenant Isolation and Persistence Boundary

Resume records are tenant-owned resources. The authenticated `tenant_id` is authoritative for resume duplicate detection, listing, candidate joins, search, and expertise persistence. `company_name` remains compatibility/display data and is not an isolation key.

Migration `017_resume_ai_tenant_isolation.sql` backfills `resumes.tenant_id`, makes it mandatory, changes duplicate uniqueness from `(company_name, file_hash)` to `(tenant_id, file_hash)`, and removes the obsolete legacy Jobs activity trigger through a forward migration. Historical migrations remain immutable.
