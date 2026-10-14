package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestShortcutHintsAcrossThemes(t *testing.T) {
	want := "Y YES  ·  N NO  ·  ESC CANCEL"
	for _, name := range []string{"violet", "cyan", "green", "amber"} {
		t.Run(name, func(t *testing.T) {
			m := Model{theme: ThemeByName(name), gameDeletePending: true}
			for title, got := range map[string]string{
				"remove footer": m.renderGamePickerFooter(122),
				"import footer": func() string { m.importConfirm = true; return m.renderGameImportFooter(122) }(),
				"hint":          m.renderShortcutHints(shortcutHint{"Y", "YES"}, shortcutHint{"N", "NO"}, shortcutHint{"ESC", "CANCEL"}),
			} {
				if plain := ansi.Strip(got); plain != want {
					t.Fatalf("%s: got %q, want %q", title, plain, want)
				}
				if lipgloss.Width(got) != lipgloss.Width(want) {
					t.Fatalf("%s: styled width changed", title)
				}
			}
		})
	}
}

func TestShortcutHintsPreserveEditorActions(t *testing.T) {
	m := Model{theme: ThemeByName("violet"), setupEditing: true}
	for _, word := range []string{"CTRL+S SAVE", "CTRL+SHIFT+V PASTE", "ESC CANCEL"} {
		if !strings.Contains(ansi.Strip(m.renderSetupFooter(122)), word) {
			t.Fatalf("Setup footer lost %s", word)
		}
	}
	m.gameEditing = true
	for _, word := range []string{"CTRL+S SAVE", "CTRL+SHIFT+V PASTE", "ESC CANCEL"} {
		if !strings.Contains(ansi.Strip(m.renderGameEditorFooter(122)), word) {
			t.Fatalf("GAMES editor footer lost %s", word)
		}
	}
}

func TestShortcutHintViewportAndReference(t *testing.T) {
	m := Model{theme: ThemeByName("green")}
	line := m.renderShortcutHintsFit(22,
		[]shortcutHint{{"CTRL+SHIFT+V", "PASTE"}, {"ENTER", "APPLY"}},
		[]shortcutHint{{"ESC", "BACK"}},
	)
	if lipgloss.Width(line) > 22 || ansi.Strip(line) != "ESC BACK" {
		t.Fatalf("bad compact hint: %q", ansi.Strip(line))
	}
	for _, k := range []string{"View + RB", "D-Pad / LS ↑↓", "A"} {
		got := m.renderShortcutReference(k, "Action")
		if !strings.HasSuffix(ansi.Strip(got), "Action") || lipgloss.Width(got) != 18+len("Action") {
			t.Fatalf("bad INFO alignment for %q: %q", k, ansi.Strip(got))
		}
	}
}
