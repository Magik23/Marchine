package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

const (
	gameFieldName = iota
	gameFieldCommand
	gameFieldArgs
	gameFieldWorkingDir
	gameFieldCount
)

type configuredGame struct {
	index int
	name  string
}

// Keep launch identities tied to their TOML indexes, not sorted row positions.
func (m Model) configuredGames() []configuredGame {
	rows := make([]configuredGame, 0, len(m.cfg.Games))
	for i, game := range m.cfg.Games {
		name := strings.TrimSpace(game.Name)
		if name == "" {
			name = fmt.Sprintf("UNNAMED GAME #%d", i+1)
		}
		rows = append(rows, configuredGame{index: i, name: name})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name)
	})
	return rows
}

// Reuse this both at startup and after editing config.toml. Invalid manual
// entries remain editable in Setup, but never become broken launch targets.
func (m *Model) reloadGames() {
	m.games = nil
	for i, game := range m.cfg.Games {
		if strings.TrimSpace(game.Name) == "" || strings.TrimSpace(game.Command) == "" {
			continue
		}
		m.games = append(m.games, library.Game{
			ID:         fmt.Sprintf("game:%d", i),
			Name:       game.Name,
			Source:     library.SourceGames,
			Command:    game.Command,
			Args:       append([]string(nil), game.Args...),
			WorkingDir: game.WorkingDir,
		})
	}
	sort.SliceStable(m.games, func(i, j int) bool {
		return strings.ToLower(m.games[i].Name) < strings.ToLower(m.games[j].Name)
	})
	m.normalizeCursor()
}

func (m *Model) openGamePicker(action int) {
	if len(m.cfg.Games) == 0 {
		m.status = "NO GAMES CONFIGURED · ADD A GAME FIRST"
		return
	}
	m.gamePickerAction = action
	m.gamePickerCursor = 0
	m.gameDeletePending = false
	m.gameDeleteAllPending = false
	m.screen = screenGamePicker
	m.status = ""
}

func (m *Model) openGameEditor(index int) {
	m.gameIndex = index
	m.gameDraft = config.GameConfig{}
	if index >= 0 {
		if index >= len(m.cfg.Games) {
			m.status = "GAME NO LONGER AVAILABLE"
			return
		}
		m.gameDraft = m.cfg.Games[index]
		m.gameDraft.Args = append([]string(nil), m.gameDraft.Args...)
	}
	m.gameField = gameFieldName
	m.gameEditing = false
	m.gameBuffer = ""
	m.screen = screenGameEditor
	m.status = "READY"
}

func (m Model) updateGamePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	rows := m.configuredGames()
	if len(rows) == 0 {
		m.screen = screenSetup
		m.status = "NO GAMES CONFIGURED"
		return m, nil
	}
	if m.gamePickerCursor >= len(rows) {
		m.gamePickerCursor = len(rows) - 1
	}
	key := msg.String()
	if m.gameDeletePending {
		switch key {
		case "y", "Y":
			if m.gameDeleteAllPending {
				m.removeAllConfiguredGames()
			} else {
				m.removeConfiguredGame(rows[m.gamePickerCursor].index)
			}
		case "n", "N", "esc":
			m.gameDeletePending = false
			m.gameDeleteAllPending = false
		}
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenSetup
		m.status = "READY"
	case "up", "k", "shift+tab":
		m.gamePickerCursor = (m.gamePickerCursor - 1 + len(rows)) % len(rows)
	case "down", "j", "tab":
		m.gamePickerCursor = (m.gamePickerCursor + 1) % len(rows)
	case "home":
		m.gamePickerCursor = 0
	case "end":
		m.gamePickerCursor = len(rows) - 1
	case "enter":
		if m.gamePickerAction == setupFieldEditGame {
			m.openGameEditor(rows[m.gamePickerCursor].index)
		} else {
			m.gameDeletePending = true
			m.gameDeleteAllPending = false
			m.status = ""
		}
	case "a", "A", "shift+a":
		if m.gamePickerAction == setupFieldRemoveGame {
			m.gameDeletePending = true
			m.gameDeleteAllPending = true
			m.status = ""
		}
	}
	return m, nil
}

func (m Model) gameFieldValue(field int) string {
	switch field {
	case gameFieldName:
		return m.gameDraft.Name
	case gameFieldCommand:
		return m.gameDraft.Command
	case gameFieldArgs:
		return formatGameArguments(m.gameDraft.Args)
	case gameFieldWorkingDir:
		return m.gameDraft.WorkingDir
	}
	return ""
}

