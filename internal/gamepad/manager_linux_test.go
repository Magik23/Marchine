//go:build linux

package gamepad

import "testing"

func TestStandaloneChordButtonsFireOnRelease(t *testing.T) {
	for _, button := range []Action{ActionView, ActionMenu, ActionTheme} {
		var state shortcutChordState
		if action, ok := state.handle(button, true); ok || action != ActionNone {
			t.Fatalf("button %v press emitted action=%v ok=%v", button, action, ok)
		}
		if action, ok := state.handle(button, false); !ok || action != button {
			t.Fatalf("button %v release emitted action=%v ok=%v", button, action, ok)
		}
		if !state.idle() {
			t.Fatalf("button %v left state active", button)
		}
	}
}

func TestChordSuppressesStandaloneActionsInEitherOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		pair [2]Action
		want Action
	}{
		{"quit-view-first", [2]Action{ActionView, ActionMenu}, ActionQuit},
		{"quit-menu-first", [2]Action{ActionMenu, ActionView}, ActionQuit},
		{"rotate-view-first", [2]Action{ActionView, ActionTheme}, ActionRotateDisplay},
		{"rotate-rb-first", [2]Action{ActionTheme, ActionView}, ActionRotateDisplay},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var state shortcutChordState
			if a, ok := state.handle(tc.pair[0], true); ok || a != ActionNone {
				t.Fatalf("first press emitted %v %v", a, ok)
			}
			if a, ok := state.handle(tc.pair[1], true); !ok || a != tc.want {
				t.Fatalf("chord emitted %v %v, want %v", a, ok, tc.want)
			}
			for _, button := range tc.pair {
				if a, ok := state.handle(button, false); ok || a != ActionNone {
					t.Fatalf("chord release emitted %v %v", a, ok)
				}
			}
			if !state.idle() {
				t.Fatal("chord did not rearm")
			}
		})
	}
}

func TestChordsDoNotRetriggerUntilBothReleased(t *testing.T) {
	for _, tc := range []struct {
		name   string
		second Action
		want   Action
	}{
		{"quit", ActionMenu, ActionQuit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var state shortcutChordState
			state.handle(ActionView, true)
			if a, ok := state.handle(tc.second, true); !ok || a != tc.want {
				t.Fatalf("first chord emitted %v %v", a, ok)
			}
			state.handle(tc.second, false)
			if a, ok := state.handle(tc.second, true); ok || a != ActionNone {
				t.Fatalf("partial release retriggered %v %v", a, ok)
			}
			state.handle(ActionView, false)
			state.handle(tc.second, false)
			state.handle(tc.second, true)
			if a, ok := state.handle(ActionView, true); !ok || a != tc.want {
				t.Fatalf("rearmed chord emitted %v %v", a, ok)
			}
		})
	}
}

func TestUnrelatedButtonsDoNotEnterChordHandling(t *testing.T) {
	for _, action := range []Action{ActionInfo, ActionHideNonArcade, ActionConfirm, ActionBack} {
		if chordButton(action) {
			t.Fatalf("button %v should not be handled as a chord", action)
		}
	}
	for _, action := range []Action{ActionView, ActionMenu, ActionTheme} {
		if !chordButton(action) {
			t.Fatalf("button %v must be handled as a chord", action)
		}
	}
}

func TestHeldViewCanRotateMoreThanOnceWithoutThemeOrSource(t *testing.T) {
	var state shortcutChordState
	state.handle(ActionView, true)
	for i := 0; i < 3; i++ {
		if a, ok := state.handle(ActionTheme, true); !ok || a != ActionRotateDisplay {
			t.Fatalf("rotation %d: emitted %v %v", i, a, ok)
		}
		if a, ok := state.handle(ActionTheme, false); ok || a != ActionNone {
			t.Fatalf("rotation %d: emitted standalone Theme on release", i)
		}
	}
	if a, ok := state.handle(ActionView, false); ok || a != ActionNone {
		t.Fatalf("chord release emitted standalone View: %v %v", a, ok)
	}
	if !state.idle() {
		t.Fatal("state was not cleared")
	}
}

func TestViewYSavePlaylistChordKeepsCategoriesAlone(t *testing.T) {
	var s shortcutChordState
	if out, emit, handled := s.handleFontButton(ActionCategories, true); emit || handled || out != ActionNone {
		t.Fatalf("standalone Y must remain Categories: %v %v %v", out, emit, handled)
	}
	s.handle(ActionView, true)
	if out, emit, handled := s.handleFontButton(ActionCategories, true); !emit || !handled || out != ActionSaveToPlaylist {
		t.Fatalf("View + Y did not save: %v %v %v", out, emit, handled)
	}
	if out, emit, handled := s.handleFontButton(ActionCategories, false); emit || !handled || out != ActionNone {
		t.Fatalf("release leaked Y action: %v %v %v", out, emit, handled)
	}
	if out, emit := s.handle(ActionView, false); emit || out != ActionNone {
		t.Fatalf("View release changed source after save: %v %v", out, emit)
	}
}
