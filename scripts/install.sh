#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

MARCHINE_BIN="$HOME/.local/bin/marchine"
DESKTOP_DIR="$HOME/.local/share/applications"
DESKTOP_FILE="$DESKTOP_DIR/marchine.desktop"

make build

install -Dm755 dist/marchine "$MARCHINE_BIN"

mkdir -p "$HOME/.config/marchine"
if [[ ! -f "$HOME/.config/marchine/config.toml" ]]; then
  cp config.example.toml "$HOME/.config/marchine/config.toml"
fi

mkdir -p "$DESKTOP_DIR"
sed "s|@MARCHINE_BIN@|$MARCHINE_BIN|g" \
  packaging/marchine.desktop.in > "$DESKTOP_FILE"
chmod 644 "$DESKTOP_FILE"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$DESKTOP_DIR" >/dev/null 2>&1 || true
fi

printf '\nInstalled Marchine to %s\n' "$MARCHINE_BIN"
printf 'Desktop entry: %s\n' "$DESKTOP_FILE"
printf 'Config: %s\n' "$HOME/.config/marchine/config.toml"
printf 'Run: marchine\n'
