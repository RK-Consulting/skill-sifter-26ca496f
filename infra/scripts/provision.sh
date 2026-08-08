#!/bin/bash
# ==============================================================================
# SkillSifter — Fresh Server Provisioning Script
# Run as: sudo ./provision.sh /path/to/skillsifter_backup_<timestamp>.tar.gz
#
# What this DOES automate:
#   - Installs Nginx, PostgreSQL, Certbot, UFW
#   - Restores the skillsifter database from the backup archive
#   - Restores SkillSifter's Nginx site config
#   - Re-enables UFW
#   - Restores the compiled binary + systemd service (as a stopgap)
#
# What this DOES NOT automate (needs your manual judgment):
#   - Pointing DNS at the new server's IP (do this BEFORE running certbot)
#   - Waiting for DNS propagation
#   - Re-issuing SSL certs (restoring old ones only works within the same
#     domain + validity window — a genuine migration should re-issue)
#   - Rotating secrets that were ever exposed before this migration
#
# ⚠️ IMPORTANT: restoring the backed-up BINARY gets SkillSifter running
# quickly for disaster recovery, but it's a snapshot as of the backup date.
# For a real, deliberate migration, prefer rebuilding fresh from source via
# SkillSifter's own repo and its own infra/scripts/deploy.sh (referenced in
# its Nginx config header) rather than relying on this stopgap binary.
#
# Scope: this script only sets up SkillSifter. If other applications also
# run on this server, they need their own separate provisioning.
# ==============================================================================
set -e

if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run as root or with sudo."
  exit 1
fi

BACKUP_ARCHIVE="$1"
if [ -z "$BACKUP_ARCHIVE" ] || [ ! -f "$BACKUP_ARCHIVE" ]; then
  echo "Usage: sudo ./provision.sh /path/to/skillsifter_backup_<timestamp>.tar.gz"
  exit 1
fi

RESTORE_DIR="/tmp/skillsifter_restore_$(date +%s)"
mkdir -p "$RESTORE_DIR"
echo "📦 Extracting backup archive to $RESTORE_DIR..."
tar -xzf "$BACKUP_ARCHIVE" -C "$RESTORE_DIR"
BACKUP_CONTENT_DIR=$(find "$RESTORE_DIR" -maxdepth 1 -type d -name "skillsifter_backup_*")

echo ""
echo "========================================"
echo "🚀 Starting fresh server provisioning — SkillSifter"
echo "========================================"

# ------------------------------------------------------------------
# 1. Core packages
# ------------------------------------------------------------------
echo "📥 Installing core packages..."
apt-get update -qq
apt-get install -y -qq \
  nginx postgresql certbot python3-certbot-nginx ufw git curl

# ------------------------------------------------------------------
# 2. PostgreSQL — restore skillsifter database
# ------------------------------------------------------------------
echo "📊 Restoring skillsifter database..."
sudo -u postgres psql -c "CREATE USER skillsifter_user WITH PASSWORD 'CHANGE_ME_SKILLSIFTER_PASSWORD';"
sudo -u postgres psql -c "CREATE DATABASE skillsifter OWNER skillsifter_user;"
sudo -u postgres pg_restore -d skillsifter "$BACKUP_CONTENT_DIR/postgres/skillsifter.dump" 2>&1 || \
  echo "  ⚠️  skillsifter restore had warnings — check output above"

echo "  ⚠️  IMPORTANT: change the placeholder password immediately:"
echo "     sudo -u postgres psql -c \"ALTER USER skillsifter_user WITH PASSWORD 'new-strong-password';\""
echo "     Then update the new password in SkillSifter's config (.env or"
echo "     systemd Environment= directives — see the unit file restored below)."

# ------------------------------------------------------------------
# 3. Nginx config
# ------------------------------------------------------------------
echo "🌐 Restoring Nginx configuration..."
cp "$BACKUP_CONTENT_DIR/nginx/api.skillsifter.in" /etc/nginx/sites-available/ 2>/dev/null \
  && echo "  ✅ Site config copied to sites-available" || echo "  ⚠️  Config not found in backup"
echo "  ⚠️  Don't enable the site or run certbot yet — do this AFTER DNS points"
echo "     to this server's new IP:"
echo "     ln -s /etc/nginx/sites-available/api.skillsifter.in /etc/nginx/sites-enabled/"
echo "     certbot --nginx -d api.skillsifter.in"

# ------------------------------------------------------------------
# 4. UFW
# ------------------------------------------------------------------
echo "🔥 Configuring firewall..."
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable
echo "  ✅ UFW enabled with SSH + Nginx allowed"

# ------------------------------------------------------------------
# 5. SkillSifter app — restore binary + systemd unit (stopgap)
# ------------------------------------------------------------------
echo "📦 Restoring SkillSifter application (stopgap — see warning at top)..."
mkdir -p /var/www/skillsifter/backend
if [ -f "$BACKUP_CONTENT_DIR/app/skillsifter_binary" ]; then
  cp "$BACKUP_CONTENT_DIR/app/skillsifter_binary" /var/www/skillsifter/backend/skillsifter
  chmod +x /var/www/skillsifter/backend/skillsifter
  echo "  ✅ Binary restored (snapshot as of backup date)"
fi
if [ -f "$BACKUP_CONTENT_DIR/app/backend.env" ]; then
  cp "$BACKUP_CONTENT_DIR/app/backend.env" /var/www/skillsifter/backend/.env
  echo "  ✅ .env restored — update the DB password here to match step 2"
fi
if [ -f "$BACKUP_CONTENT_DIR/app/skillsifter.service" ]; then
  cp "$BACKUP_CONTENT_DIR/app/skillsifter.service" /etc/systemd/system/
  systemctl daemon-reload
  systemctl enable skillsifter
  echo "  ✅ systemd service installed (not started — update password first, then:"
  echo "     systemctl start skillsifter)"
fi

echo ""
echo "========================================"
echo "✅ PROVISIONING COMPLETE — MANUAL STEPS REMAIN"
echo "========================================"
echo "Before this server is truly live, you must:"
echo "  1. Point DNS (api.skillsifter.in) at this server's new IP"
echo "  2. Wait for DNS propagation (check: dig +short api.skillsifter.in)"
echo "  3. Set a new DB password (see warning above) — never reuse old ones"
echo "  4. Update SkillSifter's config with the new password"
echo "  5. Enable the Nginx site + run certbot for a fresh SSL cert (see step 3)"
echo "  6. Start SkillSifter: systemctl start skillsifter"
echo "  7. STRONGLY CONSIDER rebuilding from source via SkillSifter's own repo"
echo "     instead of running the restored binary long-term"
echo "  8. Test end-to-end before decommissioning the old server"
echo "========================================"