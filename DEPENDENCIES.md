# Marchine dependencies

Marchine keeps a deliberately small Go dependency graph.

## Direct Go modules

```text
charm.land/bubbletea/v2         v2.0.9
charm.land/lipgloss/v2          v2.0.6
github.com/pelletier/go-toml/v2 v2.4.3
```

Exact module checksums are committed in `go.sum`, and the complete dependency snapshot is committed under `vendor/`.

Normal development and packaging commands use:

```text
-mod=vendor
```

so they do not need to resolve the Go module graph over the network. Dependency network access is intentional only when running commands such as:

```bash
make deps
make vendor-refresh
```

## Runtime dependencies

Marchine's core terminal application has no required emulator dependency.

- **MAME** is optional and required only for automatic ARCADE indexing/launching.
- **Foot** is optional and used only by the supplied local desktop-launcher template.
- **Hyprland/Omarchy** is optional; the repository includes an example window rule but does not install compositor configuration.

## Rendering

The fightstick is an embedded ANSI text asset. Marchine does not require a graphics library, browser/web renderer, image decoder, OpenGL layer, Sixel renderer, or animation runtime.

## License material

The upstream release workflow and AUR package include Marchine's MIT license and the license/notice files found in the vendored dependency tree.
