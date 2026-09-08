#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
make build
install -Dm755 dist/marchine "$HOME/.local/bin/marchine"
mkdir -p "$HOME/.config/marchine"
if [[ ! -f "$HOME/.config/marchine/config.toml" ]]; then
  cp config.example.toml "$HOME/.config/marchine/config.toml"
fi
printf '\nInstalled Marchine to %s\n' "$HOME/.local/bin/marchine"
printf 'Config: %s\n' "$HOME/.config/marchine/config.toml"
printf 'Run: marchine\n'
