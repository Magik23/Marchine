package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Name       string
	Accent     color.Color
	Accent2    color.Color
	Text       color.Color
	Muted      color.Color
	Border     color.Color
	Success    color.Color
	Warning    color.Color
	Selection  color.Color
	SelectText color.Color
}

var themeOrder = []string{"violet", "cyan", "green", "amber"}

func ThemeByName(name string) Theme {
	switch strings.ToLower(name) {
	case "cyan":
		return Theme{
			Name: "cyan", Accent: lipgloss.Color("#2DE2E6"), Accent2: lipgloss.Color("#6AE4FF"),
			Text: lipgloss.Color("#D7F8FF"), Muted: lipgloss.Color("#6E9AA6"), Border: lipgloss.Color("#188FA3"),
			Success: lipgloss.Color("#62F5A8"), Warning: lipgloss.Color("#FFD166"), Selection: lipgloss.Color("#2DE2E6"),
			SelectText: lipgloss.Color("#061014"),
		}
	case "green":
		return Theme{
			Name: "green", Accent: lipgloss.Color("#5CFF8D"), Accent2: lipgloss.Color("#2FD675"),
			Text: lipgloss.Color("#D8FFE4"), Muted: lipgloss.Color("#70A77E"), Border: lipgloss.Color("#2B8A4B"),
			Success: lipgloss.Color("#8BFFB2"), Warning: lipgloss.Color("#FFD166"), Selection: lipgloss.Color("#5CFF8D"),
			SelectText: lipgloss.Color("#06100A"),
		}
	case "amber":
		return Theme{
			Name: "amber", Accent: lipgloss.Color("#FFB83E"), Accent2: lipgloss.Color("#E8922E"),
			Text: lipgloss.Color("#FFE7B5"), Muted: lipgloss.Color("#A88958"), Border: lipgloss.Color("#9B661D"),
			Success: lipgloss.Color("#B7F774"), Warning: lipgloss.Color("#FF8A5B"), Selection: lipgloss.Color("#FFB83E"),
			SelectText: lipgloss.Color("#140D03"),
		}
	default:
		return Theme{
			Name: "violet", Accent: lipgloss.Color("#F03CFF"), Accent2: lipgloss.Color("#8D7CFF"),
			Text: lipgloss.Color("#E4DEFF"), Muted: lipgloss.Color("#8880B7"), Border: lipgloss.Color("#7A4FCC"),
			Success: lipgloss.Color("#73F8B0"), Warning: lipgloss.Color("#FFD166"), Selection: lipgloss.Color("#E83BFF"),
			SelectText: lipgloss.Color("#110717"),
		}
	}
}

func NextTheme(current string) Theme {
	idx := 0
	for i, n := range themeOrder {
		if n == strings.ToLower(current) {
			idx = i
			break
		}
	}
	return ThemeByName(themeOrder[(idx+1)%len(themeOrder)])
}
