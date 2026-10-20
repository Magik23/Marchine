# Marchine Info + Credits

Marchine keeps operational help and project credits on separate terminal screens.

## Info — `I` / controller `X`

Info gives a compact **ABOUT MARCHINE** description followed by **SOURCES**: ARCADE (MAME ROMs only), GAMES (executables, scripts and emulator commands), and PLAYLISTS (mixed ARCADE/GAMES collections). It links to the [main README](../README.md) for detailed instructions instead of repeating the user guide inside the TUI. It also shows the optional Hypr Rotate integration, keyboard/controller shortcuts, and Quick Settings for **Hide Non-Arcade** and **Theme**.

The running release number appears at the right of the Info source bar, without a second version line in the main text. Use `marchine --version` from a shell for the full build identifier.

### Main shortcuts

| Keyboard | Controller | Action |
| --- | --- | --- |
| `↑` / `↓`, `j` / `k` | D-Pad / left stick | Navigate |
| `[` / `]` | Left/Right or LT/RT | Jump by letter |
| `Enter` | A | Launch / Apply |
| `Esc` | B | Back / Cancel |
| `Tab` | View (tap) | Cycle ARCADE → GAMES → PLAYLISTS |
| `P` | View + Y | Save highlighted ARCADE/GAMES title to an existing playlist |
| `C` | Y | Arcade Categories |
| `I` | X | Info |
| `O` | Menu | Credits |
| `H` | LB | Hide Non-Arcade |
| `T` | RB | Change theme |
| `Q` | View + Menu | Quit |
| Foot `Ctrl++` / `Ctrl+-` | View + X / View + A | Increase/decrease font size |
| Foot `Ctrl+0` | — | Reset font size |
| `/` | — | Search |
| `S` | — | Setup (ARCADE first, GAMES second) |
| `R` | — | Refresh Arcade index only |

The PLAYLISTS view exclusively creates, renames, opens, and deletes playlists; it can contain both ROMs and native games. Use `N` to create a playlist or `R` to rename one in its manager. From ARCADE or GAMES, `P` or View + Y selects an existing playlist destination. Inside a playlist, `D` requests removal of a *reference*; `D` in its manager requests deletion of the entire collection. Only **keyboard `Y`** confirms removal/deletion; `N`/`Esc` cancels. `Esc`/B returns from an opened playlist to its manager; at the top-level manager it does nothing (use Tab/View to switch sources). Removing a playlist or item never deletes ROMs, executables, or configured games. Naming requires keyboard text input. See the [PLAYLISTS user guide](../README.md#playlists).

For controller font resizing, **hold View first** and tap **X** (larger) or **A** (smaller). The chord suppresses the standalone X/A actions and does not switch sources. View + Menu and View + RB preserve their separate chord behavior. Font zoom requires the focused Marchine Foot desktop window on Hyprland, using `hyprctl`; ordinary terminal font shortcuts work independently. Font changes are not saved between launches.

### Optional display rotation — separate Hyprland utility

[Hypr Rotate](https://github.com/Magik23/hypr-rotate) is an independent, general-purpose Hyprland utility, not a Marchine or MAME component. It rotates the focused monitor at runtime without writing monitor configuration or changing boot orientation. Install it separately if desired.

A **suggested** keyboard shortcut, **Super+Alt+R**, can be configured in Hyprland and works system-wide, whether Marchine is open or not. This shortcut is a user-defined binding, not a built-in Marchine key. While Marchine is running and receiving controller input, **View + RB** optionally invokes `hypr-rotate` if the command is installed and on `PATH`. Marchine works normally without it.

When Marchine's terminal is smaller than **124×48 cells**, it shows a resize/reduce-font hint instead of the full interface. View + A remains usable to decrease font size on supported Hyprland/Foot desktops.

Inside Setup's text fields, paste with Foot `Ctrl+Shift+V`; `Ctrl+S` saves even while editing. Marchine ignores `Ctrl+C` to prevent accidental exits. `S` opens Setup from Info, Credits, and Categories as well as the library; `Tab` does not switch sources from these secondary screens. Shortcut key labels follow the active theme color, while descriptions and separators remain visually distinct. GAMES bulk import and removal use **keyboard `Y` only** for their final confirmations; `Enter` and controller A never count as Yes. See [Setup](SETUP.md) for the GAMES and Arcade workflows.

### Quick Settings

- **Hide Non-Arcade** is presentation-only; it doesn't delete cached or on-disk game files.
- **Theme** offers MARCHINE, PLASMA, PHOSPHOR, and EMBER.

These settings are persistent in Marchine's configuration, unlike the temporary window/font adjustments.

## Credits — `O` / controller `Menu`

Credits contains authorship, stack, license, local-first philosophy and Marchine's terminal artwork.

The embedded artwork is static terminal text: it is not resampled or animated.
