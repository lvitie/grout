package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ToolsSettingsScreen struct {
	router *desktop.Router
}

func NewToolsSettingsScreen(router *desktop.Router) *ToolsSettingsScreen {
	return &ToolsSettingsScreen{router: router}
}

func (s *ToolsSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Tools")

	// Maintenance Group
	maintenanceGroup := adw.NewPreferencesGroup()
	maintenanceGroup.SetTitle("Maintenance")
	page.Add(maintenanceGroup)

	rebuildCacheRow := adw.NewActionRow()
	rebuildCacheRow.SetTitle("Rebuild Cache")
	rebuildCacheRow.SetSubtitle("Full resync with RomM server")
	rebuildCacheRow.SetActivatable(true)
	maintenanceGroup.Add(rebuildCacheRow)

	// Artwork Group
	artworkGroup := adw.NewPreferencesGroup()
	artworkGroup.SetTitle("Artwork")
	page.Add(artworkGroup)

	syncArtRow := adw.NewActionRow()
	syncArtRow.SetTitle("Download Missing Art")
	syncArtRow.SetActivatable(true)
	artworkGroup.Add(syncArtRow)

	navPage := adw.NewNavigationPage(page, "Tools")
	return navPage
}
