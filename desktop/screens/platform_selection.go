package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PlatformSelectionScreen struct {
	router    *desktop.Router
	platforms []romm.Platform
}

func NewPlatformSelectionScreen(router *desktop.Router) *PlatformSelectionScreen {
	cm := cache.GetCacheManager()
	platforms, _ := cm.GetPlatforms()
	return &PlatformSelectionScreen{
		router:    router,
		platforms: platforms,
	}
}

func (s *PlatformSelectionScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	for _, p := range s.platforms {
		row := adw.NewActionRow()
		row.SetTitle(p.Name)
		row.SetSubtitle(p.FSSlug)
		row.SetSelectable(true)
		listBox.Append(row)
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		router.Navigate(NewGameListScreen(router, s.platforms[idx]))
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	page := adw.NewNavigationPage(scrolled, "Platforms")
	return page
}
