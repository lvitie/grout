package screens

import (
	"grout/cache"
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SyncedGamesScreen struct {
	router *desktop.Router
}

func NewSyncedGamesScreen(router *desktop.Router) *SyncedGamesScreen {
	return &SyncedGamesScreen{router: router}
}

func (s *SyncedGamesScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	
	cm := cache.GetCacheManager()
	ids, _ := cm.GetSyncedRomIDs()
	
	for _, id := range ids {
		// Fetch ROM name from cache
		rom, _ := cm.GetRom(id)
		row := adw.NewActionRow()
		row.SetTitle(rom.Name)
		listBox.Append(row)
	}

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	page := adw.NewNavigationPage(scrolled, "Synced Games")
	return page
}
