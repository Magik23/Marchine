# Marchine architecture

## Layers

```text
cmd/marchine
    flags, config bootstrap, doctor mode, program lifecycle

internal/config
    TOML config, path expansion, atomic save

internal/library
    normalized launchable-game model

internal/mame
    automatic Arcade driver
    ROM discovery, MAME XML parsing, CatVer, cache, launch arguments

internal/ui
    Bubble Tea model/update/view
    themes, fixed responsive layout, embedded ANSI sculpture
```

## Launch model

Every selected item resolves to:

```text
command + args + optional working directory
```

Marchine uses Bubble Tea process handoff so the child application owns the terminal while running. Marchine resumes when the child exits.

GAMES entries come directly from user config. ARCADE entries store durable local metadata plus MAME launch identity.

## Arcade driver

The current automatic Arcade driver is MAME-specific internally even though Setup uses the neutral field label `EMULATOR EXECUTABLE`.

A live refresh performs bounded external calls including:

```text
mame -showconfig
mame -version
mame -listxml
```

Marchine discovers top-level `.zip` and `.7z` ROM archives in the configured/detected ROM directories, then filters MAME XML to machines whose ROM archive is present.

MAME's XML display rotation is retained. A 90°/270° machine is marked vertical and launched with `-autorol`.

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
