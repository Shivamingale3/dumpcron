#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
BIN_SRC="$REPO_DIR/dumpcron"
BIN_DST="/usr/local/bin/dumpcron"
CONFIG_DIR="/etc/dumpcron"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
STATE_DIR="/var/lib/dumpcron"
SYSTEMD_SRC="$REPO_DIR/deploy/dumpcron.service"
SYSTEMD_DST="/etc/systemd/system/dumpcron.service"
PIDEX_SRC="$REPO_DIR/config/dumpcron.conf"
PIDEX_DIR="/etc/pidex/custom.d"
PIDEX_DST="$PIDEX_DIR/dumpcron.conf"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

say()   { echo -e "$*"; }
ok()    { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}⚠${NC} $*"; }
die()   { echo -e "${RED}✗${NC} $*"; exit 1; }

say ""
say "╔══════════════════════════════════════╗"
say "║    Dumpcron v1 — Installer          ║"
say "╚══════════════════════════════════════╝"
say ""

# ── preflight ────────────────────────────────────────────────────────────────

if [[ $EUID -ne 0 ]]; then
    die "must run as root:  sudo ./install.sh"
fi

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

if [[ -z "$GOBIN" ]] || ! "$GOBIN" version &>/dev/null; then
    die "Go not found on PATH — install it first:  https://go.dev/dl/"
fi

if [[ ! -f "$REPO_DIR/go.mod" ]]; then
    die "run this script from the dumpcron repository root (deploy/install.sh)"
fi

ok "preflight checks passed (go: $GOBIN)"

# ── build ────────────────────────────────────────────────────────────────────

say "building dumpcron ..."
(
    cd "$REPO_DIR"
    "$GOBIN" build -ldflags="-s -w" -o dumpcron ./cmd/dumpcron
)
ok "build complete"

# ── install binary ───────────────────────────────────────────────────────────

install -m 755 "$BIN_SRC" "$BIN_DST"
ok "installed $BIN_DST"

# ── create directories ───────────────────────────────────────────────────────

mkdir -p "$CONFIG_DIR" "$STATE_DIR"
ok "created $CONFIG_DIR  $STATE_DIR"

# ── sample config (only if none exists) ──────────────────────────────────────

if [[ -f "$CONFIG_FILE" ]]; then
    ok "config already exists at $CONFIG_FILE — leaving untouched"
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
  #     - auth_db
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
    ok "created sample config at $CONFIG_FILE"
fi

# ── systemd unit ─────────────────────────────────────────────────────────────

cp "$SYSTEMD_SRC" "$SYSTEMD_DST"
systemctl daemon-reload
ok "installed systemd unit: dumpcron.service"

# ── PiDex config ─────────────────────────────────────────────────────────────

if [[ -d "$PIDEX_DIR" ]]; then
    cp "$PIDEX_SRC" "$PIDEX_DST"
    ok "installed PiDex config at $PIDEX_DST"
else
    warn "PiDex directory $PIDEX_DIR not found — skipping PiDex config"
    warn "To enable PiDex alerts, install PiDex, then run:"
    warn "  sudo cp dumpcron.conf $PIDEX_DST"
    warn "  sudo pidex setup  →  Manage custom services  →  dumpcron.conf"
fi

# ── done ─────────────────────────────────────────────────────────────────────

say ""
say "╔══════════════════════════════════════════════════════════════════╗"
say "║  Dumpcron v1 installed                                         ║"
say "╠══════════════════════════════════════════════════════════════════╣"
say "║                                                                ║"
say "║  Next steps:                                                   ║"
say "║  1. Edit config:   vim $CONFIG_FILE              ║"
say "║  2. Create dir:    mkdir -p /srv/backups                       ║"
say "║  3. Validate:      dumpcron validate                           ║"
say "║  4. Start:         systemctl start dumpcron                    ║"
say "║  5. Enable:        systemctl enable dumpcron                   ║"
say "║                                                                ║"
say "║  Logs:             journalctl -u dumpcron -f                   ║"
say "║                                                                ║"
if [[ -d "$PIDEX_DIR" ]]; then
say "║  PiDex:            sudo pidex setup → Manage custom services   ║"
say "║                    → dumpcron.conf → Register → Restart pidex  ║"
fi
say "╚══════════════════════════════════════════════════════════════════╝"
say ""
