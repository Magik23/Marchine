# Sculpture

Marchine v1.18 treats `internal/ui/fightstick.ans` as a fixed terminal asset.

- Embedded source geometry: **44 columns × 24 rows**.
- Display geometry trims the first **6** generated dot-only rows, leaving **44 columns × 18 rows** before any height-limited crop.
- The source `.ans` file itself remains untouched.
- The renderer derives width from the embedded file; the dimensions are not hard-coded into the layout.
- The detail panel follows the asset width, keeping only a small deliberate margin around it.
- The `.ans` file is embedded into the Go binary at build time.
- Marchine does **not** resample, dither, reindex, scale, or replace characters.
- When height is limited, Marchine may use a stable centered vertical crop of the remaining rows.
- When the terminal becomes genuinely too narrow or too short, the sculpture is hidden instead of being deformed.

The asset remains the source of truth. Replacing `fightstick.ans` with another fixed-width ANSI asset automatically changes the panel width on the next build; if a future asset has different top breathing room, adjust `sculptureTopTrimRows` deliberately rather than resampling the art.
