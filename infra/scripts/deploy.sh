#!/usr/bin/env bash
# Repeatable deploy. Run by hand on the server whenever you want to test:
#   bash infra/scripts/deploy.sh
# Deploys whichever branch is currently checked out on this server —
# does NOT force-switch branches. Check out the branch you want first.
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$APP_DIR"

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
RELEASE_VERSION="$(cat "$APP_DIR/VERSION" 2>/dev/null || echo "unknown")"
RELEASE_REVISION="$(git rev-parse HEAD)"
echo "==> SkillSifter release: v${RELEASE_VERSION} (${RELEASE_REVISION})"
echo "==> Pulling latest ${CURRENT_BRANCH}"
git fetch origin
git reset --hard "origin/${CURRENT_BRANCH}"
RELEASE_VERSION="$(cat "$APP_DIR/VERSION" 2>/dev/null || echo "unknown")"
RELEASE_REVISION="$(git rev-parse HEAD)"
export SKILLSIFTER_VERSION="$RELEASE_VERSION"
export SKILLSIFTER_REVISION="$RELEASE_REVISION"
echo "==> Deploying SkillSifter v${RELEASE_VERSION} (${RELEASE_REVISION})"

echo "==> Checking live Nginx configuration for drift"
LIVE_NGINX="/etc/nginx/sites-available/api.skillsifter.in"
REPO_NGINX="$APP_DIR/infra/nginx/api.skillsifter.in.conf"
if [ -f "$LIVE_NGINX" ] && ! cmp -s "$REPO_NGINX" "$LIVE_NGINX"; then
  echo "❌ DEPLOY ABORTED: live Nginx configuration differs from Git."
  echo "Live: $LIVE_NGINX"
  echo "Git:  $REPO_NGINX"
  echo "Review the difference, reconcile the intended change into Git,"
  echo "then run deploy.sh again. The live service was NOT touched."
  diff -u "$REPO_NGINX" "$LIVE_NGINX" || true
  exit 1
fi
echo "✅ Nginx configuration matches Git"

echo "==> Loading backend environment for deployment and integration tests"
if [ ! -f "$APP_DIR/backend/.env" ]; then
  echo "❌ DEPLOY ABORTED: backend/.env is missing."
  echo "The live service was NOT touched."
  exit 1
fi

set -a
source "$APP_DIR/backend/.env"
set +a

: "${TEST_DB_HOST:=${DB_HOST:-localhost}}"
: "${TEST_DB_PORT:=${DB_PORT:-5432}}"
: "${TEST_DB_USER:=${DB_USER:-}}"
: "${TEST_DB_PASSWORD:=${DB_PASSWORD:-}}"

if [ -z "$TEST_DB_USER" ] || [ -z "$TEST_DB_PASSWORD" ]; then
  echo "❌ DEPLOY ABORTED: TEST_DB_USER/TEST_DB_PASSWORD are not configured."
  echo "Set them in backend/.env or provide DB_USER/DB_PASSWORD there."
  echo "The live service was NOT touched."
  exit 1
fi

export TEST_DB_HOST TEST_DB_PORT TEST_DB_USER TEST_DB_PASSWORD

echo "==> Integration test database: ${TEST_DB_HOST}:${TEST_DB_PORT} as ${TEST_DB_USER}"
echo "==> Running backend test gate (fmt, vet, test) before touching the live service"
cd "$APP_DIR/backend"
go mod download

UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "❌ DEPLOY ABORTED: the following files are not gofmt-formatted:"
  echo "$UNFORMATTED"
  echo "The live service was NOT touched. Fix formatting, push, and redeploy."
  exit 1
fi

if ! go vet ./...; then
  echo "❌ DEPLOY ABORTED: go vet failed. The live service was NOT touched."
  exit 1
fi

if ! go test ./...; then
  echo "❌ DEPLOY ABORTED: tests failed. The live service was NOT touched."
  exit 1
fi

echo "✅ Test gate passed — proceeding with build and deploy"

echo "==> Building backend"
go build -o skillsifter .

echo "==> Syncing nginx config"
cp "$APP_DIR/infra/nginx/api.skillsifter.in.conf" /etc/nginx/sites-available/api.skillsifter.in
nginx -t
systemctl reload nginx

echo "==> Syncing systemd unit"
cp "$APP_DIR/infra/systemd/skillsifter.service" /etc/systemd/system/skillsifter.service
sed -i -e "s|__APP_DIR__|$APP_DIR|g" -e "s|__APP_VERSION__|$RELEASE_VERSION|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload

echo "==> Restarting service"
systemctl restart skillsifter
sleep 2
systemctl status skillsifter --no-pager

echo "==> Health check"
curl -sf http://localhost:8081/api/health-check && echo ""

echo "==> Production E2E bootstrap route check"
bootstrap_status="$(curl -sS -o /dev/null -w '%{http_code}' -X OPTIONS http://localhost:8081/api/e2e/bootstrap)"
if [ "$bootstrap_status" != "200" ] && [ "$bootstrap_status" != "204" ]; then
  echo "❌ DEPLOY FAILED: /api/e2e/bootstrap is not available for OPTIONS (HTTP $bootstrap_status)."
  echo "The deployed backend does not contain the production smoke bootstrap route."
  exit 1
fi
echo "✅ Production E2E bootstrap route is available (HTTP $bootstrap_status)"

echo "==> Deploy succeeded"
