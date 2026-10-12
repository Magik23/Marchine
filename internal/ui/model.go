package ui

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
	"github.com/magik23/marchine/internal/mame"
)

type arcadeLoadedMsg struct {
	Result mame.Result
	Err    error
}

type launchFinishedMsg struct {
	Name              string
	Err               error
	RestoreFullscreen bool
}

type displayRotateFinishedMsg struct {
	Err error
}

// Optional companion utility: Marchine neither rotates monitors itself nor
// requires hypr-rotate to be installed. The command runs asynchronously.
func rotateDisplayCmd() tea.Cmd {
	return func() tea.Msg {
		return displayRotateFinishedMsg{Err: exec.Command("hypr-rotate").Run()}
	}
}

type screenMode int

const (
	screenLibrary screenMode = iota
	screenSetup
	screenInfo
	screenCredits
	screenCategories
)

const (
	filterAll       = "ALL"
	filterArcade    = "ARCADE ONLY"
	filterVertical  = "VERTICAL"
	filterNonArcade = "NON - ARCADE"
)

const (
	setupFieldMAME = iota
	setupFieldROM
	setupFieldCount
)

const (
	infoSettingHideNonArcade = iota
	infoSettingTheme
	infoSettingCount
)

type Model struct {
	letterJump       string
	letterJumpOrigin int
	cfg              config.Config
	configPath       string
	theme            Theme
	games            []library.Game
	arcade           []library.Game
	source           library.Source
	filter           string
	cursor           int
	search           string
	searchMode       bool
	width            int
	height           int

	loading     bool
	forceRescan bool
	arcadeMeta  mame.Result
	status      string
	lastLaunch  string
	demo        bool

	screen         screenMode
	overlayHistory []screenMode
	setupField     int
	setupEditing   bool
	setupBuffer    string
	setupMAME      string
	setupROM       string

	categoryCursor int
	infoSetting    int
}

