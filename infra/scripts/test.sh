#!/usr/bin/env bash
# Build and test both frontend and backend. Exits non-zero on the first
# failure, so this is safe to use as a pre-deploy gate.
#
# Usage: bash infra/scripts/test.sh
#
# Optional env vars for the backend integration test (candidate_handlers_test.go):
#   TEST_DB_HOST, TEST_DB_PORT, TEST_DB_USER, TEST_DB_PASSWORD, TEST_DB_NAME
# If no test database is reachable, that specific test SKIPS (not fails) —
# everything else still runs.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "=================================================="
echo "  BACKEND"
echo "=================================================="
cd "$REPO_ROOT/backend"

echo "--> gofmt check (fails if any file is not gofmt-formatted)"
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "The following files are not gofmt-formatted:"
  echo "$UNFORMATTED"
  echo "Run: gofmt -w <file> to fix, then commit."
  exit 1
fi

echo "--> go vet ./..."
go vet ./...

echo "--> go build ./..."
go build ./...

echo "--> go test ./... -v"
go test ./... -v

echo ""
echo "=================================================="
echo "  FRONTEND"
echo "=================================================="
cd "$REPO_ROOT/frontend"

echo "--> npm ci"
npm ci --no-audit --no-fund

echo "--> npm run lint"
npm run lint

echo "--> npm test"
npm test

echo "--> npm run build"
npm run build

echo ""
echo "=================================================="
echo "  ALL CHECKS PASSED"
echo "=================================================="