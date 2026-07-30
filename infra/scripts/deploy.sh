#!/usr/bin/env bash
# Repeatable deploy. Run by GitHub Actions on every push to main.
# Can also be run by hand on the server: bash infra/scripts/deploy.sh
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$APP_DIR"

echo "==> Pulling latest main"
git fetch origin
git checkout main
git reset --hard origin/main

echo "==> Building backend"
cd "$APP_DIR/backend"
go mod download
go build -o skillsifter .

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