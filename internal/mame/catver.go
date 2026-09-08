package mame

import (
	"bufio"
	"os"
	"strings"
)

// ParseCatVer reads category metadata from a CatVer-style INI file.
//
// CatVer files commonly contain multiple sections:
//
//	[Category]
//	1942=Shooter / Flying Vertical
//
//	[CatVer]
//	1942=0.34b4
//
// The [CatVer] section is NOT category metadata. It records the MAME
// version in which a machine was added. Marchine must therefore never
// merge it into the category map.
//
// Some compatible metadata files use [Genre] instead of [Category].
// Marchine supports that as a fallback, but always prefers [Category]
// when it is present.
func ParseCatVer(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	categories := map[string]string{}
	genres := map[string]string{}

	section := ""

	s := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	s.Buffer(buf, 2*1024*1024)

	for s.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(s.Text(), "\ufeff"))

		if line == "" ||
			strings.HasPrefix(line, ";") ||
			strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(
				strings.TrimSpace(strings.Trim(line, "[]")),
			)
			continue
		}

		// IMPORTANT:
		// [CatVer] contains MAME version numbers, not categories.
		if section != "category" && section != "genre" {
			continue
		}

		left, right, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		name := strings.ToLower(strings.TrimSpace(left))
		value := strings.TrimSpace(right)

		if name == "" || value == "" {
			continue
		}

		switch section {
		case "category":
			categories[name] = value

		case "genre":
			genres[name] = value
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	// Real CatVer [Category] metadata always wins.
	if len(categories) > 0 {
		return categories, nil
	}

	// Support Genre-style metadata files as a fallback.
	return genres, nil
}

// BrowseCategory turns CatVer's very granular strings into a stable set of
// launcher categories. The raw CatVer value is still retained on each Game,
// so Marchine never throws away the local metadata it indexed.
func BrowseCategory(cat, name string) string {
	c := strings.ToLower(strings.TrimSpace(cat))
	n := strings.ToLower(strings.TrimSpace(name))
	both := c + " " + n

	switch {
	case strings.Contains(both, "mahjong"):
		return "MAHJONG"

	case containsAny(
		both,
		"casino",
		"gambling",
		"slot machine",
		"slots",
		"fruit machine",
		"video poker",
		"poker machine",
		"lottery",
	):
		return "CASINO / GAMBLING"

	case containsAny(both, "pachinko", "medal game"):
		return "MEDAL / PACHINKO"

	case strings.Contains(both, "hanafuda"):
		return "HANAFUDA / CARD"

	case containsAny(c, "utility", "utilities", "electromechanical"):
		return "OTHER / UTILITY"

	case strings.Contains(c, "shooter / flying vertical"),
		strings.Contains(c, "shooter / flying horizontal"),
		strings.Contains(c, "shooter / flying"):
		return "SHMUPS"

	case strings.HasPrefix(c, "fighter / versus"),
		strings.Contains(c, "fighter / 2d"),
		strings.Contains(c, "fighter / 3d"):
		return "FIGHTING"

	case strings.Contains(c, "fighter / 2.5d"),
		strings.Contains(c, "fighter / horizontal"),
		strings.Contains(c, "fighter / vertical"),
		strings.Contains(c, "beat em up"),
		strings.Contains(c, "beat 'em up"):
		return "BEAT'EM UP"

	case strings.Contains(c, "shooter / walking"),
		strings.Contains(c, "run and gun"),
		strings.Contains(c, "run & gun"):
		return "RUN&GUN"

	case strings.Contains(c, "platform"):
		return "PLATFORM"

	case strings.Contains(c, "maze"):
		return "MAZE"

	case strings.HasPrefix(c, "puzzle"),
		strings.Contains(c, " / puzzle"):
		return "PUZZLE"

	case strings.HasPrefix(c, "driving /"),
		strings.Contains(c, "racing"),
		strings.Contains(c, "race"):
		return "RACING"

	case strings.Contains(c, "sports"):
		return "SPORTS"

	case containsAny(c, "rhythm", "music", "dance"):
		return "RHYTHM"

	case containsAny(c, "lightgun", "light gun", "shooter / gun"):
		return "LIGHTGUN"

	case strings.HasPrefix(c, "shooter"):
		return "SHOOTER"

	case containsAny(c, "ball & paddle", "ball and paddle", "breakout"):
		return "BALL & PADDLE"

	case strings.HasPrefix(c, "action"):
		return "ACTION"

	case strings.Contains(c, "quiz"):
		return "QUIZ"

	case strings.Contains(c, "pinball"):
		return "PINBALL"

	case containsAny(
		c,
		"mini-games",
		"minigames",
		"multi-game",
		"multigame",
	):
		return "MINIGAMES"

	case containsAny(c, "tabletop", "board game", "card game", "cards"):
		return "TABLETOP / CARD"

	case c == "":
		return "UNCLASSIFIED"
	}

	// CatVer grows over time. Unknown families are still exposed instead of
	// disappearing: use the top-level family before the slash as a category.
	family := strings.TrimSpace(strings.SplitN(cat, "/", 2)[0])
	if family == "" {
		return "UNCLASSIFIED"
	}

	return strings.ToUpper(family)
}

// CoarseCategory is kept for compatibility with older callers/tests.
func CoarseCategory(cat string) string {
	return BrowseCategory(cat, "")
}

func IsNonArcade(cat, name string) bool {
	c := strings.ToLower(strings.TrimSpace(cat))
	n := strings.ToLower(strings.TrimSpace(name))

	// Category metadata is the strongest signal. Beyond the obvious gambling
	// families, MAME/CatVer can include home systems, computers, handhelds,
	// utilities and device-like software in the same installed ROM tree.
	if containsAny(
		c,
		"mahjong",
		"gambling",
		"casino",
		"lottery",
		"slot machine",
		"slots",
		"medal game",
		"pachinko",
		"hanafuda",
		"electromechanical",
		"utility",
		"utilities",
		"fruit machine",
		"video poker",
		"poker machine",
		"computer",
		"console",
		"handheld",
		"calculator",
		"printer",
		"terminal",
		"educational",
		"home system",
		"home computer",
		"chess machine",
	) {
		return true
	}

	// Descriptions are only used for high-confidence nuisance families. Avoid
	// broad words like "computer" here because they can legitimately occur in
	// an arcade game's title.
	return containsAny(
		n,
		"mahjong",
		"video poker",
		"slot machine",
		"fruit machine",
		"pachinko",
		"hanafuda",
		"casino",
		"lottery",
	)
}

func IsJunkCategory(cat string) bool {
	return IsNonArcade(cat, "")
}

func IsJunkName(name string) bool {
	return IsNonArcade("", name)
}

func containsAny(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}

	return false
}
