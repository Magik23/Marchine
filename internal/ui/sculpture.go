package ui

import (
	_ "embed"
	"regexp"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// The Marchine fightstick is a fixed ANSI asset generated externally.
// Marchine never resamples, dithers, reflows or replaces its glyphs.
// The panel width is derived from the embedded asset so replacing
// fightstick.ans with a smaller/larger sculpture automatically changes the
// detail-panel width instead of leaving a large empty container around it.

const minSculptureVisibleH = 12

//go:embed fightstick.ans
var fightstickANSI string

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// FightstickSculpture returns the fixed fightstick artwork at 1:1 terminal-cell
// geometry, recolored to the active theme. No runtime scaling is performed.
// The source artwork is preserved from its first row to its last row: Marchine
// does not shave generated background rows. If the viewport is genuinely
// shorter than the asset, a stable center crop is used; horizontal geometry is
// never cropped or resampled.
//
// When the caller gives us more vertical room than the artwork needs, we do
// not manufacture blank rows above or below it. This keeps the ANSI image
// flush to its native geometry and avoids black bands in the detail panel.
func FightstickSculpture(width, height int, theme Theme) string {
	artW, artH := sculptureDimensions()
	if artW <= 0 || artH <= 0 || width < artW || height < minSculptureVisibleH {
		return ""
	}

	lines := fightstickPlainLines()
	visibleRows := min(height, len(lines))
	start := 0
	if visibleRows < len(lines) {
		// Stable center viewport: cropping only, never resampling.
		start = (len(lines) - visibleRows) / 2
	}
	lines = lines[start : start+visibleRows]

	// Center horizontally only. Vertical geometry is exactly the asset/crop
	// height, with no synthetic blank band above or below the ANSI image.
	left := max(0, (width-artW)/2)
	right := max(0, width-artW-left)
	leftPad := strings.Repeat(" ", left)
	rightPad := strings.Repeat(" ", right)

	out := make([]string, 0, visibleRows)
	colorize := lipgloss.NewStyle().Foreground(theme.Accent)
	for _, line := range lines {
		out = append(out, leftPad+colorize.Render(line)+rightPad)
	}
	return strings.Join(out, "\n")
}

// sculpturePanelWidth is the exact outer width Marchine reserves for the
// detail panel. The panel interior is width-2 (its two border cells), so +4
// gives the native ANSI asset one deliberate cell of breathing room on each
// side and no hidden four-cell right gutter.
func sculpturePanelWidth() int {
	artW, _ := sculptureDimensions()
	if artW <= 0 {
		return 48
	}
	return max(32, artW+4)
}

func sculptureDimensions() (width, height int) {
	lines := fightstickPlainLines()
	for _, line := range lines {
		width = max(width, visibleANSIWidth(line))
	}
	return width, len(lines)
}

func fightstickLines() []string {
	raw := strings.ReplaceAll(fightstickANSI, "\r\n", "\n")
	raw = strings.Trim(raw, "\n")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

func fightstickPlainLines() []string {
	rawLines := fightstickLines()
	out := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		out = append(out, ansiEscape.ReplaceAllString(line, ""))
	}
	return out
}

func visibleANSIWidth(s string) int {
	return utf8.RuneCountInString(ansiEscape.ReplaceAllString(s, ""))
}
