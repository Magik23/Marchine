//go:build linux

package gamepad

import "testing"

func TestViewFontChordsSuppressStandaloneActions(t *testing.T) {
	for _, tc := range []struct{ button, want Action }{
		{ActionInfo, ActionFontIncrease},
		{ActionConfirm, ActionFontDecrease},
	} {
		var state shortcutChordState
		state.handle(ActionView, true)
		for i := 0; i < 2; i++ {
			if got, emit, handled := state.handleFontButton(tc.button, true); !handled || !emit || got != tc.want {
				t.Fatalf("button %v press %d: got %v emit=%v handled=%v", tc.button, i, got, emit, handled)
			}
			if got, emit, handled := state.handleFontButton(tc.button, false); !handled || emit || got != ActionNone {
				t.Fatalf("button %v release leaked: %v %v %v", tc.button, got, emit, handled)
			}
		}
		if got, ok := state.handle(ActionView, false); ok || got != ActionNone {
			t.Fatalf("View should not switch source after zoom: %v %v", got, ok)
		}
		if !state.idle() {
			t.Fatal("chord state remained active")
		}
	}
}

func TestFaceButtonsWithoutViewStayImmediate(t *testing.T) {
	var state shortcutChordState
	for _, button := range []Action{ActionInfo, ActionConfirm} {
		if _, _, handled := state.handleFontButton(button, true); handled {
			t.Fatalf("ordinary face button %v was intercepted", button)
		}
	}
}

func TestExistingRotateAndQuitChordsStillWork(t *testing.T) {
	for _, tc := range []struct{ button, want Action }{
		{ActionTheme, ActionRotateDisplay},
		{ActionMenu, ActionQuit},
	} {
		var state shortcutChordState
		state.handle(ActionView, true)
		if got, ok := state.handle(tc.button, true); !ok || got != tc.want {
			t.Fatalf("existing chord: got %v %v, want %v", got, ok, tc.want)
		}
	}
}
