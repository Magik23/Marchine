//go:build linux

package gamepad

import (
	"sync"
	"time"
)

const (
	rescanInterval = time.Second
	repeatTick     = 20 * time.Millisecond
)

type Event struct {
	Device Device
	Action Action
}

type Manager struct {
	events chan Event
	stop   chan struct{}
	done   chan struct{}

	closeOnce sync.Once
	readers   sync.WaitGroup
}

func NewManager() *Manager {
	m := &Manager{
		events: make(chan Event, 32),
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}

	go m.run()

	return m
}

func (m *Manager) Events() <-chan Event {
	return m.events
}

func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		close(m.stop)
	})

	<-m.done
}

type readerExit struct {
	reader *Reader
}

type inputSignal struct {
	reader *Reader
	signal Signal
}

type heldKey struct {
	path   string
	action Action
}

type holdState struct {
	device  Device
	started time.Time
	next    time.Time
}

func (m *Manager) run() {
	defer close(m.done)

	scanTicker := time.NewTicker(rescanInterval)
	defer scanTicker.Stop()

	repeatTicker := time.NewTicker(repeatTick)
	defer repeatTicker.Stop()

	live := make(map[string]*Reader)
	held := make(map[heldKey]holdState)

	gone := make(chan readerExit, 64)
	signals := make(chan inputSignal, 64)

	clearHeld := func(path string) {
		for key := range held {
			if key.path == path {
				delete(held, key)
			}
		}
	}

	emit := func(device Device, action Action) bool {
		select {
		case m.events <- Event{
			Device: device,
			Action: action,
		}:
			return true

		case <-m.stop:
			return false
		}
	}

	scan := func() {
		devices, err := Discover()
		if err != nil {
			return
		}

		seen := make(map[string]bool)

		for _, device := range devices {
			seen[device.Path] = true

			if _, exists := live[device.Path]; exists {
				continue
			}

			reader, err := OpenReader(device)
			if err != nil {
				continue
			}

			live[device.Path] = reader

			m.readers.Add(1)
			go m.readLoop(reader, signals, gone)
		}

		for path, reader := range live {
			if seen[path] {
				continue
			}

			_ = reader.Close()
			delete(live, path)
			clearHeld(path)
		}
	}

	scan()

	for {
		select {
		case <-m.stop:
			for path, reader := range live {
				_ = reader.Close()
				delete(live, path)
			}

			m.readers.Wait()
			close(m.events)
			return

		case exited := <-gone:
			path := exited.reader.Device().Path

			current, exists := live[path]
			if exists && current == exited.reader {
				_ = current.Close()
				delete(live, path)
			}

			clearHeld(path)

		case input := <-signals:
			device := input.reader.Device()
			action := input.signal.Action

			key := heldKey{
				path:   device.Path,
				action: action,
			}

			if input.signal.Pressed == false {
				delete(held, key)
				continue
			}

			if repeatable(action) {
				opposite := repeatOpposite(action)
				if opposite != ActionNone {
					delete(held, heldKey{
						path:   device.Path,
						action: opposite,
					})
				}

				now := time.Now()

				held[key] = holdState{
					device:  device,
					started: now,
					next:    now.Add(initialRepeatDelay(action)),
				}
			}

			if emit(device, action) == false {
				return
			}

		case now := <-repeatTicker.C:
			for key, state := range held {
				if now.Before(state.next) {
					continue
				}

				if emit(state.device, key.action) == false {
					return
				}

				state.next = now.Add(
					repeatInterval(
						key.action,
						now.Sub(state.started),
					),
				)

				held[key] = state
			}

		case <-scanTicker.C:
			scan()
		}
	}
}

func (m *Manager) readLoop(
	reader *Reader,
	signals chan<- inputSignal,
	gone chan<- readerExit,
) {
	defer m.readers.Done()

	for {
		signal, err := reader.ReadSignal()
		if err != nil {
			select {
			case gone <- readerExit{reader: reader}:
			case <-m.stop:
			}

			return
		}

		select {
		case signals <- inputSignal{
			reader: reader,
			signal: signal,
		}:

		case <-m.stop:
			return
		}
	}
}

func repeatable(action Action) bool {
	switch action {
	case ActionUp,
		ActionDown,
		ActionPreviousLetter,
		ActionNextLetter:
		return true

	default:
		return false
	}
}

func repeatOpposite(action Action) Action {
	switch action {
	case ActionUp:
		return ActionDown

	case ActionDown:
		return ActionUp

	case ActionPreviousLetter:
		return ActionNextLetter

	case ActionNextLetter:
		return ActionPreviousLetter

	default:
		return ActionNone
	}
}

func initialRepeatDelay(action Action) time.Duration {
	switch action {
	case ActionPreviousLetter, ActionNextLetter:
		return 350 * time.Millisecond

	default:
		return 300 * time.Millisecond
	}
}

func repeatInterval(action Action, heldFor time.Duration) time.Duration {
	switch action {
	case ActionPreviousLetter, ActionNextLetter:
		switch {
		case heldFor < time.Second:
			return 220 * time.Millisecond

		case heldFor < 2*time.Second:
			return 140 * time.Millisecond

		default:
			return 85 * time.Millisecond
		}

	default:
		switch {
		case heldFor < time.Second:
			return 110 * time.Millisecond

		case heldFor < 2*time.Second:
			return 70 * time.Millisecond

		default:
			return 45 * time.Millisecond
		}
	}
}
