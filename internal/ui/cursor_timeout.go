package ui

import (
	"bufio"
	"encoding/json"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A short inactivity timeout hides the pointer again after Hyprland makes it
// visible during a monitor transform, without using cursor.invisible. This is
// a runtime-only override owned only while the Marchine Foot window has focus.
const marchineCursorTimeout = 0.18

func parseCursorTimeout(data []byte) (float64, bool) {
	var value struct {
		Float *float64 `json:"float"`
	}
	if err := json.Unmarshal(data, &value); err != nil || value.Float == nil ||
		math.IsNaN(*value.Float) || math.IsInf(*value.Float, 0) || *value.Float < 0 {
		return 0, false
	}
	return *value.Float, true
}

func activeWindowIsMarchine(data []byte) bool {
	var window struct {
		InitialClass string `json:"initialClass"`
	}
	return json.Unmarshal(data, &window) == nil && strings.EqualFold(window.InitialClass, "marchine")
}

func cursorTimeoutEqual(a, b float64) bool { return math.Abs(a-b) < 0.001 }

func readCursorTimeout() (float64, bool) {
	data, err := exec.Command("hyprctl", "-j", "getoption", "cursor:inactive_timeout").Output()
	if err != nil {
		return 0, false
	}
	return parseCursorTimeout(data)
}

func setCursorTimeout(value float64) bool {
	arg := `hl.config({ ["cursor.inactive_timeout"] = ` + strconv.FormatFloat(value, 'f', -1, 64) + ` })`
	return exec.Command("hyprctl", "eval", arg).Run() == nil
}

func activeMarchineWindow() bool {
	data, err := exec.Command("hyprctl", "-j", "activewindow").Output()
	return err == nil && activeWindowIsMarchine(data)
}

func cursorFocusEvent(line string) bool {
	for _, prefix := range []string{"activewindow>>", "activewindowv2>>", "workspace>>", "workspacev2>>", "focusedmon>>", "focusedmonv2>>", "configreloaded>>"} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// StartMarchineCursorTimeout is disabled outside Hyprland. The temporary
// cursor.inactive_timeout is restored when Marchine loses focus, launches a
// game, or exits. We never set cursor.invisible or write Hyprland config files.
func StartMarchineCursorTimeout() func() {
	nop := func() {}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" || os.Getenv("XDG_RUNTIME_DIR") == "" {
		return nop
	}
	if _, err := exec.LookPath("hyprctl"); err != nil {
		return nop
	}

	original, ok := readCursorTimeout()
	if !ok || (original > 0 && original <= marchineCursorTimeout) {
		// Preserve an existing shorter timeout; it already does the job.
		return nop
	}

	socket := filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "hypr", os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"), ".socket2.sock")
	conn, err := net.DialTimeout("unix", socket, time.Second)
	if err != nil {
		return nop // Never set an option we cannot monitor and restore.
	}

	stopped := make(chan struct{})
	finished := make(chan struct{})
	events := make(chan struct{}, 1)
	readerStopped := make(chan struct{})
	go func() {
		defer close(readerStopped)
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			if cursorFocusEvent(scanner.Text()) {
				select {
				case events <- struct{}{}:
				default:
				}
			}
		}
	}()

	go func() {
		defer close(finished)
		owned := false
		restore := func() {
			if !owned {
				return
			}
			// Don't clobber a value changed by a user or another utility.
			if current, ok := readCursorTimeout(); ok && cursorTimeoutEqual(current, marchineCursorTimeout) {
				_ = setCursorTimeout(original)
			}
			owned = false
		}
		defer restore()

		refresh := func() {
			if GamepadSuspended() || !activeMarchineWindow() {
				restore()
				return
			}
			current, ok := readCursorTimeout()
			if !ok {
				return
			}
			if owned {
				if !cursorTimeoutEqual(current, marchineCursorTimeout) {
					// Another actor changed the timeout. Leave it alone.
					owned = false
				}
				return
			}
			if cursorTimeoutEqual(current, original) && setCursorTimeout(marchineCursorTimeout) {
				owned = true
			}
		}

		refresh()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopped:
				return
			case <-readerStopped:
				return
			case <-events:
				refresh()
			case <-ticker.C:
				refresh()
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(stopped)
			_ = conn.Close()
			<-finished
		})
	}
}
