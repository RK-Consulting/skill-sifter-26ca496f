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
  echo "DEPLOY ABORTED: live Nginx configuration differs from Git."
  echo "Live: $LIVE_NGINX"
  echo "Git:  $REPO_NGINX"
  echo "Review the diff, reconcile the intended change into Git, then redeploy."
  diff -u "$REPO_NGINX" "$LIVE_NGINX" || true
  exit 1
fi
echo "Nginx configuration matches Git"

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

# Load the production database environment before the test gate so integration
# tests use the same DB credentials as the application. Explicit TEST_DB_*
# variables still override these values inside the test bootstrap.
source "$APP_DIR/backend/.env"
export DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME

# Provision integration-test databases once. All database lifecycle changes
# are owned by this deploy script; the application DB role is not granted
# CREATEDB. Existing databases are left untouched.
ensure_test_db() {
  local db_name="$1"
  if sudo -u postgres psql -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$db_name'" | grep -q 1; then
    echo "==> Test database already exists: $db_name"
  else
    echo "==> Creating test database once: $db_name"
    sudo -u postgres psql -d postgres -v ON_ERROR_STOP=1 \
      -c "CREATE DATABASE \"$db_name\" OWNER \"$DB_USER\";"
  fi
}

ensure_test_db "${SKILLSIFTER_SCHEMA_TEST_DB:-skillsifter_schema_test}"
ensure_test_db "${SKILLSIFTER_HANDLER_TEST_DB:-skillsifter_handler_test}"

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
