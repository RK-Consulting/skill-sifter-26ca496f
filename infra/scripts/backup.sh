#!/bin/bash
# ==============================================================================
# SkillSifter — Server Backup Script
# Run as: sudo ./backup.sh
# Output: /home/harisha/backups/skillsifter_backup_<timestamp>.tar.gz
#
# Scope: this backs up SkillSifter's own database, config, and app state,
# plus shared server-level facts (UFW, installed packages, cron) that any
# fresh-server rebuild needs regardless of which app is being restored.
# It does NOT reference or depend on any other application on this server.
#
# ⚠️ ASSUMPTIONS TO VERIFY (marked inline below) — these paths were not
# directly confirmed and should be checked before relying on this script:
#   - .env location: assumed /var/www/skillsifter/backend/.env
#   - Redis usage: not confirmed whether SkillSifter uses Redis at all.
#     If it doesn't, the Redis section here is harmless but unnecessary.
#     If it does but with app-specific keys, no separate backup is needed
#     since the RDB snapshot captures the whole Redis instance regardless.
# ==============================================================================
set -e

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_ROOT="/home/harisha/backups"
WORKDIR="$BACKUP_ROOT/skillsifter_backup_$TIMESTAMP"
ARCHIVE="$BACKUP_ROOT/skillsifter_backup_$TIMESTAMP.tar.gz"

if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run as root or with sudo (needed for pg_dump, nginx, letsencrypt, ufw)."
  exit 1
fi

mkdir -p "$WORKDIR"
echo "🚀 Starting SkillSifter backup -> $WORKDIR"

# ------------------------------------------------------------------
# 1. PostgreSQL — skillsifter database only
# ------------------------------------------------------------------
echo "📊 Backing up skillsifter database..."
mkdir -p "$WORKDIR/postgres"
sudo -u postgres pg_dump -Fc skillsifter -f "$WORKDIR/postgres/skillsifter.dump" \
  && echo "  ✅ skillsifter (custom format)" || echo "  ❌ skillsifter dump failed"
sudo -u postgres pg_dump skillsifter -f "$WORKDIR/postgres/skillsifter.sql" \
  && echo "  ✅ skillsifter (plain SQL)" || echo "  ❌ skillsifter SQL dump failed"
sudo -u postgres psql -c "\du skillsifter_user" > "$WORKDIR/postgres/role.txt" 2>&1 || true

# ------------------------------------------------------------------
# 2. Nginx — SkillSifter's site config only
# ------------------------------------------------------------------
echo "🌐 Backing up SkillSifter Nginx config..."
mkdir -p "$WORKDIR/nginx"
cp /etc/nginx/sites-available/api.skillsifter.in "$WORKDIR/nginx/" 2>&1 \
  && echo "  ✅ api.skillsifter.in config" || echo "  ⚠️  config not found at expected path"

# ------------------------------------------------------------------
# 3. Let's Encrypt certificate — SkillSifter domain only
# ------------------------------------------------------------------
echo "🔒 Backing up SSL certificate..."
mkdir -p "$WORKDIR/letsencrypt"
tar -czf "$WORKDIR/letsencrypt/api_skillsifter_cert.tar.gz" \
  /etc/letsencrypt/live/api.skillsifter.in \
  /etc/letsencrypt/archive/api.skillsifter.in \
  /etc/letsencrypt/renewal/api.skillsifter.in.conf 2>/dev/null \
  && echo "  ✅ Certificate archived" || echo "  ⚠️  Cert archive had issues"

# ------------------------------------------------------------------
# 4. SkillSifter app state — binary, .env, systemd unit
# ------------------------------------------------------------------
echo "📦 Backing up SkillSifter app state..."
mkdir -p "$WORKDIR/app"
if [ -f /var/www/skillsifter/backend/skillsifter ]; then
  cp /var/www/skillsifter/backend/skillsifter "$WORKDIR/app/skillsifter_binary"
  echo "  ✅ Compiled binary"
else
  echo "  ⚠️  Binary not found at /var/www/skillsifter/backend/skillsifter — verify path"
fi
# ASSUMPTION — verify this is the real .env path before relying on it
if [ -f /var/www/skillsifter/backend/.env ]; then
  cp /var/www/skillsifter/backend/.env "$WORKDIR/app/backend.env"
  echo "  ✅ .env (SENSITIVE — contains live secrets)"
else
  echo "  ⚠️  No .env found at assumed path — SkillSifter may configure via"
  echo "     systemd Environment= directives instead; check the unit file below"
fi
cp /etc/systemd/system/skillsifter.service "$WORKDIR/app/" 2>/dev/null \
  && echo "  ✅ systemd unit file (also captures any inline env vars)" \
  || echo "  ⚠️  systemd unit not found at expected path"

# ------------------------------------------------------------------
# 5. Shared server-level facts — captured independently, not a
#    reference to any other application
# ------------------------------------------------------------------
echo "🔥 Backing up UFW rules..."
mkdir -p "$WORKDIR/server"
ufw status verbose > "$WORKDIR/server/ufw_status.txt"

echo "⏰ Backing up crontabs..."
crontab -u harisha -l > "$WORKDIR/server/harisha_crontab.txt" 2>&1 || echo "(none)" > "$WORKDIR/server/harisha_crontab.txt"
crontab -u root -l > "$WORKDIR/server/root_crontab.txt" 2>&1 || echo "(none)" > "$WORKDIR/server/root_crontab.txt"

echo "📝 Documenting installed packages..."
dpkg --get-selections > "$WORKDIR/server/apt_packages.txt"

{
  echo "Hostname: $(hostname)"
  echo "OS: $(lsb_release -d 2>/dev/null || cat /etc/os-release | grep PRETTY_NAME)"
  echo "Kernel: $(uname -r)"
  echo "Disk: $(df -h / | tail -1)"
  echo "Memory: $(free -h | grep Mem)"
  echo "Backup taken: $(date)"
} > "$WORKDIR/server/system_info.txt"
echo "  ✅ Server-level facts (UFW, cron, packages, system info)"

# ------------------------------------------------------------------
# Compress
# ------------------------------------------------------------------
echo "📦 Compressing archive..."
tar -czf "$ARCHIVE" -C "$BACKUP_ROOT" "skillsifter_backup_$TIMESTAMP"
rm -rf "$WORKDIR"

SIZE=$(du -h "$ARCHIVE" | cut -f1)
echo ""
echo "========================================"
echo "✅ SKILLSIFTER BACKUP COMPLETE"
echo "========================================"
echo "📁 Archive: $ARCHIVE"
echo "💾 Size: $SIZE"
echo ""
echo "⚠️  Contains live secrets (.env/systemd env vars, DB dump). Download"
echo "   off this server and don't retain local copies longer than needed."
echo ""
echo "📥 scp harisha@$(curl -s ifconfig.me):$ARCHIVE ./"
echo "========================================"