func NewModel(cfg config.Config, configPath string, forceRescan, demo bool) Model {
	if strings.TrimSpace(configPath) == "" {
		configPath, _ = config.ConfigPath()
	}
	setupMAME := cfg.Arcade.MAMECommand
	if p, err := exec.LookPath(setupMAME); err == nil {
		setupMAME = p
	}
	setupROM := ""
	if len(cfg.Arcade.ROMPaths) > 0 {
		setupROM = cfg.Arcade.ROMPaths[0]
	}

	m := Model{
		cfg:         cfg,
		configPath:  configPath,
		theme:       ThemeByName(cfg.Theme),
		source:      library.SourceGames,
		loading:     cfg.Arcade.Enabled && !demo,
		forceRescan: forceRescan,
		status:      "READY",
		demo:        demo,
		setupMAME:   setupMAME,
		setupROM:    setupROM,
	}
	for i, g := range cfg.Games {
		if strings.TrimSpace(g.Name) == "" || strings.TrimSpace(g.Command) == "" {
			continue
		}
		m.games = append(m.games, library.Game{
			ID:         fmt.Sprintf("game:%d", i),
			Name:       g.Name,
			Source:     library.SourceGames,
			Command:    g.Command,
			Args:       append([]string(nil), g.Args...),
			WorkingDir: g.WorkingDir,
		})
	}
	sort.SliceStable(m.games, func(i, j int) bool { return strings.ToLower(m.games[i].Name) < strings.ToLower(m.games[j].Name) })

	if demo {
		m.arcade = demoArcade()
		m.games = demoGames()
		m.loading = false
		m.arcadeMeta = mame.Result{MAMEVersion: "0.289", Warning: "DEMO MODE"}
		if cfg.Arcade.HideJunk {
			m.filter = filterArcade
		}
	}
	if strings.EqualFold(cfg.StartupSource, "arcade") && (cfg.Arcade.Enabled || demo) {
		m.source = library.SourceArcade
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.loading {
		return m.loadArcadeCmd(m.forceRescan)
	}
	return nil
}

func (m Model) loadArcadeCmd(force bool) tea.Cmd {
	cfg := m.cfg.Arcade
	return func() tea.Msg {
		res, err := mame.Load(cfg, force)
		return arcadeLoadedMsg{Result: res, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case arcadeLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.status = "ARCADE INDEX ERROR: " + msg.Err.Error()
			return m, nil
		}
		// A scan command captures the config that existed when it started. Apply
		// the model's current launch identity before exposing the result so a
		// Setup save made during indexing can never leave stale launch commands.
		mame.ApplyLaunchConfig(msg.Result.Games, m.cfg.Arcade)
		m.arcadeMeta = msg.Result
		m.arcade = msg.Result.Games
		if m.filter == "" && m.cfg.Arcade.HideJunk {
			m.filter = filterArcade
		}
		if msg.Result.Warning != "" {
			m.status = msg.Result.Warning
		} else if msg.Result.FromCache {
			m.status = fmt.Sprintf("ARCADE READY · %d GAMES · CACHE", len(m.arcade))
		} else {
			m.status = fmt.Sprintf("ARCADE READY · %d GAMES", len(m.arcade))
		}
		m.normalizeCursor()
		return m, nil

	case launchFinishedMsg:
		SetGamepadSuspended(false)

		m.lastLaunch = msg.Name
		if msg.Err != nil {
			m.status = "LAUNCH FAILED: " + msg.Err.Error()
		} else {
			m.status = "RETURNED FROM " + strings.ToUpper(msg.Name)
		}

		return m, restoreMarchineWindowCmd(msg.RestoreFullscreen)

	case displayRotateFinishedMsg:
		switch {
		case msg.Err == nil:
			m.status = "DISPLAY ROTATED"
		case errors.Is(msg.Err, exec.ErrNotFound):
			m.status = "DISPLAY ROTATE NOT INSTALLED"
		default:
			m.status = "DISPLAY ROTATE FAILED: " + msg.Err.Error()
		}
		return m, nil

	case tea.PasteMsg:
		// Bubble Tea v2 emits bracketed terminal paste as its own message.
		// Path fields are exactly where users are most likely to paste long text.
		if m.screen == screenSetup && m.setupEditing {
			m.setupBuffer += msg.Content
			return m, nil
		}
		if m.screen == screenLibrary && m.searchMode {
			m.search += msg.Content
			m.cursor = 0
			return m, nil
		}

	case GamepadMsg:
		return m.updateGamepad(msg.Action)

	case tea.KeyPressMsg:
		// Keyboard index navigation must match controller preview/confirm/cancel.
		if m.letterJump != "" && m.screen == screenLibrary && !m.searchMode {
			switch msg.String() {
			case "[":
				m.cycleJumpLetter(-1)
				return m, nil
			case "]":
				m.cycleJumpLetter(1)
				return m, nil
			case "enter":
				m.confirmJumpLetter()
				return m, nil
			case "esc":
				m.cancelJumpLetter()
				return m, nil
			default:
				m.letterJump = ""
				m.status = "READY"
			}
		}

		// Match CLIAMP-style responsive behavior: while the terminal is below
		// Marchine's minimum usable geometry, hide the app and ignore launcher
		// actions. Resizing restores the exact previous state.
		if m.width > 0 && m.height > 0 && terminalTooSmall(m.width, m.height) {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}

		if m.screen == screenCategories {
			return m.updateCategories(msg)
		}
		if m.screen == screenInfo {
			return m.updateInfo(msg)
		}
		if m.screen == screenCredits {
			return m.updateCredits(msg)
		}
		if m.screen == screenSetup {
			if m.setupEditing {
				return m.updateSetupEdit(msg)
			}
			if msg.String() == "i" || msg.String() == "I" || msg.String() == "ctrl+k" {
				m.openInfo()
				return m, nil
			}
			if msg.String() == "o" || msg.String() == "O" || msg.String() == "ctrl+g" {
				m.openCredits()
				return m, nil
			}
			return m.updateSetup(msg)
		}
		if m.searchMode {
			return m.updateSearch(msg)
		}

		key := msg.String()
		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "i", "I", "ctrl+k":
			m.openInfo()
		case "o", "O", "ctrl+g":
			m.openCredits()
		case "s", "S":
			m.openSetup()
		case "tab":
			m.toggleSource()
		case "c":
			if m.source == library.SourceArcade {
				m.openCategories()
			}
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "[":
			m.cycleJumpLetter(-1)
		case "]":
			m.cycleJumpLetter(1)
		case "pgup":
			m.moveCursor(-m.pageSize())
		case "pgdown":
			m.moveCursor(m.pageSize())
		case "home":
			m.cursor = 0
		case "end":
			if n := len(m.filtered()); n > 0 {
				m.cursor = n - 1
			}
		case "/":
			m.searchMode = true
			m.status = "SEARCH"
		case "esc":
			if m.search != "" {
				m.search = ""
				m.cursor = 0
			}
		case "t", "ctrl+t":
			m.cycleTheme()
		case "h", "H":
			m.toggleHideNonArcade()
		case "r", "R":
			if m.demo {
				m.status = "DEMO MODE · REFRESH DISABLED"
				return m, nil
			}
			if m.loading {
				m.status = "ARCADE REFRESH ALREADY IN PROGRESS"
				return m, nil
			}
			if m.cfg.Arcade.Enabled {
				m.loading = true
				m.status = "REFRESHING ARCADE LIBRARY…"
				return m, m.loadArcadeCmd(true)
			}
		case "enter":
			if g, ok := m.selected(); ok {
				m.status = "LAUNCHING · " + strings.ToUpper(g.Name)

				SetGamepadSuspended(true)

				name := g.Name
				restoreFullscreen := marchineHyprlandFullscreen()

				// Arcade/MAME is graphical, so keep Marchine's terminal UI alive
				// behind the game instead of handing the terminal over to MAME.
				if g.Source == library.SourceArcade {
					command := g.Command
					args := append([]string(nil), g.Args...)
					workingDir := g.WorkingDir

					return m, func() tea.Msg {
						cmd := exec.Command(command, args...)
						if workingDir != "" {
							cmd.Dir = workingDir
						}

						err := cmd.Run()

						return launchFinishedMsg{
							Name:              name,
							Err:               err,
							RestoreFullscreen: restoreFullscreen,
						}
					}
				}

				// Generic GAMES entries keep the existing terminal-handoff behavior.
				cmd := exec.Command(g.Command, g.Args...)
				if g.WorkingDir != "" {
					cmd.Dir = g.WorkingDir
				}

				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return launchFinishedMsg{
						Name:              name,
						Err:               err,
						RestoreFullscreen: restoreFullscreen,
					}
				})
			}
		}
	}
	return m, nil
}

