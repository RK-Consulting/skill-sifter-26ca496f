#!/usr/bin/env bash
# Repeatable deploy. Run by hand on the server whenever you want to test:
#   bash infra/scripts/deploy.sh
# Deploys whichever branch is currently checked out on this server —
# does NOT force-switch branches. Check out the branch you want first.
set -euo pipefail

APP_DIR="$(cd "$(dirname "\${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$APP_DIR"

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
echo "==> Pulling latest \${CURRENT_BRANCH}"
git fetch origin
git reset --hard "origin/\${CURRENT_BRANCH}"

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
sed -i "s|__APP_DIR__|$APP_DIR|g" /etc/systemd/system/skillsifter.service
systemctl daemon-reload

echo "==> Restarting service"
systemctl restart skillsifter
sleep 2
systemctl status skillsifter --no-pager

echo "==> Health check"
curl -sf http://localhost:8081/health-check && echo "" && echo "==> Deploy succeeded"
