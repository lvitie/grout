package screens

import (
	"fmt"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/cache"
	"grout/desktop"
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
	host := router.State().GetHost()
	deviceID := ""
	if host != nil {
		deviceID = host.DeviceID
	}
	ids := cm.GetSyncedRomIDs(deviceID)

	for _, id := range ids {
		row := adw.NewActionRow()
		row.SetTitle(fmt.Sprintf("ROM #%d", id))
		listBox.Append(row)
	}

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	page := adw.NewNavigationPage(scrolled, "Synced Games")
	return page
}
