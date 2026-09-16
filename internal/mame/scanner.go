package mame

import (
	"bufio"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

const cacheSchema = "marchine-index-v4"

var errCacheSchema = errors.New("unsupported Marchine cache schema")

const (
	showConfigTimeout = 10 * time.Second
	versionTimeout    = 5 * time.Second
	listXMLTimeout    = 3 * time.Minute
)

type Result struct {
	Games       []library.Game
	MAMEVersion string
	ROMPaths    []string
	CatVerUsed  string
	FromCache   bool
	Warning     string
	Diagnostic  string
}

type cacheFile struct {
	Schema  string       `json:"schema"`
	SavedAt time.Time    `json:"saved_at"`
	Result  cachedResult `json:"result"`
}

type cachedResult struct {
	Games       []library.Game `json:"games"`
	MAMEVersion string         `json:"mame_version"`
	ROMPaths    []string       `json:"rom_paths"`
	CatVerUsed  string         `json:"catver_used"`
	Warning     string         `json:"warning"`
}

type machineXML struct {
	Name         string `xml:"name,attr"`
	CloneOf      string `xml:"cloneof,attr"`
	IsBIOS       string `xml:"isbios,attr"`
	IsDevice     string `xml:"isdevice,attr"`
	Runnable     string `xml:"runnable,attr"`
	Description  string `xml:"description"`
	Year         string `xml:"year"`
	Manufacturer string `xml:"manufacturer"`
	Driver       struct {
		Status string `xml:"status,attr"`
	} `xml:"driver"`
	Displays []struct {
		Rotate string `xml:"rotate,attr"`
	} `xml:"display"`
}

func Load(cfg config.ArcadeConfig, force bool) (Result, error) {
	if !cfg.Enabled {
		return Result{}, nil
	}

	// Normal startup is cache-first by design. A cache written by the current
	// schema is used immediately without touching MAME, ROM paths, or CatVer.
	// A legacy cache is deliberately NOT treated as current: Marchine attempts a
	// fresh scan once, and only exposes that legacy data as an explicit degraded
	// fallback when the live source is unavailable.
	cache, _ := cachePath()
	var saved *cacheFile
	var legacy *cacheFile

	if cache != "" {
		if c, err := readCache(cache); err == nil {
			saved = &c
			if !force {
				return resultFromCache(c, cfg, "", ""), nil
			}
		} else if errors.Is(err, errCacheSchema) && len(c.Result.Games) > 0 {
			legacy = &c
		}
	}

	mameCommand := strings.TrimSpace(cfg.MAMECommand)
	if mameCommand == "" {
		mameCommand = "mame"
	}

	mamePath, err := exec.LookPath(mameCommand)
	if err != nil {
		if res, ok := cacheFallback(saved, legacy, cfg, "MAME NOT FOUND — SAVED LIBRARY KEPT", err); ok {
			return res, nil
		}
		return Result{Warning: "MAME NOT FOUND", Diagnostic: err.Error()}, nil
	}

	romPaths := append([]string(nil), cfg.ROMPaths...)
	if len(romPaths) == 0 {
		romPaths = detectROMPaths(mamePath)
	}

	roms, err := discoverROMs(romPaths)
	if err != nil {
		if res, ok := cacheFallback(saved, legacy, cfg, "REFRESH FAILED — SAVED LIBRARY KEPT", err); ok {
			return res, nil
		}
		return Result{}, err
	}

	// Do not destroy a good saved library because an external drive is
	// temporarily missing or a configured ROM path is unavailable.
	if len(roms) == 0 {
		if res, ok := cacheFallback(saved, legacy, cfg, "NO ROMS FOUND — SAVED LIBRARY KEPT", errors.New("no ROM archives found")); ok {
			return res, nil
		}

		return Result{
			MAMEVersion: readVersion(mamePath),
			ROMPaths:    romPaths,
			CatVerUsed:  detectCatVer(cfg.CatVerPath),
			Warning:     "NO ROMS FOUND — SET arcade.rom_paths",
		}, nil
	}

	version := readVersion(mamePath)
	catver := detectCatVer(cfg.CatVerPath)

	categories := map[string]string{}
	warning := ""
	if catver != "" {
		categories, err = ParseCatVer(catver)
		if err != nil {
			warning = "CATVER READ FAILED"
			categories = map[string]string{}
		}
	}

	games, err := parseInstalledMachines(mamePath, roms, categories, cfg)
	if err != nil {
		if res, ok := cacheFallback(saved, legacy, cfg, "REFRESH FAILED — SAVED LIBRARY KEPT", err); ok {
			return res, nil
		}
		return Result{}, err
	}

	res := Result{
		Games:       games,
		MAMEVersion: version,
		ROMPaths:    romPaths,
		CatVerUsed:  catver,
		Warning:     warning,
	}

	if catver == "" && res.Warning == "" {
		res.Warning = "CATVER NOT FOUND — ARCADE FILTERS LIMITED"
	}

	if cache != "" {
		if err := writeCache(cache, res); err != nil {
			res.Warning = joinWarnings(res.Warning, "CACHE SAVE FAILED")
			res.Diagnostic = "cache save: " + err.Error()
		}
	}

	return res, nil
}

func cacheFallback(saved, legacy *cacheFile, cfg config.ArcadeConfig, warning string, cause error) (Result, bool) {
	diagnostic := ""
	if cause != nil {
		diagnostic = cause.Error()
	}

	if saved != nil {
		return resultFromCache(*saved, cfg, warning, diagnostic), true
	}

	if legacy != nil {
		return resultFromCache(*legacy, cfg, "LEGACY CACHE — REFRESH REQUIRED", diagnostic), true
	}

	return Result{}, false
}

func resultFromCache(c cacheFile, cfg config.ArcadeConfig, warningOverride, diagnostic string) Result {
	r := c.Result
	warning := r.Warning
	if warningOverride != "" {
		warning = warningOverride
	}

	games := cloneGames(r.Games)
	hydrateLaunchConfig(games, cfg)

	return Result{
		Games:       games,
		MAMEVersion: r.MAMEVersion,
		ROMPaths:    append([]string(nil), r.ROMPaths...),
		CatVerUsed:  r.CatVerUsed,
		FromCache:   true,
		Warning:     warning,
		Diagnostic:  diagnostic,
	}
}

func cloneGames(in []library.Game) []library.Game {
	out := make([]library.Game, len(in))
	for i := range in {
		out[i] = in[i]
		out[i].Args = append([]string(nil), in[i].Args...)
	}
	return out
}

// ApplyLaunchConfig updates cached/in-memory Arcade entries with the current
// emulator executable and explicit ROM-path policy without requiring a rescan.
// Durable metadata remains cached; launch identity always follows current config.
func ApplyLaunchConfig(games []library.Game, cfg config.ArcadeConfig) {
	hydrateLaunchConfig(games, cfg)
}

func hydrateLaunchConfig(games []library.Game, cfg config.ArcadeConfig) {
	command := strings.TrimSpace(cfg.MAMECommand)
	if command == "" {
		command = "mame"
	}

	for i := range games {
		if games[i].Source != library.SourceArcade || strings.TrimSpace(games[i].ROM) == "" {
			continue
		}
		games[i].Command = command
		games[i].Args = LaunchArgs(games[i].ROM, cfg.ROMPaths, games[i].Vertical)
	}
}

// LaunchArgs returns MAME arguments for one indexed ROM. Explicit Marchine ROM
// paths are passed to MAME so a path that indexes successfully also launches
// successfully even when it is absent from mame.ini. Auto-rotation is based on
// MAME's own display orientation metadata, not category-name guesses.
func LaunchArgs(rom string, romPaths []string, vertical bool) []string {
	args := []string{rom}

	if paths := normalizeROMPathList(romPaths); len(paths) > 0 {
		args = append(args, "-rompath", strings.Join(paths, ";"))
	}

	if vertical {
		args = append(args, "-autorol")
	}

	return args
}

func normalizeROMPathList(paths []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))

	for _, p := range paths {
		p = config.ExpandPath(strings.TrimSpace(strings.Trim(p, `"`)))
		if p == "" {
			continue
		}
		p = filepath.Clean(p)
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}

	return out
}

