//go:build !linux

package main

import tea "charm.land/bubbletea/v2"

func startGamepads(_ *tea.Program) func() {
	return func() {}
}
