package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameListScreen struct {
	router   *desktop.Router
	platform romm.Platform
	games    []romm.Rom
}

func NewGameListScreen(router *desktop.Router, platform romm.Platform) *GameListScreen {
	cm := cache.GetCacheManager()
	games, _ := cm.GetPlatformGames(platform.ID)
	return &GameListScreen{
		router:   router,
		platform: platform,
		games:    games,
	}
}

func (s *GameListScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	for _, g := range s.games {
		row := adw.NewActionRow()
		row.SetTitle(g.Name)
		row.SetSubtitle(g.FsName)
		row.SetSelectable(true)
		listBox.Append(row)
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		router.Navigate(NewGameDetailsScreen(router, s.games[idx]))
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle(s.platform.Name, ""))

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(scrolled)

	page := adw.NewNavigationPage(box, s.platform.Name)
	return page
}