func detectROMPaths(mamePath string) []string {
	seen := map[string]bool{}
	var out []string

	add := func(p string) {
		p = config.ExpandPath(strings.TrimSpace(strings.Trim(p, `"`)))
		if p == "" {
			return
		}

		if !filepath.IsAbs(p) {
			if wd, err := os.Getwd(); err == nil {
				p = filepath.Join(wd, p)
			}
		}

		p = filepath.Clean(p)

		if !seen[p] {
			seen[p] = true

			if st, err := os.Stat(p); err == nil && st.IsDir() {
				out = append(out, p)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), showConfigTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, mamePath, "-showconfig")
	if b, err := cmd.Output(); err == nil {
		s := bufio.NewScanner(strings.NewReader(string(b)))

		for s.Scan() {
			line := strings.TrimSpace(s.Text())

			if !strings.HasPrefix(line, "rompath") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 2 {
				for _, p := range strings.Split(strings.Join(fields[1:], " "), ";") {
					add(p)
				}
			}
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		for _, p := range []string{
			filepath.Join(home, ".mame", "roms"),
			filepath.Join(home, "mame", "roms"),
			filepath.Join(home, "MAME", "roms"),
			filepath.Join(home, "ROMs", "mame"),
			filepath.Join(home, "ROMs", "MAME"),
			filepath.Join(home, "Games", "MAME", "roms"),
		} {
			add(p)
		}
	}

	return out
}

func discoverROMs(paths []string) (map[string]bool, error) {
	roms := map[string]bool{}

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("read ROM path %s: %w", dir, err)
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".zip" || ext == ".7z" {
				stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
				roms[strings.ToLower(stem)] = true
			}
		}
	}

	return roms, nil
}

