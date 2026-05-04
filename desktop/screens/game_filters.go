package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameFiltersScreen struct {
	router *desktop.Router
}

func NewGameFiltersScreen(router *desktop.Router) *GameFiltersScreen {
	return &GameFiltersScreen{router: router}
}

func (s *GameFiltersScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Filter & Sort")

	sortGroup := adw.NewPreferencesGroup()
	sortGroup.SetTitle("Sort Order")
	page.Add(sortGroup)

	sortRow := adw.NewComboRow()
	sortRow.SetTitle("Sort By")
	sortRow.SetModel(gtk.NewStringList([]string{"Name", "Release Date", "Rating"}))
	sortGroup.Add(sortRow)

	filterGroup := adw.NewPreferencesGroup()
	filterGroup.SetTitle("Filters")
	page.Add(filterGroup)

	favRow := adw.NewSwitchRow()
	favRow.SetTitle("Favorites Only")
	filterGroup.Add(favRow)

	navPage := adw.NewNavigationPage(page, "Filters")
	return navPage
}
