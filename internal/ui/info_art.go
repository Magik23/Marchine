package ui

import (
	_ "embed"
	"strings"
)

// INFO credit artwork supplied as an external ASCII asset. Keep the visible
// glyph geometry untouched; only generator-added trailing padding is removed
// so centering uses the real artwork width.
//
//go:embed info-credit-ascii.txt
var infoCreditASCII string

func creditsBBSArt() string {
	raw := strings.ReplaceAll(infoCreditASCII, "\r\n", "\n")
	lines := strings.Split(raw, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n")
}
