package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
)

type PlatformMappingScreen struct {
	router *desktop.Router
}

func NewPlatformMappingScreen(router *desktop.Router) *PlatformMappingScreen {
	return &PlatformMappingScreen{router: router}
}

func (s *PlatformMappingScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Platform Directory Mapping")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Local Mappings")
	page.Add(group)

	row := adw.NewActionRow()
	row.SetTitle("Mapping Logic")
	row.SetSubtitle("Map local directories to RomM platform slugs")
	group.Add(row)

	navPage := adw.NewNavigationPage(page, "Platform Mapping")
	return navPage
}