func (m *Model) cycleTheme() {
	m.theme = NextTheme(m.theme.Name)
	m.cfg.Theme = m.theme.Name
	m.status = "THEME · " + ThemeDisplayName(m.theme.Name)
	if m.demo {
		return
	}
	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.status = "THEME SAVE FAILED: " + err.Error()
	}
}

func (m *Model) toggleHideNonArcade() {
	// Remember the selected game, not just its list index: hiding entries
	// changes the indexes of games that remain visible.
	before := m.filtered()
	previousCursor := m.cursor
	m.cfg.Arcade.HideJunk = !m.cfg.Arcade.HideJunk

	// H changes the visibility of non-arcade games within the current
	// category (including VERTICAL); it must not discard that selection.
	// Only the broad ALL / ARCADE ONLY / NON - ARCADE views are replaced
	// when toggling Hide Non-Arcade, as before.
	switch m.filter {
	case "", filterAll, filterArcade, filterNonArcade:
		if m.cfg.Arcade.HideJunk {
			m.filter = filterArcade
		} else {
			m.filter = ""
		}
	}
	if m.cfg.Arcade.HideJunk {
		m.status = "HIDE NON-ARCADE · ON"
	} else {
		m.status = "HIDE NON-ARCADE · OFF"
	}
	m.categoryCursor = 0

	// Keep the same game selected if it survived. If it was hidden, select
	// the next surviving game in the old list order, or the previous one
	// when there is no next game. The list renderer follows m.cursor.
	after := m.filtered()
	m.cursor = 0
	if previousCursor >= 0 && previousCursor < len(before) && len(after) > 0 {
		positions := make(map[string]int, len(after))
		for i, game := range after {
			positions[game.ID] = i
		}

		if i, ok := positions[before[previousCursor].ID]; ok {
			m.cursor = i
		} else {
			found := false
			for i := previousCursor + 1; i < len(before); i++ {
				if next, ok := positions[before[i].ID]; ok {
					m.cursor = next
					found = true
					break
				}
			}
			if !found {
				for i := previousCursor - 1; i >= 0; i-- {
					if previous, ok := positions[before[i].ID]; ok {
						m.cursor = previous
						break
					}
				}
			}
		}
	}
	if m.demo {
		return
	}
	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.status = "FILTER SAVE FAILED: " + err.Error()
	}
}

