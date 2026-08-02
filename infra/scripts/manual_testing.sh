#!/usr/bin/env bash
# Manual_testing.sh — API-level regression test, replacing manual UI data
# entry for verifying CRUD actually works end to end.
#
# Why API-level, not browser automation: every real bug found tonight
# (Candidates, Jobs, Daily Tasks) was a backend schema/handler mismatch —
# something that shows up identically whether you click through the UI or
# call the API directly. Testing at the API layer catches the same bugs
# faster and more reliably, with no browser needed.
#
# What this does NOT catch: frontend rendering bugs (e.g. the Daily Tasks
# blank-page crash from a missing route) — those still need a real browser.
# This is a complement to manual testing, not a full replacement.
#
# Requirements: curl, jq (sudo apt install jq if missing)
#
# Usage:
#   BASE_URL=https://api.skillsifter.in \
#   TEST_EMAIL=you@example.com \
#   TEST_PASSWORD=yourpassword \
#   bash infra/scripts/Manual_testing.sh
#
# Never hardcode real credentials in this file — always pass them via env vars.

set -uo pipefail  # no -e: we want to keep running and report every failure, not stop at the first

BASE_URL="${BASE_URL:-https://api.skillsifter.in}"
EMAIL="${TEST_EMAIL:-}"
PASSWORD="${TEST_PASSWORD:-}"

if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
  echo "ERROR: set TEST_EMAIL and TEST_PASSWORD environment variables before running."
  echo "  TEST_EMAIL=you@example.com TEST_PASSWORD=yourpassword bash $0"
  exit 1
fi

if ! command -v jq &> /dev/null; then
  echo "ERROR: jq is required but not installed. Install with: sudo apt install jq"
  exit 1
fi

LOG_FILE="manual_testing_$(date +%Y%m%d_%H%M%S).log"
PASS_COUNT=0
FAIL_COUNT=0

# Logs to both stdout and the log file
log() {
  echo "$1" | tee -a "$LOG_FILE"
}

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  log "  ✅ PASS: $1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  log "  ❌ FAIL: $1"
}

# Runs a curl request, logs the outcome, and returns the response body via
# the global RESPONSE variable (bash can't return strings directly).
api_call() {
  local method="$1" path="$2" data="${3:-}"
  if [ -n "$data" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$path" \
      -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$data")
  else
    RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$path" \
      -H "Authorization: Bearer $TOKEN")
  fi
  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | sed '$d')
}

log "=================================================="
log "  SkillSifter API Regression Test"
log "  Target: $BASE_URL"
log "  Started: $(date)"
log "=================================================="

# ---------- Login ----------
log ""
log "--> Login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
  fail "Login — could not obtain token. Response: $LOGIN_RESPONSE"
  log ""
  log "Cannot continue without a valid token. Exiting."
  exit 1
fi
pass "Login — token obtained"

# ---------- Candidates ----------
log ""
log "--> Candidates CRUD"

CANDIDATE_PAYLOAD='{
  "name": "Regression Test Candidate",
  "email": "regtest.candidate@example.com",
  "phone": "9999999999",
  "position": "Test Engineer",
  "location": "Bangalore",
  "experience": "3 years",
  "currentCTC": "8 LPA",
  "expectedCTC": "12 LPA",
  "noticePeriod": "30 days",
  "jlptLanguage": "N/A",
  "skills": "Testing, Automation"
}'

api_call POST "/api/candidates" "$CANDIDATE_PAYLOAD"
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  CANDIDATE_ID=$(echo "$BODY" | jq -r '.data.id // empty')
  if [ -n "$CANDIDATE_ID" ]; then
    pass "Create candidate — id=$CANDIDATE_ID"
  else
    fail "Create candidate — HTTP $HTTP_CODE but no id in response: $BODY"
  fi
else
  fail "Create candidate — HTTP $HTTP_CODE: $BODY"
fi

