package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

func fakeGio(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gio"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func writeLauncher(t *testing.T, path, name string) {
	t.Helper()
	content := "[Desktop Entry]\nType=Application\nName=" + name + "\nExec=steam steam://rungameid/367520\nTerminal=false\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestImportDiscoversDesktopShortcutsAndExecutables(t *testing.T) {
	fakeGio(t)
	dir := t.TempDir()
	shortcut := filepath.Join(dir, "Hollow Knight.desktop")
	writeLauncher(t, shortcut, "Hollow Knight")
	binary := filepath.Join(dir, "Celeste.AppImage")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, skipped, err := discoverImportCandidates(dir, nil)
	if err != nil || len(rows) != 2 || skipped != 1 {
		t.Fatalf("directory import got rows=%#v skipped=%d err=%v", rows, skipped, err)
	}
	if rows[0].game.Name != "Celeste" || rows[0].game.Command != binary || rows[0].game.WorkingDir != dir {
		t.Fatalf("native executable not identified: %+v", rows[0])
	}
	if rows[1].game.Name != "Hollow Knight" || rows[1].game.Command != "gio" || len(rows[1].game.Args) != 2 || rows[1].game.Args[0] != "launch" || rows[1].game.Args[1] != shortcut {
		t.Fatalf("desktop shortcut not preserved as gio launch: %+v", rows[1])
	}
	if !rows[0].selected || !rows[1].selected {
		t.Fatal("new candidates must be selected by default")
	}

	glob, _, err := discoverImportCandidates(filepath.Join(dir, "*.desktop"), nil)
	if err != nil || len(glob) != 1 || glob[0].game.Name != "Hollow Knight" {
		t.Fatalf("glob import failed: %#v %v", glob, err)
	}
	quoted, _, err := discoverImportCandidates("\""+shortcut+"\"", nil)
	if err != nil || len(quoted) != 1 {
		t.Fatalf("quoted path with spaces failed: %#v %v", quoted, err)
	}
	already := []config.GameConfig{{Name: "Custom Name", Command: "gio", Args: []string{"launch", shortcut}}}
	dups, _, err := discoverImportCandidates(shortcut, already)
	if err != nil || len(dups) != 1 || !dups[0].duplicate || dups[0].selected {
		t.Fatalf("duplicate should be visible but unselected: %#v %v", dups, err)
	}
}

