package screens

import (
	"grout/desktop"
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SaveConflictScreen struct {
	router *desktop.Router
	rom    romm.Rom
}

func NewSaveConflictScreen(router *desktop.Router, rom romm.Rom) *SaveConflictScreen {
	return &SaveConflictScreen{
		router: router,
		rom:    rom,
	}
}

func (s *SaveConflictScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Save Conflict")
	statusPage.SetDescription("Both local and remote saves have changed since the last sync.")
	statusPage.SetIconName("dialog-warning-symbolic")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Choose which save to keep")

	localRow := adw.NewActionRow()
	localRow.SetTitle("Keep Local Save")
	localRow.SetSubtitle("Upload local save to server")
	localRow.SetActivatable(true)
	group.Add(localRow)

	remoteRow := adw.NewActionRow()
	remoteRow.SetTitle("Keep Remote Save")
	remoteRow.SetSubtitle("Overwrite local save with server data")
	remoteRow.SetActivatable(true)
	group.Add(remoteRow)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetValign(gtk.AlignCenter)
	box.Append(group)

	statusPage.SetChild(box)

	page := adw.NewNavigationPage(statusPage, "Conflict")
	return page
}