if [ -n "${CANDIDATE_ID:-}" ]; then
  api_call GET "/api/candidates"
  if echo "$BODY" | jq -e ".data[] | select(.id == $CANDIDATE_ID)" > /dev/null 2>&1; then
    pass "List candidates — created candidate appears in list"
  else
    fail "List candidates — created candidate NOT found in list"
  fi

  api_call PUT "/api/candidates/$CANDIDATE_ID" '{"name":"Regression Test Candidate Updated","email":"regtest.candidate@example.com","position":"Senior Test Engineer"}'
  if [ "$HTTP_CODE" = "200" ]; then
    pass "Update candidate — HTTP 200"
  else
    fail "Update candidate — HTTP $HTTP_CODE: $BODY"
  fi
fi

# ---------- Jobs ----------
log ""
log "--> Jobs CRUD (requires admin/manager role)"

JOB_PAYLOAD='{
  "title": "Regression Test Job",
  "department": "QA",
  "location": "Remote",
  "status": "open",
  "description": "Created by Manual_testing.sh",
  "requirements": "None — test data"
}'

api_call POST "/api/jobs" "$JOB_PAYLOAD"
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  JOB_ID=$(echo "$BODY" | jq -r '.data.id // empty')
  if [ -n "$JOB_ID" ]; then
    pass "Create job — id=$JOB_ID"
  else
    fail "Create job — HTTP $HTTP_CODE but no id in response: $BODY"
  fi
elif [ "$HTTP_CODE" = "403" ]; then
  fail "Create job — HTTP 403 (this account may not have manager/admin role)"
else
  fail "Create job — HTTP $HTTP_CODE: $BODY"
fi

if [ -n "${JOB_ID:-}" ]; then
  api_call GET "/api/jobs"
  if echo "$BODY" | jq -e ".data[] | select(.id == $JOB_ID)" > /dev/null 2>&1; then
    pass "List jobs — created job appears in list (this is the exact bug found earlier tonight)"
  else
    fail "List jobs — created job NOT found in list"
  fi
fi

# ---------- Daily Tasks ----------
log ""
log "--> Daily Tasks CRUD"

api_call GET "/api/company-users"
if [ "$HTTP_CODE" = "200" ]; then
  pass "Fetch company-users — route exists (this was completely missing earlier tonight)"
  ASSIGNEE_ID=$(echo "$BODY" | jq -r '.data[0].id // empty')
else
  fail "Fetch company-users — HTTP $HTTP_CODE: $BODY"
fi

if [ -n "${ASSIGNEE_ID:-}" ]; then
  DAILY_JOB_PAYLOAD="{\"jdNo\": 9999, \"instructions\": \"Regression test task\", \"assignedUser\": $ASSIGNEE_ID}"
  api_call POST "/api/daily-jobs" "$DAILY_JOB_PAYLOAD"
  if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    DAILY_JOB_ID=$(echo "$BODY" | jq -r '.data.id // empty')
    if [ -n "$DAILY_JOB_ID" ]; then
      pass "Create daily task — id=$DAILY_JOB_ID"
    else
      fail "Create daily task — HTTP $HTTP_CODE but no id: $BODY"
    fi
  else
    fail "Create daily task — HTTP $HTTP_CODE: $BODY"
  fi

  if [ -n "${DAILY_JOB_ID:-}" ]; then
    api_call GET "/api/daily-jobs"
    if echo "$BODY" | jq -e ".data[] | select(.id == $DAILY_JOB_ID)" > /dev/null 2>&1; then
      pass "List daily tasks — created task appears in list"
    else
      fail "List daily tasks — created task NOT found in list"
    fi
  fi
else
  fail "Skipping daily task create/list — no assignee available from company-users"
fi

# ---------- Business Development ----------
log ""
log "--> Business Development CRUD"

