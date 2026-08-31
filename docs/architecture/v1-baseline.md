# SkillSifter V1 Architecture Baseline

**Status:** Approved architecture baseline  
**Applies from:** v0.4.0  
**Target:** v1.0.0

## Purpose

SkillSifter is a multi-tenant Recruitment Operations Platform for small and mid-sized recruitment and staffing firms. It is operations-first: the system must make recruiters' daily work structured, traceable, and manageable. AI assists recruiters; it does not make hiring or submission decisions.

This document defines architectural boundaries, not feature-level implementation. Product scope belongs in `docs/product/v1-scope.md`; significant decisions require an ADR; implementation work belongs in a GitHub issue.

## Core workflow

```text
Lead → Opportunity → Client → Requirement
                                ↓
Candidate → Recruitment Assignment → Screening → Submission → Client review
                                                        ↓
                                                  Interviews → Offer → Joining
                                                                    ↓
                                               Invoice → Payment → Guarantee → Closure / Replacement
```

`Recruitment Assignment` is the central transaction. It connects one candidate to one client requirement. Candidate status describes the recruiter's current relationship with a candidate and must not replace the state of a specific assignment.

## V1 product boundary

```text
Dashboard
Recruitment: active requirements, candidate pool, submissions, interviews, offers, closures
Clients: clients, contacts, requirements, activity
Candidates: database, AI search, new candidates, requests
Business development: leads, opportunities, follow-ups
Commercial: invoices, payments, replacements
Reports
```

The system must not add major modules outside this boundary without an approved product decision.

## Target domain boundary

```text
tenants/companies, users, roles
clients, client_contacts, requirements
candidates, candidate_skills, candidate_languages, candidate_statuses
recruitment_assignments, screenings, submissions
interview_processes, interview_rounds, offers, joinings
invoices, payments, replacement_cases
leads, opportunities, follow_ups, activities
```

This is a conceptual model, not authorization to create every table at once. Tables, fields, relations, state transitions, and migrations are defined in approved milestone issues.

## Architectural constraints

- Use the existing React/TypeScript/Vite frontend, Go modular-monolith API, PostgreSQL, Nginx, and Docker architecture.
- Maintain tenant isolation for every tenant-owned operation. Cross-tenant access is a security defect.
- Treat SkillSifter as the authoritative system of record for operational data.
- Model language expertise generically as language + proficiency framework + proficiency level; do not make JLPT the universal data model.
- Preserve historical transactional facts where a later candidate or requirement update would otherwise change what was submitted or agreed.
- Base dashboard and reports on persisted, authoritative data. Do not display fabricated operational metrics.
- Keep external sourcing manual in V1.0. LinkedIn/job-portal scraping, portals, autonomous recruiting, elaborate analytics, and advanced automation are deferred to V2.0.

## AI boundary

Initial AI use cases are resume extraction and candidate-to-requirement matching, both subject to recruiter review. AstraMind, if used, remains a separate domain-agnostic service. SkillSifter performs domain interpretation and tenant scoping before sending data to it.

## Release sequence

| Milestone | Scope |
| --- | --- |
| v0.4 | Product and architecture foundation, target model, API/state/testing conventions, ADRs and governance |
| v0.5 | Leads, opportunities, clients, requirements |
| v0.6 | Candidate database, language expertise, resume handling, recruiter-assisted AI search/matching |
| v0.7 | Recruitment assignments, screening, interest, submissions, client review |
| v0.8 | Interview process, feedback, offers |
| v0.9 | Joining, invoices, payments, guarantee, replacement, closure |
| v0.9.1 | UAT and stabilization with R K Consulting |
| v0.9.2 | Production hardening: security, tenant isolation, migrations, backups, monitoring, regression |
| v1.0 | Production release |

## Governance

Architecture changes follow: problem → analysis → options → trade-offs → approved decision → ADR → GitHub issue → implementation. When an issue conflicts with this baseline, has a tenant/security concern, requires a breaking API/database change, or is ambiguous, implementation stops for an owner decision.

See [the Codex engineering rules](../../CODEX_ENGINEERING_RULES.md), [V1 scope](../product/v1-scope.md), and [ADRs](ADRs/README.md).
