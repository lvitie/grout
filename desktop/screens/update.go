package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type UpdateScreen struct {
	router *desktop.Router
}

func NewUpdateScreen(router *desktop.Router) *UpdateScreen {
	return &UpdateScreen{router: router}
}

func (s *UpdateScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Check for Updates")
	statusPage.SetDescription("Checking if a new version of Grout is available...")
	statusPage.SetIconName("software-update-available-symbolic")

	spinner := gtk.NewSpinner()
	spinner.Start()

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetValign(gtk.AlignCenter)
	box.Append(spinner)

	statusPage.SetChild(box)

	page := adw.NewNavigationPage(statusPage, "Update")
	return page
}
