# Changelog

## Unreleased

## 1.26.0
- Completed the in-app **GAMES** setup workflow: add, edit and remove named native executables, scripts and external launch commands; launch targets retain optional arguments and working directories.
- Save new/edited GAMES entries atomically to `config.toml`, update the visible GAMES library immediately, and preserve configured games across restarts. Edit/Remove use a picker, with a Y/N deletion confirmation; removing an entry never deletes executable files.
- Added quoted-argument parsing for paths with spaces, round-trip formatting, and validation for incomplete or invalid game entries; arguments are passed to the executable directly rather than interpreted by a shell.
- Reorganized Setup **ARCADE first, GAMES second**. `Ctrl+S` saves Arcade settings, whereas `R` refreshes the Arcade index without saving pending changes; GAMES needs no refresh after saving.
- Improved both Arcade and GAMES text editing: Foot `Ctrl+Shift+V` paste support and guidance; `Ctrl+S` commits and saves the active field without requiring Enter; input errors preserve editable text.
- Disabled `Ctrl+C` as a Marchine quit shortcut on all screens to avoid accidental exits; continue using `Q` or `View + Menu` to quit.
- Added controller font zoom using **View + X** (larger) and **View + A** (smaller), via Foot's existing zoom shortcuts for the focused Marchine window on Hyprland. Standard keyboard zoom still works; font changes remain session-only.
- Improved the below-minimum **124×48** terminal warning with font/window-size recovery instructions; controller font reduction remains available when the normal UI cannot render.
- Added a compact running version indicator at the right of the Info screen's source bar; detailed Git build identifiers remain available through `marchine --version`.
- Removed redundant indexed-game/CatVer/hide-mode text from the Categories header and polished the Setup headings, instructions and removal confirmation.
- Fixed user-local icon installation (128/256/512 px), and included the same icons in future GitHub release archives.
- Added focused tests for game management/persistence, Arcade refresh separation, clipboard and active-field saving, Ctrl+C handling, controller chords, font zoom and minimum terminal layout.

## 1.25.0
- Preserve the selected game when toggling Hide Non-Arcade; if it becomes hidden, advance to the next visible game or fall back to the previous one instead of jumping to the top.
- Keep the selected Arcade category (including SHMUPS and VERTICAL) when toggling Hide Non-Arcade; update the visible list and source-bar count in place.
- Show a live game count at the right of the source bar, following the selected Games/Arcade source and any category, search, or Hide Non-Arcade filtering.
- Added optional `View + RB` display rotation in Marchine via the independent [Hypr Rotate](https://github.com/Magik23/hypr-rotate) (`hypr-rotate`) utility, with chord suppression, missing-command feedback, and no desktop dependencies.
- Redesigned arcade-first controller navigation: Left/Right browse letters, View switches Arcade/Games, Menu opens Credits, LB toggles Hide Non-Arcade, RB cycles Theme, and `View + Menu` remains Quit; L3/R3 are unassigned.
- Added keyboard/controller-accessible Theme to Info Quick Settings, named theme labels, aligned shortcut legends, and fixed keyboard letter-jump confirm/cancel; Setup and Refresh remain keyboard-only.
- Improved Arcade list navigation by keeping the selection around three-fifths down the viewport during long scrolling, preserving visibility of upcoming games.
- Restored the Marchine icon in the local desktop launcher installed by the source installer.
- Added a VERTICAL Arcade category filter using MAME display-orientation metadata without changing MAME launch or rotation behavior.
- Removed automatic `-autorol` injection for vertical MAME games so Marchine no longer overrides rotation; MAME configuration and the desktop compositor control presentation.

## 1.24.0
- Simplified desktop launching: `marchine --desktop` now always starts fullscreen through Foot, removing the saved fullscreen preference and hard-coded 1180×900 desktop geometry.
- Preserved Hyprland window state across game launches: user-chosen window geometry remains untouched, while compositor fullscreen is restored when appropriate.
- Improved game-return focus and fullscreen restoration without taking ownership of runtime window sizing or placement.
- Added `View + Menu` as the controller quit chord; X remains Info, Menu remains Setup, and View is reserved as the quit-chord modifier.
- Simplified the launcher footer around primary browsing actions, Info, and Quit while keeping secondary controls documented in the Info screen.

## 1.23.1
- Fixed Staticcheck S1002 warnings in the new gamepad input code so the release workflow passes security/static-analysis validation.

## 1.23.0
- Added native Linux gamepad support using evdev, with capability-based device discovery, hotplug handling, multiple-controller support, and accelerated held navigation.
- Added fixed controller mappings for launch/back, source switching, Index, Categories, Setup, Theme, Info, and Credits; Search remains keyboard-driven.
- Added Index jumping with available first-character buckets, wraparound navigation, confirm, and cancel/restore behavior.
- Added Info-screen Quick Settings for **Fullscreen on Launch** and **Hide Non-Arcade**, navigable by keyboard or controller and persisted in Marchine's config.
- Added the `marchine --desktop` bootstrap so Marchine can read its fullscreen preference before creating its Foot window; the setting affects Marchine only and leaves MAME display configuration untouched.
- Aligned the local installer and packaged Linux/AUR desktop entry on the same Foot-based desktop launch path, with Foot declared as a package runtime dependency.
- Exposed Setup in the main footer and hardened page composition so the footer remains visible at the supported minimum terminal geometry.
- Vendored `github.com/holoplot/go-evdev` and added regression coverage for gamepad input, Quick Settings persistence, and normal/fullscreen desktop launch arguments.

## 1.22.1
- Added the official Marchine application icons in 128, 256, and 512 pixel sizes.
- Added a generic Linux desktop entry with `Terminal=true` and `Icon=marchine`, without requiring Foot.
- Updated AUR packaging to install the desktop entry and icons into the standard system application and hicolor icon paths.
- Kept the Foot-specific launcher and Omarchy/Hyprland integration as optional local examples, and extended CI validation to cover both desktop entries.

## 1.22.0
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
