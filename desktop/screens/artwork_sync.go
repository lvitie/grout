package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ArtworkSyncScreen struct {
	router *desktop.Router
}

func NewArtworkSyncScreen(router *desktop.Router) *ArtworkSyncScreen {
	return &ArtworkSyncScreen{router: router}
}

func (s *ArtworkSyncScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Artwork Sync")
	statusPage.SetIconName("image-x-generic-symbolic")

	page := adw.NewNavigationPage(statusPage, "Artwork Sync")
	return page
}
