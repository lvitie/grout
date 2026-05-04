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

// DefaultMapping provides a basic mapping for standard gamepads
func MapEvent(ev evdev.InputEvent) Action {
	// Standard Linux Gamepad API mapping
	if ev.Type != evdev.EV_KEY {
		return ActionNone
	}

	// Value 1 is press, 0 is release, 2 is repeat
	if ev.Value == 0 {
		return ActionNone
	}

	switch ev.Code {
	case 304: // BTN_SOUTH (A)
		return ActionConfirm
	case 305: // BTN_EAST (B)
		return ActionBack
	case 307: // BTN_NORTH (X)
		return ActionMenu
	case 308: // BTN_WEST (Y)
		return ActionAlt
	case 310: // BTN_TL (L1)
		return ActionL1
	case 311: // BTN_TR (R1)
		return ActionR1
	case 312: // BTN_TL2 (L2)
		return ActionL2
	case 313: // BTN_TR2 (R2)
		return ActionR2
	case 314: // BTN_SELECT
		return ActionSelect
	case 315: // BTN_START
		return ActionStart
	// D-Pad is often EV_ABS (hat), but some use EV_KEY
	case 103: // KEY_UP
		return ActionUp
	case 108: // KEY_DOWN
		return ActionDown
	case 105: // KEY_LEFT
		return ActionLeft
	case 106: // KEY_RIGHT
		return ActionRight
	}

	return ActionNone
}
