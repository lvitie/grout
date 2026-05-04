package desktop

import (
	"grout/desktop/screens"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Router struct {
	window *adw.ApplicationWindow
	nav    *adw.NavigationView
}

type Screen interface {
	Build(router *Router) gtk.Widgetter
}

func NewRouter(window *adw.ApplicationWindow) *Router {
	nav := adw.NewNavigationView()
	window.SetContent(nav)

	return &Router{
		window: window,
		nav:    nav,
	}
}

func (r *Router) ShowFirstScreen() {
	r.Navigate(screens.NewLoginScreen(r))
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
