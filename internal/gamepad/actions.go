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

	// View is chord-only: View + Menu quits; X remains the Info button.
	ActionView
	ActionSetup

	ActionPreviousSource
	ActionNextSource

	ActionPreviousLetter
	ActionNextLetter

	ActionTheme
	ActionCredits
)
