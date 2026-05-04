package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type BiosDownloadScreen struct {
	router *desktop.Router
}

func NewBiosDownloadScreen(router *desktop.Router) *BiosDownloadScreen {
	return &BiosDownloadScreen{router: router}
}

func (s *BiosDownloadScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("BIOS Download")
	statusPage.SetIconName("system-software-update-symbolic")

	page := adw.NewNavigationPage(statusPage, "BIOS")
	return page
}
