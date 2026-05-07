package controller

import "github.com/holoplot/go-evdev"

// Action represents a high-level UI action
type Action int

const (
	ActionNone Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionConfirm // A / Cross
	ActionBack    // B / Circle
	ActionMenu    // X / Square
	ActionAlt     // Y / Triangle
	ActionL1
	ActionR1
	ActionL2
	ActionR2
	ActionSelect
	ActionStart
)

// StickState tracks the current direction of each axis.
type StickState struct {
	axes map[evdev.EvCode]Action
}

func NewStickState() *StickState {
	return &StickState{axes: make(map[evdev.EvCode]Action)}
}

const stickCenter int32 = 32768
const stickDeadzone int32 = 12000

func (s *StickState) updateStick(code evdev.EvCode, value int32, neg, pos Action) {
	centered := value - stickCenter
	var current Action
	if centered < -stickDeadzone {
		current = neg
	} else if centered > stickDeadzone {
		current = pos
	}
	s.axes[code] = current
}

func (s *StickState) updateHat(code evdev.EvCode, value int32, neg, pos Action) {
	var current Action
	switch value {
	case -1:
		current = neg
	case 1:
		current = pos
	}
	s.axes[code] = current
}

// HeldAction returns the first non-None direction across all axes.
func (s *StickState) HeldAction() Action {
	for _, a := range s.axes {
		if a != ActionNone {
			return a
		}
	}
	return ActionNone
}
