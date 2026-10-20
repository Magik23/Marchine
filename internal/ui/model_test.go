package ui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/gamepad"
	"github.com/magik23/marchine/internal/library"
)

func creditArtSignature(t *testing.T) string {
	t.Helper()
	for _, line := range strings.Split(creditsBBSArt(), "\n") {
		if signature := strings.TrimSpace(line); signature != "" {
			return signature
		}
	}
	t.Fatal("credit artwork contains no visible lines")
	return ""
}

func TestDemoArcadeFiltering(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	m.filter = "SHMUPS"
	items := m.filtered()
	if len(items) < 4 {
		t.Fatalf("expected several shmups, got %d", len(items))
	}
	for _, g := range items {
		if g.Category != "SHMUPS" {
			t.Fatalf("unexpected category %q", g.Category)
		}
	}
}

func TestGamesRemainFlat(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceGames
	m.search = "term"
	items := m.filtered()
	if len(items) != 1 || items[0].Name != "Terminator 2D" {
		t.Fatalf("unexpected flat search result: %#v", items)
	}
}

func TestCompactFooterKeepsInfoAndQuitEscapeHatches(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := m.renderFooter(46)

	for _, want := range []string{"I", "INFO", "Q", "QUIT"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("compact footer lost %q: %q", want, footer)
		}
	}

	for _, forbidden := range []string{"THEME", "REINDEX", "SETUP", "CREDITS"} {
		if strings.Contains(footer, forbidden) {
			t.Fatalf("compact footer kept secondary control %q: %q", forbidden, footer)
		}
	}
}

func TestInfoAndCreditsAreSeparateScreens(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	info := m.renderInfo(120, 36)

	for _, want := range []string{
		"ABOUT MARCHINE",
		"SOURCES",
		"ARCADE: MAME ROMs only.",
		"PLAYLISTS: Mixed ARCADE and GAMES collections.",
		"Documentation: github.com/Magik23/Marchine",
		"SHORTCUTS",
		"O",
		"Credits",
		"QUICK SETTING",
		"HIDE NON-ARCADE",
	} {
		if !strings.Contains(info, want) {
			t.Fatalf("Info screen missing %q", want)
		}
	}

	for _, forbidden := range []string{
		"Pierre Dionne",
		"Albenoir Studio",
		"CREDITS\n",
		"FULLSCREEN",
	} {
		if strings.Contains(info, forbidden) {
			t.Fatalf("Info screen still contains credit material %q", forbidden)
		}
	}

	credits := m.renderCredits(160, 48)
	for _, want := range []string{
		"CREDITS",
		"Pierre Dionne",
		"Albenoir Studio",
		"MARCHINE — Game Machinery.",
	} {
		if !strings.Contains(credits, want) {
			t.Fatalf("Credits screen missing %q", want)
		}
	}
}

func TestDetailDoesNotExposeSculptureDebugUI(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	detail := m.renderDetail(42, 18)
	for _, forbidden := range []string{"CONTROL SCULPTURE", "HOLD", "ORBIT", "VIEW 0"} {
		if strings.Contains(detail, forbidden) {
			t.Fatalf("detail leaked sculpture debug label %q", forbidden)
		}
	}
}

func TestSetupUsesNeutralEmulatorLabel(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	setup := m.renderSetup(100, 36)
	if !strings.Contains(setup, "EMULATOR EXECUTABLE") {
		t.Fatalf("Setup missing neutral emulator label")
	}
}

func TestDetailNeverExposesRawLaunchCommand(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	detail := m.renderDetail(42, 18)
	for _, forbidden := range []string{"$ mame", "$ ", "mame 1942"} {
		if strings.Contains(detail, forbidden) {
			t.Fatalf("detail exposed implementation command %q: %q", forbidden, detail)
		}
	}
}

func TestFrontFooterKeepsPrimaryActionsOnly(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := m.renderFooter(180)

	for _, forbidden := range []string{
		"NAV",
		"REINDEX",
		"ROTATE",
		"THEME",
		"FILTER",
		"SETUP",
		"CREDITS",
		"MARCHINE // GAME MACHINERY",
	} {
		if strings.Contains(footer, forbidden) {
			t.Fatalf("front footer exposed secondary command %q: %q", forbidden, footer)
		}
	}

	for _, want := range []string{
		"SEARCH",
		"SOURCE",
		"CATEGORIES",
		"PLAY",
		"INFO",
		"QUIT",
	} {
		if !strings.Contains(footer, want) {
			t.Fatalf("front footer missing %q: %q", want, footer)
		}
	}
}

func TestTopNavigationOnlyShowsSources(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	tabs := m.renderSourceTabs(120)
	for _, want := range []string{"GAMES", "ARCADE", "PLAYLISTS"} {
		if !strings.Contains(tabs, want) {
			t.Fatalf("top navigation missing %q: %q", want, tabs)
		}
	}
	for _, forbidden := range []string{"SETUP", "INFO"} {
		if strings.Contains(tabs, forbidden) {
			t.Fatalf("top navigation exposed utility %q: %q", forbidden, tabs)
		}
	}
}

