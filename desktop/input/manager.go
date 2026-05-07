package input

import (
	"grout/desktop/controller"
	"grout/internal"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Callbacks struct {
	Back         func()
	TabLeft      func()
	TabRight     func()
	ToggleView   func()
	QuickMenu    func()
	ExitSearch   func() bool
	FocusContent func()
}

type Manager struct {
	window *adw.ApplicationWindow
	cb     Callbacks
	ctrl   *controller.Handler
}

func NewManager(window *adw.ApplicationWindow, cb Callbacks) *Manager {
	m := &Manager{
		window: window,
		cb:     cb,
		ctrl:   controller.NewHandler(),
	}

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseBubble)
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		if m.focusedWidgetIsEditable() {
			if keyval == gdk.KEY_Escape {
				m.cb.Back()
				return true
			}
			return false
		}
		action := mapKeyvalToAction(keyval)
		if action == controller.ActionNone {
			return false
		}
		m.handleAction(action)
		return true
	})
	window.AddController(keyCtrl)

	return m
}

func (m *Manager) Start() {
	m.ctrl.Start(func(action controller.Action) {
		glib.IdleAdd(func() {
			m.handleAction(action)
		})
	})
}

func (m *Manager) Stop() {
	m.ctrl.Stop()
}

func (m *Manager) handleAction(action controller.Action) {
	logger := internal.GetLogger()
	m.window.SetFocusVisible(true)

	// If focus is on an editable widget, intercept navigation actions
	editable := m.focusedWidgetIsEditable()
	logger.Debug("handleAction", "action", action, "editable", editable)
	if editable {
		switch action {
		case controller.ActionBack:
			if m.cb.ExitSearch != nil {
				m.cb.ExitSearch()
			}
			m.exitEditable()
			return
		case controller.ActionDown, controller.ActionConfirm:
			m.exitEditable()
			return
		case controller.ActionUp, controller.ActionLeft, controller.ActionRight:
			return
		}
		// L1, R1, Select, Start, etc. fall through to normal handling
	}

	switch action {
	case controller.ActionUp:
		m.navigate(0, -1)
	case controller.ActionDown:
		m.navigate(0, 1)
	case controller.ActionLeft:
		m.navigate(-1, 0)
	case controller.ActionRight:
		m.navigate(1, 0)
	case controller.ActionConfirm:
		focused := m.window.Focus()
		if focused != nil {
			gtk.BaseWidget(focused).Activate()
		}
	case controller.ActionBack:
		if m.cb.ExitSearch != nil && m.cb.ExitSearch() {
			// Search was active, exited
		} else {
			m.cb.Back()
		}
	case controller.ActionL1:
		if m.cb.TabLeft != nil {
			m.cb.TabLeft()
		}
	case controller.ActionR1:
		if m.cb.TabRight != nil {
			m.cb.TabRight()
		}
	case controller.ActionSelect:
		if m.cb.ToggleView != nil {
			m.cb.ToggleView()
		}
	case controller.ActionStart:
		if m.cb.QuickMenu != nil {
			m.cb.QuickMenu()
		}
	default:
		logger.Debug("Unhandled input action", "action", action)
	}
}

type indexer interface{ Index() int }

func (m *Manager) navigate(dx, dy int) {
	focused := m.window.Focus()
	if focused == nil {
		return
	}

	// If focus is on a ListBox container, jump into a row
	if lb, ok := focused.(*gtk.ListBox); ok {
		sel := lb.SelectedRow()
		if sel != nil {
			sel.GrabFocus()
			return
		}
		row := lb.RowAtIndex(0)
		if row != nil {
			row.GrabFocus()
		}
		return
	}

	// If focus is on a FlowBox container, jump into first child
	if fb, ok := focused.(*gtk.FlowBox); ok {
		child := fb.ChildAtIndex(0)
		if child != nil {
			child.GrabFocus()
		}
		return
	}

	// Get current index from focused widget
	row, ok := focused.(indexer)
	if !ok {
		return
	}
	currentIdx := row.Index()

	// Find parent container (ListBox or FlowBox)
	cur := gtk.BaseWidget(focused)
	for {
		parent := cur.Parent()
		if parent == nil {
			return
		}
		if lb, ok := parent.(*gtk.ListBox); ok {
			if dy == 0 {
				return
			}
			next := lb.RowAtIndex(currentIdx + dy)
			if next != nil {
				next.GrabFocus()
			} else {
				m.jumpListBox(lb, dy)
			}
			return
		}
		if fb, ok := parent.(*gtk.FlowBox); ok {
			m.navigateFlowBox(fb, currentIdx, dx, dy)
			return
		}
		cur = gtk.BaseWidget(parent)
	}
}

