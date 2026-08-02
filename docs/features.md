# SkillSifter — Feature Specification & Implementation Status

This documents the full functional scope of SkillSifter as currently designed, cross-checked against the actual codebase (not just what renders). Status is based on direct inspection of `src/` and `backend/`, not assumption.

**Status key:**
- ✅ **Live** — wired to the real API, verified working
- ⚠️ **Partial / Mixed** — some parts real, some hardcoded — see note
- 🔴 **Mock only** — hardcoded data, no real API call at all
- ❓ **Unverified** — exists in code, not yet manually tested end-to-end

---

## A. Dashboard (`src/components/dashboard/Dashboard.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Welcome message | ✅ Live | Pulls logged-in user's name |
| 2 | Total Candidates (stat) | ✅ Live | Real API call, `isLoading` state handled |
| 3 | Active Jobs (stat) | ✅ Live | Same pattern as above |
| 4 | Daily Tasks (stat) | ✅ Live | Same pattern |
| 5 | Business Contacts (stat) | ✅ Live | Same pattern |
| 6 | Hiring Trend (bar graph) | ❓ Unverified | Backend endpoint (`GET /api/reports/hiring`) exists and is real — need to confirm frontend actually renders its response rather than a placeholder |
| 7 | Candidate Sources (pie chart) | ❓ Unverified | Backend endpoint (`GET /api/reports/sources`) exists and is real — same caveat as above |
| 8 | Upload Files button | ❓ Unverified | UI present; unclear if resume upload has real backend handling yet, or is a placeholder for the future AI/AstraMind resume-parsing feature (see `docs/architecture.md` §11) |
| 9 | Recruitment Pipeline (Screening/Interview/Rejected counts) | 🔴 **Mock only** | Hardcoded `pipelineData` array (`24`, `12`, `8`) — no API call at all, not even a fallback pattern |
| 10 | Recent Activity feed | 🔴 **Mock only** | Hardcoded `activityData` array (Sarah Wilson, TechSolutions Inc, Alex) — no API call, "View All" button destination unverified |

## B. Candidates (`src/pages/Candidates.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Search by skill set (text + filter) | ❓ Unverified | UI present; need to confirm search actually filters via API vs. client-side only |
| 2 | Add Candidate button → form | ❓ Unverified | Form exists (`AddCandidate.tsx`); need to confirm submit wires to `POST /api/candidates` |
| 3 | Results table | ❓ Unverified | Need to confirm columns render real data, not placeholders |

## C. Jobs (`src/pages/Jobs.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Search by skill/designation | ❓ Unverified | |
| 2 | Add Job button → form | ❓ Unverified | |
| 3 | Results table (Title, Department, Location, Status, Date Posted, Actions) | ❓ Unverified | |

## D. Daily Tasks (`src/pages/DailyJobs.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Search by assignee | ❓ Unverified | |
| 2 | Add Assignment button → form | ❓ Unverified | |
| 3 | Results table (JD No., Instructions, Assigned To, Assigned Date, Actions) | ❓ Unverified | |

## E. Business Development (`src/pages/BusinessDev.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Search Clients | ❓ Unverified | |
| 2 | Add Client button → form | ❓ Unverified | |
| 3 | Results table (Client Name, Partner, Contact Person, Contact Info, Added, Actions) | ⚠️ **Partial** | Real API call attempted first; **falls back to hardcoded mock data only if the API call fails** — this is the correct defensive pattern, unlike the Dashboard bug above. Worth confirming the API call is actually succeeding in production (not silently failing into mock every time) |

## F. Interviews (`src/pages/Interviews.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Search Interview Schedules | ❓ Unverified | |
| 2 | Add/Schedule Interview button → form | ❓ Unverified | |
| 3 | Results table (Candidate, Position, Date & Time, Status, Feedback, Actions) | ❓ Unverified | |

## G. Reports & Analytics (`src/pages/Reports.tsx`)

| # | Feature | Status | Note |
|---|---|---|---|
| 1 | Dashboard-style layout | ❓ Unverified | |
| 2 | Total Candidates, Total Interviews, Interview Rate, Hiring Trend, Candidate Sources | ❓ Unverified | Same underlying `/api/reports/*` endpoints as Dashboard §A.6–7 — if those are confirmed working here, likely also fixable on the Dashboard the same way |

## Global

| Feature | Status | Note |
|---|---|---|
| Logout | ❓ Unverified | |
| Global search (search icon, top nav) | ❓ Unverified | Unclear what this searches — candidates only, or across all resource types |
| Add Candidate (top nav shortcut) | ❓ Unverified | Likely same form as B.2 |

---

## Immediate action items surfaced by this review

1. **Fix Dashboard §A.9 and §A.10** — replace hardcoded `pipelineData`/`activityData` arrays with real API calls. Highest-priority bug found here: it's the first thing any user (or investor, or pilot customer) sees, and it's currently lying about the data.
2. **Verify Dashboard §A.6–7 (charts)** — confirm they render real report data; the backend endpoints exist, so this may already work correctly, or may need frontend wiring.
3. **Systematic pass through B–G**: this document currently has a lot of ❓ — each of B.1–G.3 needs a real click-through test to convert unverifieds into confirmed ✅/⚠️/🔴, ideally logged with actual console errors from the DevTools review already in progress.