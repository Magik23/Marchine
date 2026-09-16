package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Theme         string       `toml:"theme"`
	StartupSource string       `toml:"startup_source"`
	Arcade        ArcadeConfig `toml:"arcade"`
	Games         []GameConfig `toml:"games"`
}

type ArcadeConfig struct {
	Enabled         bool     `toml:"enabled"`
	MAMECommand     string   `toml:"mame_command"`
	ROMPaths        []string `toml:"rom_paths"`
	CatVerPath      string   `toml:"catver_path"`
	HideClones      bool     `toml:"hide_clones"`
	HidePreliminary bool     `toml:"hide_preliminary"`
	HideJunk        bool     `toml:"hide_junk"`
}

type GameConfig struct {
	Name       string   `toml:"name"`
	Command    string   `toml:"command"`
	Args       []string `toml:"args"`
	WorkingDir string   `toml:"working_dir"`
}

func Default() Config {
	return Config{
		Theme:         "violet",
		StartupSource: "arcade",
		Arcade: ArcadeConfig{
			Enabled:         true,
			MAMECommand:     "mame",
			HideClones:      true,
			HidePreliminary: true,
			HideJunk:        true,
		},
	}
}

func ConfigPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "marchine", "config.toml"), nil
}

func CacheDir() (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "marchine"), nil
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		var err error
		path, err = ConfigPath()
		if err != nil {
			return cfg, err
		}
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.Normalize()
	return cfg, nil
}

func (c *Config) Normalize() {
	if c.Theme == "" {
		c.Theme = "violet"
	}
	if c.StartupSource == "" {
		c.StartupSource = "arcade"
	}
	if c.Arcade.MAMECommand == "" {
		c.Arcade.MAMECommand = "mame"
	}
	c.Arcade.MAMECommand = ExpandPath(c.Arcade.MAMECommand)
	c.Arcade.CatVerPath = ExpandPath(c.Arcade.CatVerPath)
	for i := range c.Arcade.ROMPaths {
		c.Arcade.ROMPaths[i] = ExpandPath(c.Arcade.ROMPaths[i])
	}
	for i := range c.Games {
		c.Games[i].Command = ExpandPath(c.Games[i].Command)
		c.Games[i].WorkingDir = ExpandPath(c.Games[i].WorkingDir)
		for j := range c.Games[i].Args {
			c.Games[i].Args[j] = ExpandPath(c.Games[i].Args[j])
		}
	}
}

func ExpandPath(s string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(s, "~/"))
		}
	}
	return os.ExpandEnv(s)
}

// Save writes the complete Marchine configuration atomically. The parent
// directory is created if necessary. Marchine keeps display identity separate
// from launch identity, so saving never renames ROMs, executables, or games.
func Save(path string, cfg Config) error {
	if path == "" {
		var err error
		path, err = ConfigPath()
		if err != nil {
			return err
		}
	}
	cfg.Normalize()
	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	committed = true

	// Best-effort directory sync makes the rename durable on filesystems that
	// support it. The successfully renamed config remains valid if this sync is
	// unsupported by the filesystem.
	if d, openErr := os.Open(dir); openErr == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}