BUSDEV_PAYLOAD='{
  "clientName": "Regression Test Client",
  "partnerName": "Test Partner",
  "contactPerson": "Test Contact",
  "contactNumber": "9999999999",
  "contactEmail": "regtest.client@example.com"
}'

api_call POST "/api/business-dev" "$BUSDEV_PAYLOAD"
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  BUSDEV_ID=$(echo "$BODY" | jq -r '.data.id // empty')
  if [ -n "$BUSDEV_ID" ]; then
    pass "Create business contact — id=$BUSDEV_ID"
  else
    fail "Create business contact — HTTP $HTTP_CODE but no id: $BODY"
  fi
else
  fail "Create business contact — HTTP $HTTP_CODE: $BODY"
fi

# ---------- Interviews ----------
log ""
log "--> Interviews CRUD (using the candidate created above)"

if [ -n "${CANDIDATE_ID:-}" ]; then
  INTERVIEW_PAYLOAD="{\"candidateId\": $CANDIDATE_ID, \"candidateName\": \"Regression Test Candidate\", \"position\": \"Test Engineer\", \"interviewDate\": \"$(date -u -d '+1 day' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+1d +%Y-%m-%dT%H:%M:%SZ)\", \"status\": \"scheduled\"}"
  api_call POST "/api/interviews" "$INTERVIEW_PAYLOAD"
  if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    INTERVIEW_ID=$(echo "$BODY" | jq -r '.data.id // empty')
    if [ -n "$INTERVIEW_ID" ]; then
      pass "Schedule interview — id=$INTERVIEW_ID"
    else
      fail "Schedule interview — HTTP $HTTP_CODE but no id: $BODY"
    fi
  else
    fail "Schedule interview — HTTP $HTTP_CODE: $BODY"
  fi

  if [ -n "${INTERVIEW_ID:-}" ]; then
    api_call GET "/api/interviews"
    if echo "$BODY" | jq -e ".data[] | select(.id == $INTERVIEW_ID)" > /dev/null 2>&1; then
      pass "List interviews — created interview appears in list"
    else
      fail "List interviews — created interview NOT found in list"
    fi
  fi
else
  fail "Skipping interview test — no candidate id available from earlier step"
fi

# ---------- Cleanup (best-effort, delete in dependency order) ----------
log ""
log "--> Cleanup"

[ -n "${INTERVIEW_ID:-}" ] && api_call DELETE "/api/interviews/$INTERVIEW_ID" && \
  { [ "$HTTP_CODE" = "200" ] && pass "Delete test interview" || fail "Delete test interview — HTTP $HTTP_CODE"; }

[ -n "${BUSDEV_ID:-}" ] && api_call DELETE "/api/business-dev/$BUSDEV_ID" && \
  { [ "$HTTP_CODE" = "200" ] && pass "Delete test business contact" || fail "Delete test business contact — HTTP $HTTP_CODE"; }

[ -n "${DAILY_JOB_ID:-}" ] && api_call DELETE "/api/daily-jobs/$DAILY_JOB_ID" && \
  { [ "$HTTP_CODE" = "200" ] && pass "Delete test daily task" || fail "Delete test daily task — HTTP $HTTP_CODE"; }

[ -n "${JOB_ID:-}" ] && api_call DELETE "/api/jobs/$JOB_ID" && \
  { [ "$HTTP_CODE" = "200" ] && pass "Delete test job" || fail "Delete test job — HTTP $HTTP_CODE"; }

[ -n "${CANDIDATE_ID:-}" ] && api_call DELETE "/api/candidates/$CANDIDATE_ID" && \
  { [ "$HTTP_CODE" = "200" ] && pass "Delete test candidate" || fail "Delete test candidate — HTTP $HTTP_CODE"; }

# ---------- Summary ----------
log ""
log "=================================================="
log "  SUMMARY: $PASS_COUNT passed, $FAIL_COUNT failed"
log "  Full log: $LOG_FILE"
log "=================================================="

if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi
exit 0