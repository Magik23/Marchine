#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

[[ -f packaging/PKGBUILD ]] || {
  printf 'packaging/PKGBUILD is missing. Run finalize-aur.sh after the release tag exists.\n' >&2
  exit 1
}

command -v makepkg >/dev/null 2>&1 || {
  printf 'makepkg is required for AUR verification. Run this on Arch Linux.\n' >&2
  exit 1
}

(
  cd packaging

  tmp_srcinfo="$(mktemp)"
  trap 'rm -f "$tmp_srcinfo"' EXIT
  makepkg --printsrcinfo > "$tmp_srcinfo"

  if [[ ! -f .SRCINFO ]] || ! cmp -s .SRCINFO "$tmp_srcinfo"; then
    printf '.SRCINFO is missing or stale. Regenerate with: makepkg --printsrcinfo > .SRCINFO\n' >&2
    exit 1
  fi

  makepkg --cleanbuild --force

  if command -v namcap >/dev/null 2>&1; then
    namcap PKGBUILD
    shopt -s nullglob
    packages=(marchine-*.pkg.tar.*)
    if (( ${#packages[@]} > 0 )); then
      namcap "${packages[@]}"
    fi
  else
    printf 'namcap not found; package built, but namcap validation was skipped.\n' >&2
  fi
)
