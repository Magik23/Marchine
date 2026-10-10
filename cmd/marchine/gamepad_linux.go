//go:build linux

package main

import (
	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/gamepad"
	"github.com/magik23/marchine/internal/ui"
)

// startGamepads connects the Linux controller manager to Bubble Tea.
// Controller input is optional: Marchine behaves normally with no pad present.
func startGamepads(p *tea.Program) func() {
	manager := gamepad.NewManager()

	go func() {
		for event := range manager.Events() {
			// Marchine must stay completely passive while a launched game owns
			// the session. Keep draining evdev events so gameplay input never
			// queues up and fires when Marchine resumes.
			if ui.GamepadSuspended() {
				continue
			}

			p.Send(ui.GamepadMsg{
				Action: event.Action,
			})
		}
	}()

	return manager.Close
}