func (m *Manager) navigateFlowBox(fb *gtk.FlowBox, currentIdx, dx, dy int) {
	if dx != 0 {
		targetIdx := currentIdx + dx
		if targetIdx < 0 {
			return
		}
		next := fb.ChildAtIndex(targetIdx)
		if next != nil {
			next.GrabFocus()
		}
		return
	}

	cols := m.flowBoxColumns(fb)
	targetIdx := currentIdx + dy*cols
	if targetIdx < 0 {
		return
	}
	next := fb.ChildAtIndex(targetIdx)
	if next != nil {
		next.GrabFocus()
	}
}

func (m *Manager) flowBoxColumns(fb *gtk.FlowBox) int {
	first := fb.ChildAtIndex(0)
	if first == nil {
		return 1
	}
	childWidth := gtk.BaseWidget(first).AllocatedWidth()
	if childWidth <= 0 {
		return 1
	}
	fbWidth := gtk.BaseWidget(fb).AllocatedWidth()
	cols := fbWidth / childWidth
	if cols < 1 {
		cols = 1
	}
	return cols
}

func (m *Manager) exitEditable() {
	focused := m.window.Focus()
	if focused != nil {
		gtk.BaseWidget(focused).SetCanFocus(false)
		glib.IdleAdd(func() {
			if m.cb.FocusContent != nil {
				m.cb.FocusContent()
			} else {
				m.window.ChildFocus(gtk.DirTabForward)
			}
			gtk.BaseWidget(focused).SetCanFocus(true)
		})
	}
}

func (m *Manager) jumpListBox(current *gtk.ListBox, dy int) {
	parent := gtk.BaseWidget(current).Parent()
	if parent == nil {
		return
	}
	// Collect all ListBox children of the parent
	var listBoxes []*gtk.ListBox
	child := gtk.BaseWidget(parent).FirstChild()
	for child != nil {
		if lb, ok := child.(*gtk.ListBox); ok {
			listBoxes = append(listBoxes, lb)
		}
		child = gtk.BaseWidget(child).NextSibling()
	}
	// Find current ListBox index
	currentIdx := -1
	for i, lb := range listBoxes {
		if lb.Native() == current.Native() {
			currentIdx = i
			break
		}
	}
	if currentIdx < 0 {
		return
	}
	targetIdx := currentIdx + dy
	if targetIdx < 0 || targetIdx >= len(listBoxes) {
		return
	}
	target := listBoxes[targetIdx]
	if dy > 0 {
		row := target.RowAtIndex(0)
		if row != nil {
			row.GrabFocus()
		}
	} else {
		// Focus last row
		for i := 0; ; i++ {
			row := target.RowAtIndex(i)
			if row == nil {
				if i > 0 {
					target.RowAtIndex(i - 1).GrabFocus()
				}
				break
			}
		}
	}
}

func (m *Manager) focusedWidgetIsEditable() bool {
	focused := m.window.Focus()
	if focused == nil {
		return false
	}
	switch focused.(type) {
	case *gtk.Text, *gtk.Entry, *gtk.SearchEntry, *gtk.EditableLabel:
		return true
	}
	return false
}

func mapKeyvalToAction(keyval uint) controller.Action {
	switch keyval {
	case gdk.KEY_Up, gdk.KEY_w, gdk.KEY_W:
		return controller.ActionUp
	case gdk.KEY_Down, gdk.KEY_s, gdk.KEY_S:
		return controller.ActionDown
	case gdk.KEY_Left, gdk.KEY_a, gdk.KEY_A:
		return controller.ActionLeft
	case gdk.KEY_Right, gdk.KEY_d, gdk.KEY_D:
		return controller.ActionRight
	case gdk.KEY_Return, gdk.KEY_KP_Enter, gdk.KEY_space:
		return controller.ActionConfirm
	case gdk.KEY_Escape, gdk.KEY_BackSpace:
		return controller.ActionBack
	default:
		return controller.ActionNone
	}
}
