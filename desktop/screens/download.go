package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DownloadScreen struct {
	router *desktop.Router
	title  string
}

func NewDownloadScreen(router *desktop.Router, title string) *DownloadScreen {
	return &DownloadScreen{
		router: router,
		title:  title,
	}
}

func (s *DownloadScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle(s.title)
	statusPage.SetIconName("folder-download-symbolic")

	progress := gtk.NewProgressBar()
	progress.SetShowText(true)
	progress.SetFraction(0.0)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetValign(gtk.AlignCenter)
	box.Append(progress)

	statusPage.SetChild(box)

	page := adw.NewNavigationPage(statusPage, "Downloading")
	return page
}
