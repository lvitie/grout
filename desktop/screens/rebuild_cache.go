package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type RebuildCacheScreen struct {
	router *desktop.Router
}

func NewRebuildCacheScreen(router *desktop.Router) *RebuildCacheScreen {
	return &RebuildCacheScreen{router: router}
}

func (s *RebuildCacheScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Rebuild Cache")
	statusPage.SetDescription("Wiping local cache and re-syncing from server...")
	statusPage.SetIconName("view-refresh-symbolic")

	page := adw.NewNavigationPage(statusPage, "Rebuild Cache")
	return page
}
