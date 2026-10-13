# Marchine

**Game Machinery.**

**Linux-first, terminal-native game launcher. Built on Arch. At home in Omarchy. MAME-first by design.**

[Get Marchine](https://github.com/Magik23/Marchine/releases/latest) · [Albenoir Lab](https://albenoir.com/lab/)

![Marchine browsing a local MAME library on Linux](docs/images/marchine-library.png)

Marchine is a focused, keyboard-and-controller launcher for games you already own and can run locally.

It is intentionally **not** a game encyclopedia. There is no artwork scraping, account layer, ratings wall, achievements feed, store metadata, or cloud library. Marchine indexes what you can launch, keeps useful local metadata close, and gets out of the way.

## Get Marchine

### Prebuilt Linux release

The current GitHub release workflow publishes a **Linux amd64** archive and SHA-256 checksum.

1. Open the [latest release](https://github.com/Magik23/Marchine/releases/latest).
2. Download:
   - `marchine-X.Y.Z-linux-amd64.tar.gz`
   - `marchine-X.Y.Z-linux-amd64.tar.gz.sha256`
3. Verify and install:

```bash
sha256sum -c marchine-X.Y.Z-linux-amd64.tar.gz.sha256
tar -xzf marchine-X.Y.Z-linux-amd64.tar.gz

install -Dm755 marchine "$HOME/.local/bin/marchine"

marchine --init-config
marchine
```

If `~/.local/bin` is not already on your `PATH`, add it through your shell configuration.

### Requirements

- Linux
- A Unicode-capable terminal; truecolor is recommended
- MAME for the automatic **ARCADE** source
- No MAME requirement for manually configured **GAMES** entries
- Foot only when using Marchine's packaged/local desktop launcher

Marchine is developed on Arch Linux and fits naturally into keyboard-driven desktops such as Omarchy and Hyprland. It is not tied to Omarchy itself.

## What Marchine does

Marchine has two intentionally simple sources. The source bar shows the number of games in the current view, updating with source changes, category filters, searches, and Hide Non-Arcade.

### GAMES

A flat list of things you can execute.

Display identity stays separate from launch identity:

```toml
[[games]]
name = "OpenTyrian"
command = "opentyrian"
args = []
```

You can also launch scripts, native executables, Steam commands, or other local targets.

### ARCADE

The current automatic Arcade driver is **MAME**.

Marchine discovers installed ROMs, asks MAME for canonical machine identity and metadata, optionally applies local CatVer category data, and builds a persistent local Arcade library.

Typical visible metadata is deliberately restrained:

```text
NAME                     YEAR   MANUFACTURER
DoDonPachi               1997   Cave
Battle Garegga           1996   Raizing
R-Type                   1987   Irem
```

Marchine does not rename, move, repair, or delete ROMs.

## Fast startup, explicit refresh

Marchine keeps the indexed Arcade library locally so normal startup does not need to rescan thousands of ROM files or ask MAME for metadata every time.

```text
Normal launch
    ↓
Load saved Arcade library
    ↓
Browse / search / filter
```

When your ROM collection changes, press:

```text
R
```

to **Refresh Arcade Library**.

The refresh pipeline uses the configured ROM source, MAME metadata, and optional CatVer data to rebuild the saved library.

A temporarily unavailable ROM drive or missing MAME executable should not silently erase a previously valid library.

The normal cache lives at:

```text
~/.cache/marchine/arcade.json
```

For diagnostics:

```bash
marchine --doctor
marchine --doctor --rescan
```

## Categories without the encyclopedia

![Marchine arcade categories and local metadata filters](docs/images/marchine-categories.png)

Marchine treats categories as views over one indexed library rather than separate duplicated playlists.

CatVer is optional local metadata. When available, Marchine can expose launcher-oriented groups such as:

```text
SHMUPS
FIGHTING
RUN & GUN
BEAT 'EM UP
RACING
PUZZLE
```

`HIDE NON-ARCADE` is non-destructive. It changes what the interface presents; it does not delete entries from the underlying index or touch ROM files.

## Built-in help, small by design

![Marchine built-in help, shortcuts and Linux-first workflow](docs/images/marchine-help.png)

Marchine keeps setup, shortcuts, project information, and credits inside the TUI.

### Main controls

| Keyboard | Controller | Action |
| --- | --- | --- |
| `↑` / `↓`, `j` / `k` | D-Pad / LS ↑ / ↓ | Navigate |
| `[` / `]` | D-Pad / LS ← / →; LT / RT | Prev / Next Letter |
| `Enter` | A | Launch / Apply |
| `Esc` | B | Back / Cancel |
| `Tab` | View | Games / Arcade |
| `C` | Y | Categories |
| `I` | X | Info |
| `O` | Menu | Credits |
| `H` | LB | Hide Non-Arcade |
| `T` | RB | Theme |
| — | View + RB | Rotate Display |
| `Q` | View + Menu | Quit |
| `/` | — | Search |
| `S` | — | Setup |
| `R` | — | Refresh Arcade Library |

Standalone View switches sources; standalone Menu opens Credits; standalone RB changes Theme. All three fire on release so `View + Menu` quits and `View + RB` rotates without triggering any of their individual actions. Press either button first, then the other; release both before repeating a chord. L3/R3 have no launcher action. Setup, Refresh, and Search remain keyboard-only.

**Desktop icons:** The local installer installs the Marchine icons under `~/.local/share/icons/hicolor/`.

**Running version:** The Info screen's source bar shows the compact release number derived from the running executable. Development commit details are hidden there; use `marchine --version` for the full build identifier.

**Optional display rotation:** `View + RB` invokes the [Hypr Rotate](https://github.com/Magik23/hypr-rotate) executable (`hypr-rotate`) from `PATH` if installed. Marchine does not install, configure, or depend on that utility, Hyprland, or Omarchy. Without the utility, Marchine remains fully functional and displays a nonfatal status message. Keyboard display rotation is configured separately in the desktop environment (for example `Super + Alt + R` in Hyprland), not by Marchine.

**For developers modifying bindings:** `View + RB` is a separate, optional external-command chord — not an ordinary Marchine button assignment. Its handling lives in `internal/gamepad/manager_linux.go` (`shortcutChordState` and `chordButton`), while `internal/ui/gamepad.go` dispatches `ActionRotateDisplay` and `internal/ui/model.go` executes `hypr-rotate`. To remap or remove this integration, update those paths together with the Info legend and associated tests. Simply uninstalling `hypr-rotate` prevents rotation but does **not** disable chord interception: `View + RB` will still suppress the usual View and RB actions.

Marchine also supports `PgUp`, `PgDown`, `Home`, and `End` for longer lists.

On Linux, compatible evdev controllers are discovered by capabilities rather than fixed `/dev/input/event*` paths. Hotplug and multiple connected controllers are supported, and held D-Pad / left-stick navigation accelerates automatically. Search remains keyboard-driven.

**Prev / Next Letter** uses only first-character buckets present in the current list. Keyboard `[` / `]`, joystick Left/Right, or LT/RT cycle with wraparound; `Enter` / A confirms the jump, while `Esc` / B cancels and restores the original position.

## Setup

Open Setup with:

```text
S
```

Setup lists **ARCADE** first, then **GAMES**. In GAMES, select **ADD GAME**, **EDIT GAME**, or **REMOVE GAME**. Add/Edit opens a four-field editor (Name, Executable, Arguments, Working Directory); Enter edits a field and Ctrl+S saves. Edit/Remove first shows the configured games, and Remove requires explicit Y confirmation. Entries update the Games library immediately; removing an entry never deletes files from disk.

Executables can be native Linux binaries, scripts, or commands on `PATH`. For Windows `.exe` games, provide a compatibility runner as the executable (for example, `wine`) and the Windows game path as an argument. Arguments support quoted paths containing spaces. Use Ctrl+S in the game editor to save; Esc cancels. In Arcade Setup, Ctrl+S saves the Arcade settings; R refreshes the Arcade index using the saved settings **without saving pending edits**. GAMES entries update as soon as they are saved in Add/Edit; R does not refresh GAMES.

The two primary Arcade fields are deliberately neutral:

```text
EMULATOR EXECUTABLE
PRIMARY ROM PATH
```

The current automatic driver is MAME, but the executable + ROM-path relationship is intentionally reusable for future command-line emulator drivers.

Configuration is stored at:

```text
~/.config/marchine/config.toml
```

The Info screen contains **Quick Settings** for **Hide Non-Arcade** and **Theme**. Up/Down selects a setting; Enter/A applies it. LB and RB are direct controller shortcuts. Both settings persist to the same config file. Theme labels are **MARCHINE → PLASMA → PHOSPHOR → EMBER**, while existing TOML identifiers remain `violet`, `cyan`, `green`, and `amber`.

You can create a starter configuration explicitly with:

```bash
marchine --init-config
```

An empty ROM path allows Marchine to use MAME's configured `rompath` and common Linux MAME locations.

### Optional CatVer metadata

Marchine does not scrape online game databases.

If a local CatVer file is available, it can be configured explicitly or discovered from common Linux locations such as:

```text
~/.mame/catver.ini
~/.attract/metadata/catver.ini
~/MAME-Curation/catver.ini
/usr/share/mame/catver.ini
```

## Built with

- **Go**
- **Bubble Tea**
- **Lip Gloss**
- **go-evdev** for native Linux controller input
- **MAME** for the current automatic Arcade driver
- **TOML** for local configuration
- optional **CatVer** for local category metadata

The interface is a native terminal UI. There is no Electron shell, browser renderer, artwork service, or online account layer.

## Build from source

Requirements:

- Linux
- Go **1.26 or newer**
- Git
- `make`

```bash
git clone https://github.com/Magik23/Marchine.git
cd Marchine

make test
make build

./dist/marchine --version
./dist/marchine --demo
```

Install the locally built binary:

```bash
make install
marchine --init-config
marchine
```

Marchine vendors its Go dependencies, so normal build and test targets can use the committed dependency snapshot.

### Desktop launch behavior

The packaged Linux desktop entry and the local Foot template both use `marchine --desktop`. Desktop launch opens Marchine fullscreen through Foot using the active monitor geometry. Marchine does not persist or manage a separate window-size preference.

Users can leave fullscreen and resize or place the existing Marchine window through their desktop compositor. On Hyprland, Marchine remembers only whether its window was compositor-fullscreen immediately before a game launch: if it was fullscreen, that fullscreen state is restored when the game exits; if it was windowed, the user's live geometry is left untouched. Other compositors receive no Marchine-specific runtime window management. Running `marchine` directly inside an existing terminal remains supported and bypasses the desktop bootstrap. Marchine never changes MAME fullscreen/window settings.

An optional Omarchy/Hyprland rule is provided at `packaging/omarchy/marchine.lua` to give Marchine a monitor-sized floating backing window for clean fullscreen handoffs. Packaging does not modify the user's compositor configuration.

## Project structure

```text
cmd/marchine/        program entry point and CLI flags
internal/config/     TOML configuration and path handling
internal/gamepad/    Linux evdev discovery, normalization and repeat handling
internal/library/    normalized launchable game model
internal/mame/       MAME indexing, local metadata and Arcade cache
internal/ui/         Bubble Tea model, views, themes and terminal sculpture
docs/                architecture and project notes
packaging/           distribution packaging work
scripts/             local install helpers
```

## Design boundary

Marchine deliberately separates responsibilities:

```text
MAME
  machine identity
  machine metadata
  ROM dependency rules
  emulation

CatVer
  optional local category metadata

Marchine
  local indexing
  saved records
  search and filters
  category views
  launch orchestration
  terminal interface
```

The guiding idea is simple:

> **MAME defines what the games are. CatVer helps describe what kind of games they are. Marchine remembers the result. The user decides when to refresh it.**

## Small by design

Marchine deliberately does not try to become a game database.

- no artwork scraping
- no account layer
- no ratings or reviews
- no achievement system
- no ROM-file modification
- no provider/store wall
- no encyclopedia sprawl

It indexes what you own and launches it cleanly.

## Status

Marchine is an active Linux-first project. The current automatic Arcade driver is MAME; additional command-line emulator drivers are possible when they fit Marchine's simple model:

```text
EMULATOR + ROM PATH → INDEX → EXECUTE
```

## Credits

**Marchine — Game Machinery.**

Conceived, directed, and designed by **Pierre Dionne / Albenoir Studio**.

Built in Go with Bubble Tea and Lip Gloss.

## License

Marchine is released under the [MIT License](LICENSE).
