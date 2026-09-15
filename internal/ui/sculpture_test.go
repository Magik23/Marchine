package ui

import (
	"strings"
	"testing"
)

func TestFightstickAssetHasStableGeometry(t *testing.T) {
	w, h := sculptureDimensions()
	if w <= 0 || h <= 0 {
		t.Fatalf("asset dimensions=%dx%d; expected non-empty ANSI art", w, h)
	}
	for i, line := range fightstickLines() {
		if got := visibleANSIWidth(line); got != w {
			t.Fatalf("asset line %d width=%d want=%d; source rows must stay fixed-width", i, got, w)
		}
	}
}

func TestFightstickPreservesEntireCurrentAsset(t *testing.T) {
	w, h := sculptureDimensions()
	if w != 62 || h != 22 {
		t.Fatalf("display sculpture dimensions=%dx%d want=62x22", w, h)
	}
	lines := fightstickLines()
	if len(lines) != 22 {
		t.Fatalf("display sculpture rows=%d want=22", len(lines))
	}
	first := ansiEscape.ReplaceAllString(lines[0], "")
	last := ansiEscape.ReplaceAllString(lines[len(lines)-1], "")
	if strings.Trim(first, "@") != "" || strings.Trim(last, "@") != "" {
		t.Fatalf("expected original @ background rows to be preserved at top and bottom")
	}
}

func TestFightstickNeverResamplesGlyphGeometry(t *testing.T) {
	artW, artH := sculptureDimensions()
	full := FightstickSculpture(artW+4, artH, ThemeByName("violet"))
	lines := strings.Split(full, "\n")
	if len(lines) != artH {
		t.Fatalf("full sculpture rows=%d want=%d", len(lines), artH)
	}
	// Two columns of breathing room on each side at art width + 4.
	if !strings.HasPrefix(ansiEscape.ReplaceAllString(lines[0], ""), "  ") {
		t.Fatalf("sculpture was not centered with expected padding")
	}
}

func TestFightstickUsesStableCropAndHidesWhenTooSmall(t *testing.T) {
	artW, _ := sculptureDimensions()
	if got := FightstickSculpture(artW+8, 16, ThemeByName("violet")); got == "" {
		t.Fatal("expected stable cropped sculpture at usable size")
	}
	if got := FightstickSculpture(artW-1, 24, ThemeByName("violet")); got != "" {
		t.Fatal("expected sculpture hidden when panel is narrower than the asset")
	}
	if got := FightstickSculpture(artW+8, minSculptureVisibleH-1, ThemeByName("violet")); got != "" {
		t.Fatal("expected sculpture hidden when panel is too short")
	}
}

func TestSculpturePanelWidthTracksAsset(t *testing.T) {
	artW, _ := sculptureDimensions()
	if got, want := sculpturePanelWidth(), max(32, artW+4); got != want {
		t.Fatalf("panel width=%d want=%d", got, want)
	}
}

func TestFightstickAddsNoVerticalBlackBands(t *testing.T) {
	artW, artH := sculptureDimensions()
	got := FightstickSculpture(artW+2, artH+6, ThemeByName("violet"))
	lines := strings.Split(got, "\n")
	if len(lines) != artH {
		t.Fatalf("sculpture height=%d want native art height=%d; no synthetic vertical padding", len(lines), artH)
	}
	plainFirst := ansiEscape.ReplaceAllString(lines[0], "")
	plainLast := ansiEscape.ReplaceAllString(lines[len(lines)-1], "")
	if strings.TrimSpace(plainFirst) == "" || strings.TrimSpace(plainLast) == "" {
		t.Fatalf("source edge rows should remain visible, not be replaced by blank bands")
	}
}
