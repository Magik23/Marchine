# Sculpture

Marchine treats `internal/ui/fightstick.ans` as a fixed terminal asset.

- Current embedded plain geometry: **62 columns × 22 rows**.
- The `.ans` file is embedded into the Go binary at build time.
- Source ANSI color escapes are stripped and the glyphs are recolored with the active Marchine theme.
- Marchine does **not** resample, dither, scale, morph, animate, or replace characters at runtime.
- The renderer derives dimensions from the embedded asset rather than hard-coding a separate panel geometry.
- The detail panel keeps a small deliberate horizontal margin around the native asset.
- When height is limited, Marchine can use a stable centered vertical crop.
- When width is insufficient, the sculpture is hidden rather than deformed.

The asset is the geometry source of truth. If `fightstick.ans` is deliberately replaced in the future, update the corresponding dimension tests and visually verify the detail-panel balance.
