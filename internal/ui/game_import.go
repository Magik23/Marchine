package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/magik23/marchine/internal/config"
)

// An import candidate is only a reference to a user's existing launcher.
// Scanning and previewing never execute, copy, rename or modify game files.
type importCandidate struct {
	game      config.GameConfig
	path      string
	kind      string
	selected  bool
	duplicate bool
}

func (m *Model) openGameImport() {
	m.screen = screenGameImportPath
	m.importPattern = ""
	m.importCandidates = nil
	m.importCursor = 0
	m.importConfirm = false
	m.importSkipped = 0
	m.status = "IMPORT GAMES · ENTER PATH OR PATTERN"
}

func (m Model) updateGameImportPath(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenSetup
		m.setupField = setupFieldImportGame
	case "q":
		// q is valid in a path; only Esc exits this input.
		m.importPattern += msg.Text
	case "enter":
		candidates, skipped, err := discoverImportCandidates(m.importPattern, m.cfg.Games)
		if err != nil {
			m.status = "IMPORT ERROR: " + err.Error()
			return m, nil
		}
		if len(candidates) == 0 {
			m.status = fmt.Sprintf("IMPORT: NO SUPPORTED FILES · %d SKIPPED", skipped)
			return m, nil
		}
		m.importCandidates = candidates
		m.importSkipped = skipped
		m.importCursor = 0
		m.importConfirm = false
		m.screen = screenGameImportPreview
		m.status = fmt.Sprintf("IMPORT PREVIEW · %d MATCHES", len(candidates))
	case "backspace":
		if m.importPattern != "" {
			_, size := utf8.DecodeLastRuneInString(m.importPattern)
			m.importPattern = m.importPattern[:len(m.importPattern)-size]
		}
	case "ctrl+u":
		m.importPattern = ""
	default:
		if msg.Text != "" {
			m.importPattern += msg.Text
		}
	}
	return m, nil
}

func (m Model) updateGameImportPreview(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.importConfirm {
		switch msg.String() {
		case "y", "Y":
			m.commitGameImport()
		case "n", "N", "esc":
			m.importConfirm = false
		}
		return m, nil
	}

	n := len(m.importCandidates)
	if n == 0 {
		m.screen = screenGameImportPath
		return m, nil
	}
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenGameImportPath
		m.status = "IMPORT PATH · ENTER TO PREVIEW"
	case "up", "k", "shift+tab":
		m.importCursor = (m.importCursor - 1 + n) % n
	case "down", "j", "tab":
		m.importCursor = (m.importCursor + 1) % n
	case "home":
		m.importCursor = 0
	case "end":
		m.importCursor = n - 1
	case "pgup":
		m.importCursor = max(0, m.importCursor-max(1, m.importRowsVisible()))
	case "pgdown":
		m.importCursor = min(n-1, m.importCursor+max(1, m.importRowsVisible()))
	case "space", " ", "enter":
		if !m.importCandidates[m.importCursor].duplicate {
			m.importCandidates[m.importCursor].selected = !m.importCandidates[m.importCursor].selected
		}
	case "a":
		for i := range m.importCandidates {
			if !m.importCandidates[i].duplicate {
				m.importCandidates[i].selected = true
			}
		}
	case "n":
		for i := range m.importCandidates {
			m.importCandidates[i].selected = false
		}
	case "i", "ctrl+s":
		if m.importSelectedCount() == 0 {
			m.status = "IMPORT: NO GAMES SELECTED"
		} else if m.demo {
			m.status = "IMPORT DISABLED IN DEMO MODE"
		} else {
			m.importConfirm = true
		}
	}
	return m, nil
}

func (m Model) importSelectedCount() int {
	n := 0
	for _, c := range m.importCandidates {
		if c.selected && !c.duplicate {
			n++
		}
	}
	return n
}

