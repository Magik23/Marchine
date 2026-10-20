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
3. Verify and install the executable and its icons (replace `1.28.0` with the version you downloaded):

```bash
VERSION=1.28.0
ARCHIVE="marchine-${VERSION}-linux-amd64"

sha256sum -c "${ARCHIVE}.tar.gz.sha256"
tar -xzf "${ARCHIVE}.tar.gz"
install -Dm755 "${ARCHIVE}/marchine" "$HOME/.local/bin/marchine"

for size in 128 256 512; do
  install -Dm644 "${ARCHIVE}/share/icons/hicolor/${size}x${size}/apps/marchine.png" \
    "$HOME/.local/share/icons/hicolor/${size}x${size}/apps/marchine.png"
done

"$HOME/.local/bin/marchine" --init-config
"$HOME/.local/bin/marchine"
```

The prebuilt archive includes icons but **does not automatically install a desktop entry**. The source installer below creates a Foot desktop launcher where Foot is available. Running Marchine directly in an existing terminal is supported without Foot.

If `~/.local/bin` is not already on your `PATH`, add it through your shell configuration.

### Requirements

- Linux
- A Unicode-capable terminal; truecolor is recommended
- MAME for the automatic **ARCADE** source
- No MAME requirement for manually configured **GAMES** entries
- Foot only when using Marchine's packaged/local desktop launcher

Marchine is developed on Arch Linux and fits naturally into keyboard-driven desktops such as Omarchy and Hyprland. It is not tied to Omarchy itself.

## What Marchine does

Marchine has three intentionally simple sources: **ARCADE | GAMES | PLAYLISTS**. `Tab` / controller View cycles among them. The source bar shows the appropriate game or playlist count and updates with source changes, category filters, searches, and Hide Non-Arcade.

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

### GAMES

A flat list of native Linux executables, scripts, commands on `PATH`, and standard Linux `.desktop` launchers. Use **`S` → GAMES → ADD GAME** for individual entries or **IMPORT GAMES** to select multiple local launch targets at once. **EDIT GAME** and **REMOVE GAME** manage saved entries without touching files on disk.

Each entry has a display name, executable, optional arguments, and optional working directory. Arguments are passed to the chosen executable, without shell interpretation. For example, this manually configured entry is equivalent to one created through Setup:

```toml
[[games]]
name = "OpenTyrian"
command = "opentyrian"
args = []
```

Games added, imported, or edited in Setup appear in GAMES immediately and persist across restarts; no library refresh is required. Launch targets requiring external applications or compatibility runners depend on those tools being installed separately.

### PLAYLISTS

PLAYLISTS is Marchine's third source: your own collections of **ARCADE and GAMES titles in one playable list**. A playlist may contain MAME ROMs alongside native executables, scripts, or configured emulator commands. Each game can belong to several playlists, and titles stay in the order you added them.

**The distinction is intentional:** ARCADE and GAMES browse the original libraries and let you **save to an existing playlist**; only PLAYLISTS creates, manages, opens, and plays collections. Playlists are never loaded as category filters in ARCADE or GAMES. Switch sources with `Tab` or controller **View**.

#### 1. Create a playlist

1. Switch to **PLAYLISTS** using `Tab` / View.
2. Press **`N`** to create a playlist, enter a name (for example, `Favorites`), and press `Enter` to save.
3. Repeat for other collections such as `SHMUP Essentials` or `Couch Gaming`. Empty playlists are allowed. Names cannot be blank, exceed 48 characters, or duplicate another playlist name (case-insensitive).

Playlist names are typed with a **keyboard**; `Esc` cancels without saving. You must create the playlist here **before** it appears as a save destination in ARCADE or GAMES.

#### 2. Add games from ARCADE or GAMES

1. Browse **ARCADE** or **GAMES** and highlight a title.
2. Press **`P`** or controller **View + Y** to open **SAVE TO PLAYLIST**.
3. Select an existing destination with Up/Down; press **Enter / A** to save, or **Esc / B** to cancel.
4. Repeat to add titles from either source to the **same** playlist. A game can be in multiple playlists; saving the same game to one playlist twice does not add a duplicate.

If no playlists exist, Marchine tells you to create one in PLAYLISTS first. **Saving a game does not switch the current source or open the playlist.**

#### 3. Open and play a collection

1. Switch to **PLAYLISTS** and highlight a playlist.
2. Press **Enter / A** to open its unified game list.
3. Navigate, search with `/`, and launch any available game with **Enter / A**, as in the normal library.
4. Press **Esc / B** to return to the PLAYLISTS manager. At the top-level manager, **Esc / B does not switch sources**; use `Tab` / View for that.

