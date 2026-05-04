package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ServerAddressScreen struct {
	router *desktop.Router
}

func NewServerAddressScreen(router *desktop.Router) *ServerAddressScreen {
	return &ServerAddressScreen{router: router}
}

func (s *ServerAddressScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Server Address")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Edit Host")
	page.Add(group)

	urlRow := adw.NewEntryRow()
	urlRow.SetTitle("Server URL")
	group.Add(urlRow)

	navPage := adw.NewNavigationPage(page, "Server")
	return navPage
}