func TestSourceBarCountTracksVisibleLibrary(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	const width = 122

	// GAMES has its own count, independent of the larger Arcade library.
	m.source = library.SourceGames
	gamesCount := len(m.filtered())
	gamesBar := ansiEscape.ReplaceAllString(m.renderSourceTabs(width), "")
	if want := groupedCount(gamesCount) + " GAMES"; !strings.HasSuffix(strings.TrimRight(gamesBar, " "), want) {
		t.Fatalf("Games source bar = %q; want suffix %q", gamesBar, want)
	}
	if got := lipgloss.Width(m.renderSourceTabs(width)); got != width {
		t.Fatalf("Games source bar width = %d; want %d", got, width)
	}

	m.toggleSource()
	if m.source != library.SourcePlaylists || m.screen != screenPlaylists {
		t.Fatal("switching Games to Playlists failed")
	}
	m.toggleSource()
	if m.source != library.SourceArcade {
		t.Fatal("switching Playlists to Arcade failed")
	}
	arcadeCount := len(m.filtered())
	arcadeBar := ansiEscape.ReplaceAllString(m.renderSourceTabs(width), "")
	if want := groupedCount(arcadeCount) + " GAMES"; !strings.HasSuffix(strings.TrimRight(arcadeBar, " "), want) {
		t.Fatalf("Arcade source bar = %q; want suffix %q", arcadeBar, want)
	}
	if arcadeCount == gamesCount {
		t.Fatal("test setup requires different Arcade and Games counts")
	}

	// Filtering and searching must describe the currently displayed games.
	m.filter = "SHMUPS"
	if want := groupedCount(len(m.filtered())) + " MATCHES"; !strings.HasSuffix(strings.TrimRight(ansiEscape.ReplaceAllString(m.renderSourceTabs(width), ""), " "), want) {
		t.Fatalf("category counter missing %q", want)
	}
	m.search = "no-such-mame-title-for-testing"
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(width), ""); !strings.HasSuffix(strings.TrimRight(got, " "), "0 MATCHES") {
		t.Fatalf("empty search counter = %q", got)
	}
	m.search = ""
	m.filter = ""
	m.cfg.Arcade.HideJunk = false
	if want := groupedCount(len(m.filtered())) + " GAMES"; !strings.HasSuffix(strings.TrimRight(ansiEscape.ReplaceAllString(m.renderSourceTabs(width), ""), " "), want) {
		t.Fatalf("hide non-arcade OFF should show full library counter %q", want)
	}

	// Secondary screens keep the same uncluttered source tabs.
	m.screen = screenInfo
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(width), ""); !strings.HasSuffix(got, " ]") {
		t.Fatalf("Info source bar should end with the GAMES tab, got %q", got)
	}
	if got := lipgloss.Width(m.renderSourceTabs(width)); got >= width {
		t.Fatalf("Info source tabs unexpectedly fill width with counter: got %d", got)
	}
}

func TestSourceBarCountHandlesLargeAndEmptyLibraries(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.source = library.SourceGames
	m.games = make([]library.Game, 3574)
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(122), ""); !strings.HasSuffix(strings.TrimRight(got, " "), "3,574 GAMES") {
		t.Fatalf("thousands separator lost: %q", got)
	}
	m.games = nil
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(122), ""); !strings.HasSuffix(strings.TrimRight(got, " "), "0 GAMES") {
		t.Fatalf("empty library should show zero: %q", got)
	}
	m.games = []library.Game{{Name: "One game"}}
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(122), ""); !strings.HasSuffix(strings.TrimRight(got, " "), "1 GAME") {
		t.Fatalf("single game should use singular label: %q", got)
	}
	m.search = "one"
	if got := ansiEscape.ReplaceAllString(m.renderSourceTabs(122), ""); !strings.HasSuffix(strings.TrimRight(got, " "), "1 MATCH") {
		t.Fatalf("single search hit should use singular label: %q", got)
	}
}

func TestSourceBarDoesNotClaimZeroWhileArcadeIndexes(t *testing.T) {
	m := NewModel(config.Default(), "", false, false)
	m.source = library.SourceArcade
	m.loading = true
	m.arcade = nil
	plain := ansiEscape.ReplaceAllString(m.renderSourceTabs(122), "")
	if !strings.Contains(plain, "INDEXING ARCADE") || strings.Contains(plain, "0 GAMES") {
		t.Fatalf("Arcade index counter should wait for library: %q", plain)
	}
}

func TestArcadeCategoriesMoveOutOfTopRowAndIntoMenu(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade

	tabs := m.renderSourceTabs(120)
	for _, forbidden := range []string{"ALL", "SHMUPS", "FIGHTING", "RUN&GUN", "PUZZLE"} {
		if strings.Contains(tabs, forbidden) {
			t.Fatalf("top source row exposed category %q: %q", forbidden, tabs)
		}
	}

	menu := m.renderCategories(100, 24)
	for _, want := range []string{"ARCADE CATEGORIES", "ARCADE ONLY", "SHMUPS", "FIGHTING", "RUN&GUN", "PUZZLE"} {
		if !strings.Contains(menu, want) {
			t.Fatalf("category menu missing %q", want)
		}
	}
}

func TestCategoryShortcutAppliesPlaylistFilter(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	m.width = minTerminalWidth
	m.height = 50

	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Text: "c"}))
	m = next.(Model)
	if m.screen != screenCategories {
		t.Fatalf("C should open category menu, screen=%v", m.screen)
	}

	filters := m.filters()
	idx := -1
	for i, f := range filters {
		if f == "FIGHTING" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("demo data should expose FIGHTING category")
	}
	m.categoryCursor = idx

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)
	if m.screen != screenLibrary || m.filter != "FIGHTING" {
		t.Fatalf("Enter should apply FIGHTING and return to library: screen=%v filter=%q", m.screen, m.filter)
	}
	items := m.filtered()
	if len(items) == 0 {
		t.Fatal("FIGHTING category should populate the main playlist")
	}
	for _, g := range items {
		if g.Category != "FIGHTING" {
			t.Fatalf("filtered playlist leaked category %q", g.Category)
		}
	}
}

func TestTerminalMinimumGate(t *testing.T) {
	if terminalTooSmall(minTerminalWidth, minTerminalHeight) {
		t.Fatal("exact minimum terminal size should render Marchine")
	}
	if !terminalTooSmall(minTerminalWidth-1, minTerminalHeight) {
		t.Fatal("terminal one column below minimum should be hidden")
	}
	if !terminalTooSmall(minTerminalWidth, minTerminalHeight-1) {
		t.Fatal("terminal one row below minimum should be hidden")
	}

	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.width = minTerminalWidth - 1
	m.height = minTerminalHeight
	view := m.View().Content
	if !strings.Contains(view, "Terminal too small.") || !strings.Contains(view, "124") || !strings.Contains(view, "48") {
		t.Fatalf("small-terminal view missing resize message: %q", view)
	}
}

