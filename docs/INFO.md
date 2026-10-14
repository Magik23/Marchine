# Marchine Info + Credits

Marchine keeps operational help and project credits on separate terminal screens.

## Info — `I` / controller `X`

Info shows how the launcher works, keyboard/controller shortcuts, the MAME emulator model, and Quick Settings for **Hide Non-Arcade** and **Theme**. The running release number appears discreetly at the right of the source bar on Info; `marchine --version` displays the full build identifier.

### Main shortcuts

| Keyboard | Controller | Action |
| --- | --- | --- |
| `↑` / `↓`, `j` / `k` | D-Pad / left stick | Navigate |
| `[` / `]` | Left/Right or LT/RT | Jump by letter |
| `Enter` | A | Launch / Apply |
| `Esc` | B | Back / Cancel |
| `Tab` | View (tap) | Switch ARCADE / GAMES |
| `C` | Y | Arcade Categories |
| `I` | X | Info |
| `O` | Menu | Credits |
| `H` | LB | Hide Non-Arcade |
| `T` | RB | Change theme |
| `Q` | View + Menu | Quit |
| — | View + RB | Rotate display (optional Hypr Rotate utility) |
| Foot `Ctrl++` / `Ctrl+-` | View + X / View + A | Increase/decrease font size |
| Foot `Ctrl+0` | — | Reset font size |
| `/` | — | Search |
| `S` | — | Setup (ARCADE first, GAMES second) |
| `R` | — | Refresh Arcade index only |

For controller font resizing, **hold View first** and tap **X** (larger) or **A** (smaller). The chord suppresses the standalone X/A actions and does not switch sources. View + Menu and View + RB preserve their separate chord behavior. Font zoom requires the focused Marchine Foot desktop window on Hyprland, using `hyprctl`; ordinary terminal font shortcuts work independently. Font changes are not saved between launches.

**View + RB is optional:** it requires the separately installed [Hypr Rotate](https://github.com/Magik23/hypr-rotate) utility and rotates the display only while Marchine is running and receiving the controller input. Marchine neither bundles nor requires Hypr Rotate for normal launching.

When Marchine's terminal is smaller than **124×48 cells**, it shows a resize/reduce-font hint instead of the full interface. View + A remains usable to decrease font size on supported Hyprland/Foot desktops.

Inside Setup's text fields, paste with Foot `Ctrl+Shift+V`; `Ctrl+S` saves even while editing. Marchine ignores `Ctrl+C` to prevent accidental exits. `S` opens Setup from Info, Credits, and Categories as well as the library; `Tab` does not switch sources from these secondary screens. Shortcut key labels follow the active theme color, while descriptions and separators remain visually distinct. GAMES bulk import and removal use **keyboard `Y` only** for their final confirmations; `Enter` and controller A never count as Yes. See [Setup](SETUP.md) for the GAMES and Arcade workflows.

### Quick Settings

- **Hide Non-Arcade** is presentation-only; it doesn't delete cached or on-disk game files.
- **Theme** offers MARCHINE, PLASMA, PHOSPHOR, and EMBER.

These settings are persistent in Marchine's configuration, unlike the temporary window/font adjustments.

## Credits — `O` / controller `Menu`

Credits contains authorship, stack, license, local-first philosophy and Marchine's terminal artwork.

The embedded artwork is static terminal text: it is not resampled or animated.
