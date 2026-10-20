package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGameIdentityLegacyMigrationWithoutSurpriseWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	legacy := "# preserve custom comment until save\n[[games]]\nname = 'Alpha'\ncommand = '/bin/true'\n[[games]]\nname = 'Beta'\ncommand = '/bin/false'\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	id0, id1 := cfg.Games[0].ID, cfg.Games[1].ID
	if id0 == "" || id1 == "" || id0 == id1 {
		t.Fatalf("invalid legacy IDs: %q %q", id0, id1)
	}
	after, _ := os.ReadFile(path)
	if string(after) != legacy {
		t.Fatal("loading legacy config modified the user's file")
	}
	cfg.Games[0].Name = "Alpha Renamed"
	cfg.Games[0].Command = "/usr/bin/true"
	cfg.Games = cfg.Games[1:] // Removing the first entry must not change the survivor's identity.
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Games) != 1 || loaded.Games[0].ID != id1 {
		t.Fatalf("identity shifted after save: %#v", loaded.Games)
	}
	if !strings.Contains(stringMustRead(t, path), "id =") {
		t.Fatal("stable identity was not persisted")
	}
}

func stringMustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestGameIdentityHandlesDuplicateAndMissingIDs(t *testing.T) {
	cfg := Default()
	cfg.Games = []GameConfig{
		{ID: "manual", Name: "First", Command: "/bin/true"},
		{ID: "manual", Name: "Second", Command: "/bin/true"},
		{Name: "Third", Command: "/bin/true"},
	}
	if err := cfg.EnsureGameIDs(); err != nil {
		t.Fatal(err)
	}
	if cfg.Games[0].ID != "manual" {
		t.Fatal("valid game identity overwritten")
	}
	seen := map[string]bool{}
	for _, g := range cfg.Games {
		if g.ID == "" || seen[g.ID] {
			t.Fatalf("duplicate or empty identity: %#v", cfg.Games)
		}
		seen[g.ID] = true
	}
	saved := cfg.Games[1].ID
	if err := cfg.EnsureGameIDs(); err != nil {
		t.Fatal(err)
	}
	if cfg.Games[1].ID != saved {
		t.Fatal("assigned IDs did not remain stable")
	}
}
