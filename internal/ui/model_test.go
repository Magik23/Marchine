package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/magik23/marchine/internal/config"
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

func TestCompactFooterKeepsInfoEscapeHatch(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := m.renderFooter(46)
	if !strings.Contains(footer, "I") || !strings.Contains(footer, "INFO") {
		t.Fatalf("compact footer lost Info shortcut: %q", footer)
	}
	if strings.Contains(footer, "THEME") || strings.Contains(footer, "REINDEX") {
		t.Fatalf("compact footer kept secondary controls instead of collapsing: %q", footer)
	}
}

func TestInfoAndCreditsAreSeparateScreens(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	info := m.renderInfo(120, 36)
	for _, want := range []string{"HOW IT WORKS", "SHORTCUTS", "EMULATOR MODEL", "O", "Credits", "ARCADE FILTER", "HIDE NON-ARCADE"} {
		if !strings.Contains(info, want) {
			t.Fatalf("Info screen missing %q", want)
		}
	}
	for _, forbidden := range []string{"Pierre Dionne", "Albenoir Studio", "CREDITS\n"} {
		if strings.Contains(info, forbidden) {
			t.Fatalf("Info screen still contains credit material %q", forbidden)
		}
	}
	credits := m.renderCredits(160, 48)
	for _, want := range []string{"CREDITS", "Pierre Dionne", "Albenoir Studio", "MARCHINE — Game Machinery."} {
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

func TestFrontFooterOmitsSecondaryCommands(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := m.renderFooter(180)
	for _, forbidden := range []string{"NAV", "REINDEX", "ROTATE", "THEME", "QUIT", "SETUP", "FILTER", "MARCHINE // GAME MACHINERY"} {
		if strings.Contains(footer, forbidden) {
			t.Fatalf("front footer exposed %q: %q", forbidden, footer)
		}
	}
	for _, want := range []string{"SEARCH", "SOURCE", "CATEGORIES", "PLAY", "INFO", "CREDITS"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("front footer missing %q: %q", want, footer)
		}
	}
}

func TestTopNavigationOnlyShowsSources(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	tabs := m.renderSourceTabs(120)
	for _, want := range []string{"GAMES", "ARCADE"} {
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

func TestFrontFooterCreditsIsLastItem(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "", false, true)
	m.source = library.SourceArcade
	footer := ansiEscape.ReplaceAllString(m.renderFooter(180), "")
	if !strings.HasSuffix(strings.TrimSpace(footer), "O CREDITS") {
		t.Fatalf("credits should be the final front-menu item: %q", footer)
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
		{"library", screenLibrary, "O CREDITS"},
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
	next, cmd := m.saveSetup(false)
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
	next, _ := m.saveSetup(false)
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

	next, cmd := m.saveSetup(false)
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
