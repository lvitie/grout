package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SyncMenuScreen struct {
	router *desktop.Router
}

func NewSyncMenuScreen(router *desktop.Router) *SyncMenuScreen {
	return &SyncMenuScreen{router: router}
}

func (s *SyncMenuScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Save Sync")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Sync Hub")
	page.Add(group)

	syncNowRow := adw.NewActionRow()
	syncNowRow.SetTitle("Sync All Now")
	syncNowRow.SetActivatable(true)
	group.Add(syncNowRow)

	historyRow := adw.NewActionRow()
	historyRow.SetTitle("Sync History")
	historyRow.SetActivatable(true)
	historyRow.ConnectActivated(func() {
		router.Navigate(NewSyncHistoryScreen(router))
	})
	group.Add(historyRow)

	syncedGamesRow := adw.NewActionRow()
	syncedGamesRow.SetTitle("Synced Games")
	syncedGamesRow.SetActivatable(true)
	syncedGamesRow.ConnectActivated(func() {
		router.Navigate(NewSyncedGamesScreen(router))
	})
	group.Add(syncedGamesRow)

	navPage := adw.NewNavigationPage(page, "Save Sync")
	return navPage
}
