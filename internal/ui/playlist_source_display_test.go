package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

func TestPlaylistListOmitsSourceLabelsButDetailShowsSource(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{
		Name: "Favorites",
		Items: []config.PlaylistItem{
			{ID: "mame:mslug"},
			{ID: m.games[0].ID},
		},
	}}
	m.source = library.SourcePlaylists
	m.screen = screenLibrary
	m.playlistIndex = 0

	items := m.filtered()
	if len(items) != 2 || items[0].Source != library.SourceArcade || items[1].Source != library.SourceGames {
		t.Fatalf("expected one arcade and one native game in playlist, got %#v", items)
	}

	list := ansi.Strip(m.renderList(78, 22))
	for _, game := range items {
		if !strings.Contains(list, game.Name) {
			t.Fatalf("playlist list lost game title %q: %q", game.Name, list)
		}
	}
	for _, unwanted := range []string{"[ARCADE]", "[GAMES]", "SOURCE:"} {
		if strings.Contains(list, unwanted) {
			t.Fatalf("playlist list retained source label %q: %q", unwanted, list)
		}
	}

	for i, want := range []string{"SOURCE: ARCADE", "SOURCE: GAMES"} {
		m.cursor = i
		detail := ansi.Strip(m.renderDetail(sculpturePanelWidth(), 48))
		if !strings.Contains(detail, want) {
			t.Fatalf("selected game detail missing %q: %q", want, detail)
		}
	}

	m.source = library.SourceArcade
	m.cursor = 0
	arcadeDetail := ansi.Strip(m.renderDetail(sculpturePanelWidth(), 48))
	if strings.Contains(arcadeDetail, "SOURCE:") {
		t.Fatal("source label should only appear when viewing mixed playlists")
	}
}
