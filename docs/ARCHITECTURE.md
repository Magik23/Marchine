# Marchine architecture — V1.20

## Layers

```text
cmd/marchine
    startup / flags / program lifecycle

internal/config
    TOML config, path expansion, atomic save

internal/library
    normalized launchable game model

internal/mame
    current automatic Arcade driver
    ROM discovery + local MAME metadata + cache

internal/ui
    Bubble Tea model/update/view
    themes
    responsive layout
    rigid terminal sculpture renderer
```

## Launch model

Every selected item resolves to:

```text
command + args + optional working directory
```

Marchine uses Bubble Tea process handoff so the game owns the terminal while running. Marchine resumes after the child process exits.

## Arcade driver

The current automatic indexing driver is MAME-specific internally even though Setup uses the neutral label `EMULATOR EXECUTABLE`.

The MAME driver reads local information using commands such as:

```text
mame -version
mame -listxml
```

and indexes only ROMs present in the configured ROM path(s).

## Sculpture renderer

`internal/ui/fightstick.ans` is embedded by `internal/ui/sculpture.go`. Rendering is 1:1; Marchine derives the current asset dimensions at runtime and sizes the detail column from the real ANSI width. It never resamples or reflows the character matrix. Limited vertical space uses a stable crop. Insufficient width hides the sculpture rather than deforming it.

## Minimum terminal geometry

The normal launcher is rendered only at **100×48 terminal cells or larger**. Below that threshold `View()` returns a dedicated resize panel instead of invoking any compact/stacked launcher layout. This keeps the two-column game browser visually rigid.
