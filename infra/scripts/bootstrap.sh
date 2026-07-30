#!/usr/bin/env bash
# One-time setup for a fresh droplet. Run manually, once, as root:
#   bash infra/scripts/bootstrap.sh
#
# After this runs, all future updates go through deploy.sh (via CI),
# never by hand-editing anything on the server again.
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_NAME="skillsifter"
DB_USER="skillsifter_user"

echo "==> App directory: $APP_DIR"

if [ -z "${DB_PASSWORD:-}" ]; then
  echo "ERROR: set DB_PASSWORD env var before running this script, e.g.:"
  echo "  DB_PASSWORD='...' JWT_SECRET='...' bash infra/scripts/bootstrap.sh"
  exit 1
fi

if [ -z "${JWT_SECRET:-}" ]; then
  echo "ERROR: set JWT_SECRET env var before running this script."
  echo "  Generate one with: openssl rand -hex 32"
  echo "  Then: DB_PASSWORD='...' JWT_SECRET='...' bash infra/scripts/bootstrap.sh"
  exit 1
fi

echo "==> Creating database (idempotent)"
DB_EXISTS=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'")
if [ "$DB_EXISTS" != "1" ]; then
  sudo -u postgres psql -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${DB_NAME};"
else
  echo "  database ${DB_NAME} already exists, skipping"
fi

echo "==> Creating user and granting privileges (idempotent)"
sudo -u postgres psql -v ON_ERROR_STOP=1 <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE USER ${DB_USER} WITH ENCRYPTED PASSWORD '${DB_PASSWORD}';
  END IF;
END
\$\$;

GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
SQL

echo "==> Writing backend/.env (not committed to git)"
cat > "$APP_DIR/backend/.env" <<EOF
PORT=8081
DB_HOST=localhost
DB_PORT=5432
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
JWT_SECRET=${JWT_SECRET}
EOF

echo "==> Installing systemd unit"
cp "$APP_DIR/infra/systemd/skillsifter.service" /etc/systemd/system/skillsifter.service
sed -i "s|__APP_DIR__|$APP_DIR|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload
systemctl enable skillsifter

echo "==> Installing nginx site"
cp "$APP_DIR/infra/nginx/api.skillsifter.in.conf" /etc/nginx/sites-available/api.skillsifter.in
ln -sf /etc/nginx/sites-available/api.skillsifter.in /etc/nginx/sites-enabled/api.skillsifter.in

echo "==> Note: run certbot separately before the first 'nginx -t' will pass:"
echo "    sudo certbot certonly --nginx -d api.skillsifter.in"
echo ""
echo "==> Bootstrap complete. Next steps:"
echo "    1. Get the TLS cert (certbot command above)"
echo "    2. sudo nginx -t && sudo systemctl reload nginx"
echo "    3. bash infra/scripts/deploy.sh   (builds + starts the app + runs migrations)"