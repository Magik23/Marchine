# Marchine core scope

The filename is retained for project history; this document describes the current restrained core rather than a version-specific prototype.

## In scope

- Terminal-native TUI.
- GAMES + ARCADE sources.
- Flat generic game-launcher entries.
- Automatic MAME Arcade indexing from local data.
- Name/year/manufacturer from MAME XML.
- MAME display-orientation detection and `-autorol` for vertical games.
- CatVer-backed Arcade categories and Hide Non-Arcade presentation filter.
- Search.
- Direct process handoff and return.
- In-app emulator executable + primary ROM-path setup.
- Multiple explicit ROM paths in config without Setup destroying secondary paths.
- Cache-first startup, explicit refresh, current-schema cache validation, and saved-library fallback.
- Four terminal color themes.
- Fixed embedded ANSI fightstick sculpture with asset-fitted detail panel.
- Responsive shortcut footer within the supported layout.
- Hard minimum terminal gate at **124×48** cells.
- Separate Info and Credits screens.
- Offline-capable vendored Go builds.

## Out of scope

- Artwork scraping.
- Game ratings or reviews.
- Achievements.
- Provider/store walls.
- Cloud accounts.
- ROM repair, renaming, or deletion.
- RetroArch core management.
- BIOS management.
- General console encyclopedia metadata.
- Automatically discovering CHD-only/directory-only MAME entries without a parent ROM archive.
- Automatic broad console-system drivers in the current release line.

Future emulator drivers are possible when they preserve Marchine's simple relationship:

```text
EMULATOR + LOCAL SOURCE -> INDEX -> EXECUTE
```
