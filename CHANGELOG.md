# Changelog

## Unreleased
- Production-hardening pass for the pre-AUR release.
- Fixed stale Credits tests so the complete Go test suite matches the current embedded artwork.
- Added real Arcade cache schema validation (`marchine-index-v4`), atomic cache writes, degraded-cache diagnostics, and saved-library fallback when refresh/source access fails.
- Rehydrate cached Arcade launch identity from current config so emulator/path changes never leave stale launch commands behind.
- Pass explicit Marchine ROM paths to MAME at launch; preserve multiple configured ROM paths when editing the primary path in Setup.
- Fixed the first-time Setup ROM-path panic and normalize `~`/environment paths in memory immediately after save, so refresh/launch no longer requires a restart.
- Detect vertical games from MAME XML display rotation and launch them with `-autorol`.
- Added timeouts for MAME metadata calls and single-flight UI refresh behavior.
- Made Unicode list trimming/padding terminal-cell aware.
- `--doctor` now exits nonzero for operational diagnostics.
- Removed the unused sculpture-animation config field.
- Hardened Makefile version fallback, PIE/trimpath builds, format checks, tests, and vet.
- Added CI across supported Go lines plus pinned govulncheck/staticcheck, shell validation, and desktop-entry validation.
- Hardened release archives to include documentation and bundled third-party license texts.
- Replaced the prototype AUR recipe with a checksum-finalized production template and Arch verification helpers.
- Made Foot/Omarchy desktop integration explicitly optional and kept compositor configuration outside package-managed user state.
- Reconciled current documentation with the 124×48 UI minimum and 62×22 static ANSI sculpture.

## 1.21.0
- Best-effort launch geometry request set to 120x48 and UI minimum updated to match.
- Theme cycling now works across Library, Categories, Setup, Info, and Credits.
- Hide Non-Arcade toggle now works from the main Library view with H.
- Fightstick sculpture now follows the active theme and uses the updated cropped asset.
- Info/Credits footers and Info shortcuts updated to expose theme and refresh behavior more clearly.

## 1.20.0

- Embedded the supplied `arcadestick_2_60.ans` as the current fightstick asset (60 × 25).
- Preserve every ANSI source row; removed the old fixed top-row shaving.
- Removed synthetic blank rows above and below the ANSI image.
- Slightly increased the visual weight of the top `GAME MACHINERY.` title only; the subtitle is unchanged.

## 1.19.0

- Center the embedded fightstick ANSI sculpture on both axes inside its dedicated sculpture viewport.
- Remove the hidden four-cell right gutter from the launcher list/detail panels so both columns use their full interior width.
- Tighten the detail-column width to the native ANSI asset plus one cell of deliberate breathing room on each side.
- Anchor every screen footer to the same fixed viewport baseline so Library, Info, Credits, Categories and Setup never make the bottom menu jump vertically.
- Keep the v1.18 single-key Info/Credits controls, Arcade refresh/filter system, responsive minimum-size gate, and fixed embedded ANSI geometry intact.

## 1.18.0

- Replace the visible Ctrl-chord utility shortcuts with single-key `I INFO` and `O CREDITS`; Credits remains the last main-footer item.
- Keep the old Ctrl+K / Ctrl+G bindings as hidden compatibility aliases.
- Rename the user-facing Arcade rebuild action to **Refresh Arcade Library** and keep it on `R`; INFO now documents it again.
- Add a persistent `H  HIDE NON-ARCADE [ON/OFF]` toggle to INFO.
- When Hide Non-Arcade is ON, force the clean Arcade playlist and remove `ALL`, casino/gambling, mahjong, `NON - ARCADE`, and any category family containing only non-arcade entries from CATEGORIES.
- When the toggle is OFF, expose the complete indexed runnable library and all category families again.
- Rename the aggregate `NON-ARCADE / JUNK` filter to `NON - ARCADE`.
- Expand MAME/CatVer non-arcade classification for home computers, consoles, handhelds, utilities, calculators, terminals and related non-arcade families while preserving the full index.
- Rename the utility browse bucket to `OTHER / UTILITY`.
- Bump the MAME cache schema to `marchine-index-v3` for the expanded category/classification semantics.
- Render the launcher into an exact one-cell outer gutter on all four edges and account for separator rows explicitly so right/left and top/bottom padding cannot drift.
- Give the compact `GAME MACHINERY.` / `BROWSE / SELECT / LAUNCH` identity a small additional visual-width increase without turning it into block art.
- Keep the V1.17 fixed ANSI sculpture, dedicated Credits screen, dynamic CatVer categories, hard minimum terminal gate, and local-first launch model intact.
