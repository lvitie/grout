package screens

import (
	"grout/cache"
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SyncHistoryScreen struct {
	router *desktop.Router
}

func NewSyncHistoryScreen(router *desktop.Router) *SyncHistoryScreen {
	return &SyncHistoryScreen{router: router}
}

func (s *SyncHistoryScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	
	cm := cache.GetCacheManager()
	history, _ := cm.GetSaveSyncHistory(50)
	
	for _, entry := range history {
		row := adw.NewActionRow()
		row.SetTitle(entry.RomName)
		row.SetSubtitle(entry.Action + " • " + entry.SyncedAt.Format("2006-01-02 15:04"))
		listBox.Append(row)
	}

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	page := adw.NewNavigationPage(scrolled, "Sync History")
	return page
}
