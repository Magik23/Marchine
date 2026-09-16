# MARCHINE

## Game Machinery.

Marchine is a terminal-native game launcher for a local, direct, keyboard-first Linux workflow. It is MAME-first without trying to become a game encyclopedia: no artwork scraping, ratings, achievements, cloud accounts, store walls, or ROM modification.

Marchine indexes what is locally launchable, hands the terminal to the selected game, and returns when the child process exits.

## Current interface

- **GAMES** — flat user-defined launch targets.
- **ARCADE** — automatic local MAME indexing.
- Search, CatVer-backed categories, and a non-destructive Hide Non-Arcade filter.
- Four terminal themes: violet, cyan, green, and amber.
- Static embedded ANSI fightstick sculpture; no image renderer or runtime animation.
- Separate Info, Credits, Categories, and Setup screens.
- Minimum supported terminal geometry: **124 columns × 48 rows**.
- Current embedded sculpture geometry: **62 columns × 22 rows**.

Below the minimum geometry Marchine shows a resize message instead of collapsing into a different layout.

## Controls

```text
↑ / ↓, j / k   Navigate
/              Search
Tab            Games / Arcade
C              Categories (Arcade)
Enter          Launch
R              Refresh Arcade Library
S              Setup
T              Cycle theme
H              Toggle Hide Non-Arcade
I              Info
O              Credits
Q              Quit
```

The in-app Info screen is the canonical shortcut reference.

## GAMES

GAMES is a flat list of arbitrary launch targets. Display identity stays separate from launch identity.

```toml
[[games]]
name = "Terminator 2D"
command = "/home/user/Games/Terminator2D/Terminator2D.x86_64"
args = []
working_dir = "/home/user/Games/Terminator2D"
```

A command can also be a launcher such as Steam:

```toml
[[games]]
name = "Terminator 2D"
command = "steam"
args = ["-applaunch", "1718460"]
```

## ARCADE / MAME

The current automatic Arcade driver is MAME. Marchine uses only local data:

- configured or detected ROM paths;
- `mame -version`;
- `mame -listxml`;
- optional CatVer metadata.

The list shows restrained metadata such as name, year, and manufacturer. Marchine never renames, moves, repairs, or deletes ROM files.

### ROM discovery

Marchine indexes top-level `.zip` and `.7z` archives found in the configured/detected ROM directories. Directory-only or CHD-only entries are not independently discovered unless the corresponding parent ROM archive is present.

If `arcade.rom_paths` is empty, Marchine tries MAME's configured `rompath` and common local MAME directories. If paths are explicitly configured in Marchine, those paths are also passed to MAME at launch so indexing and launching use the same source locations.

Multiple explicit ROM paths are supported in TOML. The Setup screen edits the **primary** path and preserves additional configured paths. Clearing the primary path in Setup intentionally returns to MAME rompath auto-detection.

### Portrait games

Marchine reads MAME display rotation metadata. Games whose MAME display reports 90° or 270° are launched with `-autorol`.

### Cache behavior

Arcade metadata is cached at the platform user-cache location (`~/.cache/marchine/arcade.json` on a typical Linux setup) using schema `marchine-index-v4`.

Normal startup is cache-first. A manual refresh performs a live scan. If a current saved library exists and MAME, a ROM path, or a refresh scan temporarily fails, Marchine keeps the saved library rather than destroying it. Cache writes are atomic.

## Setup

Press `S`.

```text
EMULATOR EXECUTABLE
PRIMARY ROM PATH
```

`Ctrl+S` saves. `R` saves and refreshes the Arcade library. Configuration is stored at the platform user-config location (`~/.config/marchine/config.toml` on a typical Linux setup).

A starter file is available as `config.example.toml`, or can be created with:

```bash
marchine --init-config
```

Useful diagnostics:

```bash
marchine --doctor
marchine --doctor --rescan
```

A doctor run exits nonzero when it encounters an operational diagnostic failure.

## Build from source

Requirements:

- Linux
- Go **1.26 or newer**
- Unicode/truecolor terminal recommended
- MAME only if using the automatic Arcade source

Dependencies are committed under `vendor/`, so normal build/test operations are offline-capable.

```bash
make check
make build
./dist/marchine --demo
```

`make vendor-refresh` is only for intentionally updating the Go dependency snapshot.

Source-archive builds without Git metadata use the visible version `dev`. Release/package builds inject the release version explicitly.

## Local desktop integration

From a source checkout, the upstream helper installer places Marchine in `~/.local/bin`, creates a starter config if needed, and installs the optional Foot desktop launcher only when Foot is present:

```bash
./scripts/install.sh
```

The desktop template is `packaging/marchine.desktop.in`.

An optional Omarchy/Hyprland floating-window example is provided at:

```text
packaging/omarchy/marchine.lua
```

It is intentionally **not** installed into a user's compositor configuration by packaging.

## AUR packaging

The upstream repository contains a production AUR template rather than a fake checksum:

```text
packaging/PKGBUILD.example
```

After a real release tag exists, generate the final `PKGBUILD` and `.SRCINFO` on Arch with:

```bash
./packaging/finalize-aur.sh 1.22.0
./packaging/verify-aur.sh
```

Use the actual release version in place of the example above. The AUR package itself is terminal/compositor neutral; Foot and Omarchy integration are shipped only as examples.

## Release policy

Pushes and pull requests run formatting, tests, vet, vulnerability analysis, static analysis, shell validation, and desktop-entry validation in GitHub Actions. Tag builds repeat the source checks before producing a versioned Linux amd64 archive with the MIT license, README, config example, and bundled third-party license texts.

## Philosophy

Marchine is not a frontend museum. It is a terminal-native index of things you can launch.

If a launch source already knows useful information locally, Marchine may show it. Marchine does not go onto the Internet to decorate your library.

## License

MIT. See `LICENSE`.
