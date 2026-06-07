#!/usr/bin/env bash
set -euo pipefail

REPO="Shivamingale3/dumpcron"
BIN_DST="/usr/local/bin/dumpcron"
CONFIG_DIR="/etc/dumpcron"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
STATE_DIR="/var/lib/dumpcron"
SERVICE_DIR="/etc/systemd/system"
PIDEX_DIR="/etc/pidex/custom.d"

echo "=== Dumpcron Installer ==="

ARCH="unknown"
case "$(uname -m)" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

installed=false

echo "Finding latest release..."
TAG=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" \
    | tr ',' '\n' | grep '"tag_name"' | cut -d'"' -f4)

if [ -n "$TAG" ] && [ "$ARCH" != "unknown" ]; then
    echo "Downloading Dumpcron $TAG (linux/$ARCH)..."

    url="https://github.com/$REPO/releases/download/$TAG/dumpcron-$TAG-linux-$ARCH"
    if curl -sfLo "$BIN_DST" "$url"; then
        chmod 755 "$BIN_DST"
        installed=true
        echo "Installed: $BIN_DST"
    else
        echo "Download failed, falling back to build from source..."
    fi
fi

if [ "$installed" = false ]; then
    echo "Pre-built binary not available for this system."

    GOBIN=""
    if command -v go &>/dev/null; then
        GOBIN="go"
    elif [[ -n "${SUDO_USER:-}" ]]; then
        USER_HOME="$(eval echo ~$SUDO_USER)"
        for guess in "$USER_HOME/go/bin/go" "$USER_HOME/.go/bin/go" /usr/local/go/bin/go; do
            if [[ -x "$guess" ]]; then
                GOBIN="$guess"
                break
            fi
        done
    fi

    if [[ -z "$GOBIN" ]]; then
        echo "ERROR: Go is required to build Dumpcron from source."
        echo "Install Go: https://go.dev/doc/install"
        exit 1
    fi

    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT
    echo "Building from source..."
    git clone --depth 1 "https://github.com/$REPO.git" "$TMP_DIR"
    cd "$TMP_DIR"
    "$GOBIN" build -ldflags="-s -w" -o "$BIN_DST" ./cmd/dumpcron
    echo "Installed: $BIN_DST"
fi

mkdir -p "$CONFIG_DIR" "$STATE_DIR"
echo "Created: $CONFIG_DIR  $STATE_DIR"

if [[ -f "$CONFIG_FILE" ]]; then
    echo "Config already exists at $CONFIG_FILE — leaving untouched"
else
    cat > "$CONFIG_FILE" << 'YAMLEOF'
# ┌─────────────────────────────────────────────────────┐
# │  Dumpcron configuration                            │
# │  Validate with:  dumpcron validate                  │
# │  Changes require: systemctl restart dumpcron        │
# └─────────────────────────────────────────────────────┘

backup_root: /srv/backups      # must exist and be writable
retention_days: 30             # backups older than this are deleted

jobs:
  # ── Example PostgreSQL job ─────────────────────────
  # - name: postgres_main
  #   type: postgres
  #   host: localhost
  #   port: 5432
  #   username: backup_user
  #   password: your_password
  #   databases:
  #     - app_db
  #   time: "02:00"

  # ── Example MySQL job ──────────────────────────────
  # - name: mysql_main
  #   type: mysql
  #   host: localhost
  #   port: 3306
  #   username: backup_user
  #   password: your_password
  #   databases:
  #     - users_db
  #   time: "03:00"

  # ── Example MongoDB job ────────────────────────────
  # - name: mongo_main
  #   type: mongo
  #   host: localhost
  #   port: 27017
  #   username: backup_user
  #   password: your_password
  #   databases:
  #     - analytics
  #   time: "04:00"
YAMLEOF
    echo "Created sample config at $CONFIG_FILE"
fi

cat > "$SERVICE_DIR/dumpcron.service" << SERVICE
[Unit]
Description=Dumpcron database backup scheduler
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$BIN_DST run
ExecStop=/bin/kill -s TERM \$MAINPID
Restart=on-failure
RestartSec=30
User=root

[Install]
WantedBy=multi-user.target
SERVICE

systemctl daemon-reload
echo "Installed systemd unit: dumpcron.service"

if [[ -d "$PIDEX_DIR" ]]; then
    cat > "$PIDEX_DIR/dumpcron.conf" << 'PIDEXEOF'
name = "dumpcron"
description = "Dumpcron Database Backup Scheduler"

[[events]]
name = "DUMPCRON_STARTED"
pattern = "DUMPCRON_STARTED"
severity = "INFO"
title = "Dumpcron Started"
message = "Dumpcron backup scheduler is now running"

[[events]]
name = "DUMPCRON_STOPPED"
pattern = "DUMPCRON_STOPPED"
severity = "WARNING"
title = "Dumpcron Stopped"
message = "Dumpcron backup scheduler has stopped"

[[events]]
name = "DUMPCRON_CONFIG_INVALID"
pattern = "DUMPCRON_CONFIG_INVALID"
severity = "CRITICAL"
title = "Dumpcron Config Error"
message = "Configuration validation failed — check /etc/dumpcron/config.yaml"

[[events]]
name = "DUMPCRON_DEPENDENCY_MISSING"
pattern = "DUMPCRON_DEPENDENCY_MISSING"
severity = "CRITICAL"
title = "Dumpcron Missing Dependency"
message = "A required binary is not installed — install the missing tool"

[[events]]
name = "DUMPCRON_STORAGE_INVALID"
pattern = "DUMPCRON_STORAGE_INVALID"
severity = "CRITICAL"
title = "Dumpcron Storage Error"
message = "Backup storage directory is missing or not writable"

[[events]]
name = "BACKUP_JOB_STARTED"
pattern = "BACKUP_JOB_STARTED"
severity = "INFO"
title = "Backup Job Started"
message = "A scheduled backup job has started"

[[events]]
name = "BACKUP_JOB_COMPLETED"
pattern = "BACKUP_JOB_COMPLETED"
severity = "INFO"
title = "Backup Job Completed"
message = "A scheduled backup job finished successfully"

[[events]]
name = "BACKUP_JOB_FAILED"
pattern = "BACKUP_JOB_FAILED"
severity = "CRITICAL"
title = "Backup Job Failed"
message = "A backup job completed with failures — check logs"

[[events]]
name = "DATABASE_BACKUP_FAILED"
pattern = "DATABASE_BACKUP_FAILED"
severity = "CRITICAL"
title = "Database Backup Failed"
message = "A single database backup failed — other databases continue"
PIDEXEOF
    echo "Installed PiDex config: $PIDEX_DIR/dumpcron.conf"
else
    echo "PiDex directory $PIDEX_DIR not found — skipping PiDex config"
fi

echo ""
echo "=== Dumpcron installed ==="
echo ""
echo "Next steps:"
echo "  1. Edit config:   nano $CONFIG_FILE"
echo "  2. Create dir:    mkdir -p \$(grep backup_root $CONFIG_FILE | awk '{print \$2}')"
echo "  3. Validate:      dumpcron validate"
echo "  4. Start:         systemctl start dumpcron"
echo "  5. Enable:        systemctl enable dumpcron"
echo ""
echo "Logs:  journalctl -u dumpcron -f"
if [[ -d "$PIDEX_DIR" ]]; then
echo ""
echo "PiDex: sudo pidex setup -> Manage custom services -> dumpcron.conf -> Register"
fi
