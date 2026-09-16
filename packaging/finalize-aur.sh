#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

usage() {
  printf 'Usage: %s VERSION\n' "$0" >&2
  printf 'Example: %s 1.22.0\n' "$0" >&2
  exit 2
}

[[ $# -eq 1 ]] || usage
version="${1#v}"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  printf 'Invalid release version: %s\n' "$1" >&2
  exit 2
fi

command -v curl >/dev/null 2>&1 || {
  printf 'curl is required to finalize the AUR package.\n' >&2
  exit 1
}
command -v sha256sum >/dev/null 2>&1 || {
  printf 'sha256sum is required to finalize the AUR package.\n' >&2
  exit 1
}

template="packaging/PKGBUILD.example"
pkgbuild="packaging/PKGBUILD"
srcinfo="packaging/.SRCINFO"
url="https://github.com/Magik23/Marchine/archive/refs/tags/v${version}.tar.gz"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

printf 'Fetching tagged source: %s\n' "$url"
curl --fail --location --silent --show-error "$url" --output "$tmp"
sha="$(sha256sum "$tmp" | awk '{print $1}')"

content="$(< "$template")"
content="${content//@PKGVER@/$version}"
content="${content//@SHA256@/$sha}"
printf '%s\n' "$content" > "$pkgbuild"

printf 'Generated %s\n' "$pkgbuild"
printf 'SHA-256: %s\n' "$sha"

if command -v makepkg >/dev/null 2>&1; then
  (
    cd packaging
    makepkg --printsrcinfo > .SRCINFO
  )
  printf 'Generated %s\n' "$srcinfo"
else
  rm -f "$srcinfo"
  printf 'makepkg not found; generate .SRCINFO on Arch before AUR submission.\n' >&2
fi
