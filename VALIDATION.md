# Marchine validation checklist

This file is a reproducible release checklist, not a claim that every command was run in every environment.

## Source gate

Run from the repository root:

```bash
make clean
make check
make build VERSION=production-check
./dist/marchine --version
./dist/marchine --demo
bash -n scripts/install.sh
bash -n packaging/finalize-aur.sh
bash -n packaging/verify-aur.sh
```

Expected version output for the command above:

```text
marchine production-check
```

`make check` verifies:

- `gofmt` cleanliness for `cmd/` and `internal/`;
- `go test -mod=vendor -count=1 ./...`;
- `go vet -mod=vendor ./...`.

## Toolchain matrix

CI must pass on the supported Go lines declared in `.github/workflows/ci.yml`.

Before a release, the current CI also runs:

```text
govulncheck ./...
staticcheck ./...
shellcheck scripts/install.sh packaging/finalize-aur.sh packaging/verify-aur.sh
desktop-file-validate /tmp/marchine-foot.desktop
desktop-file-validate packaging/marchine.desktop
```

## Functional smoke checks

Use a real MAME installation and a small known ROM set to verify:

1. cache-first startup opens without a forced `mame -listxml` scan;
2. `R` performs one refresh at a time;
3. explicit `rom_paths` index and launch using the same paths;
4. multiple explicit ROM paths survive editing the primary path in Setup;
5. clearing the primary Setup path restores auto-detection;
6. a vertical MAME machine launches with `-autorol`;
7. temporarily removing/unmounting the ROM source keeps a valid saved library and reports the degraded state;
8. `marchine --doctor --rescan` exits nonzero for a real operational diagnostic failure;
9. GAMES entries launch and return to Marchine cleanly.

## UI smoke checks

Verify at and above **124×48**:

- Library remains two-column;
- Name / Year / Manufacturer columns expand without starving Manufacturer;
- embedded fightstick remains fixed **62×22** terminal-cell geometry;
- all four themes remain legible;
- Info, Credits, Categories, Setup, and Library share the expected footer baseline;
- below 124×48, only the terminal-too-small view is shown.

## Release artifact gate

For a real tag `vX.Y.Z`:

```bash
make clean
make check
make build VERSION=X.Y.Z
test "$(./dist/marchine --version)" = "marchine X.Y.Z"
```

The GitHub release workflow is responsible for the final archive and SHA-256 file.

## AUR gate

Only after the real GitHub tag exists:

```bash
./packaging/finalize-aur.sh X.Y.Z
./packaging/verify-aur.sh
```

Before submitting/updating the AUR repository, confirm:

- `packaging/PKGBUILD` contains the real source SHA-256;
- `packaging/.SRCINFO` exactly matches `makepkg --printsrcinfo`;
- `makepkg --cleanbuild --force` succeeds on Arch;
- `namcap` is clean or every remaining warning is understood and intentional;
- ideally, repeat the package build in a clean Arch chroot.
