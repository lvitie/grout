package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
	"grout/internal"
)

type SettingsScreen struct {
	router *desktop.Router
	config *internal.Config
}

func NewSettingsScreen(router *desktop.Router) *SettingsScreen {
	config, _ := internal.LoadConfig()
	return &SettingsScreen{
		router: router,
		config: config,
	}
}

func (s *SettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Settings")
	page.SetIconName("preferences-system-symbolic")

	// General Group
	generalGroup := adw.NewPreferencesGroup()
	generalGroup.SetTitle("General")
	page.Add(generalGroup)

	downloadArtRow := adw.NewSwitchRow()
	downloadArtRow.SetTitle("Download Artwork")
	downloadArtRow.SetSubtitle("Automatically fetch boxart from RomM")
	downloadArtRow.SetActive(s.config.DownloadArt)
	downloadArtRow.Connect("notify::active", func() {
		s.config.DownloadArt = downloadArtRow.Active()
		internal.SaveConfig(s.config)
	})
	generalGroup.Add(downloadArtRow)

	unzipRow := adw.NewSwitchRow()
	unzipRow.SetTitle("Unzip Downloads")
	unzipRow.SetActive(s.config.UnzipDownloads)
	unzipRow.Connect("notify::active", func() {
		s.config.UnzipDownloads = unzipRow.Active()
		internal.SaveConfig(s.config)
	})
	generalGroup.Add(unzipRow)

	closeToTrayRow := adw.NewSwitchRow()
	closeToTrayRow.SetTitle("Close to Tray")
	closeToTrayRow.SetSubtitle("Keep running in the background when the window is closed")
	closeToTrayRow.SetActive(s.config.ShouldCloseToTray())
	closeToTrayRow.Connect("notify::active", func() {
		val := closeToTrayRow.Active()
		s.config.CloseToTray = &val
		internal.SaveConfig(s.config)
	})
	generalGroup.Add(closeToTrayRow)

	// Appearance Group
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("Appearance")
	page.Add(appearanceGroup)

	showBoxArtRow := adw.NewSwitchRow()
	showBoxArtRow.SetTitle("Show Box Art")
	showBoxArtRow.SetActive(s.config.ShowBoxArt)
	showBoxArtRow.Connect("notify::active", func() {
		s.config.ShowBoxArt = showBoxArtRow.Active()
		internal.SaveConfig(s.config)
	})
	appearanceGroup.Add(showBoxArtRow)

	collectionsRow := adw.NewActionRow()
	collectionsRow.SetTitle("Collections")
	collectionsRow.SetSubtitle("Configure regular, smart, and virtual collections visibility")
	collectionsRow.SetActivatable(true)
	collectionsRow.ConnectActivated(func() {
		router.Navigate(NewCollectionsSettingsScreen(router))
	})
	appearanceGroup.Add(collectionsRow)

	// More Group
	moreGroup := adw.NewPreferencesGroup()
	moreGroup.SetTitle("More")
	page.Add(moreGroup)

	saveSyncRow := adw.NewActionRow()
	saveSyncRow.SetTitle("Save Sync")
	saveSyncRow.SetSubtitle("Synchronize game saves with RomM")
	saveSyncRow.SetActivatable(true)
	saveSyncRow.ConnectActivated(func() {
		router.Navigate(NewSyncMenuScreen(router))
	})
	moreGroup.Add(saveSyncRow)

	advancedRow := adw.NewActionRow()
	advancedRow.SetTitle("Advanced")
	advancedRow.SetSubtitle("Network, logging, and experimental settings")
	advancedRow.SetActivatable(true)
	advancedRow.ConnectActivated(func() {
		router.Navigate(NewAdvancedSettingsScreen(router))
	})
	moreGroup.Add(advancedRow)

	toolsRow := adw.NewActionRow()
	toolsRow.SetTitle("Tools")
	toolsRow.SetSubtitle("Maintenance and manual synchronization")
	toolsRow.SetActivatable(true)
	toolsRow.ConnectActivated(func() {
		router.Navigate(NewToolsSettingsScreen(router))
	})
	moreGroup.Add(toolsRow)

	navPage := adw.NewNavigationPage(page, "Settings")
	return navPage
}
