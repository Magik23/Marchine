package ui

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

func TestGameArgumentsQuotesAndRoundTrip(t *testing.T) {
	input := `-window "My Game.exe" 'other file' "" /tmp/file`
	want := []string{"-window", "My Game.exe", "other file", "", "/tmp/file"}
	got, err := parseGameArguments(input)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("parseGameArguments = %#v, %v; want %#v", got, err, want)
	}
	roundTrip, err := parseGameArguments(formatGameArguments(want))
	if err != nil || !reflect.DeepEqual(roundTrip, want) {
		t.Fatalf("argument round trip = %#v, %v; want %#v", roundTrip, err, want)
	}
	for _, input := range []string{`"unterminated`, `'unterminated`, `trailing\`} {
		if _, err := parseGameArguments(input); err == nil {
			t.Fatalf("unbalanced arguments %q were accepted", input)
		}
	}
}

func TestSetupArcadeFirstAndGamesActions(t *testing.T) {
	m := NewModel(config.Default(), "", false, false)
	ui := m.renderSetup(122, 35)
	arcade := strings.Index(ui, "ARCADE SETUP")
	games := strings.Index(ui, "GAMES")
	if arcade < 0 || games < arcade {
		t.Fatalf("Setup should have ARCADE before GAMES")
	}
	for _, want := range []string{"ADD GAME", "EDIT GAME", "REMOVE GAME"} {
		if !strings.Contains(ui, want) {
			t.Fatalf("Setup missing %q", want)
		}
	}
	m.openSetup()
	m.setupField = setupFieldAddGame
	next, _ := m.updateSetup(specialKey(tea.KeyEnter))
	if next.(Model).screen != screenGameEditor {
		t.Fatal("Add Game did not open the editor")
	}
}

func TestSetupRefreshNeverSavesPendingArcadeEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := config.Default()
	cfg.Arcade.ROMPaths = []string{"/roms/saved"}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	m := NewModel(cfg, path, false, false)
	m.loading = false
	m.openSetup()
	m.setupField = setupFieldROM
	m.setupROM = "/roms/unsaved"
	m.setupMAME = "/tmp/unsaved-mame"

	next, cmd := m.updateSetup(textKey("r"))
	m = next.(Model)
	if cmd == nil || !m.loading {
		t.Fatal("R should start Arcade refresh without saving")
	}
	if m.cfg.Arcade.MAMECommand != cfg.Arcade.MAMECommand ||
		!reflect.DeepEqual(m.cfg.Arcade.ROMPaths, cfg.Arcade.ROMPaths) {
		t.Fatal("R applied pending Arcade edits to running config")
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Arcade.MAMECommand != cfg.Arcade.MAMECommand ||
		!reflect.DeepEqual(saved.Arcade.ROMPaths, cfg.Arcade.ROMPaths) {
		t.Fatal("R changed saved Arcade settings")
	}

	// The Games section has no refresh action: games appear on save.
	m.loading = false
	m.setupField = setupFieldAddGame
	next, cmd = m.updateSetup(textKey("r"))
	if cmd != nil || next.(Model).loading {
		t.Fatal("R in Games Setup should not start Arcade refresh")
	}
}

func TestSetupAndRemovalShowConciseInstructions(t *testing.T) {
	m := NewModel(config.Default(), "", false, true)
	setup := m.renderSetup(122, 35)
	for _, want := range []string{"ARCADE SETUP", "GAMES SETUP", "Ctrl+S saves Arcade settings", "R refreshes Arcade", "Ctrl+S saves a game in Add/Edit"} {
		if !strings.Contains(setup, want) {
			t.Fatalf("Setup missing %q", want)
		}
	}
	for _, forbidden := range []string{"Enter edits Arcade paths", "R saves and refreshes"} {
		if strings.Contains(setup, forbidden) {
			t.Fatalf("obsolete Setup hint %q", forbidden)
		}
	}
	m.cfg.Games = []config.GameConfig{{Name: "Example", Command: "/bin/true"}}
	m.openGamePicker(setupFieldRemoveGame)
	m.gameDeletePending = true
	confirm := m.renderGamePicker(122, 35)
	for _, want := range []string{"REMOVE 'Example' FROM MARCHINE?", "Y YES", "N NO"} {
		if !strings.Contains(confirm, want) {
			t.Fatalf("Remove confirmation missing %q", want)
		}
	}
	if strings.Contains(confirm, "Files remain untouched") {
		t.Fatal("Remove confirmation contains redundant file disclaimer")
	}
}

func TestAddEditRemoveConfiguredGame(t *testing.T) {
	path := filepath.Join(t.TempDir(), "marchine", "config.toml")
	cfg := config.Default()
	cfg.Arcade.ROMPaths = []string{"/roms/one", "/roms/two"}
	cfg.Games = []config.GameConfig{{Name: "Zebra", Command: "zebra"}}
	m := NewModel(cfg, path, false, false)
	m.loading = false
	m.source = library.SourceGames

	// Add and persist a game with an executable path and quoted argument.
	m.openGameEditor(-1)
	m.gameDraft = config.GameConfig{
		Name:       "Alpha",
		Command:    "/tmp/alpha.x86_64",
		Args:       []string{"--path", "/tmp/with spaces"},
		WorkingDir: "/tmp",
	}
	m.saveConfiguredGame()
	if m.screen != screenSetup || len(m.games) != 2 || m.games[0].Name != "Alpha" {
		t.Fatalf("new game did not show immediately: %#v", m.games)
	}
	loaded, err := config.Load(path)
	if err != nil || len(loaded.Games) != 2 || loaded.Games[1].Name != "Alpha" {
		t.Fatalf("added game not persisted: %#v, %v", loaded.Games, err)
	}
	if !reflect.DeepEqual(loaded.Arcade.ROMPaths, cfg.Arcade.ROMPaths) {
		t.Fatalf("game save altered Arcade paths: %#v", loaded.Arcade.ROMPaths)
	}

	// Picker displays sorted names, but preserves the original config index.
	m.openGamePicker(setupFieldEditGame)
	rows := m.configuredGames()
	if rows[0].name != "Alpha" || rows[0].index != 1 {
		t.Fatalf("picker lost config index: %#v", rows)
	}
	m.gamePickerCursor = 0
	next, _ := m.updateGamePicker(specialKey(tea.KeyEnter))
	m = next.(Model)
	if m.screen != screenGameEditor || m.gameIndex != 1 {
		t.Fatalf("editing wrong game: screen=%d index=%d", m.screen, m.gameIndex)
	}
	m.gameDraft.Name = "Beta"
	m.gameDraft.Command = "wine"
	m.gameDraft.Args = []string{"/tmp/game.exe"}
	m.saveConfiguredGame()
	if len(m.games) != 2 || m.games[0].Name != "Beta" || m.games[0].Command != "wine" {
		t.Fatalf("edit didn't refresh library: %#v", m.games)
	}

	// Removal is non-destructive and requires explicit Y confirmation.
	m.openGamePicker(setupFieldRemoveGame)
	m.gamePickerCursor = 0
	next, _ = m.updateGamePicker(specialKey(tea.KeyEnter))
	m = next.(Model)
	if !m.gameDeletePending {
		t.Fatal("Remove Game should request confirmation")
	}
	next, _ = m.updateGamePicker(textKey("n"))
	m = next.(Model)
	if len(m.cfg.Games) != 2 || m.gameDeletePending {
		t.Fatal("N should cancel deletion")
	}
	next, _ = m.updateGamePicker(specialKey(tea.KeyEnter))
	m = next.(Model)
	next, _ = m.updateGamePicker(textKey("y"))
	m = next.(Model)
	if len(m.games) != 1 || m.games[0].Name != "Zebra" {
		t.Fatalf("removal failed: %#v", m.games)
	}
	loaded, err = config.Load(path)
	if err != nil || len(loaded.Games) != 1 || loaded.Games[0].Name != "Zebra" {
		t.Fatalf("removal not persisted: %#v, %v", loaded.Games, err)
	}
}

func TestGameEditorValidationAndCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	m := NewModel(config.Default(), path, false, false)
	m.openGameEditor(-1)
	m.saveConfiguredGame()
	if m.screen != screenGameEditor || !strings.Contains(m.status, "NAME IS REQUIRED") {
		t.Fatalf("missing name wasn't rejected: %q", m.status)
	}
	m.gameDraft.Name = "Invalid"
	m.saveConfiguredGame()
	if m.screen != screenGameEditor || !strings.Contains(m.status, "EXECUTABLE IS REQUIRED") {
		t.Fatalf("missing executable wasn't rejected: %q", m.status)
	}
	next, _ := m.updateGameEditor(specialKey(tea.KeyEscape))
	m = next.(Model)
	if m.screen != screenSetup || len(m.cfg.Games) != 0 {
		t.Fatal("Esc should cancel without saving")
	}
}

func TestNewGameScreensMaintainFooterAtMinimumGeometry(t *testing.T) {
	cfg := config.Default()
	cfg.Games = []config.GameConfig{{Name: "First Game", Command: "/bin/true"}}
	m := NewModel(cfg, "", false, true)
	m.width, m.height = minTerminalWidth, minTerminalHeight
	m.openGamePicker(setupFieldEditGame)
	plain := ansiEscape.ReplaceAllString(m.View().Content, "")
	if !strings.Contains(plain, "FIRST") && !strings.Contains(plain, "First Game") {
		t.Fatalf("picker lost game name at minimum geometry")
	}
	if !strings.Contains(plain, "ESC SETUP") {
		t.Fatalf("picker lost footer at minimum geometry")
	}
	m.openGameEditor(0)
	plain = ansiEscape.ReplaceAllString(m.View().Content, "")
	for _, want := range []string{"NAME", "EXECUTABLE", "ARGUMENTS", "WORKING DIRECTORY", "ESC CANCEL"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("editor lost %q at minimum geometry", want)
		}
	}
}

// Bracketed terminal paste is a dedicated Bubble Tea message. Ctrl+S must
// commit the active input before saving, without a separate Enter press.
func TestGameEditorPasteAndSaveActiveField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	m := NewModel(config.Default(), path, false, false)
	m.openGameEditor(-1)
	m.gameDraft.Name = "Marchine Test"
	m.gameField = gameFieldCommand
	m.gameEditing = true
	m.gameBuffer = ""

	next, cmd := m.Update(tea.PasteMsg{Content: "/home/user/My Games/marchine"})
	m = next.(Model)
	if cmd != nil || m.gameBuffer != "/home/user/My Games/marchine" {
		t.Fatalf("paste was not inserted into active field: %q", m.gameBuffer)
	}

	// Ctrl+C must never quit an active editor or lose the unsaved field.
	next, cmd = m.updateGameEditor(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	m = next.(Model)
	if cmd != nil || !m.gameEditing || m.gameBuffer != "/home/user/My Games/marchine" {
		t.Fatal("Ctrl+C exited or erased active game field")
	}

	next, cmd = m.updateGameEditor(tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl}))
	m = next.(Model)
	if cmd != nil || m.screen != screenSetup {
		t.Fatalf("Ctrl+S in active field did not save: screen=%d", m.screen)
	}
	saved, err := config.Load(path)
	if err != nil || len(saved.Games) != 1 || saved.Games[0].Command != "/home/user/My Games/marchine" {
		t.Fatalf("active game field was not persisted: %#v, %v", saved.Games, err)
	}
}

func TestGameEditorInvalidArgumentsStayEditableOnSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	m := NewModel(config.Default(), path, false, false)
	m.openGameEditor(-1)
	m.gameDraft.Name = "Test"
	m.gameDraft.Command = "code"
	m.gameField = gameFieldArgs
	m.gameEditing = true
	m.gameBuffer = `"unterminated`
	next, cmd := m.updateGameEditor(tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl}))
	m = next.(Model)
	if cmd != nil || m.screen != screenGameEditor || !m.gameEditing || m.gameBuffer != `"unterminated` {
		t.Fatal("invalid arguments unexpectedly saved or discarded active input")
	}
	if !strings.Contains(m.status, "GAME ARGUMENTS:") {
		t.Fatalf("expected validation error, got %q", m.status)
	}
}

func TestArcadeSetupPasteAndSaveActiveField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	m := NewModel(config.Default(), path, false, false)
	m.openSetup()
	m.setupField = setupFieldMAME
	m.setupEditing = true
	m.setupBuffer = ""
	next, cmd := m.Update(tea.PasteMsg{Content: "/opt/mame custom/mame"})
	m = next.(Model)
	if cmd != nil || m.setupBuffer != "/opt/mame custom/mame" {
		t.Fatalf("Arcade field paste failed: %q", m.setupBuffer)
	}
	next, cmd = m.updateSetupEdit(tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl}))
	m = next.(Model)
	if cmd != nil || m.setupEditing || m.cfg.Arcade.MAMECommand != "/opt/mame custom/mame" {
		t.Fatal("Ctrl+S in Arcade field did not apply and save")
	}
	saved, err := config.Load(path)
	if err != nil || saved.Arcade.MAMECommand != "/opt/mame custom/mame" {
		t.Fatalf("Arcade paste was not persisted: %#v, %v", saved.Arcade, err)
	}
}
