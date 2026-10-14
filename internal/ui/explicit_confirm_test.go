package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/gamepad"
)

// A second Enter (or controller A, which translates into Enter) must never
// commit a delete. Only an explicit keyboard Y may accept the warning.
func TestRemoveConfirmationNeedsKeyboardY(t *testing.T) {
	for _, tc := range []struct {
		name          string
		removeAll     bool
		wantRemaining int
	}{
		{"one", false, 1},
		{"all", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			cfg := config.Default()
			cfg.Games = []config.GameConfig{{Name: "First", Command: "/bin/true"}, {Name: "Second", Command: "/bin/true"}}
			if err := config.Save(path, cfg); err != nil {
				t.Fatal(err)
			}
			m := NewModel(cfg, path, false, false)
			m.width, m.height = minTerminalWidth, minTerminalHeight
			m.openGamePicker(setupFieldRemoveGame)
			var next tea.Model
			if tc.removeAll {
				next, _ = m.updateGamePicker(textKey("a"))
			} else {
				next, _ = m.updateGamePicker(specialKey(tea.KeyEnter))
			}
			m = next.(Model)
			if !m.gameDeletePending || m.gameDeleteAllPending != tc.removeAll {
				t.Fatal("expected remove confirmation")
			}
			next, _ = m.updateGamePicker(specialKey(tea.KeyEnter))
			m = next.(Model)
			if !m.gameDeletePending || len(m.cfg.Games) != 2 {
				t.Fatal("Enter incorrectly confirmed removal")
			}
			next, _ = m.updateGamepad(gamepad.ActionConfirm)
			m = next.(Model)
			if !m.gameDeletePending || len(m.cfg.Games) != 2 {
				t.Fatal("controller A incorrectly confirmed removal")
			}
			saved, err := config.Load(path)
			if err != nil || len(saved.Games) != 2 {
				t.Fatalf("game config changed before Y: %#v, %v", saved.Games, err)
			}
			rendered := ansi.Strip(m.renderGamePicker(minTerminalWidth, minTerminalHeight))
			if strings.Contains(rendered, "ENTER / Y YES") || strings.Contains(rendered, "ENTER YES") || !strings.Contains(rendered, "Y YES") {
				t.Fatal("remove confirmation legend still advertises Enter")
			}
			next, _ = m.updateGamePicker(textKey("y"))
			m = next.(Model)
			if m.gameDeletePending || len(m.cfg.Games) != tc.wantRemaining {
				t.Fatalf("Y failed: games=%d pending=%v", len(m.cfg.Games), m.gameDeletePending)
			}
		})
	}
}

// The importer must not commit when Enter or controller Confirm is pressed on
// its final prompt; the regular Enter preview/toggle behavior is unaffected.
func TestImportConfirmationNeedsKeyboardY(t *testing.T) {
	dir := t.TempDir()
	executable := filepath.Join(dir, "test-game.sh")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	m := NewModel(config.Default(), path, false, false)
	m.width, m.height = minTerminalWidth, minTerminalHeight
	m.screen = screenGameImportPreview
	m.importCandidates = []importCandidate{{game: config.GameConfig{Name: "Test Game", Command: executable}, path: executable, kind: "EXECUTABLE", selected: true}}
	m.importConfirm = true
	next, _ := m.updateGameImportPreview(specialKey(tea.KeyEnter))
	m = next.(Model)
	if !m.importConfirm || len(m.cfg.Games) != 0 {
		t.Fatal("Enter incorrectly confirmed import")
	}
	next, _ = m.updateGamepad(gamepad.ActionConfirm)
	m = next.(Model)
	if !m.importConfirm || len(m.cfg.Games) != 0 {
		t.Fatal("controller A incorrectly confirmed import")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("config written before Y: %v", err)
	}
	footer := ansi.Strip(m.renderGameImportFooter(minTerminalWidth))
	if strings.Contains(footer, "ENTER IMPORT") || !strings.Contains(footer, "Y YES") {
		t.Fatal("import legend still advertises Enter")
	}
	next, _ = m.updateGameImportPreview(textKey("n"))
	m = next.(Model)
	if m.importConfirm || len(m.cfg.Games) != 0 {
		t.Fatal("N should cancel import")
	}
	m.importConfirm = true
	next, _ = m.updateGameImportPreview(textKey("y"))
	m = next.(Model)
	if m.importConfirm || len(m.cfg.Games) != 1 {
		t.Fatalf("Y failed: games=%d pending=%v", len(m.cfg.Games), m.importConfirm)
	}
	saved, err := config.Load(path)
	if err != nil || len(saved.Games) != 1 {
		t.Fatalf("Y did not persist import: %v, %#v", err, saved.Games)
	}
}
