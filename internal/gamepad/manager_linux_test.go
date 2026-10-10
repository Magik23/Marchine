//go:build linux

package gamepad

import "testing"

func TestQuitChordSuppressesIndividualActions(t *testing.T) {
	var state quitChordState

	if action, ok := state.handle(ActionView, true); ok || action != ActionNone {
		t.Fatalf("View press emitted action=%v ok=%v", action, ok)
	}

	if action, ok := state.handle(ActionSetup, true); !ok || action != ActionQuit {
		t.Fatalf("View + Menu emitted action=%v ok=%v, want Quit", action, ok)
	}

	if action, ok := state.handle(ActionView, false); ok || action != ActionNone {
		t.Fatalf("View release after chord emitted action=%v ok=%v", action, ok)
	}

	if action, ok := state.handle(ActionSetup, false); ok || action != ActionNone {
		t.Fatalf("Menu release after chord emitted action=%v ok=%v", action, ok)
	}
}

func TestQuitChordKeepsMenuAndReservesView(t *testing.T) {
	var view quitChordState

	if action, ok := view.handle(ActionView, true); ok || action != ActionNone {
		t.Fatalf("View press emitted action=%v ok=%v", action, ok)
	}
	if action, ok := view.handle(ActionView, false); ok || action != ActionNone {
		t.Fatalf("View alone should be reserved for the quit chord, action=%v ok=%v", action, ok)
	}

	var menu quitChordState

	if action, ok := menu.handle(ActionSetup, true); ok || action != ActionNone {
		t.Fatalf("Menu press emitted action=%v ok=%v", action, ok)
	}
	if action, ok := menu.handle(ActionSetup, false); !ok || action != ActionSetup {
		t.Fatalf("Menu release emitted action=%v ok=%v, want Setup", action, ok)
	}
}

func TestQuitChordWorksInEitherPressOrder(t *testing.T) {
	var state quitChordState

	if action, ok := state.handle(ActionSetup, true); ok || action != ActionNone {
		t.Fatalf("Menu press emitted action=%v ok=%v", action, ok)
	}

	if action, ok := state.handle(ActionView, true); !ok || action != ActionQuit {
		t.Fatalf("Menu + View emitted action=%v ok=%v, want Quit", action, ok)
	}
}

func TestQuitChordUsesViewNotXInfo(t *testing.T) {
	if quitChordAction(ActionInfo) {
		t.Fatal("X / Info must not participate in the View + Menu quit chord")
	}
	if !quitChordAction(ActionView) || !quitChordAction(ActionSetup) {
		t.Fatal("View and Menu must participate in the quit chord")
	}
}
