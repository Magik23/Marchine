package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/gamepad"
	"github.com/magik23/marchine/internal/library"
)

func TestPlaylistRootBackDoesNotSwitchSource(t *testing.T) {
	m := playlistTestModel(t)
	m.openPlaylistManager()
	m = playlistKey(m, "esc")
	if m.source != library.SourcePlaylists || m.screen != screenPlaylists {
		t.Fatalf("ESC switched away from the PLAYLISTS root: source=%q screen=%v", m.source, m.screen)
	}
	result, _ := m.Update(GamepadMsg{Action: gamepad.ActionBack})
	m = result.(Model)
	if m.source != library.SourcePlaylists || m.screen != screenPlaylists {
		t.Fatal("controller B switched away from the PLAYLISTS root")
	}
	m = playlistKey(m, "tab")
	if m.source != library.SourceArcade || m.screen != screenLibrary {
		t.Fatal("TAB should continue to switch PLAYLISTS to ARCADE")
	}
}

func TestPlaylistNestedBackStillReturnsToManager(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites"}}
	m.openPlaylistManager()
	m = playlistKey(m, "enter")
	if m.screen != screenLibrary || m.source != library.SourcePlaylists {
		t.Fatal("playlist did not open")
	}
	m = playlistKey(m, "esc")
	if m.screen != screenPlaylists || m.source != library.SourcePlaylists {
		t.Fatal("ESC did not return from a playlist to the manager")
	}
}

func TestPlaylistCountsAndSourceBarRightInset(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites", Items: []config.PlaylistItem{{ID: "mame:mslug"}}}}
	m.openPlaylistManager()
	const width = 122

	for _, tc := range []struct {
		count int
		want  string
	}{{0, "0 PLAYLISTS"}, {1, "1 PLAYLIST"}, {3, "3 PLAYLISTS"}} {
		playlists := make([]config.PlaylistConfig, tc.count)
		copy(playlists, m.cfg.Playlists)
		m.cfg.Playlists = playlists
		plain := ansi.Strip(m.renderSourceTabs(width))
		if !strings.HasSuffix(plain, tc.want+"  ") {
			t.Fatalf("source bar = %q; want count %q with two-cell inset", plain, tc.want)
		}
		if lipgloss.Width(m.renderSourceTabs(width)) != width {
			t.Fatalf("source bar width exceeds viewport for count %d", tc.count)
		}
	}

	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites", Items: []config.PlaylistItem{{ID: "mame:mslug"}}}}
	plainPanel := ansi.Strip(m.renderPlaylists(width, 20))
	if strings.Contains(plainPanel, "PLAYLISTS  //") || !strings.Contains(plainPanel, "1 GAME") {
		t.Fatalf("playlist panel did not use clean title and singular item count: %q", plainPanel)
	}
	if got, want := lipgloss.Width(m.renderPlaylists(width, 20)), lipgloss.Width(m.renderList(width, 20)); got != want {
		t.Fatalf("playlist and library frame widths disagree: %d vs %d", got, want)
	}
	if got, want := lipgloss.Height(m.renderPlaylists(width, 20)), lipgloss.Height(m.renderList(width, 20)); got != want {
		t.Fatalf("playlist and library frame heights disagree: %d vs %d", got, want)
	}
}

func TestSourceFooterOrderIsConsistent(t *testing.T) {
	m := playlistTestModel(t)
	m.cfg.Playlists = []config.PlaylistConfig{{Name: "Favorites"}}
	for _, source := range []library.Source{library.SourceArcade, library.SourceGames, library.SourcePlaylists} {
		m.source = source
		m.screen = screenLibrary
		if source == library.SourcePlaylists {
			m.screen = screenPlaylists
		}
		footer := ""
		if source == library.SourcePlaylists {
			footer = ansi.Strip(m.renderPlaylistFooter(122))
		} else {
			footer = ansi.Strip(m.renderFooter(122))
		}
		local, tab, global := strings.Index(footer, "ENTER"), strings.Index(footer, "TAB SOURCE"), strings.Index(footer, "I INFO")
		if local < 0 || tab <= local || global <= tab || !strings.Contains(footer, "Q QUIT") {
			t.Fatalf("unexpected footer order for source %v: %q", source, footer)
		}
	}
}
