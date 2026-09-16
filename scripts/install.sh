#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

MARCHINE_BIN="$HOME/.local/bin/marchine"
CONFIG_DIR="$HOME/.config/marchine"
CONFIG_FILE="$CONFIG_DIR/config.toml"
DESKTOP_DIR="$HOME/.local/share/applications"
DESKTOP_FILE="$DESKTOP_DIR/marchine.desktop"

make build
install -Dm755 dist/marchine "$MARCHINE_BIN"

mkdir -p "$CONFIG_DIR"
if [[ ! -f "$CONFIG_FILE" ]]; then
  install -m644 config.example.toml "$CONFIG_FILE"
fi

desktop_status="not installed (Foot not found)"
if command -v foot >/dev/null 2>&1; then
  mkdir -p "$DESKTOP_DIR"
  desktop_content="$(< packaging/marchine.desktop.in)"
  desktop_content="${desktop_content//@MARCHINE_BIN@/$MARCHINE_BIN}"
  printf '%s\n' "$desktop_content" > "$DESKTOP_FILE"
  chmod 644 "$DESKTOP_FILE"
  desktop_status="$DESKTOP_FILE"

  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$DESKTOP_DIR" >/dev/null 2>&1 || true
  fi
fi

printf '\nInstalled Marchine to %s\n' "$MARCHINE_BIN"
printf 'Config: %s\n' "$CONFIG_FILE"
printf 'Desktop entry: %s\n' "$desktop_status"
printf 'Run: %s\n' "$MARCHINE_BIN"
