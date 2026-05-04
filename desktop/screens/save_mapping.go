package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SaveMappingScreen struct {
	router *desktop.Router
}

func NewSaveMappingScreen(router *desktop.Router) *SaveMappingScreen {
	return &SaveMappingScreen{router: router}
}

func (s *SaveMappingScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Save Slot Mapping")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Global Mapping")
	page.Add(group)

	slot1Row := adw.NewEntryRow()
	slot1Row.SetTitle("Slot 1")
	group.Add(slot1Row)

	navPage := adw.NewNavigationPage(page, "Save Mapping")
	return navPage
}
