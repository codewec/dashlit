#!/usr/bin/env bash

set -Eeuo pipefail

readonly INSTALL_DIR="/opt/dashlit"
readonly DATA_DIR="/var/lib/dashlit"
readonly CONFIG_DIR="/etc/dashlit"
readonly SERVICE_FILE="/etc/systemd/system/dashlit.service"
readonly UPDATE_COMMAND="/usr/local/bin/dashlit-update"

purge=false

log() {
  printf '[DashLit] %s\n' "$*"
}

fail() {
  printf '[DashLit] Error: %s\n' "$*" >&2
  exit 1
}

case "${1:-}" in
  "") ;;
  --purge) purge=true ;;
  -h | --help)
    printf 'Usage: %s [--purge]\n' "${0##*/}"
    printf '  --purge  Also remove configuration, persistent data, and the system user.\n'
    exit 0
    ;;
  *) fail "unknown option: $1" ;;
esac

if [[ $# -gt 1 ]]; then
  fail "only one option is supported"
fi

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  fail "run this script as root"
fi

if command -v systemctl >/dev/null 2>&1; then
  systemctl disable --now dashlit 2>/dev/null || true
fi

rm -f -- "$SERVICE_FILE" "$UPDATE_COMMAND"
rm -rf -- "$INSTALL_DIR"

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload
  systemctl reset-failed dashlit 2>/dev/null || true
fi

if [[ "$purge" == true ]]; then
  rm -rf -- "$CONFIG_DIR" "$DATA_DIR"

  if id dashlit >/dev/null 2>&1; then
    userdel dashlit
  fi
  if getent group dashlit >/dev/null 2>&1; then
    groupdel dashlit
  fi

  log "DashLit, its configuration, and persistent data were removed"
else
  log "DashLit was removed"
  log "Configuration was preserved in ${CONFIG_DIR}"
  log "Persistent data was preserved in ${DATA_DIR}"
  log "Run this script with --purge to remove them as well"
fi
