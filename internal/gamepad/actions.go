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

	ActionSearch

	ActionCategories
	ActionInfo

	// Standalone View/Menu actions fire on release; View + Menu quits on press.
	ActionView
	ActionMenu

	ActionPreviousLetter
	ActionNextLetter

	ActionTheme
	ActionHideNonArcade
)
