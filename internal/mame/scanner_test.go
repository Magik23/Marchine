package mame

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
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
  <machine name="ddonpach" runnable="yes"><description>DoDonPachi</description><year>1997</year><manufacturer>Cave</manufacturer><driver status="good"/><display rotate="270"/></machine>
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
		res.Games[0].Category != "SHMUPS" ||
		!res.Games[0].Vertical {
		t.Fatalf(
			"unexpected first game: %#v",
			res.Games[0],
		)
	}

	wantLaunchArgs := []string{"ddonpach", "-rompath", filepath.Clean(roms), "-autorol"}
	if !reflect.DeepEqual(res.Games[0].Args, wantLaunchArgs) {
		t.Fatalf("explicit ROM path/orientation missing from launch args: got %#v want %#v", res.Games[0].Args, wantLaunchArgs)
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

func TestLaunchArgsPreserveMultipleExplicitROMPaths(t *testing.T) {
	paths := []string{"/mnt/arcade-a", "/mnt/arcade-b", "/mnt/arcade-a"}
	got := LaunchArgs("1942", paths, true)
	want := []string{"1942", "-rompath", "/mnt/arcade-a;/mnt/arcade-b", "-autorol"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LaunchArgs=%#v want=%#v", got, want)
	}
}

func TestCurrentCacheRehydratesLaunchIdentityFromConfig(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(d, "cache"))

	path, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}

	cached := Result{
		Games: []library.Game{{
			ID:       "mame:1942",
			Name:     "1942",
			Source:   library.SourceArcade,
			ROM:      "1942",
			Vertical: true,
			Command:  "/old/mame",
			Args:     []string{"1942"},
		}},
		MAMEVersion: "0.289",
		ROMPaths:    []string{"/old/roms"},
	}
	if err := writeCache(path, cached); err != nil {
		t.Fatal(err)
	}

	cfg := config.ArcadeConfig{
		Enabled:     true,
		MAMECommand: "/new/mame",
		ROMPaths:    []string{"/new/roms", "/more/roms"},
	}
	res, err := Load(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.FromCache || len(res.Games) != 1 {
		t.Fatalf("unexpected cache result: %#v", res)
	}
	if res.Games[0].Command != "/new/mame" {
		t.Fatalf("cached command=%q want current config", res.Games[0].Command)
	}
	wantArgs := []string{"1942", "-rompath", "/new/roms;/more/roms", "-autorol"}
	if !reflect.DeepEqual(res.Games[0].Args, wantArgs) {
		t.Fatalf("cached launch args=%#v want=%#v", res.Games[0].Args, wantArgs)
	}
}

func TestLegacyCacheIsExplicitDegradedFallback(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(d, "cache"))

	path, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	legacy := cacheFile{
		// Intentionally no Schema: this represents caches from before v4.
		SavedAt: time.Now(),
		Result: cachedResult{Games: []library.Game{{
			ID:      "mame:1942",
			Name:    "1942",
			Source:  library.SourceArcade,
			ROM:     "1942",
			Command: "/old/mame",
		}}},
	}
	b, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.ArcadeConfig{Enabled: true, MAMECommand: filepath.Join(d, "missing-mame")}
	res, err := Load(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.FromCache || len(res.Games) != 1 {
		t.Fatalf("legacy fallback lost library: %#v", res)
	}
	if res.Warning != "LEGACY CACHE — REFRESH REQUIRED" {
		t.Fatalf("legacy warning=%q", res.Warning)
	}
	if res.Diagnostic == "" {
		t.Fatal("legacy degraded fallback should retain the live-source diagnostic")
	}
}

func TestForcedScanFailureKeepsCurrentCache(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fixture")
	}

	d := t.TempDir()
	roms := filepath.Join(d, "roms")
	if err := os.MkdirAll(roms, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roms, "1942.zip"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(d, "mame")
	goodXML := `<?xml version="1.0"?><mame><machine name="1942" runnable="yes"><description>1942</description><year>1984</year><manufacturer>Capcom</manufacturer><driver status="good"/><display rotate="90"/></machine></mame>`
	writeFakeMAME(t, fake, roms, goodXML)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(d, "cache"))

	cfg := config.ArcadeConfig{Enabled: true, MAMECommand: fake, ROMPaths: []string{roms}}
	first, err := Load(cfg, true)
	if err != nil || len(first.Games) != 1 {
		t.Fatalf("initial scan failed: res=%#v err=%v", first, err)
	}

	broken := "#!/usr/bin/env bash\ncase \"${1:-}\" in\n-version) echo fake;;\n-showconfig) echo 'rompath " + shellQuote(roms) + "';;\n-listxml) printf '<mame><machine'; exit 1;;\n*) exit 2;;\nesac\n"
	if err := os.WriteFile(fake, []byte(broken), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Load(cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.FromCache || len(res.Games) != 1 {
		t.Fatalf("failed refresh should preserve cache: %#v", res)
	}
	if res.Warning != "REFRESH FAILED — SAVED LIBRARY KEPT" {
		t.Fatalf("warning=%q", res.Warning)
	}
	if res.Diagnostic == "" {
		t.Fatal("refresh fallback should retain diagnostic detail")
	}
}

func TestCacheFileCarriesCurrentSchema(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "arcade.json")
	if err := writeCache(path, Result{MAMEVersion: "0.289"}); err != nil {
		t.Fatal(err)
	}
	c, err := readCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Schema != cacheSchema {
		t.Fatalf("schema=%q want=%q", c.Schema, cacheSchema)
	}
	matches, err := filepath.Glob(filepath.Join(d, ".arcade-*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("atomic cache write left temp files: %#v", matches)
	}
}
