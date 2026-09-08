package ui

import (
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
	Name string
	Err  error
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
	filterNonArcade = "NON - ARCADE"
)

const (
	setupFieldMAME = iota
	setupFieldROM
	setupFieldCount
)

type Model struct {
	cfg        config.Config
	configPath string
	theme      Theme
	games      []library.Game
	arcade     []library.Game
	source     library.Source
	filter     string
	cursor     int
	search     string
	searchMode bool
	width      int
	height     int

	loading     bool
	forceRescan bool
	arcadeMeta  mame.Result
	status      string
	lastLaunch  string
	demo        bool

	screen        screenMode
	infoReturn    screenMode
	creditsReturn screenMode
	setupField    int
	setupEditing  bool
	setupBuffer   string
	setupMAME     string
	setupROM      string

	categoryCursor int
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
		m.arcadeMeta = msg.Result
		m.arcade = msg.Result.Games
		if m.filter == "" && m.cfg.Arcade.HideJunk {
			m.filter = filterArcade
		}
		if m.setupROM == "" && len(msg.Result.ROMPaths) > 0 {
			m.setupROM = msg.Result.ROMPaths[0]
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
		m.lastLaunch = msg.Name
		if msg.Err != nil {
			m.status = "LAUNCH FAILED: " + msg.Err.Error()
		} else {
			m.status = "RETURNED FROM " + strings.ToUpper(msg.Name)
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

	case tea.KeyPressMsg:
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
			if m.cfg.Arcade.Enabled {
				m.loading = true
				m.status = "REFRESHING ARCADE LIBRARY…"
				return m, m.loadArcadeCmd(true)
			}
		case "enter":
			if g, ok := m.selected(); ok {
				m.status = "LAUNCHING · " + strings.ToUpper(g.Name)
				cmd := exec.Command(g.Command, g.Args...)
				if g.WorkingDir != "" {
					cmd.Dir = g.WorkingDir
				}
				name := g.Name
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return launchFinishedMsg{Name: name, Err: err}
				})
			}
		}
	}
	return m, nil
}

func (m *Model) cycleTheme() {
	m.theme = NextTheme(m.theme.Name)
	m.cfg.Theme = m.theme.Name
	m.status = "THEME · " + strings.ToUpper(m.theme.Name)
	if m.demo {
		return
	}
	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.status = "THEME SAVE FAILED: " + err.Error()
	}
}

func (m *Model) toggleHideNonArcade() {
	m.cfg.Arcade.HideJunk = !m.cfg.Arcade.HideJunk
	if m.cfg.Arcade.HideJunk {
		m.filter = filterArcade
		m.status = "HIDE NON-ARCADE · ON"
	} else {
		m.filter = ""
		m.status = "HIDE NON-ARCADE · OFF"
	}
	m.categoryCursor = 0
	m.cursor = 0
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

func (m *Model) openInfo() {
	m.infoReturn = m.screen
	if m.infoReturn == screenInfo || m.infoReturn == screenCredits {
		m.infoReturn = screenLibrary
	}
	m.screen = screenInfo
	m.searchMode = false
}

func (m Model) updateInfo(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "o", "O", "ctrl+g":
		m.openCredits()
	case "t", "ctrl+t":
		m.cycleTheme()
	case "h", "H":
		m.toggleHideNonArcade()
	case "esc", "i", "I", "ctrl+k":
		if m.infoReturn == screenInfo || m.infoReturn == screenCredits {
			m.infoReturn = screenLibrary
		}
		m.screen = m.infoReturn
		if m.status == "" {
			m.status = "READY"
		}
	}
	return m, nil
}

func (m *Model) openCredits() {
	m.creditsReturn = m.screen
	if m.creditsReturn == screenCredits || m.creditsReturn == screenInfo {
		m.creditsReturn = screenLibrary
	}
	m.screen = screenCredits
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
		if m.creditsReturn == screenCredits || m.creditsReturn == screenInfo {
			m.creditsReturn = screenLibrary
		}
		m.screen = m.creditsReturn
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
	if strings.TrimSpace(m.setupROM) == "" {
		if len(m.cfg.Arcade.ROMPaths) > 0 {
			m.setupROM = m.cfg.Arcade.ROMPaths[0]
		} else if len(m.arcadeMeta.ROMPaths) > 0 {
			m.setupROM = m.arcadeMeta.ROMPaths[0]
		}
	}
	m.status = "ARCADE SETUP"
}

func (m Model) updateSetup(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c":
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
	rom := strings.TrimSpace(m.setupROM)
	if rom == "" {
		m.cfg.Arcade.ROMPaths = nil
	} else {
		m.cfg.Arcade.ROMPaths = []string{rom}
	}
	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.status = "SETUP SAVE FAILED: " + err.Error()
		return m, nil
	}
	m.status = "SETUP SAVED"
	if refresh {
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
	hasNonArcade := false
	for _, g := range m.arcade {
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
