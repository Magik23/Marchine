package ui

import (
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/gamepad"
)

var gamepadSuspended atomic.Bool

func SetGamepadSuspended(suspended bool) {
	gamepadSuspended.Store(suspended)
}

func GamepadSuspended() bool {
	return gamepadSuspended.Load()
}

type GamepadMsg struct {
	Action gamepad.Action
}

func (m Model) updateGamepad(action gamepad.Action) (tea.Model, tea.Cmd) {
	if action == gamepad.ActionQuit {
		return m, tea.Quit
	}

	if action == gamepad.ActionRotateDisplay {
		m.status = "ROTATING DISPLAY…"
		return m, rotateDisplayCmd()
	}

	// Zoom must remain available even when the terminal is too small to
	// render the launcher. Foot processes the injected key combination.
	if action == gamepad.ActionFontIncrease || action == gamepad.ActionFontDecrease {
		return m, footFontZoomCmd(action == gamepad.ActionFontIncrease)
	}

	if m.width > 0 && m.height > 0 && terminalTooSmall(m.width, m.height) {
		return m, nil
	}

	// The joystick's horizontal axis and LT/RT share the letter index.
	if action == gamepad.ActionLeft {
		action = gamepad.ActionPreviousLetter
	} else if action == gamepad.ActionRight {
		action = gamepad.ActionNextLetter
	}

	if action == gamepad.ActionPreviousLetter {
		if m.screen == screenLibrary && !m.searchMode {
			m.cycleJumpLetter(-1)
		}
		return m, nil
	}

	if action == gamepad.ActionNextLetter {
		if m.screen == screenLibrary && !m.searchMode {
			m.cycleJumpLetter(1)
		}
		return m, nil
	}

	if m.letterJump != "" {
		switch action {
		case gamepad.ActionConfirm:
			m.confirmJumpLetter()
			return m, nil

		case gamepad.ActionBack:
			m.cancelJumpLetter()
			return m, nil

		default:
			m.letterJump = ""
		}
	}

	// Controller library navigation wraps at the ends.
	if m.screen == screenLibrary && !m.searchMode {
		switch action {
		case gamepad.ActionUp:
			m.moveGamepadCursor(-1)
			return m, nil
		case gamepad.ActionDown:
			m.moveGamepadCursor(1)
			return m, nil
		}
	}

	key, ok := m.gamepadKey(action)
	if !ok {
		return m, nil
	}

	return m.Update(key)
}

func (m *Model) moveGamepadCursor(delta int) {
	n := len(m.filtered())

	if n == 0 {
		m.cursor = 0
		return
	}

	m.cursor = (m.cursor + delta) % n

	if m.cursor < 0 {
		m.cursor += n
	}
}

func (m Model) gamepadKey(action gamepad.Action) (tea.KeyPressMsg, bool) {
	if (m.screen == screenSetup && m.setupEditing) ||
		(m.screen == screenGameEditor && m.gameEditing) ||
		(m.screen == screenGameImportPath) {
		switch action {
		case gamepad.ActionConfirm:
			return specialKey(tea.KeyEnter), true
		case gamepad.ActionBack:
			return specialKey(tea.KeyEscape), true
		default:
			return tea.KeyPressMsg{}, false
		}
	}

	// In import preview, A toggles a row; Menu opens the final import
	// confirmation. View/other library controls remain unchanged.
	if m.screen == screenGameImportPreview && action == gamepad.ActionMenu {
		return textKey("i"), true
	}
	// Menu requests Remove All. Only keyboard Y accepts the final
	// confirmation; controller Confirm maps to Enter, which is ignored.
	if m.screen == screenGamePicker && m.gamePickerAction == setupFieldRemoveGame && action == gamepad.ActionMenu {
		return textKey("a"), true
	}

	switch action {
	case gamepad.ActionUp:
		return specialKey(tea.KeyUp), true

	case gamepad.ActionDown:
		return specialKey(tea.KeyDown), true

	case gamepad.ActionConfirm:
		return specialKey(tea.KeyEnter), true

	case gamepad.ActionBack:
		return specialKey(tea.KeyEscape), true

	case gamepad.ActionSearch:
		return tea.KeyPressMsg{}, false

	case gamepad.ActionCategories:
		if m.screen == screenLibrary || m.screen == screenCategories {
			return textKey("c"), true
		}

	case gamepad.ActionInfo:
		return textKey("i"), true

	case gamepad.ActionView:
		if m.screen == screenLibrary {
			return specialKey(tea.KeyTab), true
		}

	case gamepad.ActionMenu:
		return textKey("o"), true

	case gamepad.ActionHideNonArcade:
		return textKey("h"), true

	case gamepad.ActionTheme:
		return textKey("t"), true
	}

	return tea.KeyPressMsg{}, false
}

func (m *Model) cycleJumpLetter(delta int) {
	letters := m.availableJumpLetters()

	if len(letters) == 0 {
		m.letterJump = ""
		m.status = "NO LETTERS AVAILABLE"
		return
	}

	if m.letterJump == "" {
		m.letterJumpOrigin = m.cursor
	}

	current := m.letterJump

	if current == "" {
		if game, ok := m.selected(); ok {
			current = gameInitial(game.Name)
		}
	}

	index := -1

	for i, letter := range letters {
		if letter == current {
			index = i
			break
		}
	}

	if index < 0 {
		if delta > 0 {
			index = 0
		} else {
			index = len(letters) - 1
		}
	} else {
		index = (index + delta + len(letters)) % len(letters)
	}

	m.letterJump = letters[index]

	for i, game := range m.filtered() {
		if gameInitial(game.Name) == m.letterJump {
			m.cursor = i
			break
		}
	}

	m.status = "JUMP TO · " + m.letterJump
}

func (m *Model) confirmJumpLetter() {
	target := m.letterJump

	if target == "" {
		return
	}

	m.letterJump = ""
	m.status = "JUMPED TO · " + target
}

func (m *Model) cancelJumpLetter() {
	m.cursor = m.letterJumpOrigin
	m.normalizeCursor()
	m.letterJump = ""
	m.status = "READY"
}

func (m Model) availableJumpLetters() []string {
	var present [26]bool
	hasOther := false

	for _, game := range m.filtered() {
		initial := gameInitial(game.Name)

		if initial == "#" {
			hasOther = true
			continue
		}

		if len(initial) != 1 {
			continue
		}

		index := int(initial[0] - 'A')

		if index >= 0 && index < 26 {
			present[index] = true
		}
	}

	letters := make([]string, 0, 27)

	if hasOther {
		letters = append(letters, "#")
	}

	for i, exists := range present {
		if exists {
			letters = append(letters, string(rune('A'+i)))
		}
	}

	return letters
}

func gameInitial(name string) string {
	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	r, _ := utf8.DecodeRuneInString(name)
	upper := unicode.ToUpper(r)

	if upper >= 'A' && upper <= 'Z' {
		return string(upper)
	}

	return "#"
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func textKey(text string) tea.KeyPressMsg {
	runes := []rune(text)
	if len(runes) == 0 {
		return tea.KeyPressMsg{}
	}

	return tea.KeyPressMsg(tea.Key{
		Text: text,
		Code: runes[0],
	})
}
