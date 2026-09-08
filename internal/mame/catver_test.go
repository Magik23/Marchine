package mame

import "testing"

func TestBrowseCategory(t *testing.T) {
	cases := map[string]string{
		"Shooter / Flying Vertical":   "SHMUPS",
		"Shooter / Flying Horizontal": "SHMUPS",
		"Fighter / Versus":            "FIGHTING",
		"Fighter / 2.5D":              "BEAT'EM UP",
		"Shooter / Walking":           "RUN&GUN",
		"Driving / Race":              "RACING",
		"Puzzle / Drop":               "PUZZLE",
		"Platform / Run Jump":         "PLATFORM",
		"Sports / Baseball":           "SPORTS",
		"Maze / Collect":              "MAZE",
		"Casino / Cards":              "CASINO / GAMBLING",
		"Mahjong / Tiles":             "MAHJONG",
		"Utilities / Test":            "OTHER / UTILITY",
	}
	for in, want := range cases {
		if got := BrowseCategory(in, ""); got != want {
			t.Fatalf("BrowseCategory(%q)=%q want %q", in, got, want)
		}
	}
}

func TestNonArcadeClassification(t *testing.T) {
	for _, cat := range []string{"Casino / Cards", "Mahjong / Tiles", "Medal Game", "Electromechanical", "Computer / Home", "Console / System", "Handheld / Electronic"} {
		if !IsNonArcade(cat, "") {
			t.Fatalf("expected %q to be non-arcade/junk", cat)
		}
	}
	if IsNonArcade("Shooter / Flying Vertical", "Metal Slug") {
		t.Fatal("did not expect normal arcade game to be non-arcade")
	}
	if !IsJunkName("Super Mahjong 2000") {
		t.Fatal("legacy IsJunkName compatibility should still detect mahjong")
	}
}
