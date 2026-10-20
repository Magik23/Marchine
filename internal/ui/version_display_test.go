package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestCompactVersion(t *testing.T) {
	cases := map[string]string{
		"1.25.0":                  "v1.25.0",
		"v1.25.0":                 "v1.25.0",
		"1.25.0-2-gabc1234":       "v1.25.0",
		"1.25.0-2-gabc1234-dirty": "v1.25.0",
		"1.25.0-dirty":            "v1.25.0",
		"dev":                     "dev",
	}

	for input, want := range cases {
		if got := compactVersion(input); got != want {
			t.Errorf("compactVersion(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestInfoSourceBarShowsVersion(t *testing.T) {
	m := Model{
		Version: "1.25.0-2-gabc1234-dirty",
		theme:   ThemeByName(""),
		screen:  screenInfo,
	}

	tabs := m.renderSourceTabs(122)

	if !strings.Contains(tabs, "v1.25.0") {
		t.Fatal("Info source bar does not show compact version")
	}
	if got := lipgloss.Width(tabs); got != 120 {
		t.Fatalf("Info version should leave two columns to the right: width = %d", got)
	}
	if strings.Contains(m.renderSourceTabs(25), "v1.25.0") {
		t.Fatal("Version should hide when space is insufficient")
	}
	if strings.Contains(m.renderInfo(124, 36), "MARCHINE v") {
		t.Fatal("Old Info heading version is still displayed")
	}

	m.screen = screenLibrary
	if strings.Contains(m.renderSourceTabs(122), "v1.25.0") {
		t.Fatal("Version should only appear on Info")
	}
}

// The running version is in Info's source bar, not duplicated in its content.
func TestInfoDoesNotDuplicateVersionOrDescribeUnsupportedEmulators(t *testing.T) {
	m := Model{Version: "1.27.0-2-gabcdef-dirty", theme: ThemeByName("")}
	info := m.renderInfo(124, 48)
	for _, unwanted := range []string{"Version:", "marchine --version", "EMULATOR MODEL", "(Default)"} {
		if strings.Contains(info, unwanted) {
			t.Fatalf("Info contains obsolete text %q", unwanted)
		}
	}
	for _, required := range []string{
		"ABOUT MARCHINE", "SOURCES", "ARCADE: MAME ROMs only.",
		"GAMES: Executables, scripts, emulator commands.",
		"PLAYLISTS: Mixed ARCADE and GAMES collections.",
		"Documentation: github.com/Magik23/Marchine",
		"System key: Super+Alt+R (Suggested)",
		"View+RB: Hypr Rotate (if installed).",
	} {
		if !strings.Contains(info, required) {
			t.Fatalf("Info missing %q", required)
		}
	}
}
