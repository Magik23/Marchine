# Marchine Setup

Open Setup with:

```text
S
```

Fields:

```text
EMULATOR EXECUTABLE
PRIMARY ROM PATH
```

The current automatic Arcade driver is MAME. The neutral executable label is intentional so future command-line emulator drivers can reuse Marchine's launch model.

Controls:

```text
↑/↓ or Tab     Select field
Enter          Edit/apply field
Ctrl+U         Clear current edit buffer
Ctrl+S         Save
R              Save + Refresh Arcade Library
T              Cycle theme
Esc / S        Return
```

Configuration is saved atomically to the platform user-config location, normally:

```text
~/.config/marchine/config.toml
```

## ROM-path behavior

- Empty `rom_paths` means MAME rompath / Marchine auto-detection.
- TOML supports multiple explicit ROM paths.
- The Setup UI edits the first **primary** path and preserves additional configured paths.
- Clearing the primary field intentionally clears the explicit list and restores auto-detection.
- Explicit Marchine ROM paths are also passed to MAME when a game launches.

Saving Setup immediately rehydrates in-memory Arcade launch arguments, even if you do not refresh metadata. `R` requests a metadata refresh after saving; if a refresh is already running Marchine does not start a second one.