func (m *Model) openCategories() {
	if m.source != library.SourceArcade {
		return
	}
	filters := m.filters()
	m.categoryCursor = 0
	for i, f := range filters {
		if f == m.activeFilter() {
			m.categoryCursor = i
			break
		}
	}
	m.screen = screenCategories
	m.searchMode = false
	m.status = "CATEGORIES"
}

func (m Model) updateCategories(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	filters := m.filters()
	if len(filters) == 0 {
		m.screen = screenLibrary
		m.status = "READY"
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "c":
		m.screen = screenLibrary
		m.status = "READY"
	case "i", "I", "ctrl+k":
		m.openInfo()
	case "o", "O", "ctrl+g":
		m.openCredits()
	case "t", "ctrl+t":
		m.cycleTheme()
	case "h", "H":
		m.toggleHideNonArcade()
	case "up", "k":
		m.categoryCursor--
		if m.categoryCursor < 0 {
			m.categoryCursor = len(filters) - 1
		}
	case "down", "j":
		m.categoryCursor++
		if m.categoryCursor >= len(filters) {
			m.categoryCursor = 0
		}
	case "home":
		m.categoryCursor = 0
	case "end":
		m.categoryCursor = len(filters) - 1
	case "enter":
		choice := filters[m.categoryCursor]
		if choice == "ALL" {
			m.filter = ""
		} else {
			m.filter = choice
		}
		m.search = ""
		m.cursor = 0
		m.screen = screenLibrary
		m.status = "READY"
	}
	return m, nil
}

// openOverlay remembers the current screen so Esc can return to where the
// user actually came from. Switching back to the immediate previous overlay
// unwinds history instead of building an Info/Credits loop.
func (m *Model) openOverlay(target screenMode) bool {
	if m.screen == target {
		return true
	}
	n := len(m.overlayHistory)
	if n > 0 && m.overlayHistory[n-1] == target {
		m.overlayHistory = m.overlayHistory[:n-1]
		m.screen = target
		return true
	}
	m.overlayHistory = append(m.overlayHistory, m.screen)
	m.screen = target
	return false
}

func (m *Model) closeOverlay() {
	n := len(m.overlayHistory)
	if n == 0 {
		m.screen = screenLibrary
		return
	}
	m.screen = m.overlayHistory[n-1]
	m.overlayHistory = m.overlayHistory[:n-1]
}

func (m *Model) openInfo() {
	returning := m.openOverlay(screenInfo)
	m.searchMode = false
	if !returning {
		m.infoSetting = infoSettingHideNonArcade
	}
}

func (m Model) updateInfo(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k", "shift+tab":
		m.infoSetting = (m.infoSetting - 1 + infoSettingCount) % infoSettingCount
	case "down", "j", "tab":
		m.infoSetting = (m.infoSetting + 1) % infoSettingCount
	case "enter":
		switch m.infoSetting {
		case infoSettingHideNonArcade:
			m.toggleHideNonArcade()
		case infoSettingTheme:
			m.cycleTheme()
		}
	case "o", "O", "ctrl+g":
		m.openCredits()
	case "t", "ctrl+t":
		m.cycleTheme()
	case "h", "H":
		m.toggleHideNonArcade()
	case "esc", "i", "I", "ctrl+k":
		m.closeOverlay()
		if m.status == "" {
			m.status = "READY"
		}
	}
	return m, nil
}

func (m *Model) openCredits() {
	m.openOverlay(screenCredits)
	m.searchMode = false
}

func (m Model) updateCredits(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "i", "I", "ctrl+k":
		m.openInfo()
	case "t", "ctrl+t":
		m.cycleTheme()
	case "esc", "o", "O", "ctrl+g":
		m.closeOverlay()
		m.status = "READY"
	}
	return m, nil
}

