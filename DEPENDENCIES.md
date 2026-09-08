# Marchine V1.6 dependencies

Marchine pins three direct Go modules:

```text
charm.land/bubbletea/v2        v2.0.9
charm.land/lipgloss/v2         v2.0.6
github.com/pelletier/go-toml/v2 v2.4.3
```

Exact checksums are committed in `go.sum`.

The normal build performs:

```text
go mod tidy
go mod vendor
go build -mod=vendor
```

So the first successful build creates a complete local dependency snapshot under `vendor/` and subsequent builds can use it without resolving the module graph again.

The sculpture renderer itself uses only the Go standard library. No graphics library, web renderer, image asset, OpenGL layer or external animation runtime is involved.
