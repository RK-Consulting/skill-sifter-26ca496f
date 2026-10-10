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
  echo "WARNING: live Nginx configuration differs from Git."
  echo "Preserving the live Nginx configuration for this deployment."
  diff -u "$REPO_NGINX" "$LIVE_NGINX" || true
  NGINX_DRIFT=true
else
  NGINX_DRIFT=false
  echo "Nginx configuration matches Git"
fi

echo "==> Loading backend environment for deployment and integration tests"
if [ ! -f "$APP_DIR/backend/.env" ]; then
  echo "DEPLOY ABORTED: backend/.env is missing. The live service was NOT touched."
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
  echo "DEPLOY ABORTED: TEST_DB_USER/TEST_DB_PASSWORD are not configured. The live service was NOT touched."
  exit 1
fi
export TEST_DB_HOST TEST_DB_PORT TEST_DB_USER TEST_DB_PASSWORD

echo "==> Integration test database: ${TEST_DB_HOST}:${TEST_DB_PORT} as ${TEST_DB_USER}"
echo "==> Running backend test gate (fmt, vet, test) before touching the live service"
cd "$APP_DIR/backend"
go mod download

UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "DEPLOY ABORTED: unformatted Go files:"
  echo "$UNFORMATTED"
  echo "The live service was NOT touched. Fix formatting, push, and redeploy."
  exit 1
fi

if ! go vet ./...; then
  echo "DEPLOY ABORTED: go vet failed. The live service was NOT touched."
  exit 1
fi
if ! go test ./...; then
  echo "DEPLOY ABORTED: tests failed. The live service was NOT touched."
  exit 1
fi

echo "Test gate passed — proceeding with build and deploy"
echo "==> Building backend"
go build -o skillsifter .

if [ "$NGINX_DRIFT" = "true" ]; then
  echo "==> Keeping existing live Nginx configuration"
else
  echo "==> Syncing nginx config"
  cp "$APP_DIR/infra/nginx/api.skillsifter.in.conf" /etc/nginx/sites-available/api.skillsifter.in
  nginx -t
  systemctl reload nginx
fi

echo "==> Syncing systemd unit"
cp "$APP_DIR/infra/systemd/skillsifter.service" /etc/systemd/system/skillsifter.service
sed -i -e "s|__APP_DIR__|$APP_DIR|g" -e "s|__APP_VERSION__|$RELEASE_VERSION|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload

echo "==> Restarting service"
systemctl restart skillsifter
sleep 2
systemctl status skillsifter --no-pager

echo "==> Health check"
curl --fail --silent --show-error --max-time 10 http://localhost:8081/api/health-check
echo ""

echo "==> Production E2E bootstrap and reset route contract checks"
for route in bootstrap reset; do
  status="$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -X OPTIONS "http://localhost:8081/api/e2e/$route" || true)"
  if [ "$status" != "204" ]; then
    echo "DEPLOY FAILED: /api/e2e/$route OPTIONS returned HTTP ${status:-unknown}; expected 204."
    echo "The service restarted, but production E2E prerequisites are not healthy."
    exit 1
  fi
  echo "Production E2E /api/e2e/$route route is available (HTTP $status)"
done

echo "==> Deploy succeeded"
