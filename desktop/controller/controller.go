package controller

import (
	"fmt"
	"grout/internal"
	"github.com/holoplot/go-evdev"
	"github.com/diamondburned/gotk4/pkg/core/glib"
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

	devices, err := evdev.ListInputDevices()
	if err != nil {
		logger.Error("Failed to list input devices", "error", err)
		return
	}

	var gamepad *evdev.InputDevice
	for _, dev := range devices {
		// Heuristic: check for buttons common on gamepads
		if dev.Capabilities[evdev.EV_KEY] != nil {
			// This is very basic; a real app might let user select device
			gamepad = dev
			logger.Info("Found potential gamepad", "name", dev.Name, "path", dev.Path)
			break
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
