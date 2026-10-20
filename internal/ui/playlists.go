package ui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

// PLAYLISTS is a third, user-curated source. Its library is unified: MAME ROMs
// and configured GAMES launch through their existing mechanisms. Playlist
// management never changes the underlying game/ROM files or launch commands.
func (m *Model) openPlaylistManager() {
	m.source = library.SourcePlaylists
	m.screen = screenPlaylists
	m.playlistIndex = -1
	m.playlistRemovePending = false
	m.playlistDeletePending = false
	m.searchMode = false
	m.search = ""
	m.cursor = 0
	m.normalizePlaylistCursor()
}

func (m *Model) normalizePlaylistCursor() {
	if len(m.cfg.Playlists) == 0 {
		m.playlistCursor = 0
	} else if m.playlistCursor >= len(m.cfg.Playlists) {
		m.playlistCursor = len(m.cfg.Playlists) - 1
	}
}

func (m *Model) openPlaylistSave() {
	if m.source == library.SourcePlaylists || m.screen != screenLibrary {
		return
	}
	game, ok := m.selected()
	if !ok {
		m.status = "NO GAME SELECTED"
		return
	}
	if len(m.cfg.Playlists) == 0 {
		m.status = "CREATE A PLAYLIST IN PLAYLISTS FIRST"
		return
	}
	m.playlistSaveGame = game
	m.playlistCursor = 0
	m.screen = screenPlaylistSave
	m.status = "SAVE TO PLAYLIST"
}

// Missing references remain visible and nonlaunchable, retaining the original
// display name so temporarily disconnected drives/ROMs don't erase curation.
func (m Model) playlistGames() []library.Game {
	if m.playlistIndex < 0 || m.playlistIndex >= len(m.cfg.Playlists) {
		return nil
	}
	byID := make(map[string]library.Game, len(m.arcade)+len(m.games))
	for _, game := range m.arcade {
		byID[game.ID] = game
	}
	for _, game := range m.games {
		byID[game.ID] = game
	}
	entries := m.cfg.Playlists[m.playlistIndex].Items
	out := make([]library.Game, 0, len(entries))
	for _, entry := range entries {
		if game, ok := byID[entry.ID]; ok {
			out = append(out, game)
		} else {
			source := library.SourceGames
			if strings.HasPrefix(entry.ID, "mame:") {
				source = library.SourceArcade
			}
			name := entry.Name
			if name == "" {
				name = entry.ID
			}
			out = append(out, library.Game{ID: entry.ID, Name: name + " [UNAVAILABLE]", Source: source})
		}
	}
	return out
}

func (m *Model) savePlaylists(next config.Config, success string) bool {
	if m.demo {
		m.status = "PLAYLIST EDIT DISABLED IN DEMO MODE"
		return false
	}
	if err := config.Save(m.configPath, next); err != nil {
		m.status = "PLAYLIST SAVE FAILED: " + err.Error()
		return false
	}
	m.cfg = next
	m.status = success
	return true
}

func (m Model) updatePlaylists(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.normalizePlaylistCursor()
	key := msg.String()
	if m.playlistDeletePending {
		switch key {
		case "y", "Y":
			idx := m.playlistCursor
			next := m.cfg
			next.Playlists = append([]config.PlaylistConfig(nil), m.cfg.Playlists[:idx]...)
			next.Playlists = append(next.Playlists, m.cfg.Playlists[idx+1:]...)
			if m.savePlaylists(next, "PLAYLIST DELETED") {
				m.normalizePlaylistCursor()
			}
			m.playlistDeletePending = false
		case "n", "N", "esc":
			m.playlistDeletePending = false
		}
		return m, nil
	}
	switch key {
	case "q":
		return m, tea.Quit
	case "tab":
		m.toggleSource()
	case "esc":
		// Already at the PLAYLISTS root: back must not switch sources.
	case "up", "k", "shift+tab":
		if len(m.cfg.Playlists) > 0 {
			m.playlistCursor = (m.playlistCursor - 1 + len(m.cfg.Playlists)) % len(m.cfg.Playlists)
		}
	case "down", "j":
		if len(m.cfg.Playlists) > 0 {
			m.playlistCursor = (m.playlistCursor + 1) % len(m.cfg.Playlists)
		}
	case "home":
		m.playlistCursor = 0
	case "end":
		if len(m.cfg.Playlists) > 0 {
			m.playlistCursor = len(m.cfg.Playlists) - 1
		}
	case "enter":
		if len(m.cfg.Playlists) == 0 {
			m.beginPlaylistName(-1)
		} else {
			m.playlistIndex = m.playlistCursor
			m.cursor, m.search = 0, ""
			m.screen = screenLibrary
			m.status = "PLAYLIST · " + m.cfg.Playlists[m.playlistIndex].Name
		}
	case "n", "N":
		m.beginPlaylistName(-1)
	case "r", "R":
		if len(m.cfg.Playlists) > 0 {
			m.beginPlaylistName(m.playlistCursor)
		}
	case "d", "D":
		if len(m.cfg.Playlists) > 0 {
			m.playlistDeletePending = true
			m.status = "CONFIRM PLAYLIST DELETION"
		}
	case "i", "I", "ctrl+k":
		m.openInfo()
	}
	return m, nil
}

