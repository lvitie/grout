package desktop

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Router struct {
	window *adw.ApplicationWindow
	nav    *adw.NavigationView
	state  *AppState
}

type Screen interface {
	Build(router *Router) gtk.Widgetter
}

func NewRouter(window *adw.ApplicationWindow, state *AppState) *Router {
	nav := adw.NewNavigationView()
	window.SetContent(nav)

	return &Router{
		window: window,
		nav:    nav,
		state:  state,
	}
}

func (r *Router) Navigate(screen Screen) {
	widget := screen.Build(r)
	if page, ok := widget.(*adw.NavigationPage); ok {
		r.nav.Push(page)
	}
}

func (r *Router) Back() {
	r.nav.Pop()
}

func (r *Router) Window() *gtk.Window {
	return &r.window.Window
}

func (r *Router) State() *AppState {
	return r.state
}
