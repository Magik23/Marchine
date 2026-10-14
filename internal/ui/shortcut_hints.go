package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// shortcutHint controls presentation only. The input handlers remain untouched.
type shortcutHint struct {
	key    string
	action string
}

func (m Model) renderShortcutHint(h shortcutHint) string {
	key := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true)
	action := lipgloss.NewStyle().Foreground(m.theme.Text)
	return key.Render(h.key) + action.Render(" "+h.action)
}

func (m Model) renderShortcutHints(hints ...shortcutHint) string {
	items := make([]string, 0, len(hints))
	for _, hint := range hints {
		items = append(items, m.renderShortcutHint(hint))
	}
	sep := lipgloss.NewStyle().Foreground(m.theme.Muted).Render("  ·  ")
	return strings.Join(items, sep)
}

// Use the same existing width-aware fallback patterns as Marchine's footers.
func (m Model) renderShortcutHintsFit(width int, variants ...[]shortcutHint) string {
	if len(variants) == 0 {
		return ""
	}
	for _, hints := range variants {
		result := m.renderShortcutHints(hints...)
		if lipgloss.Width(result) <= width {
			return result
		}
	}
	return m.renderShortcutHints(variants[len(variants)-1]...)
}

// INFO uses a compact two-column layout. Only the key, not its padding, is accented.
func (m Model) renderShortcutReference(control, action string) string {
	key := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true)
	label := lipgloss.NewStyle().Foreground(m.theme.Text)
	gap := max(1, 18-lipgloss.Width(control))
	return key.Render(control) + strings.Repeat(" ", gap) + label.Render(action)
}
