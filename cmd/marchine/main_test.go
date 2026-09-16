package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magik23/marchine/internal/config"
)

func TestEnsureStarterConfigCreatesAndPreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")

	if err := ensureStarterConfig(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "MARCHINE — Game Machinery.") {
		t.Fatalf("starter config missing expected header: %q", string(b))
	}

	const custom = "theme = \"amber\"\n"
	if err := os.WriteFile(path, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureStarterConfig(path); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != custom {
		t.Fatalf("existing config was overwritten: %q", string(b))
	}
}

func TestRunDoctorAllowsDisabledArcade(t *testing.T) {
	cfg := config.Default()
	cfg.Arcade.Enabled = false
	if err := runDoctor(cfg, false, filepath.Join(t.TempDir(), "config.toml")); err != nil {
		t.Fatalf("doctor returned error for disabled Arcade: %v", err)
	}
}

func TestRunDoctorFailsOnOperationalDiagnostic(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	cfg := config.Default()
	cfg.Arcade.MAMECommand = filepath.Join(t.TempDir(), "definitely-missing-mame")
	if err := runDoctor(cfg, true, filepath.Join(t.TempDir(), "config.toml")); err == nil {
		t.Fatal("doctor should return an error when MAME cannot be found and no cache exists")
	}
}
