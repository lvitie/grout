package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CollectionsSettingsScreen struct {
	router *desktop.Router
}

func NewCollectionsSettingsScreen(router *desktop.Router) *CollectionsSettingsScreen {
	return &CollectionsSettingsScreen{router: router}
}

func (s *CollectionsSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Collections Visibility")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Show Collections")
	page.Add(group)

	regularRow := adw.NewSwitchRow()
	regularRow.SetTitle("Regular Collections")
	group.Add(regularRow)

	smartRow := adw.NewSwitchRow()
	smartRow.SetTitle("Smart Collections")
	group.Add(smartRow)

	virtualRow := adw.NewSwitchRow()
	virtualRow.SetTitle("Virtual Collections")
	group.Add(virtualRow)

	navPage := adw.NewNavigationPage(page, "Collections")
	return navPage
}