func TestLibraryViewKeepsSingleCellOuterGutter(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	m.width = minTerminalWidth
	m.height = minTerminalHeight

	view := m.View().Content
	if got, want := lipgloss.Width(view), minTerminalWidth; got != want {
		t.Fatalf("view right gutter mismatch: got rendered width %d, want %d", got, want)
	}
	if got, want := lipgloss.Height(view), minTerminalHeight; got != want {
		t.Fatalf("view bottom gutter mismatch: got rendered height %d, want %d", got, want)
	}
	firstLine, _, _ := strings.Cut(view, "\n")
	if strings.TrimSpace(firstLine) != "" {
		t.Fatalf("view should keep one blank row above the launcher: %q", firstLine)
	}
	lines := strings.Split(view, "\n")
	if strings.TrimSpace(lines[len(lines)-1]) != "" {
		t.Fatalf("view should keep one blank row below the launcher: %q", lines[len(lines)-1])
	}
	for _, line := range lines[1 : len(lines)-1] {
		plain := ansiEscape.ReplaceAllString(line, "")
		if len(plain) == 0 || plain[0] != ' ' || plain[len(plain)-1] != ' ' {
			t.Fatalf("view should keep matching left/right gutters: %q", plain)
		}
	}
}

func TestCreditASCIIIsHighInCreditsScreen(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	credits := m.renderCredits(160, 48)
	plain := ansiEscape.ReplaceAllString(credits, "")
	lines := strings.Split(plain, "\n")
	creditLine := -1
	artLine := -1
	signature := creditArtSignature(t)
	for i, line := range lines {
		if creditLine < 0 && strings.Contains(line, "CREDITS") {
			creditLine = i
		}
		if artLine < 0 && strings.Contains(line, signature) {
			artLine = i
		}
	}
	if creditLine < 0 || artLine < 0 {
		t.Fatalf("missing credits heading/current art: credit=%d art=%d signature=%q", creditLine, artLine, signature)
	}
	if artLine-creditLine > 4 {
		t.Fatalf("credit ASCII still sits too low: heading line=%d art line=%d", creditLine, artLine)
	}
}

func TestCreditsUsesCurrentEmbeddedASCIIAsset(t *testing.T) {
	art := creditsBBSArt()
	if strings.TrimSpace(art) == "" {
		t.Fatal("embedded credit artwork is empty")
	}
	if !strings.Contains(art, creditArtSignature(t)) {
		t.Fatal("embedded credit artwork lost its current signature")
	}
}

func TestHeaderUsesBrowseSelectLaunch(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	header := ansiEscape.ReplaceAllString(m.renderHeader(150, 40), "")
	if !strings.Contains(header, "G  A  M  E     M  A  C  H  I  N  E  R  Y  .") || !strings.Contains(header, "BROWSE     /     SELECT     /     LAUNCH") {
		t.Fatalf("header missing updated identity block: %q", header)
	}
	if strings.Contains(header, "PRESERVE") || strings.Contains(header, "EXECUTE") {
		t.Fatalf("header kept obsolete slogan: %q", header)
	}
}

func TestWideCreditsPlaceASCIIBesideCopy(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	credits := ansiEscape.ReplaceAllString(m.renderCredits(160, 48), "")
	lines := strings.Split(credits, "\n")
	signature := creditArtSignature(t)
	for _, line := range lines {
		if strings.Contains(line, "MARCHINE — Game Machinery.") && strings.Contains(line, signature) {
			return
		}
	}
	t.Fatalf("wide CREDITS did not place current credit ASCII beside credit copy")
}

func TestFrontFooterQuitIsLastItem(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := ansiEscape.ReplaceAllString(m.renderFooter(180), "")

	if !strings.HasSuffix(strings.TrimSpace(footer), "Q QUIT") {
		t.Fatalf("quit should be the final front-menu item: %q", footer)
	}
}

func TestWideComponentsUseFullAllocatedWidth(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	const w = 120
	if got := lipgloss.Width(m.renderSearch(w)); got != w {
		t.Fatalf("search right edge drift: got %d want %d", got, w)
	}
	if got := lipgloss.Width(m.renderMain(w, 28)); got != w {
		t.Fatalf("main right edge drift: got %d want %d", got, w)
	}
	if got := lipgloss.Width(m.renderInfo(w, 48)); got != w {
		t.Fatalf("info right edge drift: got %d want %d", got, w)
	}
	if got := lipgloss.Width(m.renderCredits(w, 48)); got != w {
		t.Fatalf("credits right edge drift: got %d want %d", got, w)
	}
}

func TestCompleteArcadeLibraryCanSeparateJunk(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade

	// Hide Non-Arcade ON keeps the complete index in memory but the visible
	// playlist contains only arcade entries.
	if m.activeFilter() != filterArcade {
		t.Fatalf("default demo arcade view should start clean, got %q", m.activeFilter())
	}
	for _, g := range m.filtered() {
		if g.NonArcade {
			t.Fatalf("ARCADE ONLY leaked non-arcade entry %#v", g)
		}
	}

	m.cfg.Arcade.HideJunk = false
	m.filter = ""
	all := m.filtered()
	foundMahjong := false
	foundCasino := false
	for _, g := range all {
		foundMahjong = foundMahjong || g.Category == "MAHJONG"
		foundCasino = foundCasino || g.Category == "CASINO / GAMBLING"
	}
	if !foundMahjong || !foundCasino {
		t.Fatalf("ALL should expose complete indexed library; mahjong=%v casino=%v", foundMahjong, foundCasino)
	}

	m.filter = filterNonArcade
	nonArcade := m.filtered()
	if len(nonArcade) < 2 {
		t.Fatalf("NON - ARCADE should expose classified noise, got %d", len(nonArcade))
	}
	for _, g := range nonArcade {
		if !g.NonArcade {
			t.Fatalf("non-arcade filter leaked normal arcade entry %#v", g)
		}
	}
}

