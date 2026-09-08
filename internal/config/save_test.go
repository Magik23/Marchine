package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveCreatesConfig(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "nested", "config.toml")
	cfg := Default()
	cfg.Arcade.MAMECommand = "/usr/bin/mame"
	cfg.Arcade.ROMPaths = []string{"/roms/mame"}
	if err := Save(p, cfg); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "/usr/bin/mame") || !strings.Contains(s, "/roms/mame") {
		t.Fatalf("saved config missing arcade paths: %s", s)
	}
}