func (m *Model) beginPlaylistName(idx int) {
	if m.demo {
		m.status = "PLAYLIST EDIT DISABLED IN DEMO MODE"
		return
	}
	m.playlistEditIndex = idx
	m.playlistName = ""
	if idx >= 0 {
		m.playlistName = m.cfg.Playlists[idx].Name
	}
	m.screen = screenPlaylistName
	m.status = "PLAYLIST NAME"
}

func (m Model) updatePlaylistName(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenPlaylists
		m.status = "READY"
	case "enter":
		name := strings.TrimSpace(m.playlistName)
		if name == "" {
			m.status = "PLAYLIST NAME IS REQUIRED"
			return m, nil
		}
		if utf8.RuneCountInString(name) > 48 {
			m.status = "PLAYLIST NAME TOO LONG (48 MAX)"
			return m, nil
		}
		for i, list := range m.cfg.Playlists {
			if i != m.playlistEditIndex && strings.EqualFold(name, strings.TrimSpace(list.Name)) {
				m.status = "PLAYLIST NAME ALREADY EXISTS"
				return m, nil
			}
		}
		next := m.cfg
		next.Playlists = append([]config.PlaylistConfig(nil), m.cfg.Playlists...)
		if m.playlistEditIndex < 0 {
			next.Playlists = append(next.Playlists, config.PlaylistConfig{Name: name})
		} else if m.playlistEditIndex < len(next.Playlists) {
			next.Playlists[m.playlistEditIndex].Name = name
		} else {
			m.status = "PLAYLIST NO LONGER EXISTS"
			return m, nil
		}
		if m.savePlaylists(next, "PLAYLIST SAVED · "+name) {
			if m.playlistEditIndex < 0 {
				m.playlistCursor = len(m.cfg.Playlists) - 1
			}
			m.screen = screenPlaylists
		}
	case "backspace":
		if m.playlistName != "" {
			_, size := utf8.DecodeLastRuneInString(m.playlistName)
			m.playlistName = m.playlistName[:len(m.playlistName)-size]
		}
	case "ctrl+u":
		m.playlistName = ""
	default:
		if msg.Text != "" {
			m.playlistName += msg.Text
		}
	}
	return m, nil
}

func (m Model) updatePlaylistSave(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.normalizePlaylistCursor()
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenLibrary
		m.status = "READY"
	case "up", "k", "shift+tab":
		m.playlistCursor = (m.playlistCursor - 1 + len(m.cfg.Playlists)) % len(m.cfg.Playlists)
	case "down", "j", "tab":
		m.playlistCursor = (m.playlistCursor + 1) % len(m.cfg.Playlists)
	case "home":
		m.playlistCursor = 0
	case "end":
		m.playlistCursor = len(m.cfg.Playlists) - 1
	case "enter":
		idx := m.playlistCursor
		game := m.playlistSaveGame
		for _, item := range m.cfg.Playlists[idx].Items {
			if item.ID == game.ID {
				m.status = "ALREADY IN · " + m.cfg.Playlists[idx].Name
				return m, nil
			}
		}
		next := m.cfg
		next.Playlists = append([]config.PlaylistConfig(nil), m.cfg.Playlists...)
		next.Playlists[idx].Items = append(append([]config.PlaylistItem(nil), next.Playlists[idx].Items...), config.PlaylistItem{ID: game.ID, Name: game.Name})
		if m.savePlaylists(next, "SAVED TO · "+m.cfg.Playlists[idx].Name) {
			m.screen = screenLibrary
		}
	}
	return m, nil
}

func (m *Model) removePlaylistSelected() {
	if m.playlistIndex < 0 || m.playlistIndex >= len(m.cfg.Playlists) {
		return
	}
	game, ok := m.selected()
	if !ok {
		return
	}
	next := m.cfg
	next.Playlists = append([]config.PlaylistConfig(nil), m.cfg.Playlists...)
	old := next.Playlists[m.playlistIndex].Items
	next.Playlists[m.playlistIndex].Items = make([]config.PlaylistItem, 0, len(old))
	for _, item := range old {
		if item.ID != game.ID {
			next.Playlists[m.playlistIndex].Items = append(next.Playlists[m.playlistIndex].Items, item)
		}
	}
	if m.savePlaylists(next, "REMOVED FROM PLAYLIST · "+game.Name) {
		m.normalizeCursor()
	}
}

