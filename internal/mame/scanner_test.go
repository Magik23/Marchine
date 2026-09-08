package mame

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/magik23/marchine/internal/config"
)

func TestLoadIndexesAndCuratesFakeMAME(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fixture")
	}

	d := t.TempDir()

	roms := filepath.Join(d, "roms")
	if err := os.MkdirAll(roms, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, n := range []string{
		"ddonpach.zip",
		"sf2ce.zip",
		"sf2ceua.zip",
		"mahjong.zip",
		"badgame.zip",
	} {
		if err := os.WriteFile(
			filepath.Join(roms, n),
			nil,
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	catver := filepath.Join(d, "catver.ini")

	if err := os.WriteFile(
		catver,
		[]byte(`[Category]
ddonpach=Shooter / Flying Vertical
sf2ce=Fighter / Versus
sf2ceua=Fighter / Versus
mahjong=Mahjong / Tiles
badgame=Shooter / Flying Horizontal

[CatVer]
ddonpach=0.68
sf2ce=0.35b6
sf2ceua=0.35b6
mahjong=0.37b5
badgame=0.99
`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(d, "mame")

	xml := `<?xml version="1.0"?>
<mame>
  <machine name="ddonpach" runnable="yes"><description>DoDonPachi</description><year>1997</year><manufacturer>Cave</manufacturer><driver status="good"/></machine>
  <machine name="sf2ce" runnable="yes"><description>Street Fighter II': Champion Edition</description><year>1992</year><manufacturer>Capcom</manufacturer><driver status="good"/></machine>
  <machine name="sf2ceua" cloneof="sf2ce" runnable="yes"><description>Street Fighter II': Champion Edition (USA)</description><year>1992</year><manufacturer>Capcom</manufacturer><driver status="good"/></machine>
  <machine name="mahjong" runnable="yes"><description>Super Mahjong</description><year>1990</year><manufacturer>Example</manufacturer><driver status="good"/></machine>
  <machine name="badgame" runnable="yes"><description>Bad Game</description><year>1999</year><manufacturer>Example</manufacturer><driver status="preliminary"/></machine>
</mame>`

	writeFakeMAME(
		t,
		fake,
		roms,
		xml,
	)

	t.Setenv(
		"XDG_CACHE_HOME",
		filepath.Join(d, "cache"),
	)

	cfg := config.ArcadeConfig{
		Enabled:         true,
		MAMECommand:     fake,
		ROMPaths:        []string{roms},
		CatVerPath:      catver,
		HideClones:      true,
		HidePreliminary: true,
		HideJunk:        true,
	}

	res, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	// Marchine keeps the complete installed runnable library. HideJunk is a
	// presentation preference handled by the UI, not a destructive index rule.
	if len(res.Games) != 3 {
		t.Fatalf(
			"got %d games: %#v",
			len(res.Games),
			res.Games,
		)
	}

	if res.FromCache {
		t.Fatal(
			"forced scan must not report FromCache",
		)
	}

	if res.Games[0].Name != "DoDonPachi" ||
		res.Games[0].Year != "1997" ||
		res.Games[0].Manufacturer != "Cave" ||
		res.Games[0].Category != "SHMUPS" {
		t.Fatalf(
			"unexpected first game: %#v",
			res.Games[0],
		)
	}

	if res.Games[1].Category != "FIGHTING" {
		t.Fatalf(
			"unexpected second game: %#v",
			res.Games[1],
		)
	}

	if res.Games[2].Category != "MAHJONG" ||
		!res.Games[2].NonArcade ||
		res.Games[2].RawCategory !=
			"Mahjong / Tiles" {
		t.Fatalf(
			"mahjong should remain indexed and classified: %#v",
			res.Games[2],
		)
	}
}

func TestNormalLoadUsesSavedLibraryWithoutTouchingMAMEOrROMs(
	t *testing.T,
) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fixture")
	}

	d := t.TempDir()

	roms := filepath.Join(d, "roms")
	if err := os.MkdirAll(roms, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(roms, "ddonpach.zip"),
		nil,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	catver := filepath.Join(d, "catver.ini")

	if err := os.WriteFile(
		catver,
		[]byte(`[Category]
ddonpach=Shooter / Flying Vertical
`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(d, "mame")

	xml := `<?xml version="1.0"?>
<mame>
  <machine name="ddonpach" runnable="yes"><description>DoDonPachi</description><year>1997</year><manufacturer>Cave</manufacturer><driver status="good"/></machine>
</mame>`

	writeFakeMAME(
		t,
		fake,
		roms,
		xml,
	)

	t.Setenv(
		"XDG_CACHE_HOME",
		filepath.Join(d, "cache"),
	)

	cfg := config.ArcadeConfig{
		Enabled:     true,
		MAMECommand: fake,
		ROMPaths:    []string{roms},
		CatVerPath:  catver,
	}

	first, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(first.Games) != 1 {
		t.Fatalf(
			"first scan got %d games",
			len(first.Games),
		)
	}

	// Remove both the emulator and ROM directory. A normal launch must still
	// return the saved library without probing either source.
	if err := os.Remove(fake); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(roms); err != nil {
		t.Fatal(err)
	}

	cached, err := Load(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if !cached.FromCache {
		t.Fatal(
			"normal load should use the saved cache",
		)
	}

	if len(cached.Games) != 1 ||
		cached.Games[0].ROM != "ddonpach" {
		t.Fatalf(
			"unexpected cached games: %#v",
			cached.Games,
		)
	}
}

func TestNewROMIsIgnoredUntilExplicitRefresh(
	t *testing.T,
) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fixture")
	}

	d := t.TempDir()

	roms := filepath.Join(d, "roms")
	if err := os.MkdirAll(roms, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(roms, "ddonpach.zip"),
		nil,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	catver := filepath.Join(d, "catver.ini")

	if err := os.WriteFile(
		catver,
		[]byte(`[Category]
ddonpach=Shooter / Flying Vertical
sf2ce=Fighter / Versus
`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(d, "mame")

	xml := `<?xml version="1.0"?>
<mame>
  <machine name="ddonpach" runnable="yes"><description>DoDonPachi</description><year>1997</year><manufacturer>Cave</manufacturer><driver status="good"/></machine>
  <machine name="sf2ce" runnable="yes"><description>Street Fighter II': Champion Edition</description><year>1992</year><manufacturer>Capcom</manufacturer><driver status="good"/></machine>
</mame>`

	writeFakeMAME(
		t,
		fake,
		roms,
		xml,
	)

	t.Setenv(
		"XDG_CACHE_HOME",
		filepath.Join(d, "cache"),
	)

	cfg := config.ArcadeConfig{
		Enabled:     true,
		MAMECommand: fake,
		ROMPaths:    []string{roms},
		CatVerPath:  catver,
	}

	initial, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(initial.Games) != 1 {
		t.Fatalf(
			"initial scan got %d games",
			len(initial.Games),
		)
	}

	// A new ROM appears on disk.
	if err := os.WriteFile(
		filepath.Join(roms, "sf2ce.zip"),
		nil,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Normal startup must keep the saved playlist and avoid noticing the new ROM.
	normal, err := Load(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if !normal.FromCache {
		t.Fatal(
			"normal load should come from cache",
		)
	}

	if len(normal.Games) != 1 {
		t.Fatalf(
			"normal load should still have 1 cached game, got %d",
			len(normal.Games),
		)
	}

	// Explicit refresh is the point where Marchine looks at the ROM folder again.
	refreshed, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if refreshed.FromCache {
		t.Fatal(
			"explicit refresh should rebuild the library",
		)
	}

	if len(refreshed.Games) != 2 {
		t.Fatalf(
			"refresh should discover 2 games, got %d",
			len(refreshed.Games),
		)
	}
}

func TestForcedRefreshKeepsSavedLibraryWhenROMsDisappear(
	t *testing.T,
) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fixture")
	}

	d := t.TempDir()

	roms := filepath.Join(d, "roms")
	if err := os.MkdirAll(roms, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(roms, "ddonpach.zip"),
		nil,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	catver := filepath.Join(d, "catver.ini")

	if err := os.WriteFile(
		catver,
		[]byte(`[Category]
ddonpach=Shooter / Flying Vertical
`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(d, "mame")

	xml := `<?xml version="1.0"?>
<mame>
  <machine name="ddonpach" runnable="yes"><description>DoDonPachi</description><year>1997</year><manufacturer>Cave</manufacturer><driver status="good"/></machine>
</mame>`

	writeFakeMAME(
		t,
		fake,
		roms,
		xml,
	)

	t.Setenv(
		"XDG_CACHE_HOME",
		filepath.Join(d, "cache"),
	)

	cfg := config.ArcadeConfig{
		Enabled:     true,
		MAMECommand: fake,
		ROMPaths:    []string{roms},
		CatVerPath:  catver,
	}

	first, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(first.Games) != 1 {
		t.Fatalf(
			"first scan got %d games",
			len(first.Games),
		)
	}

	if err := os.RemoveAll(roms); err != nil {
		t.Fatal(err)
	}

	refreshed, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if !refreshed.FromCache {
		t.Fatal(
			"missing ROM path should preserve the saved library",
		)
	}

	if len(refreshed.Games) != 1 {
		t.Fatalf(
			"saved library should remain intact, got %d games",
			len(refreshed.Games),
		)
	}

	if refreshed.Warning !=
		"NO ROMS FOUND — SAVED LIBRARY KEPT" {
		t.Fatalf(
			"unexpected warning: %q",
			refreshed.Warning,
		)
	}
}

func writeFakeMAME(
	t *testing.T,
	path,
	roms,
	xml string,
) {
	t.Helper()

	script := "#!/usr/bin/env bash\n" +
		"set -e\n" +
		"case \"${1:-}\" in\n" +
		"  -version) echo '0.289 (fake)' ;;\n" +
		"  -showconfig) echo 'rompath " +
		shellQuote(roms) +
		"' ;;\n" +
		"  -listxml) cat <<'XML'\n" +
		xml +
		"\nXML\n" +
		"  ;;\n" +
		"  *) exit 2 ;;\n" +
		"esac\n"

	if err := os.WriteFile(
		path,
		[]byte(script),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
}

func shellQuote(s string) string {
	return "'" +
		strings.ReplaceAll(
			s,
			"'",
			"'\\''",
		) +
		"'"
}
