//go:build linux

package gamepad

import (
	"fmt"

	evdev "github.com/holoplot/go-evdev"
	"golang.org/x/sys/unix"
)

type Reader struct {
	device     Device
	input      *evdev.InputDevice
	normalizer *Normalizer
}

func OpenReader(device Device) (*Reader, error) {
	input, err := evdev.OpenWithFlags(device.Path, unix.O_RDONLY)
	if err != nil {
		return nil, fmt.Errorf("open controller %s: %w", device.Path, err)
	}

	absInfo, err := input.AbsInfos()
	if err != nil {
		_ = input.Close()
		return nil, fmt.Errorf("read controller axis info %s: %w", device.Path, err)
	}

	if err := input.NonBlock(); err != nil {
		_ = input.Close()
		return nil, fmt.Errorf("configure controller %s: %w", device.Path, err)
	}

	return &Reader{
		device:     device,
		input:      input,
		normalizer: NewNormalizer(absInfo),
	}, nil
}

func (r *Reader) Device() Device {
	return r.device
}

func (r *Reader) ReadSignal() (Signal, error) {
	for {
		event, err := r.input.ReadOne()
		if err != nil {
			return Signal{}, fmt.Errorf(
				"read controller %s: %w",
				r.device.Path,
				err,
			)
		}

		signal, ok := r.normalizer.HandleSignal(*event)
		if ok {
			return signal, nil
		}
	}
}

func (r *Reader) ReadAction() (Action, error) {
	for {
		signal, err := r.ReadSignal()
		if err != nil {
			return ActionNone, err
		}

		if signal.Pressed {
			return signal.Action, nil
		}
	}
}

func (r *Reader) Close() error {
	return r.input.Close()
}
