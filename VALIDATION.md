# Validation — Marchine V1.20

Validated in the build environment with:

- `gofmt` on changed Go sources;
- `go test ./...` passing;
- `go vet ./...` passing;
- static `dist/marchine` built from this V1.20 source;
- `./dist/marchine --version` returning `marchine 1.20.0`;
- supplied `arcadestick_2_60.ans` embedded as `internal/ui/fightstick.ans`;
- visible ANSI geometry validated at exactly **60 columns × 25 rows**;
- no fixed top-row shaving: the first and last source rows remain part of the rendered asset;
- no synthetic blank rows added above or below the ANSI image when the panel is taller than the asset;
- stable center-crop fallback retained only when the terminal is genuinely too short for the full asset;
- detail-column width derived from the current ANSI width;
- top `GAME MACHINERY.` title given a small visual-weight/width increase while `BROWSE / SELECT / LAUNCH` remains unchanged;
- common footer baseline across Library, Info, Credits, Categories and Setup preserved;
- CLIAMP-style terminal-too-small gate adjusted to **112 × 48** so the new 60-column sculpture never forces the launcher into the defensive stacked layout.

The release ZIP includes `vendor/`, so normal build/test commands use the vendored dependency snapshot.
