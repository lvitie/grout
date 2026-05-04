package controller

import (
	"github.com/holoplot/go-evdev"
	"github.com/diamondburned/gotk4/pkg/glib"
)

type Action int

const (
	ActionUp Action = iota
	ActionDown
	ActionLeft
	ActionRight
	ActionConfirm
	ActionBack
	ActionMenu
	ActionAlt
)

type Handler struct {
	actionCh chan Action
}

func NewHandler() *Handler {
	return &Handler{
		actionCh: make(chan Action, 10),
	}
}

func (h *Handler) Start(callback func(Action)) {
	go func() {
		for action := range h.actionCh {
			a := action
			glib.IdleAdd(func() {
				callback(a)
			})
		}
	}()

	// In a real implementation, we would poll /dev/input/event* here.
	// For now, it's a skeleton.
}

func (h *Handler) Actions() <-chan Action {
	return h.actionCh
}
