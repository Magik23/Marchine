//go:build linux

package gamepad

import evdev "github.com/holoplot/go-evdev"

const defaultStickDeadZone = 0.45

type Signal struct {
	Action  Action
	Pressed bool
}

type Normalizer struct {
	absInfo      map[evdev.EvCode]evdev.AbsInfo
	axisState    map[evdev.EvCode]int8
	triggerState map[evdev.EvCode]bool
}

func NewNormalizer(absInfo map[evdev.EvCode]evdev.AbsInfo) *Normalizer {
	return &Normalizer{
		absInfo:      absInfo,
		axisState:    make(map[evdev.EvCode]int8),
		triggerState: make(map[evdev.EvCode]bool),
	}
}

// Handle keeps the original press-only interface available.
func (n *Normalizer) Handle(event evdev.InputEvent) (Action, bool) {
	signal, ok := n.HandleSignal(event)
	if !ok || !signal.Pressed {
		return ActionNone, false
	}

	return signal.Action, true
}

func (n *Normalizer) HandleSignal(event evdev.InputEvent) (Signal, bool) {
	switch event.Type {
	case evdev.EV_KEY:
		return n.handleButton(event.Code, event.Value)

	case evdev.EV_ABS:
		switch event.Code {
		case evdev.ABS_HAT0X, evdev.ABS_HAT0Y, evdev.ABS_X, evdev.ABS_Y:
			return n.handleAxis(event.Code, event.Value)

		case evdev.ABS_Z, evdev.ABS_RZ:
			return n.handleTrigger(event.Code, event.Value)
		}
	}

	return Signal{}, false
}

func (n *Normalizer) handleButton(code evdev.EvCode, value int32) (Signal, bool) {
	if value == 2 {
		return Signal{}, false
	}

	if value < 0 || value > 1 {
		return Signal{}, false
	}

	var action Action

	switch code {
	case evdev.BTN_SOUTH:
		action = ActionConfirm

	case evdev.BTN_EAST:
		action = ActionBack

	// Physical Xbox X opens Info directly; View is tracked separately for the quit chord.
	case evdev.BTN_NORTH:
		action = ActionInfo

	case evdev.BTN_WEST:
		action = ActionCategories

	case evdev.BTN_SELECT:
		action = ActionView

	case evdev.BTN_START:
		action = ActionSetup

	case evdev.BTN_TL:
		action = ActionPreviousSource

	case evdev.BTN_TR:
		action = ActionNextSource

	case evdev.BTN_THUMBL:
		action = ActionTheme

	case evdev.BTN_THUMBR:
		action = ActionCredits

	default:
		return Signal{}, false
	}

	return Signal{
		Action:  action,
		Pressed: value == 1,
	}, true
}

func (n *Normalizer) handleAxis(code evdev.EvCode, value int32) (Signal, bool) {
	var position int8

	switch code {
	case evdev.ABS_HAT0X, evdev.ABS_HAT0Y:
		position = sign(value)

	case evdev.ABS_X, evdev.ABS_Y:
		position = n.analogPosition(code, value)
	}

	previous := n.axisState[code]

	if position == previous {
		return Signal{}, false
	}

	n.axisState[code] = position

	if position == 0 {
		action := axisAction(code, previous)

		if action == ActionNone {
			return Signal{}, false
		}

		return Signal{
			Action:  action,
			Pressed: false,
		}, true
	}

	action := axisAction(code, position)

	if action == ActionNone {
		return Signal{}, false
	}

	return Signal{
		Action:  action,
		Pressed: true,
	}, true
}

func (n *Normalizer) handleTrigger(code evdev.EvCode, value int32) (Signal, bool) {
	pressed := n.triggerPressed(code, value)
	previous := n.triggerState[code]

	if pressed == previous {
		return Signal{}, false
	}

	n.triggerState[code] = pressed

	action := ActionPreviousLetter

	if code == evdev.ABS_RZ {
		action = ActionNextLetter
	}

	return Signal{
		Action:  action,
		Pressed: pressed,
	}, true
}

func axisAction(code evdev.EvCode, position int8) Action {
	switch code {
	case evdev.ABS_HAT0X, evdev.ABS_X:
		if position < 0 {
			return ActionLeft
		}
		if position > 0 {
			return ActionRight
		}

	case evdev.ABS_HAT0Y, evdev.ABS_Y:
		if position < 0 {
			return ActionUp
		}
		if position > 0 {
			return ActionDown
		}
	}

	return ActionNone
}

func (n *Normalizer) triggerPressed(code evdev.EvCode, value int32) bool {
	info, ok := n.absInfo[code]

	if !ok || info.Maximum <= info.Minimum {
		return value > 127
	}

	threshold := float64(info.Minimum) +
		float64(info.Maximum-info.Minimum)*0.50

	return float64(value) >= threshold
}

func (n *Normalizer) analogPosition(code evdev.EvCode, value int32) int8 {
	info, ok := n.absInfo[code]

	if !ok || info.Maximum <= info.Minimum {
		if value <= -16384 {
			return -1
		}
		if value >= 16384 {
			return 1
		}
		return 0
	}

	minimum := float64(info.Minimum)
	maximum := float64(info.Maximum)
	center := (minimum + maximum) / 2
	halfRange := (maximum - minimum) / 2

	if halfRange <= 0 {
		return 0
	}

	normalized := (float64(value) - center) / halfRange

	if normalized <= -defaultStickDeadZone {
		return -1
	}

	if normalized >= defaultStickDeadZone {
		return 1
	}

	return 0
}

func sign(value int32) int8 {
	if value < 0 {
		return -1
	}

	if value > 0 {
		return 1
	}

	return 0
}