func parseInstalledMachines(
	mamePath string,
	roms map[string]bool,
	categories map[string]string,
	cfg config.ArcadeConfig,
) ([]library.Game, error) {
	if len(roms) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), listXMLTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, mamePath, "-listxml")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(stdout)
	var games []library.Game

	for {
		tok, err := dec.Token()

		if err == io.EOF {
			break
		}

		if err != nil {
			terminateProcess(cmd)
			if ctx.Err() != nil {
				return nil, fmt.Errorf("mame -listxml timed out after %s: %w", listXMLTimeout, ctx.Err())
			}
			return nil, fmt.Errorf("parse mame -listxml: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok || (se.Name.Local != "machine" && se.Name.Local != "game") {
			continue
		}

		var mx machineXML
		if err := dec.DecodeElement(&mx, &se); err != nil {
			terminateProcess(cmd)
			if ctx.Err() != nil {
				return nil, fmt.Errorf("mame -listxml timed out after %s: %w", listXMLTimeout, ctx.Err())
			}
			return nil, err
		}

		name := strings.ToLower(mx.Name)

		if !roms[name] {
			continue
		}

		if truthy(mx.IsBIOS) ||
			truthy(mx.IsDevice) ||
			strings.EqualFold(mx.Runnable, "no") {
			continue
		}

		if cfg.HidePreliminary &&
			strings.EqualFold(mx.Driver.Status, "preliminary") {
			continue
		}

		if cfg.HideClones &&
			mx.CloneOf != "" &&
			roms[strings.ToLower(mx.CloneOf)] {
			continue
		}

		cat := categories[name]

		// Keep the complete installed runnable library. "hide_junk" is a
		// presentation preference handled by the UI's ARCADE ONLY filter rather
		// than a destructive indexing rule.
		rawCategory := strings.TrimSpace(cat)
		browseCategory := BrowseCategory(rawCategory, mx.Description)
		nonArcade := IsNonArcade(rawCategory, mx.Description)
		vertical := isVerticalMachine(mx)

		games = append(games, library.Game{
			ID:           "mame:" + name,
			Name:         cleanDescription(mx.Description, name),
			Source:       library.SourceArcade,
			Year:         strings.TrimSpace(mx.Year),
			Manufacturer: cleanManufacturer(mx.Manufacturer),
			Category:     browseCategory,
			RawCategory:  rawCategory,
			NonArcade:    nonArcade,
			Vertical:     vertical,
			ROM:          name,
			Command:      mamePath,
			Args:         LaunchArgs(name, cfg.ROMPaths, vertical),
			CloneOf:      mx.CloneOf,
		})
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("mame -listxml timed out after %s: %w", listXMLTimeout, ctx.Err())
		}
		return nil, fmt.Errorf(
			"mame -listxml failed: %w: %s",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	sort.SliceStable(games, func(i, j int) bool {
		return strings.ToLower(games[i].Name) <
			strings.ToLower(games[j].Name)
	})

	return games, nil
}

func terminateProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func isVerticalMachine(mx machineXML) bool {
	for _, display := range mx.Displays {
		switch strings.TrimSpace(display.Rotate) {
		case "90", "270":
			return true
		}
	}
	return false
}

func truthy(v string) bool {
	return strings.EqualFold(v, "yes") ||
		strings.EqualFold(v, "true") ||
		v == "1"
}

func cleanDescription(s, fallback string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return fallback
	}

	return s
}

func cleanManufacturer(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return "—"
	}

	// Preserve MAME's local manufacturer/publisher string as the source of truth.
	return s
}

func readVersion(mamePath string) string {
	ctx, cancel := context.WithTimeout(context.Background(), versionTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, mamePath, "-version")

	b, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	line := strings.TrimSpace(string(b))

	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}

	return line
}

func detectCatVer(explicit string) string {
	candidates := []string{}

	if explicit != "" {
		candidates = append(candidates, explicit)
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(
			candidates,
			filepath.Join(home, ".mame", "catver.ini"),
			filepath.Join(home, ".attract", "metadata", "catver.ini"),
			filepath.Join(home, "MAME-Curation", "catver.ini"),
		)
	}

	candidates = append(
		candidates,
		"/usr/share/mame/catver.ini",
	)

	for _, p := range candidates {
		p = config.ExpandPath(p)

		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}

	return ""
}

func cachePath() (string, error) {
	d, err := config.CacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, "arcade.json"), nil
}

func readCache(path string) (cacheFile, error) {
	var c cacheFile

	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}

	if c.Schema != cacheSchema {
		return c, fmt.Errorf("%w: got %q, want %q", errCacheSchema, c.Schema, cacheSchema)
	}

	return c, nil
}

func writeCache(path string, r Result) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	c := cacheFile{
		Schema:  cacheSchema,
		SavedAt: time.Now().UTC(),
		Result: cachedResult{
			Games:       r.Games,
			MAMEVersion: r.MAMEVersion,
			ROMPaths:    r.ROMPaths,
			CatVerUsed:  r.CatVerUsed,
			Warning:     r.Warning,
		},
	}

	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	tmp, err := os.CreateTemp(dir, ".arcade-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if err = tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err = tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpName, path); err != nil {
		return err
	}

	// Best-effort directory sync makes the rename durable on filesystems that
	// support it. A failure here does not invalidate the already-atomic cache.
	if d, openErr := os.Open(dir); openErr == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

func joinWarnings(parts ...string) string {
	var out []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, " · ")
}
