package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

func TestSetupFromSecondaryScreens(t *testing.T) {
	for _, tc := range []struct {
		name   string
		screen screenMode
	}{
		{"info", screenInfo},
		{"credits", screenCredits},
		{"categories", screenCategories},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(config.Default(), "", false, true)
			m.width, m.height = 140, 52
			m.source = library.SourceArcade
			m.screen = tc.screen
			m.overlayHistory = []screenMode{screenLibrary}
			// Local Tab behavior must not change the source or screen.
			next, _ := m.Update(specialKey(tea.KeyTab))
			m = next.(Model)
			if m.screen != tc.screen || m.source != library.SourceArcade {
				t.Fatalf("Tab changed source or screen: %v %v", m.source, m.screen)
			}
			next, _ = m.Update(textKey("s"))
			m = next.(Model)
			if m.screen != screenSetup || len(m.overlayHistory) != 0 {
				t.Fatalf("S failed to open clean Setup: %v history=%v", m.screen, m.overlayHistory)
			}
			next, _ = m.Update(specialKey(tea.KeyEscape))
			m = next.(Model)
			if m.screen != screenLibrary || m.source != library.SourceArcade {
				t.Fatalf("Esc did not restore library: %v %v", m.source, m.screen)
			}
		})
	}
}

func TestSetupShortcutClearsNestedOverlayHistory(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 140, 52
	m.openInfo()
	m.openCredits()
	next, _ := m.Update(textKey("s"))
	m = next.(Model)
	if m.screen != screenSetup || len(m.overlayHistory) != 0 {
		t.Fatalf("stale overlay history after Setup: %v %v", m.screen, m.overlayHistory)
	}
}