func (m Model) updateGameEditor(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.gameEditing {
		switch key {
		case "esc":
			m.gameEditing = false
			m.gameBuffer = ""
		case "enter", "ctrl+s":
			value := strings.TrimSpace(m.gameBuffer)
			switch m.gameField {
			case gameFieldName:
				m.gameDraft.Name = value
			case gameFieldCommand:
				m.gameDraft.Command = value
			case gameFieldArgs:
				args, err := parseGameArguments(value)
				if err != nil {
					m.status = "GAME ARGUMENTS: " + err.Error()
					return m, nil
				}
				m.gameDraft.Args = args
			case gameFieldWorkingDir:
				m.gameDraft.WorkingDir = value
			}
			m.gameEditing = false
			m.gameBuffer = ""
			if key == "ctrl+s" {
				// Save the just-edited value too, without requiring Enter first.
				m.saveConfiguredGame()
			} else {
				m.gameField = (m.gameField + 1) % gameFieldCount
				m.status = "FIELD UPDATED · CTRL+S TO SAVE GAME"
			}
		case "backspace":
			if m.gameBuffer != "" {
				_, size := utf8.DecodeLastRuneInString(m.gameBuffer)
				m.gameBuffer = m.gameBuffer[:len(m.gameBuffer)-size]
			}
		case "ctrl+u":
			m.gameBuffer = ""
		default:
			if msg.Text != "" {
				m.gameBuffer += msg.Text
			}
		}
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		if m.gameIndex >= 0 {
			m.screen = screenGamePicker
		} else {
			m.screen = screenSetup
		}
		m.status = "GAME EDIT CANCELLED"
	case "up", "k", "shift+tab":
		m.gameField = (m.gameField - 1 + gameFieldCount) % gameFieldCount
	case "down", "j", "tab":
		m.gameField = (m.gameField + 1) % gameFieldCount
	case "enter":
		m.gameEditing = true
		m.gameBuffer = m.gameFieldValue(m.gameField)
	case "ctrl+s":
		m.saveConfiguredGame()
	}
	return m, nil
}

func (m *Model) saveConfiguredGame() {
	if m.demo {
		m.status = "GAME SAVE DISABLED IN DEMO MODE"
		return
	}
	if strings.TrimSpace(m.gameDraft.Name) == "" {
		m.status = "GAME NAME IS REQUIRED"
		return
	}
	if strings.TrimSpace(m.gameDraft.Command) == "" {
		m.status = "GAME EXECUTABLE IS REQUIRED"
		return
	}

	next := m.cfg
	next.Games = append([]config.GameConfig(nil), m.cfg.Games...)
	isAdd := m.gameIndex < 0
	if isAdd {
		next.Games = append(next.Games, m.gameDraft)
	} else {
		if m.gameIndex >= len(next.Games) {
			m.status = "GAME NO LONGER AVAILABLE"
			return
		}
		next.Games[m.gameIndex] = m.gameDraft
	}
	next.Normalize()
	if err := config.Save(m.configPath, next); err != nil {
		m.status = "GAME SAVE FAILED: " + err.Error()
		return
	}
	m.cfg = next
	m.reloadGames()
	m.screen = screenSetup
	if isAdd {
		m.setupField = setupFieldAddGame
		m.status = "GAME ADDED · " + m.gameDraft.Name
	} else {
		m.setupField = setupFieldEditGame
		m.status = "GAME UPDATED · " + m.gameDraft.Name
	}
}

func (m *Model) removeConfiguredGame(index int) {
	if m.demo {
		m.status = "GAME REMOVE DISABLED IN DEMO MODE"
		m.gameDeletePending = false
		m.gameDeleteAllPending = false
		return
	}
	if index < 0 || index >= len(m.cfg.Games) {
		m.status = "GAME NO LONGER AVAILABLE"
		return
	}
	name := m.cfg.Games[index].Name
	next := m.cfg
	next.Games = append([]config.GameConfig(nil), m.cfg.Games[:index]...)
	next.Games = append(next.Games, m.cfg.Games[index+1:]...)
	if err := config.Save(m.configPath, next); err != nil {
		m.status = "GAME REMOVE FAILED: " + err.Error()
		m.gameDeletePending = false
		m.gameDeleteAllPending = false
		return
	}
	m.cfg = next
	m.reloadGames()
	m.screen = screenSetup
	m.setupField = setupFieldRemoveGame
	m.gameDeletePending = false
	m.gameDeleteAllPending = false
	m.status = "GAME REMOVED · " + name
}

