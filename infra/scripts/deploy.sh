#!/usr/bin/env bash
# Repeatable deploy. Run by hand on the server whenever you want to test:
#   bash infra/scripts/deploy.sh
# Deploys whichever branch is currently checked out on this server —
# does NOT force-switch branches. Check out the branch you want first.
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$APP_DIR"

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
echo "==> Pulling latest ${CURRENT_BRANCH}"
git fetch origin
git reset --hard "origin/${CURRENT_BRANCH}"

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

echo "==> Building backend Docker image"
GIT_SHA="$(git rev-parse --short HEAD)"
docker build -t "skillsifter:${GIT_SHA}" -t skillsifter:latest "$APP_DIR/backend"
echo "    tagged skillsifter:${GIT_SHA} and skillsifter:latest"

echo "==> Running database migrations"
source "$APP_DIR/backend/.env"
export PGPASSWORD="$DB_PASSWORD"

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename TEXT PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
);
SQL

for f in "$APP_DIR"/backend/database/migrations/*.sql; do
  fname="$(basename "$f")"
  already_applied=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tA \
    -c "SELECT 1 FROM schema_migrations WHERE filename = '${fname}'")
  if [ "$already_applied" != "1" ]; then
    echo "  applying $fname"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -f "$f"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 \
      -c "INSERT INTO schema_migrations(filename) VALUES ('${fname}')"
  else
    echo "  skipping $fname (already applied)"
  fi
done

echo "==> Syncing nginx config"
cp "$APP_DIR/infra/nginx/api.skillsifter.in.conf" /etc/nginx/sites-available/api.skillsifter.in
nginx -t
systemctl reload nginx

echo "==> Syncing systemd unit"
cp "$APP_DIR/infra/systemd/skillsifter.service" /etc/systemd/system/skillsifter.service
sed -i "s|__APP_DIR__|$APP_DIR|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload

echo "==> Restarting service"
systemctl restart skillsifter
sleep 2
systemctl status skillsifter --no-pager

echo "==> Health check"
curl -sf http://localhost:8081/health-check && echo "" && echo "==> Deploy succeeded"