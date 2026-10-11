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

// View is a modifier for both Quit and Rotate Display. The standalone
// View/Menu/RB actions fire on release, so neither part of a chord leaks
// through as a source switch, Credits screen, or Theme change.
//
// View + RB is an OPTIONAL EXTERNAL-UTILITY shortcut, not a normal Marchine
// controller binding. To change/remove it, review handle() and chordButton()
// together; changing the physical RB mapping alone is not sufficient.
// Preserve the independent View + Menu Quit chord and release suppression.
type shortcutChordState struct {
	viewPressed  bool
	menuPressed  bool
	themePressed bool
	consumed     bool
}

func (s *shortcutChordState) handle(action Action, pressed bool) (Action, bool) {
	var wasPressed bool
	switch action {
	case ActionView:
		wasPressed = s.viewPressed
		s.viewPressed = pressed
	case ActionMenu:
		wasPressed = s.menuPressed
		s.menuPressed = pressed
	case ActionTheme:
		wasPressed = s.themePressed
		s.themePressed = pressed
	default:
		return ActionNone, false
	}

	if pressed {
		if s.viewPressed {
			switch {
			case s.menuPressed && !s.consumed:
				s.consumed = true
				return ActionQuit, true
			case s.themePressed && !s.menuPressed && (!s.consumed || !wasPressed):
				// While View remains held, each fresh RB press can rotate
				// again without firing the standalone Theme action.
				s.consumed = true
				return ActionRotateDisplay, true
			}
		}
		return ActionNone, false
	}

	standalone := wasPressed && !s.consumed
	if !s.viewPressed && !s.menuPressed && !s.themePressed {
		s.consumed = false
	}
	if standalone {
		return action, true
	}
	return ActionNone, false
}

func (s shortcutChordState) idle() bool {
	return !s.viewPressed && !s.menuPressed && !s.themePressed
}

func chordButton(action Action) bool {
	return action == ActionView || action == ActionMenu || action == ActionTheme
}

func (m *Manager) run() {
	defer close(m.done)

	scanTicker := time.NewTicker(rescanInterval)
	defer scanTicker.Stop()

	repeatTicker := time.NewTicker(repeatTick)
	defer repeatTicker.Stop()

	live := make(map[string]*Reader)
	held := make(map[heldKey]holdState)
	chords := make(map[string]*shortcutChordState)

	gone := make(chan readerExit, 64)
	signals := make(chan inputSignal, 64)

	clearHeld := func(path string) {
		for key := range held {
			if key.path == path {
				delete(held, key)
			}
		}

		delete(chords, path)
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

			if chordButton(action) {
				state := chords[device.Path]
				if state == nil {
					state = &shortcutChordState{}
					chords[device.Path] = state
				}

				output, ok := state.handle(action, input.signal.Pressed)
				if state.idle() {
					delete(chords, device.Path)
				}

				if ok && !emit(device, output) {
					return
				}
				continue
			}

			key := heldKey{
				path:   device.Path,
				action: action,
			}

			if !input.signal.Pressed {
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

			if !emit(device, action) {
				return
			}

		case now := <-repeatTicker.C:
			for key, state := range held {
				if now.Before(state.next) {
					continue
				}

				if !emit(state.device, key.action) {
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
		ActionLeft,
		ActionRight,
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

	case ActionLeft:
		return ActionRight

	case ActionRight:
		return ActionLeft

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
	case ActionLeft, ActionRight, ActionPreviousLetter, ActionNextLetter:
		return 350 * time.Millisecond

	default:
		return 300 * time.Millisecond
	}
}

func repeatInterval(action Action, heldFor time.Duration) time.Duration {
	switch action {
	case ActionLeft, ActionRight, ActionPreviousLetter, ActionNextLetter:
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