// Remove all configured GAMES entries in one atomic config save.
// ARCADE metadata, ROMs, executables and desktop shortcut files are untouched.
func (m *Model) removeAllConfiguredGames() {
	m.gameDeletePending = false
	m.gameDeleteAllPending = false
	if m.demo {
		m.status = "GAME REMOVE DISABLED IN DEMO MODE"
		return
	}
	count := len(m.cfg.Games)
	if count == 0 {
		m.status = "NO GAMES CONFIGURED"
		return
	}
	next := m.cfg
	next.Games = nil
	if err := config.Save(m.configPath, next); err != nil {
		m.status = "GAME REMOVE FAILED: " + err.Error()
		return
	}
	m.cfg = next
	m.reloadGames()
	m.screen = screenSetup
	m.setupField = setupFieldRemoveGame
	m.status = fmt.Sprintf("GAME LIBRARY CLEARED · %d ENTRIES REMOVED", count)
}

// Single-command-line arguments with quotes; do not invoke a shell.
// This keeps quoted paths intact (including when they contain spaces).
func parseGameArguments(raw string) ([]string, error) {
	var result []string
	var current strings.Builder
	var quote rune
	escaped := false
	started := false
	flush := func() {
		if started {
			result = append(result, current.String())
			current.Reset()
			started = false
		}
	}
	for _, r := range raw {
		if escaped {
			current.WriteRune(r)
			escaped = false
			started = true
			continue
		}
		if quote == '\'' {
			if r == '\'' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		if r == '\\' {
			escaped = true
			started = true
			continue
		}
		if quote == '"' {
			if r == '"' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			started = true
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		started = true
		current.WriteRune(r)
	}
	if quote != 0 || escaped {
		return nil, fmt.Errorf("unclosed quote or trailing backslash")
	}
	flush()
	return result, nil
}

func formatGameArguments(args []string) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "" || strings.ContainsAny(arg, " \t\n\r\"'\\") {
			parts = append(parts, strconv.Quote(arg))
		} else {
			parts = append(parts, arg)
		}
	}
	return strings.Join(parts, " ")
}