func (m *Model) openSetup() {
	m.screen = screenSetup
	m.setupEditing = false
	m.setupBuffer = ""
	if strings.TrimSpace(m.setupMAME) == "" {
		m.setupMAME = m.cfg.Arcade.MAMECommand
	}
	if strings.TrimSpace(m.setupROM) == "" && len(m.cfg.Arcade.ROMPaths) > 0 {
		m.setupROM = m.cfg.Arcade.ROMPaths[0]
	}
	m.status = "ARCADE SETUP"
}

func (m Model) updateSetup(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "s":
		m.screen = screenLibrary
		m.status = "READY"
	case "up", "k", "shift+tab":
		m.setupField = (m.setupField - 1 + setupFieldCount) % setupFieldCount
	case "down", "j", "tab":
		m.setupField = (m.setupField + 1) % setupFieldCount
	case "enter":
		m.setupEditing = true
		if m.setupField == setupFieldMAME {
			m.setupBuffer = m.setupMAME
		} else {
			m.setupBuffer = m.setupROM
		}
	case "ctrl+s":
		return m.saveSetup(false)
	case "t", "ctrl+t":
		m.cycleTheme()
	case "r":
		return m.saveSetup(true)
	}
	return m, nil
}

func (m Model) updateSetupEdit(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.setupEditing = false
		m.setupBuffer = ""
		return m, nil
	case "enter":
		if m.setupField == setupFieldMAME {
			m.setupMAME = strings.TrimSpace(m.setupBuffer)
		} else {
			m.setupROM = strings.TrimSpace(m.setupBuffer)
		}
		m.setupEditing = false
		m.setupBuffer = ""
		m.status = "FIELD UPDATED · CTRL+S TO SAVE"
		return m, nil
	case "backspace":
		if m.setupBuffer != "" {
			_, size := utf8.DecodeLastRuneInString(m.setupBuffer)
			m.setupBuffer = m.setupBuffer[:len(m.setupBuffer)-size]
		}
		return m, nil
	case "ctrl+u":
		m.setupBuffer = ""
		return m, nil
	}
	if msg.Text != "" {
		m.setupBuffer += msg.Text
	}
	return m, nil
}

func (m Model) saveSetup(refresh bool) (tea.Model, tea.Cmd) {
	if m.demo {
		m.status = "DEMO MODE · SETTINGS NOT WRITTEN"
		return m, nil
	}
	mameCommand := strings.TrimSpace(m.setupMAME)
	if mameCommand == "" {
		mameCommand = "mame"
	}
	m.cfg.Arcade.Enabled = true
	m.cfg.Arcade.MAMECommand = mameCommand

	// Setup intentionally edits only the primary ROM path. Preserve any
	// additional paths that were configured manually instead of silently
	// collapsing a multi-path configuration to one entry. Clearing the primary
	// field is explicit: it restores MAME rompath auto-detection and clears the
	// configured list.
	rom := strings.TrimSpace(m.setupROM)
	if rom == "" {
		m.cfg.Arcade.ROMPaths = nil
	} else {
		paths := []string{rom}
		seen := map[string]bool{rom: true}
		if len(m.cfg.Arcade.ROMPaths) > 1 {
			for _, existing := range m.cfg.Arcade.ROMPaths[1:] {
				existing = strings.TrimSpace(existing)
				if existing == "" || seen[existing] {
					continue
				}
				seen[existing] = true
				paths = append(paths, existing)
			}
		}
		m.cfg.Arcade.ROMPaths = paths
	}

	// Keep the in-memory model normalized as well as the persisted copy. Save
	// normalizes its value argument, but Setup must also use expanded ~/env paths
	// immediately for refresh/launch without requiring a Marchine restart.
	m.cfg.Normalize()

	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.status = "SETUP SAVE FAILED: " + err.Error()
		return m, nil
	}

	// Saving launch identity should take effect immediately, even if the user
	// chooses not to refresh metadata.
	mame.ApplyLaunchConfig(m.arcade, m.cfg.Arcade)
	m.status = "SETUP SAVED"

	if refresh {
		if m.loading {
			m.status = "SETUP SAVED · REFRESH ALREADY IN PROGRESS"
			return m, nil
		}
		m.loading = true
		m.source = library.SourceArcade
		m.status = "SETUP SAVED · REFRESHING ARCADE LIBRARY…"
		return m, m.loadArcadeCmd(true)
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc", "enter":
		m.searchMode = false
		m.status = "READY"
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "backspace":
		if m.search != "" {
			_, size := utf8.DecodeLastRuneInString(m.search)
			m.search = m.search[:len(m.search)-size]
			m.cursor = 0
		}
		return m, nil
	case "ctrl+u":
		m.search = ""
		m.cursor = 0
		return m, nil
	}
	if msg.Text != "" {
		m.search += msg.Text
		m.cursor = 0
	}
	return m, nil
}