Marchine uses the original ARCADE or GAMES launch configuration automatically. A missing game reference stays visible as **`[UNAVAILABLE]`**, rather than being silently replaced or deleted; it cannot launch until that source game is restored.

#### 4. Rename, remove, or delete

| Where | Keyboard | Controller | Action |
| --- | --- | --- | --- |
| PLAYLISTS manager | `N` | — | Create a playlist |
| PLAYLISTS manager | `R` | — | Rename the selected playlist |
| PLAYLISTS manager | `Enter` | A | Open the selected playlist |
| PLAYLISTS manager | `D` | — | Request deletion of the selected playlist |
| Inside a playlist | `D` | — | Request removal of the selected game *from this playlist* |
| In a deletion/removal prompt | **`Y`** | — | Confirm (keyboard only) |
| In a deletion/removal prompt | `N` / `Esc` | B (cancel) | Cancel |
| Inside a playlist | `Esc` | B | Back to the PLAYLISTS manager |
| At the PLAYLISTS manager | `Tab` | View | Switch to the next source |

Deleting a playlist or removing an entry changes **only the saved collection**. It never deletes a ROM, an executable, a desktop shortcut, or a configured GAMES entry. Keyboard confirmation is intentional for destructive collection actions; naming, renaming, and deletion aren't fully controller-only in this release.

#### Storage and upgrades

Playlists and GAMES entries are stored in `~/.config/marchine/config.toml`. Each configured GAMES entry now has a **stable `id`**, so editing, removing, or reordering other games doesn't redirect playlist references to the wrong executable. MAME entries use their ROM identity.

When upgrading from earlier Marchine versions, existing `[[games]]` entries without IDs **load normally and receive stable IDs in memory**. Marchine writes them on the next successful configuration save (for example, when creating a playlist); merely launching the new version does not rewrite the file. **No reimport is necessary.** Saves are atomic. Manual TOML edits remain possible, but saving reserializes the file, so back up custom formatting/comments and your configuration before deliberately downgrading.

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

Marchine treats categories as filters over the indexed MAME library, separate from user-created collections in PLAYLISTS.

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

Marchine keeps setup, shortcuts, project information, and credits inside the TUI. Shortcut hints use accent-colored keys, regular action text, and muted separators drawn from the **active theme** (Marchine, Plasma, Phosphor, or Ember), without decorative keycap boxes.

### Main controls

| Keyboard | Controller | Action |
| --- | --- | --- |
| `↑` / `↓`, `j` / `k` | D-Pad / LS ↑ / ↓ | Navigate |
| `[` / `]` | D-Pad / LS ← / →; LT / RT | Prev / Next Letter |
| `Enter` | A | Launch / Apply |
| `Esc` | B | Back / Cancel (no source change at top level) |
| `Tab` | View | Cycle Arcade / Games / Playlists |
| `P` | View + Y | Save selected ARCADE/GAMES game to playlist |
| `C` | Y | Categories |
| `I` | X | Info |
| `O` | Menu | Credits |
| `H` | LB | Hide Non-Arcade |
| `T` | RB | Theme |
| — | View + RB | Rotate Display |
| `Ctrl++` / `Ctrl+-` / `Ctrl+0` | View + X / View + A | Increase / decrease / reset Foot font (controller: zoom only) |
| `Q` | View + Menu | Quit |
| `/` | — | Search |
| `S` | — | Setup |
| `R` | — | Refresh Arcade Library |

