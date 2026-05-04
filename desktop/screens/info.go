package screens

import (
	"fmt"
	"grout/desktop"
	"grout/version"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type InfoScreen struct {
	router *desktop.Router
}

func NewInfoScreen(router *desktop.Router) *InfoScreen {
	return &InfoScreen{router: router}
}

func (s *InfoScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Grout")
	statusPage.SetDescription(fmt.Sprintf("Version %s\n\n© 2026 RomM Team", version.Get().Version))
	statusPage.SetIconName("help-about-symbolic")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Links")

	githubRow := adw.NewActionRow()
	githubRow.SetTitle("GitHub Repository")
	githubRow.SetActivatable(true)
	group.Add(githubRow)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetVAlign(gtk.AlignCenter)
	box.Append(group)

	statusPage.SetChild(box)

	page := adw.NewNavigationPage(statusPage, "About")
	return page
}