func TestImportSkipsUnsafeOrInvalidShortcuts(t *testing.T) {
	fakeGio(t)
	dir := t.TempDir()
	for name, source := range map[string]string{
		"hidden.desktop":    "[Desktop Entry]\nType=Application\nHidden=true\nName=Hidden\nExec=/bin/true\n",
		"nodisplay.desktop": "[Desktop Entry]\nType=Application\nNoDisplay=true\nName=NoDisplay\nExec=/bin/true\n",
		"link.desktop":      "[Desktop Entry]\nType=Link\nName=Link\nURL=https://example.org\n",
		"empty.desktop":     "[Desktop Entry]\nType=Application\nName=No Command\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rows, skipped, err := discoverImportCandidates(filepath.Join(dir, "*.desktop"), nil)
	if err != nil || len(rows) != 0 || skipped != 4 {
		t.Fatalf("invalid shortcuts accepted: rows=%#v skipped=%d err=%v", rows, skipped, err)
	}
	if _, _, err := discoverImportCandidates("[", nil); err == nil {
		t.Fatal("invalid glob should fail safely")
	}
}

func TestImportLargeScrollAndSelectionPersist(t *testing.T) {
	m := NewModel(config.Default(), filepath.Join(t.TempDir(), "config.toml"), false, false)
	m.width, m.height = minTerminalWidth, minTerminalHeight
	m.openGameImport()
	m.importCandidates = make([]importCandidate, 200)
	for i := range m.importCandidates {
		m.importCandidates[i] = importCandidate{
			game: config.GameConfig{Name: fmt.Sprintf("Game %03d", i), Command: fmt.Sprintf("/games/%03d", i)},
			kind: "EXECUTABLE", selected: false,
		}
	}
	m.screen = screenGameImportPreview
	m.importCandidates[1].selected = true
	m.importCandidates[86].selected = true
	m.importCandidates[197].selected = true
	m.importCursor = 197
	if m.importSelectedCount() != 3 {
		t.Fatal("selection not retained outside viewport")
	}
	plain := ansiEscape.ReplaceAllString(m.View().Content, "")
	for _, text := range []string{"3 SELECTED", "198/200", "Game 197", "ESC PATH"} {
		if !strings.Contains(plain, text) {
			t.Fatalf("large-list preview missing %q", text)
		}
	}
	if strings.Contains(plain, "Game 001") {
		t.Fatal("200-row preview leaked off-screen rows")
	}
	next, _ := m.updateGameImportPreview(specialKey(tea.KeyHome))
	m = next.(Model)
	if m.importCursor != 0 || m.importSelectedCount() != 3 {
		t.Fatalf("jump to top lost selections: cursor=%d selected=%d", m.importCursor, m.importSelectedCount())
	}
	next, _ = m.updateGameImportPreview(textKey("n"))
	m = next.(Model)
	if m.importSelectedCount() != 0 {
		t.Fatal("clear all failed")
	}
	next, _ = m.updateGameImportPreview(textKey("a"))
	m = next.(Model)
	if m.importSelectedCount() != 200 {
		t.Fatalf("select all expected 200, got %d", m.importSelectedCount())
	}
}

func TestImportOneAtomicSaveAndCancellation(t *testing.T) {
	fakeGio(t)
	dir := t.TempDir()
	writeLauncher(t, filepath.Join(dir, "Steam Game.desktop"), "Steam Game")
	writeLauncher(t, filepath.Join(dir, "GOG Game.desktop"), "GOG Game")
	path := filepath.Join(t.TempDir(), "marchine", "config.toml")
	cfg := config.Default()
	cfg.Arcade.ROMPaths = []string{"/mame/one", "/mame/two"}
	cfg.Games = []config.GameConfig{{Name: "Existing", Command: "/bin/true"}}
	m := NewModel(cfg, path, false, false)
	m.source = library.SourceGames
	m.openGameImport()
	m.importPattern = filepath.Join(dir, "*.desktop")
	next, _ := m.updateGameImportPath(specialKey(tea.KeyEnter))
	m = next.(Model)
	if m.screen != screenGameImportPreview || m.importSelectedCount() != 2 {
		t.Fatalf("no preview after discovery: screen=%d selected=%d status=%s", m.screen, m.importSelectedCount(), m.status)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("preview must never save config")
	}
	// First back out to path entry, then return to preview, keeping config untouched.
	next, _ = m.updateGameImportPreview(specialKey(tea.KeyEscape))
	m = next.(Model)
	if m.screen != screenGameImportPath {
		t.Fatal("Esc should return to path entry")
	}
	next, _ = m.updateGameImportPath(specialKey(tea.KeyEnter))
	m = next.(Model)
	next, _ = m.updateGameImportPreview(textKey("i"))
	m = next.(Model)
	if !m.importConfirm {
		t.Fatal("import should require explicit confirmation")
	}
	next, _ = m.updateGameImportPreview(textKey("n"))
	m = next.(Model)
	if m.importConfirm {
		t.Fatal("N should cancel import confirmation")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("cancelling import must not save config")
	}
	next, _ = m.updateGameImportPreview(textKey("i"))
	m = next.(Model)
	next, _ = m.updateGameImportPreview(textKey("y"))
	m = next.(Model)
	if m.screen != screenSetup || m.setupField != setupFieldImportGame || len(m.games) != 3 {
		t.Fatalf("commit did not refresh games: screen=%d games=%d status=%s", m.screen, len(m.games), m.status)
	}
	loaded, err := config.Load(path)
	if err != nil || len(loaded.Games) != 3 || len(loaded.Arcade.ROMPaths) != 2 {
		t.Fatalf("atomic config save lost entries or arcade paths: %#v %v", loaded, err)
	}
	// Reimport identical shortcut paths: show duplicate rows but never add again.
	m.openGameImport()
	m.importPattern = filepath.Join(dir, "*.desktop")
	next, _ = m.updateGameImportPath(specialKey(tea.KeyEnter))
	m = next.(Model)
	if len(m.importCandidates) != 2 || m.importSelectedCount() != 0 || !m.importCandidates[0].duplicate {
		t.Fatalf("reimport did not identify duplicates: %#v", m.importCandidates)
	}
}

func TestGameImportDemoDoesNotSave(t *testing.T) {
	m := NewModel(config.Default(), filepath.Join(t.TempDir(), "config.toml"), false, true)
	m.openGameImport()
	m.importCandidates = []importCandidate{{game: config.GameConfig{Name: "Example", Command: "/bin/true"}, selected: true}}
	m.screen = screenGameImportPreview
	next, _ := m.updateGameImportPreview(textKey("i"))
	m = next.(Model)
	if m.importConfirm || !strings.Contains(m.status, "DEMO") {
		t.Fatal("demo import should be explicitly disabled")
	}
}

func TestImportNameFromGameFolderAndSkipWindowsTargets(t *testing.T) {
	if name := nativeImportName("/home/user/Games/Hollow Knight/start.sh"); name != "Hollow Knight" {
		t.Fatalf("generic launch script should use its game directory: %q", name)
	}
	if name := nativeImportName("/home/user/Games/Celeste/Celeste.AppImage"); name != "Celeste" {
		t.Fatalf("AppImage name unexpectedly changed: %q", name)
	}
	for _, file := range []string{"Game.exe", "shortcut.lnk", "internet.url", "installer.msi", "lib.so"} {
		if nativeImportFile(file) {
			t.Fatalf("unsupported Windows/shared target accepted: %q", file)
		}
	}
}

func TestDesktopHandoffClassification(t *testing.T) {
	if !isDesktopGameLaunch("gio", []string{"launch", "/home/test/Steam Game.desktop"}) {
		t.Fatal("desktop launcher should be classified as handoff")
	}
	if !isDesktopGameLaunch("/usr/bin/gio", []string{"launch", "/tmp/gog.desktop"}) {
		t.Fatal("absolute gio command should be classified as handoff")
	}
	if isDesktopGameLaunch("wine", []string{"game.exe"}) || isDesktopGameLaunch("gio", []string{"open", "game.desktop"}) {
		t.Fatal("ordinary executable incorrectly classified as desktop handoff")
	}
}
