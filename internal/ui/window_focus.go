package ui

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const marchineWindowSelector = `initialclass:^marchine$`

type hyprlandWindowState struct {
	InitialClass string `json:"initialClass"`
	Fullscreen   int    `json:"fullscreen"`
}

func marchineHyprlandFullscreen() bool {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return false
	}

	if _, err := exec.LookPath("hyprctl"); err != nil {
		return false
	}

	output, err := exec.Command("hyprctl", "-j", "activewindow").Output()
	if err != nil {
		return false
	}

	return parseMarchineHyprlandFullscreen(output)
}

func parseMarchineHyprlandFullscreen(output []byte) bool {
	var state hyprlandWindowState
	if err := json.Unmarshal(output, &state); err != nil {
		return false
	}

	return strings.EqualFold(state.InitialClass, "marchine") && state.Fullscreen >= 2
}

func restoreMarchineWindowCmd(restoreFullscreen bool) tea.Cmd {
	return func() tea.Msg {
		if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
			return nil
		}

		if _, err := exec.LookPath("hyprctl"); err != nil {
			return nil
		}

		// Restore compositor fullscreen while Marchine is still behind the game.
		// Focusing afterward avoids briefly exposing the underlying floating size.
		if restoreFullscreen {
			_ = exec.Command(
				"hyprctl",
				"dispatch",
				`hl.dsp.window.fullscreen_state({ window = "`+marchineWindowSelector+`", internal = 2, client = -1, action = "set" })`,
			).Run()
		}

		_ = exec.Command(
			"hyprctl",
			"dispatch",
			`hl.dsp.focus({ window = "`+marchineWindowSelector+`" })`,
		).Run()

		return nil
	}
}
