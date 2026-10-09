//go:build linux

package gamepad

import (
	"fmt"
	"sort"

	evdev "github.com/holoplot/go-evdev"
	"golang.org/x/sys/unix"
)

// Device describes a readable Linux input device that looks like a gamepad.
type Device struct {
	Path string
	Name string
}

// Discover finds readable game controllers by capabilities rather than
// controller name, brand, vendor ID, or device path.
func Discover() ([]Device, error) {
	paths, err := evdev.ListDevicePaths()
	if err != nil {
		return nil, fmt.Errorf("list input devices: %w", err)
	}

	devices := make([]Device, 0)

	for _, candidate := range paths {
		dev, err := evdev.OpenWithFlags(
			candidate.Path,
			unix.O_RDONLY|unix.O_NONBLOCK,
		)
		if err != nil {
			continue
		}

		isPad := isGamepad(dev)
		_ = dev.Close()

		if isPad {
			devices = append(devices, Device{
				Path: candidate.Path,
				Name: candidate.Name,
			})
		}
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Path < devices[j].Path
	})

	return devices, nil
}

func isGamepad(dev *evdev.InputDevice) bool {
	keys := eventSet(dev.CapableEvents(evdev.EV_KEY))
	axes := eventSet(dev.CapableEvents(evdev.EV_ABS))

	hasFaceButton :=
		hasCode(keys, evdev.BTN_SOUTH) ||
			hasCode(keys, evdev.BTN_EAST) ||
			hasCode(keys, evdev.BTN_NORTH) ||
			hasCode(keys, evdev.BTN_WEST)

	hasDPad :=
		hasCode(axes, evdev.ABS_HAT0X) &&
			hasCode(axes, evdev.ABS_HAT0Y)

	hasLeftStick :=
		hasCode(axes, evdev.ABS_X) &&
			hasCode(axes, evdev.ABS_Y)

	return hasFaceButton && (hasDPad || hasLeftStick)
}

func eventSet(codes []evdev.EvCode) map[evdev.EvCode]struct{} {
	set := make(map[evdev.EvCode]struct{}, len(codes))

	for _, code := range codes {
		set[code] = struct{}{}
	}

	return set
}

func hasCode(set map[evdev.EvCode]struct{}, code evdev.EvCode) bool {
	_, ok := set[code]
	return ok
}
