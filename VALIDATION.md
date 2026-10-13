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
6. MAME games launch without Marchine-injected rotation flags and follow MAME's own display configuration;
7. temporarily removing/unmounting the ROM source keeps a valid saved library and reports the degraded state;
8. `marchine --doctor --rescan` exits nonzero for a real operational diagnostic failure;
9. GAMES entries can be added, edited, removed, saved and reloaded after restarting; removal never deletes the executable;
10. the GAMES source launches both a directly configured executable and an executable with arguments/working directory, and returns to Marchine according to the launched process's behavior;
11. Foot `Ctrl+Shift+V` paste works in active game/Arcade fields and `Ctrl+S` saves the active value without pressing Enter first;
12. `Ctrl+C` does not quit Marchine, while `Q` and View + Menu still do;
13. saving Arcade settings and refreshing Arcade are separate operations; Setup `R` doesn't save pending edits or refresh GAMES;
14. View + X / View + A adjust Foot font size on a focused Marchine Hyprland desktop window, preserving standalone View/X/A behaviors and previous View chords.

## UI smoke checks

Verify at and above **124×48**:

- Library remains two-column;
- Name / Year / Manufacturer columns expand without starving Manufacturer;
- embedded fightstick remains fixed **62×22** terminal-cell geometry;
- all four themes remain legible;
- Info, Credits, Categories, Setup, and Library share the expected footer baseline;
- below 124×48, only the terminal-too-small recovery message is shown; View+A can still decrease font size;
- Info's source bar shows a compact version label, without Git commit hashes;
- Setup displays ARCADE before GAMES, and game picker/editor footers remain visible at the minimum geometry.

## Release artifact gate

For a real tag `vX.Y.Z`:

```bash
make clean
make check
make build VERSION=X.Y.Z
test "$(./dist/marchine --version)" = "marchine X.Y.Z"
```

The GitHub release workflow is responsible for the final archive and SHA-256 file. After publication, verify the archive contains the executable at `marchine-X.Y.Z-linux-amd64/marchine` plus all three icon files under its `share/icons/hicolor/` tree. Check the README's prebuilt installation instructions against that exact layout, then apply `docs/releases/vX.Y.Z.md` as the GitHub Release body (the workflow currently generates notes by default).

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
