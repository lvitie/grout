package controller

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/holoplot/go-evdev"
	"grout/internal"
)

type Handler struct {
	stopCh chan struct{}
}

func NewHandler() *Handler {
	return &Handler{
		stopCh: make(chan struct{}),
	}
}

func (h *Handler) Start(callback func(Action)) {
	logger := internal.GetLogger()

	files, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		logger.Error("Failed to glob input devices", "error", err)
		return
	}

	var devices []*evdev.InputDevice
	var permissionDenied int
	for _, path := range files {
		dev, err := evdev.Open(path)
		if err != nil {
			if os.IsPermission(err) {
				permissionDenied++
			}
			continue
		}
		devices = append(devices, dev)
	}
	if permissionDenied > 0 {
		logger.Warn("Could not open input devices (permission denied). Add your user to the 'input' group for gamepad support",
			"denied", permissionDenied, "opened", len(devices))
	}

	var gamepad *evdev.InputDevice
	for _, dev := range devices {
		hasAbs := false
		for _, t := range dev.CapableTypes() {
			if t == evdev.EV_ABS {
				hasAbs = true
				break
			}
		}
		if hasAbs {
			gamepad = dev
			name, _ := dev.Name()
			logger.Info("Found potential gamepad", "name", name, "path", dev.Path())
			break
		} else {
			dev.Close()
		}
	}

	if gamepad == nil {
		logger.Warn("No gamepad found")
		return
	}

	if err := gamepad.NonBlock(); err != nil {
		logger.Error("Failed to set gamepad to non-blocking mode", "error", err)
		gamepad.Close()
		return
	}

	// Shared held direction, protected by mutex
	var mu sync.Mutex
	var heldDir Action

	// Repeat goroutine — fires held direction on a timer
	go func() {
		const initialDelay = 500 * time.Millisecond
		const repeatRate = 150 * time.Millisecond

		var lastDir Action
		var holdingSince time.Time
		var lastFired time.Time

		for {
			select {
			case <-h.stopCh:
				return
			default:
			}

			mu.Lock()
			dir := heldDir
			mu.Unlock()

			if dir != lastDir {
				lastDir = dir
				holdingSince = time.Now()
				lastFired = time.Time{}
			}

			if dir != ActionNone {
				now := time.Now()
				held := now.Sub(holdingSince)
				if held >= initialDelay {
					sinceLastFire := now.Sub(lastFired)
					if lastFired.IsZero() || sinceLastFire >= repeatRate {
						lastFired = now
						callback(dir)
					}
				}
			}

			time.Sleep(16 * time.Millisecond)
		}
	}()

	// Event loop — reads events, fires on direction change, updates heldDir
	go func() {
		defer gamepad.Close()

		stickState := NewStickState()

		for {
			select {
			case <-h.stopCh:
				return
			default:
			}

			ev, err := gamepad.ReadOne()
			if err != nil {
				if errors.Is(err, syscall.EAGAIN) {
					time.Sleep(16 * time.Millisecond)
				} else {
					logger.Error("Error reading gamepad event", "error", err)
					return
				}
				continue
			}

			// Buttons — fire immediately
			if ev.Type == evdev.EV_KEY {
				action := mapButton(ev)
				if action != ActionNone {
					callback(action)
				}
				continue
			}

			if ev.Type != evdev.EV_ABS {
				continue
			}

			// Track axis state
			prevDir := stickState.HeldAction()
			mapAxis(ev, stickState)
			newDir := stickState.HeldAction()

			// Update shared held direction
			mu.Lock()
			heldDir = newDir
			mu.Unlock()

			// Fire on direction change
			if newDir != prevDir && newDir != ActionNone {
				callback(newDir)
			}
		}
	}()
}

func mapAxis(ev *evdev.InputEvent, state *StickState) {
	switch ev.Code {
	case evdev.ABS_HAT0X:
		state.updateHat(ev.Code, ev.Value, ActionLeft, ActionRight)
	case evdev.ABS_HAT0Y:
		state.updateHat(ev.Code, ev.Value, ActionUp, ActionDown)
	case evdev.ABS_X, evdev.ABS_RX:
		state.updateStick(ev.Code, ev.Value, ActionLeft, ActionRight)
	case evdev.ABS_Y, evdev.ABS_RY:
		state.updateStick(ev.Code, ev.Value, ActionUp, ActionDown)
	}
}

func mapButton(ev *evdev.InputEvent) Action {
	if ev.Value == 0 {
		return ActionNone
	}
	switch ev.Code {
	case 304:
		return ActionConfirm
	case 305:
		return ActionBack
	case 307:
		return ActionMenu
	case 308:
		return ActionAlt
	case 310:
		return ActionL1
	case 311:
		return ActionR1
	case 312:
		return ActionL2
	case 313:
		return ActionR2
	case 314:
		return ActionSelect
	case 315:
		return ActionStart
	}
	return ActionNone
}

func (h *Handler) Stop() {
	close(h.stopCh)
}
