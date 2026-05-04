package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SaveSyncScreen struct {
	router *desktop.Router
}

func NewSaveSyncScreen(router *desktop.Router) *SaveSyncScreen {
	return &SaveSyncScreen{router: router}
}

func (s *SaveSyncScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Save Synchronization")
	statusPage.SetDescription("Synchronizing your game saves with RomM...")
	statusPage.SetIconName("view-refresh-symbolic")

	progress := gtk.NewProgressBar()
	progress.SetFraction(0.0)
	progress.SetShowText(true)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetVAlign(gtk.AlignCenter)
	box.Append(progress)

	statusPage.SetChild(box)

	page := adw.NewNavigationPage(statusPage, "Save Sync")
	return page
}
