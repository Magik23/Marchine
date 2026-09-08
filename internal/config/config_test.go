package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPreservesDefaults(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "config.toml")
	if err := os.WriteFile(p, []byte("theme = \"amber\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme != "amber" {
		t.Fatalf("theme=%q", cfg.Theme)
	}
	if cfg.Arcade.MAMECommand != "mame" {
		t.Fatalf("mame command default lost: %q", cfg.Arcade.MAMECommand)
	}
}
