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
LIVE_RATE_LIMITS="/etc/nginx/conf.d/skillsifter-rate-limits.conf"
REPO_RATE_LIMITS="$APP_DIR/infra/nginx/skillsifter-rate-limits.conf"
if [ -f "$LIVE_RATE_LIMITS" ] && ! cmp -s "$REPO_RATE_LIMITS" "$LIVE_RATE_LIMITS"; then
  echo "DEPLOY ABORTED: live Nginx rate-limit configuration differs from Git."
  echo "Live: $LIVE_RATE_LIMITS"
  echo "Git:  $REPO_RATE_LIMITS"
  diff -u "$REPO_RATE_LIMITS" "$LIVE_RATE_LIMITS" || true
  exit 1
fi
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

echo "==> Preparing least-privilege runtime account and storage"
if ! getent group skillsifter >/dev/null 2>&1; then
  groupadd --system skillsifter
fi
if ! id -u skillsifter >/dev/null 2>&1; then
  useradd --system --gid skillsifter --home-dir /var/lib/skillsifter --create-home --shell /usr/sbin/nologin skillsifter
fi
install -d -o skillsifter -g skillsifter -m 0750 /var/lib/skillsifter/resumes
chown skillsifter:skillsifter /var/lib/skillsifter /var/lib/skillsifter/resumes
chmod 0750 /var/lib/skillsifter /var/lib/skillsifter/resumes
chown root:skillsifter "$APP_DIR/backend/.env"
chmod 0640 "$APP_DIR/backend/.env"

echo "==> Building backend"
go build -o skillsifter .

echo "==> Syncing nginx rate-limit policy and site config"
cp "$APP_DIR/infra/nginx/skillsifter-rate-limits.conf" /etc/nginx/conf.d/skillsifter-rate-limits.conf
cp "$APP_DIR/infra/nginx/api.skillsifter.in.conf" /etc/nginx/sites-available/api.skillsifter.in
nginx -t
systemctl reload nginx

echo "==> Syncing systemd unit"
cp "$APP_DIR/infra/systemd/skillsifter.service" /etc/systemd/system/skillsifter.service
sed -i -e "s|__APP_DIR__|$APP_DIR|g" -e "s|__APP_VERSION__|$RELEASE_VERSION|g" -e "s|__APP_REVISION__|$RELEASE_REVISION|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload

echo "==> Restarting service"
systemctl restart skillsifter
sleep 2
systemctl status skillsifter --no-pager

echo "==> Health and deployed revision check"
health_payload="$(curl --fail --silent --show-error --max-time 10 http://localhost:8081/api/health-check)"
printf '%s\n' "$health_payload"
if ! python3 -c 'import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get("status") == "OK" and d.get("version") == sys.argv[1] and d.get("revision") == sys.argv[2] else 1)' "$RELEASE_VERSION" "$RELEASE_REVISION" <<< "$health_payload"; then
  echo "DEPLOY FAILED: health endpoint version/revision does not match this deployment."
  echo "Expected version=$RELEASE_VERSION revision=$RELEASE_REVISION"
  exit 1
fi

echo "==> Production E2E reset preflight contract check"
status="$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -X OPTIONS "http://localhost:8081/api/e2e/reset" || true)"
if [ "$status" != "204" ]; then
  echo "DEPLOY FAILED: /api/e2e/reset OPTIONS returned HTTP ${status:-unknown}; expected 204."
  echo "The service restarted, but production E2E reset prerequisite is not healthy."
  exit 1
fi
echo "Production E2E /api/e2e/reset OPTIONS preflight is healthy (HTTP $status)"

echo "==> Deploy succeeded"
