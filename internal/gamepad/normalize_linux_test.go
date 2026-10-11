//go:build linux

package gamepad

import (
	"testing"

	evdev "github.com/holoplot/go-evdev"
)

func TestButtonMapping(t *testing.T) {
	n := NewNormalizer(nil)

	tests := []struct {
		code   evdev.EvCode
		action Action
	}{
		{evdev.BTN_SOUTH, ActionConfirm},
		{evdev.BTN_EAST, ActionBack},
		{evdev.BTN_NORTH, ActionInfo},
		{evdev.BTN_WEST, ActionCategories},
		{evdev.BTN_SELECT, ActionView},
		{evdev.BTN_START, ActionMenu},
		{evdev.BTN_TL, ActionHideNonArcade},
		{evdev.BTN_TR, ActionTheme},
		{evdev.BTN_TL2, ActionPreviousLetter},
		{evdev.BTN_TR2, ActionNextLetter},
	}

	for _, test := range tests {
		action, ok := n.Handle(evdev.InputEvent{
			Type:  evdev.EV_KEY,
			Code:  test.code,
			Value: 1,
		})

		if !ok || action != test.action {
			t.Fatalf(
				"code %d: got action=%v ok=%v, want action=%v",
				test.code,
				action,
				ok,
				test.action,
			)
		}
	}
}

func TestButtonReleaseIgnored(t *testing.T) {
	n := NewNormalizer(nil)

	action, ok := n.Handle(evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  evdev.BTN_SOUTH,
		Value: 0,
	})

	if ok || action != ActionNone {
		t.Fatalf("release produced action=%v ok=%v", action, ok)
	}
}

func TestDPadRequiresNeutralBeforeSameDirectionAgain(t *testing.T) {
	n := NewNormalizer(nil)

	action, ok := n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_HAT0Y,
		Value: -1,
	})

	if !ok || action != ActionUp {
		t.Fatalf("first up: got action=%v ok=%v", action, ok)
	}

	action, ok = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_HAT0Y,
		Value: -1,
	})

	if ok || action != ActionNone {
		t.Fatalf("repeated up produced action=%v ok=%v", action, ok)
	}

	_, _ = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_HAT0Y,
		Value: 0,
	})

	action, ok = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_HAT0Y,
		Value: -1,
	})

	if !ok || action != ActionUp {
		t.Fatalf("re-armed up: got action=%v ok=%v", action, ok)
	}
}

func TestLeftStickDeadZoneAndDirections(t *testing.T) {
	n := NewNormalizer(map[evdev.EvCode]evdev.AbsInfo{
		evdev.ABS_X: {
			Minimum: -32768,
			Maximum: 32767,
		},
		evdev.ABS_Y: {
			Minimum: -32768,
			Maximum: 32767,
		},
	})

	action, ok := n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_Y,
		Value: 8000,
	})

	if ok || action != ActionNone {
		t.Fatalf("dead-zone movement produced action=%v ok=%v", action, ok)
	}

	action, ok = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_Y,
		Value: -25000,
	})

	if !ok || action != ActionUp {
		t.Fatalf("stick up: got action=%v ok=%v", action, ok)
	}

	_, _ = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_Y,
		Value: 0,
	})

	action, ok = n.Handle(evdev.InputEvent{
		Type:  evdev.EV_ABS,
		Code:  evdev.ABS_X,
		Value: 25000,
	})

	if !ok || action != ActionRight {
		t.Fatalf("stick right: got action=%v ok=%v", action, ok)
	}
}

func TestThumbstickClicksHaveNoAction(t *testing.T) {
	n := NewNormalizer(nil)
	for _, code := range []evdev.EvCode{evdev.BTN_THUMBL, evdev.BTN_THUMBR} {
		if action, ok := n.Handle(evdev.InputEvent{Type: evdev.EV_KEY, Code: code, Value: 1}); ok || action != ActionNone {
			t.Fatalf("button %v produced action=%v ok=%v", code, action, ok)
		}
	}
}