func (m *Model) toggleSource() {
	if m.source == library.SourceGames {
		m.source = library.SourceArcade
		if m.cfg.Arcade.HideJunk {
			m.filter = filterArcade
		} else {
			m.filter = ""
		}
	} else {
		m.source = library.SourceGames
		m.filter = ""
	}
	m.search = ""
	m.cursor = 0
}

func (m *Model) moveCursor(delta int) {
	n := len(m.filtered())
	if n == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
}

func (m *Model) normalizeCursor() {
	n := len(m.filtered())
	if n == 0 {
		m.cursor = 0
	} else if m.cursor >= n {
		m.cursor = n - 1
	}
}

func (m Model) pageSize() int {
	if m.height <= 24 {
		return 7
	}
	return max(8, m.height-16)
}

func (m Model) selected() (library.Game, bool) {
	items := m.filtered()
	if len(items) == 0 || m.cursor < 0 || m.cursor >= len(items) {
		return library.Game{}, false
	}
	return items[m.cursor], true
}

func (m Model) activeFilter() string {
	if m.source == library.SourceArcade && m.cfg.Arcade.HideJunk {
		if m.filter == "" || m.filter == filterAll || m.filter == filterNonArcade {
			return filterArcade
		}
	}
	if m.filter == "" {
		return filterAll
	}
	return m.filter
}

func (m Model) filters() []string {
	if len(m.arcade) == 0 {
		if m.cfg.Arcade.HideJunk {
			return []string{filterArcade}
		}
		return []string{filterAll, filterArcade}
	}

	seen := map[string]bool{}
	arcadeSeen := map[string]bool{}
	hasVertical := false
	hasNonArcade := false
	for _, g := range m.arcade {
		if g.Vertical && (!m.cfg.Arcade.HideJunk || !g.NonArcade) {
			hasVertical = true
		}
		if g.Category != "" {
			seen[g.Category] = true
			if !g.NonArcade {
				arcadeSeen[g.Category] = true
			}
		}
		if g.NonArcade {
			hasNonArcade = true
		}
	}

	out := []string{filterArcade}
	if !m.cfg.Arcade.HideJunk {
		out = []string{filterAll, filterArcade}
	}

	preferred := []string{
		"ACTION", "BALL & PADDLE", "BEAT'EM UP", "FIGHTING", "LIGHTGUN",
		"MAZE", "PLATFORM", "PUZZLE", "RACING", "RHYTHM", "RUN&GUN",
		"SHMUPS", "SHOOTER", "SPORTS", "MINIGAMES", "QUIZ", "PINBALL",
		"CASINO / GAMBLING", "MAHJONG", "MEDAL / PACHINKO",
		"HANAFUDA / CARD", "TABLETOP / CARD", "OTHER / UTILITY",
		"UNCLASSIFIED",
	}
	used := map[string]bool{}
	for _, f := range preferred {
		if !seen[f] || (m.cfg.Arcade.HideJunk && !arcadeSeen[f]) {
			continue
		}
		out = append(out, f)
		used[f] = true
	}

	// CatVer can introduce category families Marchine does not know yet. Keep
	// those visible too, but when Hide Non-Arcade is ON omit category families
	// that contain no actual arcade entries.
	var extra []string
	for f := range seen {
		if used[f] || (m.cfg.Arcade.HideJunk && !arcadeSeen[f]) {
			continue
		}
		extra = append(extra, f)
	}
	sort.Strings(extra)
	out = append(out, extra...)

	if hasVertical {
		out = append(out, filterVertical)
	}

	if !m.cfg.Arcade.HideJunk && hasNonArcade {
		out = append(out, filterNonArcade)
	}
	return out
}