func TestHideNonArcadePreservesCategoryAndLiveCount(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.source = library.SourceArcade
	m.filter = "SHMUPS"
	m.width, m.height = 140, 52

	// Include a non-arcade entry in the same category to ensure H changes
	// the contents of that category rather than switching to the full list.
	m.arcade = append(m.arcade, library.Game{
		ID: "mame:shmup-demo-nonarcade", Name: "Non-arcade Shmup",
		Source: library.SourceArcade, Category: "SHMUPS", NonArcade: true,
	})
	arcadeOnly := len(m.filtered())
	if arcadeOnly == 0 {
		t.Fatal("test requires arcade shmups")
	}

	for _, want := range []struct {
		action tea.KeyPressMsg
		hidden bool
		count  int
	}{
		{textKey("h"), false, arcadeOnly + 1},
		{textKey("h"), true, arcadeOnly},
	} {
		next, _ := m.Update(want.action)
		m = next.(Model)
		if m.screen != screenLibrary || m.filter != "SHMUPS" || m.cfg.Arcade.HideJunk != want.hidden {
			t.Fatalf("H lost category: screen=%v filter=%q hidden=%v", m.screen, m.filter, m.cfg.Arcade.HideJunk)
		}
		if got := len(m.filtered()); got != want.count {
			t.Fatalf("SHMUPS count = %d; want %d", got, want.count)
		}
		for _, g := range m.filtered() {
			if g.Category != "SHMUPS" || (want.hidden && g.NonArcade) {
				t.Fatalf("SHMUPS filter leaked item: %#v", g)
			}
		}
		if suffix := groupedCount(want.count) + " MATCHES"; !strings.HasSuffix(strings.TrimRight(ansiEscape.ReplaceAllString(m.renderSourceTabs(122), ""), " "), suffix) {
			t.Fatalf("source bar did not update to %q", suffix)
		}
	}

	// Controller LB and Info Quick Settings use the same toggle handler.
	next, _ := m.Update(GamepadMsg{Action: gamepad.ActionHideNonArcade})
	m = next.(Model)
	if m.filter != "SHMUPS" || m.cfg.Arcade.HideJunk {
		t.Fatal("controller LB should show non-arcade entries without leaving SHMUPS")
	}
	m.openInfo()
	next, _ = m.Update(specialKey(tea.KeyEnter))
	m = next.(Model)
	if m.filter != "SHMUPS" || !m.cfg.Arcade.HideJunk {
		t.Fatal("Info quick setting should preserve SHMUPS category")
	}
}

// Hide Non-Arcade must not throw the cursor back to game one. Selection is
// tracked by game ID because the list indexes change when entries are hidden.
func TestHideNonArcadeKeepsSelectionAndFindsNearestSurvivor(t *testing.T) {
	games := []library.Game{
		{ID: "mame:a", Name: "A", Category: "SHMUPS", Source: library.SourceArcade},
		{ID: "mame:b", Name: "B", Category: "SHMUPS", Source: library.SourceArcade, NonArcade: true},
		{ID: "mame:c", Name: "C", Category: "SHMUPS", Source: library.SourceArcade},
		{ID: "mame:d", Name: "D", Category: "SHMUPS", Source: library.SourceArcade, NonArcade: true},
		{ID: "mame:e", Name: "E", Category: "SHMUPS", Source: library.SourceArcade},
		{ID: "mame:f", Name: "F", Category: "SHMUPS", Source: library.SourceArcade, NonArcade: true},
	}

	makeModel := func(cursor int) Model {
		m := NewModel(config.Default(), "", false, true)
		m.source = library.SourceArcade
		m.filter = "SHMUPS"
		m.cfg.Arcade.HideJunk = false
		m.arcade = append([]library.Game(nil), games...)
		m.cursor = cursor
		return m
	}

	for _, tc := range []struct {
		name       string
		start      int
		wantID     string
		wantCursor int
	}{
		{"surviving selection moves with its ID", 2, "mame:c", 1},
		{"hidden selection advances to next game", 1, "mame:c", 1},
		{"hidden selection prefers next over previous", 3, "mame:e", 2},
		{"last hidden selection falls back to previous", 5, "mame:e", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := makeModel(tc.start)
			next, _ := m.Update(textKey("h"))
			m = next.(Model)
			if m.filter != "SHMUPS" || !m.cfg.Arcade.HideJunk {
				t.Fatalf("H changed category or did not enable hiding: filter=%q hidden=%v", m.filter, m.cfg.Arcade.HideJunk)
			}
			g, ok := m.selected()
			if !ok || g.ID != tc.wantID || m.cursor != tc.wantCursor {
				t.Fatalf("selection after H = %#v (ok=%v cursor=%d); want %q at %d", g, ok, m.cursor, tc.wantID, tc.wantCursor)
			}
		})
	}

	// Turning Hide Non-Arcade back OFF must still point at the same game,
	// even though earlier non-arcade entries are inserted back into the list.
	m := makeModel(4) // E
	m.toggleHideNonArcade()
	m.toggleHideNonArcade()
	g, ok := m.selected()
	if !ok || g.ID != "mame:e" || m.cursor != 4 || m.cfg.Arcade.HideJunk {
		t.Fatalf("selection not preserved after round trip: %#v, cursor=%d", g, m.cursor)
	}

	// Keyboard, controller, and Info Quick Settings share the same selection logic.
	for _, tc := range []struct {
		name   string
		toggle func(Model) Model
	}{
		{"controller LB", func(m Model) Model {
			next, _ := m.Update(GamepadMsg{Action: gamepad.ActionHideNonArcade})
			return next.(Model)
		}},
		{"Info Quick Settings", func(m Model) Model {
			m.openInfo()
			next, _ := m.Update(specialKey(tea.KeyEnter))
			return next.(Model)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.toggle(makeModel(4))
			g, ok := m.selected()
			if !ok || g.ID != "mame:e" || m.cursor != 2 {
				t.Fatalf("%s lost the selected game: %#v (cursor=%d)", tc.name, g, m.cursor)
			}
		})
	}

	// If the filter becomes empty, the cursor must remain a valid zero.
	m = makeModel(1)
	m.arcade = []library.Game{games[1], games[3]}
	m.cursor = 1
	m.toggleHideNonArcade()
	if m.cursor != 0 || len(m.filtered()) != 0 {
		t.Fatalf("empty filtered list left cursor=%d, count=%d", m.cursor, len(m.filtered()))
	}
}

