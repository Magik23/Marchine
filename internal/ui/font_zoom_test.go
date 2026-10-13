package ui

import (
	"strings"
	"testing"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/gamepad"
)

func TestFontZoomCmdAvailableWhenTooSmall(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 70, 20
	for _, action := range []gamepad.Action{gamepad.ActionFontIncrease, gamepad.ActionFontDecrease} {
		next, cmd := m.updateGamepad(action)
		if cmd == nil {
			t.Fatalf("font zoom blocked by small window: %v", action)
		}
		if next.(Model).screen != m.screen {
			t.Fatal("font zoom changed screen")
		}
	}
}

func TestFontZoomTargetsFocusedMarchine(t *testing.T) {
	for _, tc := range []struct {
		increase bool
		key      string
	}{
		{true, `key = "equal"`},
		{false, `key = "minus"`},
	} {
		lua := fontZoomLua(tc.increase)
		if !strings.Contains(lua, `w.class == "marchine"`) || !strings.Contains(lua, tc.key) || !strings.Contains(lua, `send_shortcut`) {
			t.Fatalf("font zoom Lua is not scoped to Marchine: %s", lua)
		}
	}
}

func TestTooSmallViewIncludesFontRecoveryHint(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 123, 47
	view := m.View().Content
	for _, want := range []string{"Terminal too small.", "124", "48", "Ctrl+-", "View+A"} {
		if !strings.Contains(view, want) {
			t.Fatalf("too-small view missing %q", want)
		}
	}
}

func TestInfoIncludesFontShortcuts(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	info := m.renderInfo(122, 36)
	for _, want := range []string{"Ctrl+ + / - / 0", "View + X / A"} {
		if !strings.Contains(info, want) {
			t.Errorf("Info missing %q", want)
		}
	}
}
