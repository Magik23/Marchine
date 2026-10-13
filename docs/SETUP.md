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

Select **ADD GAME**, **EDIT GAME**, or **REMOVE GAME** with the arrows/Tab and press Enter. Edit/Remove show the existing entries in a selection list. Removal requires **Y YES** confirmation; `N` or `Esc` cancels. Removing a Marchine entry never deletes any game/application files.

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

Saving changes `~/.config/marchine/config.toml` and updates the visible GAMES list **immediately**, without restarting Marchine or pressing `R`. In the GAMES section, `R` has no action.

## Common controls and storage

- `Q` quits Marchine when not editing text. `Ctrl+C` is intentionally ignored, including in active fields.
- `T` cycles theme from Setup.
- Foot uses `Ctrl+Shift+C` for copying selected terminal text and `Ctrl+Shift+V` for pasting.
- Configuration is saved atomically, normally to `~/.config/marchine/config.toml`. Manual TOML editing remains supported.
