# Marchine Setup — V1.20

Open with:

```text
S
```

Fields:

```text
EMULATOR EXECUTABLE
ROM PATH
```

The current automatic Arcade driver is MAME. The neutral field label is intentional so future drivers can reuse the same launcher model without redesigning Setup.

Controls:

```text
↑/↓ or Tab     Select field
Enter          Edit/apply field
Ctrl+U         Clear current edit buffer
Ctrl+S         Save
R              Save + Refresh
Esc            Return
I              Info
```

Configuration is saved atomically to:

```text
~/.config/marchine/config.toml
```

An empty ROM path means: use MAME's configured `rompath` / auto-detection.