func TestHideNonArcadePreservesVerticalAndHandlesSpecialFilters(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.source = library.SourceArcade
	m.filter = filterVertical
	m.toggleHideNonArcade()
	if m.filter != filterVertical || m.cfg.Arcade.HideJunk {
		t.Fatal("Hide Non-Arcade OFF should preserve VERTICAL")
	}
	m.toggleHideNonArcade()
	if m.filter != filterVertical || !m.cfg.Arcade.HideJunk {
		t.Fatal("Hide Non-Arcade ON should preserve VERTICAL")
	}

	// The special NON - ARCADE filter cannot coexist with hiding non-arcade
	// entries. Continue showing the broad Arcade Only view instead.
	m.cfg.Arcade.HideJunk = false
	m.filter = filterNonArcade
	m.toggleHideNonArcade()
	if m.filter != filterArcade || !m.cfg.Arcade.HideJunk {
		t.Fatalf("NON - ARCADE + hide should become ARCADE ONLY, got %q", m.filter)
	}
}

func TestCategoryMenuIncludesComprehensiveSpecialFilters(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade

	clean := strings.Join(m.filters(), "|")
	for _, want := range []string{"ARCADE ONLY", "SHMUPS", "FIGHTING", "PUZZLE"} {
		if !strings.Contains(clean, want) {
			t.Fatalf("clean category filters missing %q: %s", want, clean)
		}
	}
	for _, hidden := range []string{"ALL", "MAHJONG", "CASINO / GAMBLING", "NON - ARCADE"} {
		if strings.Contains(clean, hidden) {
			t.Fatalf("Hide Non-Arcade ON should remove %q from categories: %s", hidden, clean)
		}
	}

	m.cfg.Arcade.HideJunk = false
	full := strings.Join(m.filters(), "|")
	for _, want := range []string{"ALL", "ARCADE ONLY", "SHMUPS", "FIGHTING", "PUZZLE", "MAHJONG", "CASINO / GAMBLING", "NON - ARCADE"} {
		if !strings.Contains(full, want) {
			t.Fatalf("full category filters missing %q: %s", want, full)
		}
	}
}

func TestInfoSingleKeysAndHideNonArcadeToggle(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	m.width = 140
	m.height = 52

	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'i', Text: "i"}))
	m = next.(Model)
	if m.screen != screenInfo {
		t.Fatalf("I should open Info, screen=%v", m.screen)
	}

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'h', Text: "h"}))
	m = next.(Model)
	if m.cfg.Arcade.HideJunk {
		t.Fatal("H should turn Hide Non-Arcade OFF")
	}
	if m.activeFilter() != filterAll {
		t.Fatalf("toggle OFF should expose ALL, got %q", m.activeFilter())
	}

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'o', Text: "o"}))
	m = next.(Model)
	if m.screen != screenCredits {
		t.Fatalf("O should open Credits, screen=%v", m.screen)
	}
}

func TestAllScreensRemainUsableAtMinimumGeometry(t *testing.T) {
	cases := []struct {
		name string
		mode screenMode
		want string
	}{
		{"library", screenLibrary, "Q QUIT"},
		{"info", screenInfo, "O CREDITS"},
		{"credits", screenCredits, "I INFO"},
		{"categories", screenCategories, "ENTER APPLY"},
		{"setup", screenSetup, "ESC RETURN"},
	}
	for _, tc := range cases {
		m := NewModel(config.Default(), "", false, true)
		m.width = minTerminalWidth
		m.height = minTerminalHeight
		m.screen = tc.mode
		plain := ansiEscape.ReplaceAllString(m.View().Content, "")
		if !strings.Contains(plain, tc.want) {
			t.Fatalf("%s minimum layout lost %q", tc.name, tc.want)
		}
	}
}

func TestHeaderKeepsFullIdentityAtMinimumWidth(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	header := ansiEscape.ReplaceAllString(m.renderHeader(minTerminalWidth-2, minTerminalHeight-2), "")
	if !strings.Contains(header, "G  A  M  E") || !strings.Contains(header, "BROWSE") {
		t.Fatalf("minimum supported width fell back to compact header: %q", header)
	}
}

func TestFooterBaselineIsIdenticalAcrossScreens(t *testing.T) {
	cfg := config.Default()
	modes := []screenMode{screenLibrary, screenInfo, screenCredits, screenCategories, screenSetup}
	var baseline int = -1
	for _, mode := range modes {
		m := NewModel(cfg, "", false, true)
		m.source = library.SourceArcade
		m.width = 140
		m.height = 52
		m.screen = mode
		plain := ansiEscape.ReplaceAllString(m.View().Content, "")
		lines := strings.Split(plain, "\n")
		lastNonBlank := -1
		for i := len(lines) - 2; i >= 1; i-- { // preserve the one-cell outer gutter
			if strings.TrimSpace(lines[i]) != "" {
				lastNonBlank = i
				break
			}
		}
		if lastNonBlank < 0 {
			t.Fatalf("screen %v rendered no content", mode)
		}
		if baseline < 0 {
			baseline = lastNonBlank
		} else if lastNonBlank != baseline {
			t.Fatalf("screen %v footer baseline=%d want=%d", mode, lastNonBlank, baseline)
		}
	}
}

