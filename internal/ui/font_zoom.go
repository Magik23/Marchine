package ui

import (
	"fmt"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Foot does not expose a font-zoom terminal escape sequence. For Marchine's
// Hyprland desktop window, send Foot its *existing* Control+=/- shortcuts.
// Check the active window inside the compositor so a background controller
// cannot zoom an unrelated application. Nothing persists to any config file.
type footFontZoomFinishedMsg struct{ Err error }

func fontZoomLua(increase bool) string {
	key := "minus"
	if increase {
		key = "equal"
	}
	return fmt.Sprintf(`local w = hl.get_active_window()
if w ~= nil and w.class == "marchine" then
    hl.dispatch(hl.dsp.send_shortcut({ mods = "CTRL", key = %q }))
end`, key)
}

func footFontZoomCmd(increase bool) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("hyprctl", "eval", fontZoomLua(increase)).CombinedOutput()
		if err != nil {
			return footFontZoomFinishedMsg{Err: fmt.Errorf("hyprctl: %w: %s", err, strings.TrimSpace(string(out)))}
		}
		return footFontZoomFinishedMsg{}
	}
}
