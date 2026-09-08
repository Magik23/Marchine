# Marchine V1.20 scope

## In scope

- Terminal-native TUI.
- GAMES + ARCADE sources.
- Flat generic game launcher entries.
- Automatic MAME Arcade indexing.
- Name/year/manufacturer from local MAME data.
- CatVer-backed Arcade categories + Hide Non-Arcade presentation filter.
- Search.
- Direct process handoff and return.
- In-app emulator executable + ROM path setup.
- Refresh Arcade library.
- Four terminal color themes.
- Fixed embedded ANSI fightstick sculpture with asset-fitted responsive panel width.
- Responsive shortcut footer within the supported layout.
- CLIAMP-style minimum terminal gate at 100×48 cells.
- Separate Info and Credits screens with single-key access.

## Out of scope

- Artwork scraping.
- Game ratings or reviews.
- Achievements.
- Provider/store walls.
- Cloud accounts.
- ROM repair/renaming/deletion.
- RetroArch core management.
- BIOS management.
- General console encyclopedia metadata.
- Automatic NES/SNES/Genesis drivers in V1.13.

Future emulator drivers are possible, but only when they can preserve Marchine's simple relationship:

```text
EMULATOR + ROM PATH -> INDEX -> EXECUTE
```
