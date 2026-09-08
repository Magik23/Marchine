package mame

import (
	"bufio"
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

type Result struct {
	Games       []library.Game
	MAMEVersion string
	ROMPaths    []string
	CatVerUsed  string
	FromCache   bool
	Warning     string
}

type cacheFile struct {
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
}

func Load(cfg config.ArcadeConfig, force bool) (Result, error) {
	if !cfg.Enabled {
		return Result{}, nil
	}

	// Normal startup is cache-first by design. If a saved Arcade library exists,
	// load it immediately without touching MAME, ROM paths, or CatVer.
	//
	// A fresh scan only happens when:
	//   - no saved cache exists yet, or
	//   - force is true (R in Marchine / --rescan on the CLI).
	cache, _ := cachePath()
	var saved cacheFile
	hasSaved := false

	if cache != "" {
		if c, err := readCache(cache); err == nil {
			saved = c
			hasSaved = true

			if !force {
				return resultFromCache(c, ""), nil
			}
		}
	}

	mamePath, err := exec.LookPath(cfg.MAMECommand)
	if err != nil {
		if hasSaved {
			return resultFromCache(saved, "MAME NOT FOUND — SAVED LIBRARY KEPT"), nil
		}
		return Result{Warning: "MAME NOT FOUND"}, nil
	}

	romPaths := cfg.ROMPaths
	if len(romPaths) == 0 {
		romPaths = detectROMPaths(mamePath)
	}

	roms, err := discoverROMs(romPaths)
	if err != nil {
		return Result{}, err
	}

	// Do not destroy a good saved library because an external drive is
	// temporarily missing or a configured ROM path is unavailable.
	if len(roms) == 0 {
		if hasSaved {
			return resultFromCache(saved, "NO ROMS FOUND — SAVED LIBRARY KEPT"), nil
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
		_ = writeCache(cache, res)
	}

	return res, nil
}

func resultFromCache(c cacheFile, warningOverride string) Result {
	r := c.Result
	warning := r.Warning
	if warningOverride != "" {
		warning = warningOverride
	}

	return Result{
		Games:       r.Games,
		MAMEVersion: r.MAMEVersion,
		ROMPaths:    r.ROMPaths,
		CatVerUsed:  r.CatVerUsed,
		FromCache:   true,
		Warning:     warning,
	}
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

	cmd := exec.Command(mamePath, "-showconfig")
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

	cmd := exec.Command(mamePath, "-listxml")

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
			_ = cmd.Process.Kill()
			return nil, fmt.Errorf("parse mame -listxml: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok || (se.Name.Local != "machine" && se.Name.Local != "game") {
			continue
		}

		var mx machineXML
		if err := dec.DecodeElement(&mx, &se); err != nil {
			_ = cmd.Process.Kill()
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

		// Keep the complete installed runnable library. "hide_junk" is now a
		// presentation preference handled by the UI's ARCADE ONLY filter rather
		// than a destructive indexing rule. This keeps casino/mahjong/etc.
		// available as explicit categories when the user wants to inspect them.
		rawCategory := strings.TrimSpace(cat)
		browseCategory := BrowseCategory(rawCategory, mx.Description)
		nonArcade := IsNonArcade(rawCategory, mx.Description)

		games = append(games, library.Game{
			ID:           "mame:" + name,
			Name:         cleanDescription(mx.Description, name),
			Source:       library.SourceArcade,
			Year:         strings.TrimSpace(mx.Year),
			Manufacturer: cleanManufacturer(mx.Manufacturer),
			Category:     browseCategory,
			RawCategory:  rawCategory,
			NonArcade:    nonArcade,
			ROM:          name,
			Command:      mamePath,
			Args:         []string{name},
			CloneOf:      mx.CloneOf,
		})
	}

	if err := cmd.Wait(); err != nil {
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
	cmd := exec.Command(mamePath, "-version")

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

	err = json.Unmarshal(b, &c)
	return c, err
}

func writeCache(path string, r Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	c := cacheFile{
		SavedAt: time.Now(),
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

	return os.WriteFile(path, b, 0o644)
}
