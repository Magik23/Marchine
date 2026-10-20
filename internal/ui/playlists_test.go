package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/gamepad"
	"github.com/magik23/marchine/internal/library"
)

func playlistTestModel(t *testing.T) Model {
	t.Helper()
	cfg := config.Default()
	cfg.Games = []config.GameConfig{
		{Name: "Hollow Knight", Command: "/bin/true"},
		{Name: "Streets of Rage 4", Command: "/bin/false"},
	}
	m := NewModel(cfg, filepath.Join(t.TempDir(), "config.toml"), false, false)
	m.arcade = demoArcade()
	m.loading = false
	m.width, m.height = minTerminalWidth, minTerminalHeight
	return m
}

func playlistKey(m Model, key string) Model {
	var msg tea.KeyPressMsg
	if key == "enter" {
		msg = specialKey(tea.KeyEnter)
	} else if key == "esc" {
		msg = specialKey(tea.KeyEscape)
	} else if key == "tab" {
		msg = specialKey(tea.KeyTab)
	} else {
		msg = textKey(key)
	}
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestUnifiedPlaylistCreationSavingLaunchingAndPersistence(t *testing.T) {
	m := playlistTestModel(t)
	// Three source tabs and a dedicated manager. No playlist is loaded in ARCADE or GAMES.
	m.source = library.SourceArcade
	m = playlistKey(m, "tab")
	if m.source != library.SourceGames || m.screen != screenLibrary {
		t.Fatal("ARCADE -> GAMES failed")
	}
	m = playlistKey(m, "tab")
	if m.source != library.SourcePlaylists || m.screen != screenPlaylists {
		t.Fatal("GAMES -> PLAYLISTS failed")
	}
	m = playlistKey(m, "n")
	if m.screen != screenPlaylistName {
		t.Fatal("NEW did not open name entry")
	}
	m = playlistKey(m, "Favorites")
	m = playlistKey(m, "enter")
	if len(m.cfg.Playlists) != 1 || m.cfg.Playlists[0].Name != "Favorites" || m.screen != screenPlaylists {
		t.Fatalf("create failed: %#v", m.cfg.Playlists)
	}
	m = playlistKey(m, "tab")
	if m.source != library.SourceArcade || m.screen != screenLibrary {
		t.Fatal("PLAYLISTS -> ARCADE failed")
	}
	// Save an Arcade ROM into Favorites (not into ARCADE's category filter).
	for i, g := range m.filtered() {
		if g.ID == "mame:mslug" {
			m.cursor = i
			break
		}
	}
	m = playlistKey(m, "p")
	if m.screen != screenPlaylistSave {
		t.Fatal("P did not open save destination picker")
	}
	m = playlistKey(m, "enter")
	if m.screen != screenLibrary || len(m.cfg.Playlists[0].Items) != 1 || m.cfg.Playlists[0].Items[0].ID != "mame:mslug" {
		t.Fatalf("ROM not saved: %#v", m.cfg.Playlists)
	}
	// Save a native executable to the *same* playlist.
	m = playlistKey(m, "tab")
	m.cursor = 0
	native, ok := m.selected()
	if !ok || native.Source != library.SourceGames {
		t.Fatalf("expected GAMES entry: %#v", native)
	}
	m = playlistKey(m, "p")
	m = playlistKey(m, "enter")
	if len(m.cfg.Playlists[0].Items) != 2 || m.cfg.Playlists[0].Items[1].ID != native.ID {
		t.Fatalf("GAMES entry not saved: %#v", m.cfg.Playlists)
	}
	m = playlistKey(m, "p")
	m = playlistKey(m, "enter")
	if !strings.Contains(m.status, "ALREADY") || len(m.cfg.Playlists[0].Items) != 2 {
		t.Fatal("duplicate add was not rejected")
	}
	m = playlistKey(m, "esc")
	m = playlistKey(m, "tab")
	m = playlistKey(m, "enter")
	items := m.filtered()
	if m.screen != screenLibrary || len(items) != 2 || items[0].Source != library.SourceArcade || items[1].Source != library.SourceGames {
		t.Fatalf("unified playlist returned incorrect entries: %#v", items)
	}
	if items[0].Command != "mame" || items[1].Command != "/bin/true" {
		t.Fatalf("original launch configurations lost: %#v", items)
	}
	if !strings.Contains(m.View().Content, "PLAYLISTS") {
		t.Fatal("third tab not visible")
	}
	loaded, err := config.Load(m.configPath)
	if err != nil || len(loaded.Playlists) != 1 || len(loaded.Playlists[0].Items) != 2 || "game:"+loaded.Games[0].ID != native.ID {
		t.Fatalf("persisted state invalid: %#v %v", loaded, err)
	}
}

func TestPlaylistsKeepIdentityThroughGameEditAndRemoval(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites", Items: []config.PlaylistItem{{ID: "game:" + m.cfg.Games[1].ID, Name: "Streets of Rage 4"}}}}
	id := m.cfg.Games[1].ID
	if err := config.Save(m.configPath, m.cfg); err != nil {
		t.Fatal(err)
	}
	m.openGameEditor(1)
	m.gameDraft.Name = "Renamed Native"
	m.saveConfiguredGame()
	if m.cfg.Games[1].ID != id {
		t.Fatal("edit changed GAMES identity")
	}
	m.openPlaylistManager()
	m = playlistKey(m, "enter")
	if got := m.filtered(); len(got) != 1 || got[0].Name != "Renamed Native" || got[0].Command != "/bin/false" {
		t.Fatalf("playlist no longer follows edited game: %#v", got)
	}
	m.removeConfiguredGame(1)
	m.openPlaylistManager()
	m = playlistKey(m, "enter")
	ghost := m.filtered()
	if len(ghost) != 1 || !strings.Contains(ghost[0].Name, "UNAVAILABLE") || ghost[0].Command != "" {
		t.Fatalf("missing reference was silently retargeted: %#v", ghost)
	}
	m = playlistKey(m, "enter")
	if !strings.Contains(m.status, "UNAVAILABLE") {
		t.Fatal("missing reference could be launched")
	}
}

func TestPlaylistDeletionRemovalAndNoFileDeletion(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites", Items: []config.PlaylistItem{{ID: "mame:mslug", Name: "Metal Slug"}}}}
	m.openPlaylistManager()
	m = playlistKey(m, "d")
	m = playlistKey(m, "n")
	if len(m.cfg.Playlists) != 1 {
		t.Fatal("cancel deleted playlist")
	}
	m = playlistKey(m, "enter")
	m = playlistKey(m, "d")
	m = playlistKey(m, "y")
	if len(m.cfg.Playlists[0].Items) != 0 || len(m.arcade) == 0 {
		t.Fatal("removing reference modified game library")
	}
	m = playlistKey(m, "esc")
	m = playlistKey(m, "d")
	m = playlistKey(m, "y")
	if len(m.cfg.Playlists) != 0 || len(m.cfg.Games) != 2 || len(m.arcade) == 0 {
		t.Fatal("deleting playlist modified source libraries")
	}
}

func TestControllerPlaylistSaveShortcut(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites"}}
	m.source = library.SourceGames
	next, _ := m.Update(GamepadMsg{Action: gamepad.ActionSaveToPlaylist})
	if next.(Model).screen != screenPlaylistSave {
		t.Fatal("controller save chord did not open playlist picker")
	}
	next, _ = next.(Model).Update(GamepadMsg{Action: gamepad.ActionBack})
	if next.(Model).screen != screenLibrary {
		t.Fatal("controller B did not return to source")
	}
}

func TestLegacyIDsArePersistedOnFirstPlaylistSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	legacy := "[[games]]\nname = 'Original'\ncommand = '/bin/true'\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(cfg, path, false, false)
	oldID := m.games[0].ID
	if oldID == "" {
		t.Fatal("legacy game has no stable identity")
	}
	m.openPlaylistManager()
	m = playlistKey(m, "n")
	m = playlistKey(m, "Favorites")
	m = playlistKey(m, "enter")
	if m.screen != screenPlaylists {
		t.Fatalf("playlist creation failed: %s", m.status)
	}
	newCfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(newCfg.Games) != 1 || "game:"+newCfg.Games[0].ID != oldID || len(newCfg.Playlists) != 1 {
		t.Fatalf("legacy ID migration failed on first playlist save: %#v", newCfg)
	}
}
