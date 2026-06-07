#!/usr/bin/env bash
set -euo pipefail

BIN="/usr/local/bin/dumpcron"
CONFIG_DIR="/etc/dumpcron"
STATE_DIR="/var/lib/dumpcron"
SERVICE="/etc/systemd/system/dumpcron.service"
PIDEX_CONF="/etc/pidex/custom.d/dumpcron.conf"

echo "=== Dumpcron Uninstall ==="

systemctl disable --now dumpcron 2>/dev/null || true

rm -f "$BIN"
rm -f "$SERVICE"
rm -f "$PIDEX_CONF"
rm -rf "$STATE_DIR"

systemctl daemon-reload

if [ -d "$CONFIG_DIR" ]; then
    read -r -p "Remove $CONFIG_DIR? [y/N]: " answer
    if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
        rm -rf "$CONFIG_DIR"
        echo "Removed $CONFIG_DIR"
    fi
fi

echo "Dumpcron uninstalled."
