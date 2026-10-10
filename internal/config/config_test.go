package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadIgnoresRemovedAnimateSculptureKey(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "config.toml")
	contents := "theme = \"green\"\nanimate_sculpture = true\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("legacy config key should remain forward-compatible: %v", err)
	}
	if cfg.Theme != "green" {
		t.Fatalf("theme=%q", cfg.Theme)
	}
}

func TestLoadIgnoresAndDropsRemovedFullscreenKey(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "config.toml")

	contents := "theme = \"green\"\nfullscreen = true\n"
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("removed fullscreen key should remain forward-compatible: %v", err)
	}

	if cfg.Theme != "green" {
		t.Fatalf("theme=%q", cfg.Theme)
	}

	if err := Save(p, cfg); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(saved), "fullscreen") {
		t.Fatalf("removed fullscreen key survived config save: %q", string(saved))
	}
}
