# Marchine architecture

## Layers

```text
cmd/marchine
    flags, config bootstrap, doctor mode, program lifecycle

internal/config
    TOML config, path expansion, stable GAMES IDs, playlists, atomic save

internal/library
    normalized launchable-game model and three source identifiers

internal/gamepad
    Linux evdev input, controller actions and modifier chords

internal/mame
    automatic Arcade driver
    ROM discovery, MAME XML parsing, CatVer, cache, launch arguments

internal/ui
    Bubble Tea model/update/view, ARCADE/GAMES browser and PLAYLISTS manager
    themes, fixed responsive layout, embedded ANSI sculpture
```

## Launch model

Every selected item resolves to:

```text
command + args + optional working directory
```

Marchine uses Bubble Tea process handoff so the child application owns the terminal while running. Marchine resumes when the child exits.

GAMES entries are persisted as `[[games]]` configuration. In-app Add/Import/Edit/Remove operations use the existing atomic config save and rebuild the GAMES list immediately. Direct launch arguments are a process argument vector, not a shell expression. ARCADE entries store durable local metadata plus MAME launch identity.

### GAMES import and external desktop shortcuts

`internal/ui/game_import.go` accepts an exact file, the immediate files in a directory, or a simple `filepath.Glob` pattern (`**` recursive scans are not supported). Discovery checks candidates without launching anything. Supported imports include executable local files and standard readable Linux `Type=Application` `.desktop` shortcuts. Preview keeps selected states across a fixed-height, scrollable viewport; existing entries are deduplicated by launch identity (command, arguments, working directory).

Import commits selected new entries in one atomic configuration save. Removing one or all configured GAMES entries changes only `[[games]]` and never deletes the referenced files or changes ARCADE settings. Both import and remove warning screens require an explicit keyboard `Y`—not Enter or controller A.

A `.desktop` item is invoked through `gio launch`, retaining its original path. Because that command can quickly hand off to another process, Marchine avoids forcibly reclaiming focus as soon as the `gio` process returns. These shortcuts depend on the host's GIO utility and their external launch targets.

Keyboard/controller shortcut hints are rendered centrally by `internal/ui/shortcut_hints.go`, using the active theme's accent, text and muted colors.

## Unified PLAYLISTS and durable game identity

Marchine has three peer sources: **ARCADE** indexes MAME ROMs, **GAMES** lists configured launch targets, and **PLAYLISTS** opens mixed, user-curated collections. ARCADE/GAMES offer only **Save to Playlist** (`P` / controller View + Y); creation, renaming, deletion, browsing and playback of a collection are owned by PLAYLISTS. The save picker accepts existing destinations only and does not change the active browsing source.

`internal/config/config.go` stores playlist definitions as `[[playlists]]` with ordered `[[playlists.items]]` references containing the original display name and a stable game identity. MAME references use `mame:<rom>`; configured GAMES references use `game:<id>` in the runtime library, while each `[[games]]` TOML entry stores only its bare generated ID. The ID is a 32-character lowercase hexadecimal value generated from 16 cryptographically random bytes.

`config.Load` accepts older GAMES entries without an `id` and assigns unique IDs **in memory only**. The next successful `config.Save` persists both IDs and playlists; starting Marchine does not rewrite configuration merely to migrate it. `config.Save` reserializes TOML and writes through a temporary file and rename, so hand-written formatting and comments may change when a save occurs. Keep backups before downgrading to an earlier version that does not preserve the new fields.

`internal/ui/playlists.go` resolves saved references against currently indexed ARCADE and configured GAMES entries by ID. Missing entries remain visible as `[UNAVAILABLE]` rather than disappearing or accidentally launching another game. Playlist deletion/removal modifies only stored references, requires an explicit keyboard `Y`, and never copies, moves or deletes the original ROMs, executables or desktop shortcuts. Playlist naming is keyboard-driven; controller navigation and launch remain supported.

At the PLAYLISTS manager root, `Esc`/controller B does not switch sources; only `Tab`/View cycles source tabs. Inside an opened playlist, `Esc`/B returns to the manager. Source counts, borders and contextual shortcut hints use the shared layout conventions.

## Optional Hyprland cursor behavior

`internal/ui/cursor_timeout.go` temporarily applies a short `cursor:inactive_timeout` at runtime only while the Marchine Foot window is focused and gamepad handling is not suspended. It restores the previous value on blur, game handoff or exit when Marchine still owns the override, without changing `cursor.invisible` or writing Hyprland configuration files. It does nothing outside a supported Hyprland session. Optional `View + RB` display rotation remains delegated to the independently installed `hypr-rotate` executable.

## Arcade driver

The current automatic Arcade driver is MAME-specific internally even though Setup uses the neutral field label `EMULATOR EXECUTABLE`.

A live refresh performs bounded external calls including:

```text
mame -showconfig
mame -version
mame -listxml
```

Marchine discovers top-level `.zip` and `.7z` ROM archives in the configured/detected ROM directories, then filters MAME XML to machines whose ROM archive is present.

Marchine does not inject MAME display-rotation flags. Display orientation and rotation remain under MAME's own configuration.

If Marchine has explicit `arcade.rom_paths`, the same paths are passed to MAME using `-rompath` when launching. Auto-detected MAME paths are not unnecessarily forced back onto MAME; in auto mode MAME remains authoritative for its own configured rompath.

## Cache model

The Arcade cache is `arcade.json` under the platform user cache directory and carries schema:

```text
marchine-index-v4
```

Startup is cache-first. A forced refresh updates the cache atomically. A valid current cache is preserved as a fallback when a refresh fails or an external ROM source is temporarily unavailable.

Cached metadata is separated from current launch configuration: when a cache is loaded, Marchine rehydrates MAME command/ROM-path launch arguments from the current config so editing Setup does not leave stale execution identity behind.

Legacy caches are never silently treated as current-schema data. If the live source is unavailable they may be exposed only as an explicit degraded fallback requiring refresh.

## Refresh concurrency

Only one Arcade refresh may be started from the UI at a time. Repeated refresh requests while one scan is in progress do not spawn additional `mame -listxml` processes.

## Sculpture renderer

`internal/ui/fightstick.ans` is embedded by `internal/ui/sculpture.go`.

The current plain geometry is **62×22 terminal cells**. Rendering is 1:1: Marchine does not resample, dither, morph, or reflow the character matrix. The detail panel derives its width from the asset. Limited vertical space may use a stable center crop; insufficient width hides the sculpture rather than deforming it.

## Minimum terminal geometry

The normal launcher is rendered only at **124×48 terminal cells or larger**. Below that threshold `View()` returns a dedicated resize panel instead of switching to a compact/stacked game-browser layout.
