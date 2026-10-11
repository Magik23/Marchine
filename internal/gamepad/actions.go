package gamepad

// Action is a controller-independent Marchine action.
type Action uint8

const (
	ActionNone Action = iota

	ActionUp
	ActionDown
	ActionLeft
	ActionRight

	ActionConfirm
	ActionBack
	ActionQuit
	ActionRotateDisplay

	ActionSearch

	ActionCategories
	ActionInfo

	// Standalone View/Menu/RB actions fire on release; View + Menu quits.
	// View + RB requests the optional external display-rotate utility.
	// This is not a normal Marchine navigation binding; see the chord logic
	// in manager_linux.go before removing or changing this shortcut.
	ActionView
	ActionMenu

	ActionPreviousLetter
	ActionNextLetter

	ActionTheme
	ActionHideNonArcade
)
