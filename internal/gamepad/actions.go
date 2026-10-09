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

	ActionSearch

	ActionCategories
	ActionInfo
	ActionSetup

	ActionPreviousSource
	ActionNextSource

	ActionPreviousLetter
	ActionNextLetter

	ActionTheme
	ActionCredits
)
