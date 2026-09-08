# MARCHINE

## Game Machinery.

Marchine is a terminal-native game launcher for a local, direct, keyboard-first Linux workflow.

It is intentionally not a game encyclopedia. Marchine does not scrape artwork, ratings, achievements, reviews, store metadata, or cloud accounts. It indexes what you can launch and gets out of the way.

## Prototype V1.20 — native ANSI geometry + clean Arcade filtering

V1.20 keeps the established launcher/category workflow and tightens the sculpture presentation: the supplied 60×25 ANSI fightstick is embedded exactly, every source row is preserved, and Marchine adds no synthetic blank rows above or below the artwork.

### V1.20 interface rules

- one embedded ANSI fightstick sculpture with immutable glyph geometry and no runtime resampling;
- exact one-cell outer gutter on every side of the terminal canvas;
- top navigation contains only **GAMES** and **ARCADE**;
- `C` opens the scrolling CatVer-backed Arcade category menu;
- `I` opens INFO and `O` opens CREDITS; legacy Ctrl shortcuts still work but are no longer shown;
- `R` refreshes the Arcade library from MAME/local metadata;
- INFO exposes `H  HIDE NON-ARCADE [ON/OFF]` and persists it in the existing `hide_junk` config field;
- with Hide Non-Arcade **ON**, Marchine shows only arcade entries and removes `ALL`, casino/gambling, mahjong, and other all-non-arcade families from the category menu;
- with it **OFF**, `ALL`, `CASINO / GAMBLING`, `MAHJONG`, `NON - ARCADE`, and any other indexed category families become available again;
- the complete installed runnable MAME index is retained either way; the toggle changes presentation, not ROM files or index contents;
- unknown future CatVer category families remain visible dynamically;
- INFO and CREDITS remain separate screens;
- CREDITS stays the final main-footer item;
- the top `GAME MACHINERY.` identity is given a little more visual width without becoming a second giant logo;
- minimum supported terminal geometry is **112 columns × 48 rows**; below that Marchine shows only the CLIAMP-style resize message.

## Sources

### GAMES

A flat list of launchable things. Display identity stays separate from launch identity.

```toml
[[games]]
name = "Terminator 2D"
command = "/home/user/Games/Terminator2D/Terminator2D.x86_64"
args = []
working_dir = "/home/user/Games/Terminator2D"
```

### ARCADE

The current automatic Arcade driver is MAME.

Marchine reads local MAME data and shows only useful browsing metadata:

```text
NAME                     YEAR   MANUFACTURER
DoDonPachi                1997   Cave
Battle Garegga            1996   Raizing
R-Type                    1987   Irem
```

No ROMs are renamed, moved, repaired, or deleted.

## Setup

Press `S`.

The in-app Setup screen exposes:

```text
EMULATOR EXECUTABLE
ROM PATH
```

The automatic Arcade driver is MAME, while the executable + ROM-path relationship is intentionally reusable for future command-line emulator drivers.

`Ctrl+S` saves. `R` inside Setup saves and refreshes the Arcade library.

## Front footer

The normal library screen intentionally stays minimal:

```text
ENTER PLAY  ·  / SEARCH  ·  TAB SOURCE  ·  C CATEGORIES  ·  I INFO  ·  O CREDITS
```

Within the supported 112×48-or-larger layout, the footer still shortens if needed; below the minimum size the launcher is hidden entirely rather than collapsed.

## Full controls

The complete reference remains in `I INFO`:

```text
↑ / ↓, j / k   Navigate
/              Search
Tab            Games / Arcade
C              Categories (Arcade)
Enter          Launch
R              Refresh Arcade Library
S              Setup
I              Info
O              Credits
Q              Quit

INFO:
H              Toggle Hide Non-Arcade

Category menu:
↑ / ↓, j / k   Select category
Enter          Apply category
Esc / C        Return
```

## Build

Requirements:

- Linux
- Go 1.23.2+ (the Makefile uses `GOTOOLCHAIN=auto`)
- Unicode/truecolor terminal recommended
- MAME only for the automatic Arcade driver

```bash
make build
./dist/marchine --demo
```

The release ZIP includes `vendor/`, so normal `make build`, `make test`, and `make demo` do not need to download dependencies. Use `make vendor-refresh` only when intentionally updating the dependency graph.

For the real library:

```bash
./dist/marchine
```

## Philosophy

Marchine is not a frontend museum. It is a terminal-native index of things you can launch.

If a launch source already knows useful information locally, Marchine may show it. Marchine does not go onto the Internet to decorate your library.

## License

MIT. See `LICENSE`.