func (m Model) renderGamePicker(width, height int) string {
	t := m.theme
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(t.Muted)
	text := lipgloss.NewStyle().Foreground(t.Text)
	selected := lipgloss.NewStyle().Foreground(t.SelectText).Background(t.Selection).Bold(true)
	warn := lipgloss.NewStyle().Foreground(t.Warning)
	innerW := max(46, width-8)
	rows := m.configuredGames()
	title := "EDIT GAME"
	if m.gamePickerAction == setupFieldRemoveGame {
		title = "REMOVE GAME"
	}
	var b strings.Builder
	b.WriteString(accent.Render(title))
	b.WriteString(muted.Render(fmt.Sprintf("  // %d CONFIGURED", len(rows))))
	b.WriteString("\n\n")
	if m.gameDeletePending && len(rows) > 0 {
		if m.gameDeleteAllPending {
			b.WriteString(warn.Render(fmt.Sprintf("REMOVE ALL %d GAMES FROM MARCHINE?", len(rows))))
			b.WriteByte('\n')
			b.WriteString(muted.Render("Only Marchine entries will be removed; no files deleted."))
		} else {
			idx := min(m.gamePickerCursor, len(rows)-1)
			b.WriteString(warn.Render("REMOVE '" + trim(rows[idx].name, innerW-26) + "' FROM MARCHINE?"))
		}
		b.WriteString("\n")
		b.WriteString(m.renderShortcutHints(shortcutHint{"Y", "YES"}, shortcutHint{"N", "NO"}, shortcutHint{"ESC", "CANCEL"}))
		b.WriteString("\n\n")
	}
	// Reserve room for borders, title, and possible delete confirmation.
	rowsVisible := max(1, height-10)
	if m.gameDeletePending {
		rowsVisible = max(1, rowsVisible-4)
	}
	cursor := min(m.gamePickerCursor, max(0, len(rows)-1))
	start := max(0, cursor-rowsVisible+1)
	if start > 0 && cursor < start+rowsVisible/2 {
		start = max(0, cursor-rowsVisible/2)
	}
	end := min(len(rows), start+rowsVisible)
	for i := start; i < end; i++ {
		line := trim(rows[i].name, innerW-2)
		if i == cursor {
			b.WriteString(selected.Render("> " + pad(line, innerW-2)))
		} else {
			b.WriteString(text.Render("  " + pad(line, innerW-2)))
		}
		b.WriteByte('\n')
	}
	if strings.HasPrefix(m.status, "GAME REMOVE FAILED") {
		b.WriteString(warn.Render(trim(m.status, innerW)))
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(t.Border).
		Width(max(50, width)).Height(max(16, height)).Padding(1, 2).Render(b.String())
}

func (m Model) renderGameEditor(width, height int) string {
	t := m.theme
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(t.Muted)
	text := lipgloss.NewStyle().Foreground(t.Text)
	selected := lipgloss.NewStyle().Foreground(t.SelectText).Background(t.Selection).Bold(true)
	warn := lipgloss.NewStyle().Foreground(t.Warning)
	innerW := max(46, width-8)
	title := "ADD GAME"
	if m.gameIndex >= 0 {
		title = "EDIT GAME"
	}
	var b strings.Builder
	b.WriteString(accent.Render(title))
	b.WriteString("\n\n")
	fields := []struct{ label, hint string }{
		{"NAME", "Game display name"},
		{"EXECUTABLE", "Native binary / script / runner (wine, etc.)"},
		{"ARGUMENTS", "Optional; quote paths containing spaces"},
		{"WORKING DIRECTORY", "Optional game directory"},
	}
	for i, field := range fields {
		b.WriteString(muted.Render(field.label))
		b.WriteByte('\n')
		value := m.gameFieldValue(i)
		isHint := value == ""
		if isHint {
			value = field.hint
		}
		if m.gameEditing && m.gameField == i {
			value = m.gameBuffer + "▌"
			isHint = false
		}
		line := trim(value, innerW-2)
		if i == m.gameField {
			b.WriteString(selected.Render("> " + pad(line, innerW-2)))
		} else if isHint {
			b.WriteString(muted.Render("  " + pad(line, innerW-2)))
		} else {
			b.WriteString(text.Render("  " + pad(line, innerW-2)))
		}
		b.WriteString("\n\n")
	}
	b.WriteString(m.renderShortcutHints(shortcutHint{"Enter", "applies"}, shortcutHint{"Ctrl+S", "saves"}, shortcutHint{"Ctrl+Shift+V", "pastes"}, shortcutHint{"Esc", "cancels"}))
	if strings.HasPrefix(m.status, "GAME ") {
		b.WriteByte('\n')
		b.WriteString(warn.Render(trim(m.status, innerW)))
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(t.Border).
		Width(max(50, width)).Height(max(16, height)).Padding(1, 2).Render(b.String())
}

func (m Model) renderGamePickerFooter(width int) string {
	if m.gameDeletePending {
		return m.renderShortcutHints(shortcutHint{"Y", "YES"}, shortcutHint{"N", "NO"}, shortcutHint{"ESC", "CANCEL"})
	}
	if m.gamePickerAction == setupFieldRemoveGame {
		return m.renderShortcutHintsFit(width,
			[]shortcutHint{{"↑↓/TAB", "SELECT"}, {"ENTER", "REMOVE"}, {"A/MENU", "REMOVE ALL"}, {"ESC", "SETUP"}},
			[]shortcutHint{{"ENTER", "REMOVE"}, {"A/MENU", "REMOVE ALL"}, {"ESC", "SETUP"}},
		)
	}
	return m.renderShortcutHints(shortcutHint{"↑↓/TAB", "SELECT"}, shortcutHint{"ENTER", "EDIT"}, shortcutHint{"ESC", "SETUP"})
}

func (m Model) renderGameEditorFooter(width int) string {
	if m.gameEditing {
		return m.renderShortcutHintsFit(width,
			[]shortcutHint{{"ENTER", "APPLY"}, {"CTRL+S", "SAVE"}, {"CTRL+SHIFT+V", "PASTE"}, {"ESC", "CANCEL"}},
			[]shortcutHint{{"ENTER", "APPLY"}, {"ESC", "CANCEL"}},
		)
	}
	return m.renderShortcutHintsFit(width,
		[]shortcutHint{{"↑↓/TAB", "FIELD"}, {"ENTER", "EDIT"}, {"CTRL+S", "SAVE GAME"}, {"ESC", "CANCEL"}},
		[]shortcutHint{{"ENTER", "EDIT"}, {"CTRL+S", "SAVE GAME"}, {"ESC", "CANCEL"}},
	)
}