// Playlist manager and its nested dialogs use the same full-width framed
// viewport as the other Marchine screens. The short selectable list remains
// intentionally compact; the surrounding border must not be compacted.
func (m Model) renderPlaylistPanel(width, height int, title string, rows []string, cursor int, notice string) string {
	t := m.theme
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(t.Muted)
	normal := lipgloss.NewStyle().Foreground(t.Text)
	selected := lipgloss.NewStyle().Foreground(t.SelectText).Background(t.Selection).Bold(true)
	warn := lipgloss.NewStyle().Foreground(t.Warning)

	// Border + two cells of padding on each side. The content never
	// consumes cells that belong to the enclosing frame.
	innerW := max(1, width-6)
	listW := min(innerW, 60)
	var b strings.Builder
	b.WriteString(accent.Render(trim(title, innerW)))
	b.WriteString("\n\n")
	if notice != "" {
		b.WriteString(warn.Render(trim(notice, innerW)))
		b.WriteString("\n\n")
	}
	if len(rows) == 0 && m.screen == screenPlaylists {
		b.WriteString(muted.Render(trim("NO PLAYLISTS YET · PRESS N TO CREATE ONE", innerW)))
	}
	visible := max(1, height-7)
	if notice != "" {
		visible = max(1, visible-2)
	}
	start := max(0, cursor-visible+1)
	for i := start; i < len(rows) && i < start+visible; i++ {
		row := pad(trim(rows[i], listW-2), listW-2)
		if i == cursor {
			b.WriteString(selected.Render("> " + row))
		} else {
			b.WriteString(normal.Render("  " + row))
		}
		b.WriteByte('\n')
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(t.Border).
		Width(max(20, width)).Height(max(10, height)).Padding(1, 2).Render(b.String())
}

func (m Model) renderPlaylists(width, height int) string {
	rows := make([]string, 0, len(m.cfg.Playlists))
	for _, p := range m.cfg.Playlists {
		label := "GAMES"
		if len(p.Items) == 1 {
			label = "GAME"
		}
		count := groupedCount(len(p.Items)) + " " + label
		const nameWidth = 44
		rows = append(rows, pad(trim(p.Name, nameWidth), nameWidth)+"  "+count)
	}
	notice := ""
	if m.playlistDeletePending {
		notice = "DELETE '" + m.cfg.Playlists[m.playlistCursor].Name + "'?  Y YES · N NO"
	}
	if strings.Contains(m.status, "FAILED") && notice == "" {
		notice = m.status
	}
	return m.renderPlaylistPanel(width, height, "PLAYLISTS", rows, m.playlistCursor, notice)
}

func (m Model) renderPlaylistName(width, height int) string {
	heading := "NEW PLAYLIST"
	if m.playlistEditIndex >= 0 {
		heading = "RENAME PLAYLIST"
	}
	content := m.playlistName + "▌"
	notice := ""
	if strings.Contains(m.status, "REQUIRED") || strings.Contains(m.status, "EXISTS") || strings.Contains(m.status, "LONG") || strings.Contains(m.status, "FAILED") {
		notice = m.status
	}
	return m.renderPlaylistPanel(width, height, heading, []string{content}, 0, notice)
}

func (m Model) renderPlaylistSave(width, height int) string {
	rows := make([]string, 0, len(m.cfg.Playlists))
	for _, p := range m.cfg.Playlists {
		marker := ""
		for _, item := range p.Items {
			if item.ID == m.playlistSaveGame.ID {
				marker = " [ADDED]"
				break
			}
		}
		rows = append(rows, p.Name+marker)
	}
	notice := ""
	if strings.Contains(m.status, "ALREADY") || strings.Contains(m.status, "FAILED") {
		notice = m.status
	}
	return m.renderPlaylistPanel(width, height, "SAVE TO PLAYLIST · "+trim(m.playlistSaveGame.Name, 32), rows, m.playlistCursor, notice)
}

func (m Model) renderPlaylistFooter(width int) string {
	switch m.screen {
	case screenPlaylists:
		if m.playlistDeletePending {
			return m.renderShortcutHints(shortcutHint{"Y", "DELETE"}, shortcutHint{"N/ESC", "CANCEL"})
		}
		return m.renderShortcutHintsFit(width,
			[]shortcutHint{{"↑↓", "SELECT"}, {"ENTER", "OPEN"}, {"N", "NEW"}, {"R", "RENAME"}, {"D", "DELETE"}, {"TAB", "SOURCE"}, {"I", "INFO"}, {"Q", "QUIT"}},
			[]shortcutHint{{"ENTER", "OPEN"}, {"N", "NEW"}, {"R", "RENAME"}, {"D", "DELETE"}, {"TAB", "SOURCE"}, {"I", "INFO"}, {"Q", "QUIT"}},
			[]shortcutHint{{"ENTER", "OPEN"}, {"N", "NEW"}, {"TAB", "SOURCE"}, {"I", "INFO"}, {"Q", "QUIT"}},
		)
	case screenPlaylistName:
		return m.renderShortcutHints(shortcutHint{"ENTER", "SAVE"}, shortcutHint{"ESC", "CANCEL"})
	case screenPlaylistSave:
		return m.renderShortcutHints(shortcutHint{"↑↓", "SELECT"}, shortcutHint{"ENTER/A", "ADD"}, shortcutHint{"ESC/B", "CANCEL"})
	}
	return ""
}