func matchesArcadeFilter(g library.Game, filter string, hideNonArcade bool) bool {
	if hideNonArcade && g.NonArcade {
		return false
	}
	switch filter {
	case "", filterAll:
		return true
	case filterArcade:
		return !g.NonArcade
	case filterVertical:
		return g.Vertical
	case filterNonArcade:
		return g.NonArcade
	default:
		return g.Category == filter
	}
}

func (m Model) filtered() []library.Game {
	var src []library.Game
	if m.source == library.SourceArcade {
		src = m.arcade
	} else {
		src = m.games
	}
	q := strings.ToLower(strings.TrimSpace(m.search))
	filter := m.filter
	if q == "" && m.source != library.SourceArcade {
		return src
	}
	out := make([]library.Game, 0, len(src))
	for _, g := range src {
		if m.source == library.SourceArcade && !matchesArcadeFilter(g, filter, m.cfg.Arcade.HideJunk) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(g.Name + " " + g.ROM + " " + g.Manufacturer + " " + g.Year + " " + g.Category + " " + g.RawCategory)
			if !containsAll(hay, q) {
				continue
			}
		}
		out = append(out, g)
	}
	return out
}

func containsAll(hay, query string) bool {
	for _, token := range strings.Fields(query) {
		if !strings.Contains(hay, token) {
			return false
		}
	}
	return true
}

func demoArcade() []library.Game {
	rows := []struct {
		n, y, m, c, raw, r string
		nonArcade          bool
	}{
		{"1942", "1984", "Capcom", "SHMUPS", "Shooter / Flying Vertical", "1942", false},
		{"Batsugun", "1993", "Toaplan", "SHMUPS", "Shooter / Flying Vertical", "batsugun", false},
		{"Battle Garegga", "1996", "Raizing", "SHMUPS", "Shooter / Flying Vertical", "bgaregga", false},
		{"DoDonPachi", "1997", "Cave", "SHMUPS", "Shooter / Flying Vertical", "ddonpach", false},
		{"ESP Ra.De.", "1998", "Cave", "SHMUPS", "Shooter / Flying Vertical", "esprade", false},
		{"Metal Slug", "1996", "SNK", "RUN&GUN", "Shooter / Walking", "mslug", false},
		{"R-Type", "1987", "Irem", "SHMUPS", "Shooter / Flying Horizontal", "rtype", false},
		{"Street Fighter II': Champion Edition", "1992", "Capcom", "FIGHTING", "Fighter / Versus", "sf2ce", false},
		{"Tetris", "1988", "Atari", "PUZZLE", "Puzzle / Drop", "tetris", false},
		{"Super Mahjong", "1990", "Example", "MAHJONG", "Mahjong / Tiles", "mahjong", true},
		{"Video Poker", "1991", "Example", "CASINO / GAMBLING", "Casino / Cards", "vpoker", true},
	}
	out := make([]library.Game, 0, len(rows))
	for _, r := range rows {
		out = append(out, library.Game{ID: "mame:" + r.r, Name: r.n, Year: r.y, Manufacturer: r.m, Category: r.c, RawCategory: r.raw, NonArcade: r.nonArcade, ROM: r.r, Source: library.SourceArcade, Command: "mame", Args: []string{r.r}})
	}
	return out
}

func demoGames() []library.Game {
	return []library.Game{
		{ID: "game:1", Name: "Hades", Source: library.SourceGames, Command: "hades"},
		{ID: "game:2", Name: "OpenTyrian", Source: library.SourceGames, Command: "opentyrian"},
		{ID: "game:3", Name: "Quake", Source: library.SourceGames, Command: "quake"},
		{ID: "game:4", Name: "Terminator 2D", Source: library.SourceGames, Command: "steam", Args: []string{"-applaunch", "1718460"}},
		{ID: "game:5", Name: "VVVVVV", Source: library.SourceGames, Command: "vvvvvv"},
	}
}