func (m *Model) commitGameImport() {
	if m.demo {
		m.importConfirm = false
		m.status = "IMPORT DISABLED IN DEMO MODE"
		return
	}
	next := m.cfg
	next.Games = append([]config.GameConfig(nil), m.cfg.Games...)
	seen := make(map[string]bool, len(next.Games)+len(m.importCandidates))
	for _, g := range next.Games {
		seen[gameImportKey(g)] = true
	}
	added, skipped := 0, 0
	for _, c := range m.importCandidates {
		if !c.selected || c.duplicate {
			continue
		}
		// A shortcut can disappear after preview (or be replaced by a directory).
		st, err := os.Stat(c.path)
		if err != nil || !st.Mode().IsRegular() || (c.kind == "EXECUTABLE" && st.Mode().Perm()&0o111 == 0) {
			skipped++
			continue
		}
		key := gameImportKey(c.game)
		if seen[key] {
			skipped++
			continue
		}
		next.Games = append(next.Games, c.game)
		seen[key] = true
		added++
	}
	if added == 0 {
		m.importConfirm = false
		m.status = "IMPORT: NO NEW GAMES TO SAVE"
		return
	}
	if err := config.Save(m.configPath, next); err != nil {
		m.importConfirm = false
		m.status = "IMPORT SAVE FAILED: " + err.Error()
		return
	}
	m.cfg = next
	m.reloadGames()
	m.importConfirm = false
	m.screen = screenSetup
	m.setupField = setupFieldImportGame
	m.status = fmt.Sprintf("IMPORT COMPLETE · %d ADDED · %d SKIPPED", added, skipped+m.importSkipped)
}

func (m Model) importRowsVisible() int {
	// Calculate the same body height as View() so PgUp/PgDown advance
	// by exactly one visible playlist rather than an approximate page.
	w, h := m.width, m.height
	if w <= 0 {
		w = minTerminalWidth
	}
	if h <= 0 {
		h = minTerminalHeight
	}
	cw, ch := w-2, h-2
	bodyH := max(12, ch-lipgloss.Height(m.renderHeader(cw, ch))-
		lipgloss.Height(m.renderSourceTabs(cw))-lipgloss.Height(m.renderGameImportFooter(cw))-3)
	return max(1, bodyH-14)
}

func (m Model) renderGameImport(width, height int) string {
	t := m.theme
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(t.Muted)
	text := lipgloss.NewStyle().Foreground(t.Text)
	selected := lipgloss.NewStyle().Foreground(t.SelectText).Background(t.Selection).Bold(true)
	warn := lipgloss.NewStyle().Foreground(t.Warning)
	innerW := max(46, width-8)
	var b strings.Builder
	b.WriteString(accent.Render("IMPORT GAMES"))
	b.WriteString(muted.Render("  // PATH / PATTERN · NO FILE BROWSER"))
	b.WriteString("\n\n")
	if m.screen == screenGameImportPath {
		b.WriteString(muted.Render("PATH OR PATTERN"))
		b.WriteByte('\n')
		line := m.importPattern + "▌"
		// Keep the caret visible while entering long paths; the prefix is
		// less important than the editable tail in a single-line field.
		if runes := []rune(line); len(runes) > innerW-3 {
			line = "…" + string(runes[len(runes)-(innerW-4):])
		}
		b.WriteString(selected.Render("> " + pad(trim(line, innerW-2), innerW-2)))
		b.WriteString("\n\n")
		b.WriteString(text.Render("~/Desktop/*.desktop"))
		b.WriteByte('\n')
		b.WriteString(muted.Render("~/Games/*.AppImage  ·  ~/Games/*/start.sh"))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("Exact path, immediate directory, or simple glob. No recursive scan."))
		b.WriteByte('\n')
		b.WriteString(m.renderShortcutHints(shortcutHint{"Enter", "previews matches"}, shortcutHint{"Esc", "returns to Setup"}))
		if strings.HasPrefix(m.status, "IMPORT ERROR") || strings.HasPrefix(m.status, "IMPORT: ") {
			b.WriteString("\n\n")
			b.WriteString(warn.Render(trim(m.status, innerW)))
		}
	} else {
		total := len(m.importCandidates)
		b.WriteString(muted.Render("MATCHED  " + trim(m.importPattern, innerW-28)))
		b.WriteString("\n")
		b.WriteString(accent.Render(fmt.Sprintf("%d SELECTED  ·  %d/%d  ·  %d SKIPPED", m.importSelectedCount(), m.importCursor+1, total, m.importSkipped)))
		b.WriteString("\n\n")
		rows := min(m.importRowsVisible(), max(1, height-14))
		// Track the highlighted row around 3/5 of the visible playlist,
		// matching the main ARCADE browsing experience.
		anchor := max(1, rows*3/5)
		start := min(max(0, m.importCursor-anchor), max(0, total-rows))
		end := min(total, start+rows)
		for i := start; i < end; i++ {
			c := m.importCandidates[i]
			mark := "[ ]"
			if c.duplicate {
				mark = "[-]"
			} else if c.selected {
				mark = "[x]"
			}
			kind := c.kind
			if c.duplicate {
				kind = "ALREADY ADDED"
			}
			nameW := max(12, innerW-20)
			line := mark + " " + pad(trim(c.game.Name, nameW-4), nameW-4) + " " + kind
			line = pad(trim(line, innerW-2), innerW-2)
			if i == m.importCursor {
				b.WriteString(selected.Render("> " + line))
			} else {
				b.WriteString(text.Render("  " + line))
			}
			b.WriteByte('\n')
		}
		b.WriteString("\n")
		if total > 0 {
			b.WriteString(muted.Render("PATH  " + trim(m.importCandidates[m.importCursor].path, innerW-7)))
			b.WriteByte('\n')
		}
		if m.importConfirm {
			b.WriteString(warn.Render(fmt.Sprintf("IMPORT %d SELECTED GAMES INTO MARCHINE?", m.importSelectedCount())))
			b.WriteString("\n")
			b.WriteString(m.renderShortcutHints(shortcutHint{"Y", "YES"}, shortcutHint{"N", "NO"}, shortcutHint{"ESC", "CANCEL"}))
		} else {
			b.WriteString(m.renderShortcutHints(shortcutHint{"SPACE/ENTER", "TOGGLE"}, shortcutHint{"A", "ALL"}, shortcutHint{"N", "NONE"}, shortcutHint{"I", "IMPORT"}))
		}
		if strings.HasPrefix(m.status, "IMPORT SAVE FAILED") || m.status == "IMPORT: NO GAMES SELECTED" || m.status == "IMPORT DISABLED IN DEMO MODE" || m.status == "IMPORT: NO NEW GAMES TO SAVE" {
			b.WriteByte('\n')
			b.WriteString(warn.Render(trim(m.status, innerW)))
		}
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(t.Border).
		Width(max(50, width)).Height(max(16, height)).Padding(1, 2).Render(b.String())
}