func TestMainPanelsUseTheirFullInteriorWidth(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	list := ansiEscape.ReplaceAllString(m.renderList(60, 24), "")
	lines := strings.Split(list, "\n")
	if len(lines) < 3 {
		t.Fatal("list panel unexpectedly short")
	}
	// The header separator should now reach directly from border to border;
	// older builds left four invisible cells on the right of every row.
	if !strings.HasSuffix(lines[2], "│") {
		t.Fatalf("list row did not reach right border: %q", lines[2])
	}
}

func TestRefreshIsSingleFlight(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, false)
	m.width = minTerminalWidth
	m.height = minTerminalHeight
	m.loading = false

	next, firstCmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'r', Text: "r"}))
	m = next.(Model)
	if firstCmd == nil || !m.loading {
		t.Fatal("first refresh should start exactly one scan command")
	}

	next, secondCmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'r', Text: "r"}))
	m = next.(Model)
	if secondCmd != nil {
		t.Fatal("second refresh while loading should not start another scan")
	}
	if m.status != "ARCADE REFRESH ALREADY IN PROGRESS" {
		t.Fatalf("unexpected single-flight status %q", m.status)
	}
}

func TestSetupPreservesAdditionalROMPaths(t *testing.T) {
	d := t.TempDir()
	path := d + "/config.toml"
	cfg := config.Default()
	cfg.Arcade.ROMPaths = []string{"/roms/primary", "/roms/secondary", "/roms/third"}

	m := NewModel(cfg, path, false, false)
	m.loading = false
	m.setupROM = "/roms/replacement"
	next, cmd := m.saveSetup()
	if cmd != nil {
		t.Fatal("save-only Setup should not start a command")
	}
	m = next.(Model)

	want := []string{"/roms/replacement", "/roms/secondary", "/roms/third"}
	if len(m.cfg.Arcade.ROMPaths) != len(want) {
		t.Fatalf("ROM paths=%#v want=%#v", m.cfg.Arcade.ROMPaths, want)
	}
	for i := range want {
		if m.cfg.Arcade.ROMPaths[i] != want[i] {
			t.Fatalf("ROM paths=%#v want=%#v", m.cfg.Arcade.ROMPaths, want)
		}
	}

	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if saved.Arcade.ROMPaths[i] != want[i] {
			t.Fatalf("saved ROM paths=%#v want=%#v", saved.Arcade.ROMPaths, want)
		}
	}
}

func TestSetupEmptyROMPathRestoresAutoDetect(t *testing.T) {
	d := t.TempDir()
	path := d + "/config.toml"
	cfg := config.Default()
	cfg.Arcade.ROMPaths = []string{"/roms/a", "/roms/b"}

	m := NewModel(cfg, path, false, false)
	m.loading = false
	m.setupROM = ""
	next, _ := m.saveSetup()
	m = next.(Model)
	if len(m.cfg.Arcade.ROMPaths) != 0 {
		t.Fatalf("empty primary path should restore auto-detect, got %#v", m.cfg.Arcade.ROMPaths)
	}
}

func TestTrimAndPadUseTerminalCellWidth(t *testing.T) {
	wide := "東京🎮 Arcade"
	trimmed := trim(wide, 7)
	if got := lipgloss.Width(trimmed); got > 7 {
		t.Fatalf("trimmed Unicode width=%d want<=7: %q", got, trimmed)
	}

	padded := pad("東京", 8)
	if got := lipgloss.Width(padded); got != 8 {
		t.Fatalf("padded Unicode width=%d want=8: %q", got, padded)
	}
}

func TestSetupNormalizesRuntimeLaunchPathsImmediately(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := config.Default()
	m := NewModel(cfg, path, false, false)
	m.loading = false
	m.arcade = []library.Game{{
		ID:     "mame:1942",
		Name:   "1942",
		Source: library.SourceArcade,
		ROM:    "1942",
	}}
	m.setupMAME = "~/bin/mame"
	m.setupROM = "$HOME/roms"

	next, cmd := m.saveSetup()
	if cmd != nil {
		t.Fatal("save-only Setup should not start a command")
	}
	m = next.(Model)

	wantMAME := filepath.Join(home, "bin", "mame")
	wantROM := filepath.Join(home, "roms")
	if m.cfg.Arcade.MAMECommand != wantMAME {
		t.Fatalf("runtime MAME command=%q want=%q", m.cfg.Arcade.MAMECommand, wantMAME)
	}
	if len(m.cfg.Arcade.ROMPaths) != 1 || m.cfg.Arcade.ROMPaths[0] != wantROM {
		t.Fatalf("runtime ROM paths=%#v want=%q", m.cfg.Arcade.ROMPaths, wantROM)
	}
	if got := m.arcade[0].Command; got != wantMAME {
		t.Fatalf("in-memory Arcade command=%q want=%q", got, wantMAME)
	}
	wantArgs := []string{"1942", "-rompath", wantROM}
	if len(m.arcade[0].Args) != len(wantArgs) {
		t.Fatalf("in-memory Arcade args=%#v want=%#v", m.arcade[0].Args, wantArgs)
	}
	for i := range wantArgs {
		if m.arcade[0].Args[i] != wantArgs[i] {
			t.Fatalf("in-memory Arcade args=%#v want=%#v", m.arcade[0].Args, wantArgs)
		}
	}
}

func TestInfoQuickSettingKeyboardAndGamepad(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	m.width = 140
	m.height = 52

	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'i', Text: "i"}))
	m = next.(Model)

	if m.screen != screenInfo {
		t.Fatalf("I should open Info, screen=%v", m.screen)
	}

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)

	if m.cfg.Arcade.HideJunk {
		t.Fatal("Enter should turn Hide Non-Arcade OFF")
	}

	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionConfirm})
	m = next.(Model)

	if !m.cfg.Arcade.HideJunk {
		t.Fatal("controller A should turn Hide Non-Arcade ON")
	}

	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionBack})
	m = next.(Model)

	if m.screen != screenLibrary {
		t.Fatalf("controller B should return from Info, screen=%v", m.screen)
	}
}

func TestHideNonArcadeQuickSettingPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := config.Default()
	cfg.Arcade.Enabled = false

	m := NewModel(cfg, path, false, false)
	m.width = 140
	m.height = 52
	m.openInfo()

	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)

	if m.cfg.Arcade.HideJunk {
		t.Fatal("Hide Non-Arcade should be OFF in memory")
	}

	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Arcade.HideJunk {
		t.Fatal("Hide Non-Arcade OFF was not persisted")
	}
}

func TestQuitIsAvailableFromSetupAndController(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width = minTerminalWidth
	m.height = minTerminalHeight
	m.openSetup()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if cmd == nil {
		t.Fatal("Q should quit from Setup when not editing a field")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("Q did not return a Bubble Tea quit message")
	}

	m = NewModel(config.Default(), "", false, true)
	m.width = minTerminalWidth - 1
	m.height = minTerminalHeight

	_, cmd = m.Update(GamepadMsg{Action: gamepad.ActionQuit})
	if cmd == nil {
		t.Fatal("controller quit chord should work even below minimum geometry")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("controller quit action did not return a Bubble Tea quit message")
	}
}

func TestInfoDocumentsQuitControls(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	info := ansiEscape.ReplaceAllString(m.renderInfo(140, 52), "")
	normalized := strings.Join(strings.Fields(info), " ")

	for _, want := range []string{
		"Q Quit",
		"X Info",
		"View + Menu Quit",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("Info screen missing control %q", want)
		}
	}

	if strings.Contains(normalized, "X / View") {
		t.Fatalf("Info screen still documents redundant View-to-Info mapping: %q", info)
	}
}

func TestViewChangesSourceAndMenuOpensCredits(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = minTerminalWidth, minTerminalHeight
	original := m.source
	next, _ := m.Update(GamepadMsg{Action: gamepad.ActionView})
	m = next.(Model)
	if m.source == original {
		t.Fatal("standalone View should switch Games / Arcade")
	}
	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionMenu})
	m = next.(Model)
	if m.screen != screenCredits {
		t.Fatal("standalone Menu should open Credits")
	}
}

func TestInfoThemeSettingAndShoulderShortcuts(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 140, 52
	m.openInfo()
	if m.infoSetting != infoSettingHideNonArcade {
		t.Fatal("Info should initially select Hide Non-Arcade")
	}
	next, _ := m.Update(GamepadMsg{Action: gamepad.ActionDown})
	m = next.(Model)
	if m.infoSetting != infoSettingTheme {
		t.Fatal("Down should select Theme")
	}
	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionConfirm})
	m = next.(Model)
	if m.theme.Name != "cyan" || ThemeDisplayName(m.theme.Name) != "PLASMA" {
		t.Fatalf("Theme should cycle to PLASMA; got %q", m.theme.Name)
	}
	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionHideNonArcade})
	m = next.(Model)
	if m.cfg.Arcade.HideJunk {
		t.Fatal("LB shortcut should toggle Hide Non-Arcade OFF")
	}
	next, _ = m.Update(GamepadMsg{Action: gamepad.ActionTheme})
	m = next.(Model)
	if m.theme.Name != "green" {
		t.Fatal("RB should cycle Theme independently of Info selection")
	}
}

func TestKeyboardIndexConfirmAndCancel(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 140, 52
	m.source = library.SourceArcade
	original := m.cursor
	next, cmd := m.Update(textKey("]"))
	m = next.(Model)
	if cmd != nil || m.letterJump == "" {
		t.Fatal("keyboard ] should open letter preview without launching")
	}
	next, cmd = m.Update(specialKey(tea.KeyEscape))
	m = next.(Model)
	if cmd != nil || m.letterJump != "" || m.cursor != original {
		t.Fatal("keyboard Escape should cancel preview and restore cursor")
	}
	next, _ = m.Update(textKey("["))
	m = next.(Model)
	if m.letterJump == "" {
		t.Fatal("keyboard [ should open letter preview")
	}
	next, cmd = m.Update(specialKey(tea.KeyEnter))
	m = next.(Model)
	if cmd != nil || m.letterJump != "" {
		t.Fatal("keyboard Enter should confirm preview, not launch game")
	}
}

func TestThemeDisplayNamesPreserveConfigIdentifiers(t *testing.T) {
	cases := map[string]string{"violet": "MARCHINE", "cyan": "PLASMA", "green": "PHOSPHOR", "amber": "EMBER"}
	for stored, label := range cases {
		if got := ThemeDisplayName(stored); got != label {
			t.Fatalf("theme %s label=%s want=%s", stored, got, label)
		}
	}
}

func TestOverlayBackReturnsToOrigin(t *testing.T) {
	for _, origin := range []screenMode{screenLibrary, screenCategories, screenSetup, screenInfo} {
		m := NewModel(config.Default(), "", false, true)
		m.width, m.height = 140, 52
		m.screen = origin
		m.openCredits()
		if m.screen != screenCredits {
			t.Fatalf("origin %v: Credits did not open", origin)
		}
		next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
		m = next.(Model)
		if m.screen != origin {
			t.Fatalf("origin %v: Esc from Credits returned to %v", origin, m.screen)
		}
	}
}

func TestInfoCreditsNavigationUnwindsWithoutLoops(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	m.width, m.height = 140, 52
	pressEsc := func() {
		next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
		m = next.(Model)
	}

	m.openInfo()
	m.infoSetting = infoSettingTheme
	m.openCredits()
	m.openInfo() // Returning to the previous Info should not add a history entry.
	if m.screen != screenInfo || m.infoSetting != infoSettingTheme || len(m.overlayHistory) != 1 {
		t.Fatalf("Credits -> Info did not restore selected setting and history: %+v", m.overlayHistory)
	}
	m.openCredits()
	pressEsc()
	if m.screen != screenInfo {
		t.Fatalf("Esc from Credits should return to Info, got %v", m.screen)
	}
	pressEsc()
	if m.screen != screenLibrary || len(m.overlayHistory) != 0 {
		t.Fatalf("Esc from Info should return to Library and clear history: screen=%v history=%v", m.screen, m.overlayHistory)
	}

	m.openCredits()
	m.openInfo()
	pressEsc()
	if m.screen != screenCredits {
		t.Fatalf("Esc from Info should return to Credits, got %v", m.screen)
	}
	pressEsc()
	if m.screen != screenLibrary || len(m.overlayHistory) != 0 {
		t.Fatalf("Esc from Credits should return to Library and clear history: screen=%v history=%v", m.screen, m.overlayHistory)
	}
}

