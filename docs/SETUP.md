# Marchine Setup

Press `S` to open Setup. **ARCADE comes first, followed by GAMES**, matching the launcher source bar.

## ARCADE Setup

The current automatic Arcade driver is MAME. Two editable fields are provided:

```text
EMULATOR EXECUTABLE
PRIMARY ROM PATH
```

The neutral executable label allows future command-line emulator drivers to reuse the model.

| Key | Behavior |
| --- | --- |
| `↑` / `↓`, `Tab` | Select a field |
| `Enter` | Edit the selected field; while typing, apply its text |
| `Ctrl+Shift+V` | Paste into the active field in Foot |
| `Ctrl+U` | Clear the active field buffer |
| `Ctrl+S` | Save Arcade settings, including the currently edited value |
| `R` | **Refresh Arcade only**, using the last saved settings; never save pending edits |
| `Esc` | Cancel the current field edit or return to the library |

`Ctrl+S` does not force a metadata refresh. After saving, current MAME command/ROM-path launch settings are applied in memory immediately. Refresh with `R` when the indexed ROM collection or metadata needs rebuilding. Only one refresh runs at a time.

### ROM-path behavior

- An empty primary ROM path restores MAME `rompath` / Marchine auto-detection.
- TOML supports multiple explicit ROM paths.
- Setup edits the **first** path and normally preserves additional configured paths; intentionally clearing the primary field resets the explicit list to auto-detection.
- Explicit Marchine ROM paths are passed to MAME when launching Arcade games.

## GAMES Setup

GAMES is a general-purpose executable launcher. It does not require MAME or online game accounts.

Select **ADD GAME**, **IMPORT GAMES**, **EDIT GAME**, or **REMOVE GAME** with the arrows/Tab and press Enter. Edit/Remove show existing entries in a selection list. In **REMOVE GAME**, `Enter` requests removal of the highlighted entry; `A` (controller **Menu**) requests **Remove All**. Both display a confirmation: **keyboard `Y` is the only way to proceed**; `N` or `Esc` cancels. `Enter` and controller A do not confirm removal. Remove All clears only configured GAMES entries in one atomic save; it never changes ARCADE settings or deletes executables, desktop shortcuts, or game files.

Add/Edit provide four fields:

| Field | Meaning |
| --- | --- |
| **NAME** | Display name in the GAMES library |
| **EXECUTABLE** | Executable path or a command available on `PATH` |
| **ARGUMENTS** | Optional arguments passed directly to that executable |
| **WORKING DIRECTORY** | Optional directory in which to start the process |

Press `Enter` to edit a field. `Enter` applies its text and advances to the next field. **`Ctrl+S` saves immediately, even if you are still typing.** `Ctrl+Shift+V` pastes in Foot; `Ctrl+U` clears the active input, and `Esc` cancels the current field edit. Missing required fields or malformed quoting are reported without silently discarding the entry.

For example, to launch VS Code with a fresh window:

```text
NAME               Visual Studio Code
EXECUTABLE         code
ARGUMENTS          --new-window
WORKING DIRECTORY  (leave empty)
```

Quote an argument when a filename contains spaces. Marchine splits the entered arguments into individual values; it **does not run a shell** or interpret shell operators (`&&`, `|`, redirection, etc.). To start a Windows `.exe`, specify an installed runner such as `wine` as EXECUTABLE and the game path as an argument. A script may require executable permission or an interpreter.

Saving changes `~/.config/marchine/config.toml` and updates the visible GAMES list **immediately**, without restarting Marchine or pressing `R`. In the GAMES section, `R` has no action. GAMES entries have persistent internal IDs: editing an entry preserves its ID so that references from [PLAYLISTS](../README.md#playlists) continue to resolve to the same game. Older configurations acquire IDs automatically on the next normal save; no reimport is necessary.

### Batch import — IMPORT GAMES

Use **IMPORT GAMES** to bring in multiple native executables, executable scripts, AppImages, and standard Linux `.desktop` launchers. Import is path-driven: it has **no file browser**, performs no installation or downloads, and never executes a candidate during discovery.

1. Enter one path or filename pattern, for example `~/.local/share/applications/*.desktop`, `~/Games/*.AppImage`, `~/Games/*/start.sh`, or a directory of launchers. A directory lists its immediate files, **not** nested folders.
2. Press **Enter** to preview matches in an Arcade-style scrollable playlist. Large sets (including 200+ matches) remain navigable; selection persists off-screen. New matches start selected; already-imported entries are marked `[-]` and cannot be selected again.
3. Use **Up/Down** to scroll, **Space/Enter** to toggle the highlighted entry, **A** to select all, **N** to clear all, and **Home/End/PageUp/PageDown** for fast navigation. The selected count stays visible.
4. Press **I** or **Ctrl+S** to request import. At the final prompt, **only keyboard `Y` saves**; `N`/`Esc` cancels. `Enter` and controller A do not confirm. Before the prompt, controller A toggles a candidate and controller Menu requests import. `Esc` from the preview returns to the path input without modifying configuration.

Standard Linux `.desktop` shortcuts are launched through `gio launch <path-to-shortcut>`; the GIO utility and anything referenced by the shortcut must already be available. Shortcuts are **referenced**, not copied, so moving or deleting the original shortcut may break its GAMES entry. A desktop launcher may return immediately after handing off to another process; Marchine avoids forcibly restoring focus at that moment. Deleting the Marchine entry never deletes the shortcut. Windows `.lnk` files are not supported. For Windows `.exe` files, use ADD GAME and configure a suitable installed runner instead of importing the `.exe` directly.

A relative path is interpreted relative to Marchine's process working directory. Simple Go `filepath.Glob` patterns are supported, but recursive `**` is **not**. Only executable files and readable `Type=Application` desktop shortcuts are candidates; unrelated text files, hidden desktop entries, and directories are ignored.


The import screen does not switch sources with `Tab`. Outside active field editing, `S` opens Setup from Info, Credits, and Categories as well as from the main library. Across screens, shortcut hints separate theme-colored **keys** from normal-colored **actions**, with muted separators.

## Common controls and storage

- `Q` quits Marchine when not editing text. `Ctrl+C` is intentionally ignored, including in active fields.
- `T` cycles theme from Setup.
- Foot uses `Ctrl+Shift+C` for copying selected terminal text and `Ctrl+Shift+V` for pasting.
- Configuration is saved atomically, normally to `~/.config/marchine/config.toml`. Manual TOML editing remains supported.