func (m Model) renderGameImportFooter(width int) string {
	if m.screen == screenGameImportPath {
		return m.renderShortcutHintsFit(width,
			[]shortcutHint{{"TYPE/PASTE", "PATH"}, {"ENTER", "PREVIEW"}, {"CTRL+U", "CLEAR"}, {"ESC", "SETUP"}},
			[]shortcutHint{{"ENTER", "PREVIEW"}, {"ESC", "SETUP"}},
		)
	}
	if m.importConfirm {
		return m.renderShortcutHints(shortcutHint{"Y", "YES"}, shortcutHint{"N", "NO"}, shortcutHint{"ESC", "CANCEL"})
	}
	return m.renderShortcutHintsFit(width,
		[]shortcutHint{{"↑↓", "SCROLL"}, {"SPACE/ENTER", "TOGGLE"}, {"A", "ALL"}, {"N", "NONE"}, {"I / CTRL+S", "IMPORT"}, {"ESC", "PATH"}},
		[]shortcutHint{{"SPACE/ENTER", "TOGGLE"}, {"A", "ALL"}, {"N", "NONE"}, {"I / CTRL+S", "IMPORT"}, {"ESC", "PATH"}},
		[]shortcutHint{{"SPACE/ENTER", "TOGGLE"}, {"I / CTRL+S", "IMPORT"}, {"ESC", "PATH"}},
	)
}

// gameImportKey compares actual launch identities rather than names. Different
// games may legitimately share the same display title.
func gameImportKey(g config.GameConfig) string {
	return g.Command + "\x00" + strings.Join(g.Args, "\x00") + "\x00" + g.WorkingDir
}

