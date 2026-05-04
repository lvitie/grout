package controller

import (
	"grout/internal"
	"github.com/holoplot/go-evdev"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"path/filepath"
)

type Handler struct {
	actionCh chan Action
	stopCh   chan struct{}
}

func NewHandler() *Handler {
	return &Handler{
		actionCh: make(chan Action, 10),
		stopCh:   make(chan struct{}),
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
	for _, path := range files {
		dev, err := evdev.Open(path)
		if err != nil {
			continue
		}
		devices = append(devices, dev)
	}

	var gamepad *evdev.InputDevice
	for _, dev := range devices {
		// Heuristic: check for buttons common on gamepads
		isGamepad := false
		for _, t := range dev.CapableTypes() {
			if t == evdev.EV_KEY {
				isGamepad = true
				break
			}
		}

		if isGamepad {
			// This is very basic; a real app might let user select device
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

	go func() {
		defer gamepad.Close()
		for {
			select {
			case <-h.stopCh:
				return
			default:
				ev, err := gamepad.ReadOne()
				if err != nil {
					logger.Error("Error reading gamepad event", "error", err)
					return
				}

				if action := MapEvent(*ev); action != ActionNone {
					glib.IdleAdd(func() {
						callback(action)
					})
				}
			}
		}
	}()
}

func (h *Handler) Stop() {
	close(h.stopCh)
}