For playlist-specific creation, rename, removal, and confirmation controls, see [PLAYLISTS](#playlists). Collections open and play only in PLAYLISTS, never as filters on ARCADE or GAMES.

Standalone View switches sources; standalone Menu opens Credits; standalone RB changes Theme. All three fire on release so `View + Menu` quits and `View + RB` rotates without triggering any of their individual actions. Press either button first, then the other; release both before repeating a chord. **For font zoom, hold View first, then tap X (larger) or A (smaller)**; normal X/A actions are suppressed during the chord. L3/R3 have no launcher action. Setup, Refresh, and Search remain keyboard-only.

**Desktop icons:** The source installer puts Marchine icons in `~/.local/share/icons/hicolor/` and creates a desktop entry when Foot is available. Release archives include the same 128/256/512 px icons; see the prebuilt install commands above.

**Running version:** The Info screen's source bar shows the compact release number derived from the running executable. Development commit details are hidden there; use `marchine --version` for the full build identifier. The Info page also links to this README for detailed guidance.

**Font zoom:** Foot already supports `Ctrl++`/`Ctrl+=` (larger), `Ctrl+-` (smaller) and `Ctrl+0` (reset). The View+X/A controller chords send those existing Foot zoom keys to the **focused Marchine desktop window on Hyprland**; they require `hyprctl` and do not change global Foot settings. Outside Hyprland or when run inside another terminal, use that terminal's own font shortcuts. No font size or window size is saved by Marchine. If the terminal falls below 124×48 cells, a resize hint replaces the full UI; View+A is still accepted to shrink the font.

**Optional display rotation:** `View + RB` invokes the [Hypr Rotate](https://github.com/Magik23/hypr-rotate) executable (`hypr-rotate`) from `PATH` if installed. Marchine does not install, configure, or depend on that utility, Hyprland, or Omarchy. Without the utility, Marchine remains fully functional and displays a nonfatal status message. Keyboard display rotation is configured separately in the desktop environment (for example `Super + Alt + R` in Hyprland), not by Marchine.

**For developers modifying bindings:** `View + RB` is a separate, optional external-command chord — not an ordinary Marchine button assignment. Its handling lives in `internal/gamepad/manager_linux.go` (`shortcutChordState` and `chordButton`), while `internal/ui/gamepad.go` dispatches `ActionRotateDisplay` and `internal/ui/model.go` executes `hypr-rotate`. To remap or remove this integration, update those paths together with the Info legend and associated tests. Simply uninstalling `hypr-rotate` prevents rotation but does **not** disable chord interception: `View + RB` will still suppress the usual View and RB actions.

Marchine also supports `PgUp`, `PgDown`, `Home`, and `End` for longer lists.

On Linux, compatible evdev controllers are discovered by capabilities rather than fixed `/dev/input/event*` paths. Hotplug and multiple connected controllers are supported, and held D-Pad / left-stick navigation accelerates automatically. Search remains keyboard-driven.

**Prev / Next Letter** uses only first-character buckets present in the current list. Keyboard `[` / `]`, joystick Left/Right, or LT/RT cycle with wraparound; `Enter` / A confirms the jump, while `Esc` / B cancels and restores the original position.

## Setup

Open Setup with `S` from the library, Info, Credits, or Categories. Source switching with `Tab` remains restricted to the library.

Setup lists **ARCADE** first, then **GAMES**. In GAMES, select **ADD GAME**, **IMPORT GAMES**, **EDIT GAME**, or **REMOVE GAME**. Add/Edit uses four fields (Name, Executable, Arguments, Working Directory); `Enter` edits/applies a field and `Ctrl+S` saves. In REMOVE GAME, `Enter` requests removal of the highlighted entry, while `A` (controller Menu) requests **Remove All**. Both removal actions require a separate **keyboard `Y`** confirmation. Removing entries never deletes executables, shortcuts, ROMs, or Arcade settings.

**Bulk import:** Enter an exact path, immediate directory, or glob such as `~/.local/share/applications/*.desktop` or `~/Games/*.AppImage`. Marchine previews candidates in a scrolling, multi-select list (including large collections). Use `Space`/`Enter` to toggle, `A` to select all, `N` for none, and `I`/`Ctrl+S` to request import. **Only keyboard `Y` confirms** the final save; `N` or `Esc` cancels. Standard Linux `.desktop` launchers use the system `gio launch` mechanism; they remain in their original locations. See the [Setup guide](docs/SETUP.md#batch-import--import-games).

Executables can be native Linux binaries, scripts, or commands on `PATH`. For Windows `.exe` games, provide a compatibility runner as the executable (for example, `wine`) and the Windows game path as an argument. Arguments support quoted paths containing spaces. While editing a field, use **Ctrl+Shift+V** to paste clipboard text in Foot. **Ctrl+S** applies the current text and saves immediately (no separate Enter required); Enter applies and moves to the next field, and Esc cancels the field edit. Foot uses Ctrl+Shift+C to copy selected terminal text. Marchine ignores Ctrl+C on all screens; use Q to quit. In Arcade Setup, Ctrl+S saves the Arcade settings; R refreshes the Arcade index using the saved settings **without saving pending edits**. GAMES entries update as soon as they are saved in Add/Edit; R does not refresh GAMES.

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

For the **full local desktop installation** (binary, default config if absent, icons, and a Foot desktop entry when Foot is installed), run:

```bash
./scripts/install.sh
"$HOME/.local/bin/marchine" --version
```

The installer preserves an existing `~/.config/marchine/config.toml`. If you want only the binary with no desktop integration, `make install` remains available.

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
