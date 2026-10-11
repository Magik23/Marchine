//go:build linux

package gamepad

import "testing"

func TestStandaloneViewAndMenuFireOnRelease(t *testing.T) {
	for _, button := range []Action{ActionView, ActionMenu} {
		var state quitChordState
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

func TestQuitChordSuppressesStandaloneActionsInEitherOrder(t *testing.T) {
	for _, pair := range [][2]Action{{ActionView, ActionMenu}, {ActionMenu, ActionView}} {
		var state quitChordState
		if a, ok := state.handle(pair[0], true); ok || a != ActionNone {
			t.Fatalf("first press emitted %v %v", a, ok)
		}
		if a, ok := state.handle(pair[1], true); !ok || a != ActionQuit {
			t.Fatalf("chord emitted %v %v, want Quit", a, ok)
		}
		for _, button := range pair {
			if a, ok := state.handle(button, false); ok || a != ActionNone {
				t.Fatalf("chord release emitted %v %v", a, ok)
			}
		}
		if !state.idle() {
			t.Fatal("chord did not rearm")
		}
	}
}

func TestQuitChordDoesNotRetriggerUntilBothReleased(t *testing.T) {
	var state quitChordState
	state.handle(ActionView, true)
	if a, ok := state.handle(ActionMenu, true); !ok || a != ActionQuit {
		t.Fatalf("first chord emitted %v %v", a, ok)
	}
	state.handle(ActionMenu, false)
	if a, ok := state.handle(ActionMenu, true); ok || a != ActionNone {
		t.Fatalf("partial release retriggered %v %v", a, ok)
	}
	state.handle(ActionView, false)
	state.handle(ActionMenu, false)
	state.handle(ActionMenu, true)
	if a, ok := state.handle(ActionView, true); !ok || a != ActionQuit {
		t.Fatalf("rearmed chord emitted %v %v", a, ok)
	}
}

func TestQuitChordDoesNotUseInfo(t *testing.T) {
	if quitChordAction(ActionInfo) || !quitChordAction(ActionView) || !quitChordAction(ActionMenu) {
		t.Fatal("only View and Menu may form quit chord")
	}
}