func discoverImportCandidates(pattern string, existing []config.GameConfig) ([]importCandidate, int, error) {
	pattern = strings.TrimSpace(pattern)
	if len(pattern) >= 2 && ((pattern[0] == '"' && pattern[len(pattern)-1] == '"') || (pattern[0] == '\'' && pattern[len(pattern)-1] == '\'')) {
		pattern = pattern[1 : len(pattern)-1]
	}
	pattern = config.ExpandPath(pattern)
	if pattern == "" {
		return nil, 0, fmt.Errorf("enter a path or pattern")
	}
	var paths []string
	if strings.ContainsAny(pattern, "*?[") {
		var err error
		paths, err = filepath.Glob(pattern)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid path pattern: %w", err)
		}
	} else {
		st, err := os.Stat(pattern)
		if err != nil {
			return nil, 0, err
		}
		if st.IsDir() {
			entries, err := os.ReadDir(pattern)
			if err != nil {
				return nil, 0, err
			}
			for _, ent := range entries {
				paths = append(paths, filepath.Join(pattern, ent.Name()))
			}
		} else {
			paths = []string{pattern}
		}
	}
	const maxImportMatches = 10000
	if len(paths) > maxImportMatches {
		return nil, 0, fmt.Errorf("too many matches (%d); refine the pattern", len(paths))
	}
	_, gioErr := exec.LookPath("gio")
	seen := make(map[string]bool, len(existing)+len(paths))
	missingGio := false
	for _, g := range existing {
		seen[gameImportKey(g)] = true
	}
	var candidates []importCandidate
	skipped := 0
	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			skipped++
			continue
		}
		st, err := os.Stat(abs)
		if err != nil || !st.Mode().IsRegular() {
			skipped++
			continue
		}
		var c importCandidate
		c.path = abs
		if strings.EqualFold(filepath.Ext(abs), ".desktop") {
			if gioErr != nil {
				missingGio = true
				skipped++
				continue
			}
			name, ok := desktopImportName(abs)
			if !ok {
				skipped++
				continue
			}
			c.kind = "DESKTOP"
			c.game = config.GameConfig{Name: name, Command: "gio", Args: []string{"launch", abs}}
		} else {
			if st.Mode().Perm()&0o111 == 0 || !nativeImportFile(abs) {
				skipped++
				continue
			}
			c.kind = "EXECUTABLE"
			c.game = config.GameConfig{Name: nativeImportName(abs), Command: abs, WorkingDir: filepath.Dir(abs)}
		}
		c.game.Name = strings.TrimSpace(strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return ' '
			}
			return r
		}, c.game.Name))
		key := gameImportKey(c.game)
		c.duplicate = seen[key]
		c.selected = !c.duplicate
		seen[key] = true
		candidates = append(candidates, c)
	}
	if missingGio && len(candidates) == 0 {
		return nil, skipped, fmt.Errorf("gio is required to import .desktop shortcuts")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := strings.ToLower(candidates[i].game.Name), strings.ToLower(candidates[j].game.Name)
		if a == b {
			return candidates[i].path < candidates[j].path
		}
		return a < b
	})
	return candidates, skipped, nil
}

// This parser reads only metadata; it deliberately does NOT interpret Exec= or
// execute arbitrary contents while scanning. gio owns desktop launch semantics.
func desktopImportName(path string) (string, bool) {
	st, err := os.Stat(path)
	if err != nil || st.Size() > 1024*1024 {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	inEntry := false
	fields := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "[Desktop Entry]" {
			inEntry = true
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inEntry = false
			continue
		}
		if !inEntry || line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			fields[strings.TrimSpace(line[:idx])] = strings.TrimSpace(line[idx+1:])
		}
	}
	if fields["Type"] != "Application" || strings.EqualFold(fields["Hidden"], "true") || strings.EqualFold(fields["NoDisplay"], "true") {
		return "", false
	}
	if fields["Exec"] == "" && !strings.EqualFold(fields["DBusActivatable"], "true") {
		return "", false
	}
	name := fields["Name"]
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return name, true
}

// Exclude Windows-specific shortcuts/executables and shared libraries from
// automatic discovery. Manual ADD GAME still permits Wine/Proton wrappers.
func nativeImportFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".exe", ".lnk", ".url", ".bat", ".cmd", ".msi", ".dll", ".so":
		return false
	default:
		return true
	}
}

// Repeated generic launch script names should become distinguishable titles:
// ~/Games/Celeste/start.sh -> Celeste, not 200 entries named "start".
func nativeImportName(path string) string {
	file := filepath.Base(path)
	name := strings.TrimSuffix(file, filepath.Ext(file))
	switch strings.ToLower(name) {
	case "start", "run", "launch", "play", "game", "start-game", "run-game":
		if folder := filepath.Base(filepath.Dir(path)); folder != "." && folder != string(filepath.Separator) {
			return folder
		}
	}
	return name
}

// gio launch is a desktop activation request, not a reliable game-lifetime
// process. The caller must not steal focus back immediately when it returns.
func isDesktopGameLaunch(command string, args []string) bool {
	return filepath.Base(command) == "gio" && len(args) >= 2 && args[0] == "launch" && strings.EqualFold(filepath.Ext(args[1]), ".desktop")
}