func TestInfoCreditsFootersOnlyShowContextualControls(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	info := ansiEscape.ReplaceAllString(m.renderInfoFooter(122), "")
	for _, want := range []string{"↑ / ↓ SELECT", "ENTER APPLY", "ESC RETURN", "O CREDITS"} {
		if !strings.Contains(info, want) {
			t.Fatalf("Info footer missing %q: %q", want, info)
		}
	}
	if strings.Contains(info, "THEME") || strings.Contains(info, "HIDE NON-ARCADE") {
		t.Fatalf("Info footer duplicates Quick Settings: %q", info)
	}

	credits := ansiEscape.ReplaceAllString(m.renderCreditsFooter(122), "")
	for _, want := range []string{"ESC RETURN", "I INFO"} {
		if !strings.Contains(credits, want) {
			t.Fatalf("Credits footer missing %q: %q", want, credits)
		}
	}
	if strings.Contains(credits, "THEME") {
		t.Fatalf("Credits footer unexpectedly advertises Theme: %q", credits)
	}
}

// Rotation is an optional external utility, not a Marchine or MAME dependency.
func TestControllerRotationMissingUtilityIsNonfatal(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	m := NewModel(config.Default(), "", false, true)
	updated, cmd := m.updateGamepad(gamepad.ActionRotateDisplay)
	if cmd == nil {
		t.Fatal("rotation must return a non-blocking command")
	}
	result, ok := cmd().(displayRotateFinishedMsg)
	if !ok || !errors.Is(result.Err, exec.ErrNotFound) {
		t.Fatalf("missing utility: %#v", result)
	}
	updated, followup := updated.(Model).Update(result)
	if followup != nil || updated.(Model).status != "DISPLAY ROTATE NOT INSTALLED" {
		t.Fatalf("missing utility should set status without quitting: %#v", updated)
	}
}

func TestControllerRotationExecutesInstalledUtility(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "called")
	script := filepath.Join(dir, "hypr-rotate")
	// No shell interpolation: a tiny executable under the isolated test PATH.
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf x > "+marker+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	m := NewModel(config.Default(), "", false, true)
	updated, cmd := m.updateGamepad(gamepad.ActionRotateDisplay)
	if cmd == nil {
		t.Fatal("rotation must return a command")
	}
	result, ok := cmd().(displayRotateFinishedMsg)
	if !ok || result.Err != nil {
		t.Fatalf("rotation command: %#v", result)
	}
	updated, followup := updated.(Model).Update(result)
	if followup != nil || updated.(Model).status != "DISPLAY ROTATED" {
		t.Fatalf("successful rotation status: %#v", updated)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("hypr-rotate was not invoked: %v", err)
	}
}

// Ctrl+C is not a Marchine quit shortcut. Foot provides terminal clipboard
// shortcuts; Ctrl+C should be harmless in every screen and editing mode.
func TestCtrlCNeverQuitsMarchine(t *testing.T) {
	ctrlC := tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})
	cases := []struct {
		name    string
		prepare func(*Model)
	}{
		{name: "library"},
		{name: "library search", prepare: func(m *Model) {
			m.searchMode = true
			m.search = "unfinished"
		}},
		{name: "letter jump", prepare: func(m *Model) {
			m.letterJump = "A"
		}},
		{name: "small terminal", prepare: func(m *Model) {
			m.width = minTerminalWidth - 1
			m.height = minTerminalHeight
		}},
		{name: "categories", prepare: func(m *Model) {
			m.screen = screenCategories
		}},
		{name: "info", prepare: func(m *Model) {
			m.screen = screenInfo
		}},
		{name: "credits", prepare: func(m *Model) {
			m.screen = screenCredits
		}},
		{name: "arcade setup", prepare: func(m *Model) {
			m.openSetup()
		}},
		{name: "arcade field", prepare: func(m *Model) {
			m.openSetup()
			m.setupEditing = true
			m.setupBuffer = "/tmp/unfinished"
		}},
		{name: "game picker", prepare: func(m *Model) {
			m.screen = screenGamePicker
			m.cfg.Games = []config.GameConfig{{Name: "Test", Command: "true"}}
		}},
		{name: "game removal confirmation", prepare: func(m *Model) {
			m.screen = screenGamePicker
			m.gameDeletePending = true
			m.cfg.Games = []config.GameConfig{{Name: "Test", Command: "true"}}
		}},
		{name: "game editor", prepare: func(m *Model) {
			m.openGameEditor(-1)
		}},
		{name: "game field", prepare: func(m *Model) {
			m.openGameEditor(-1)
			m.gameEditing = true
			m.gameBuffer = "/tmp/unfinished"
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(config.Default(), "", false, true)
			m.width = minTerminalWidth
			m.height = minTerminalHeight
			if tc.prepare != nil {
				tc.prepare(&m)
			}
			before := m
			next, cmd := m.Update(ctrlC)
			after := next.(Model)
			if cmd != nil {
				t.Fatal("Ctrl+C unexpectedly produced a command")
			}
			if after.screen != before.screen || after.setupEditing != before.setupEditing ||
				after.gameEditing != before.gameEditing || after.setupBuffer != before.setupBuffer ||
				after.gameBuffer != before.gameBuffer || after.search != before.search ||
				after.letterJump != before.letterJump || after.gameDeletePending != before.gameDeletePending {
				t.Fatal("Ctrl+C changed screen, in-progress text, or selection")
			}
		})
	}
}
