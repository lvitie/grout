package screens

import (
	"grout/desktop"
	"grout/internal"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
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
	downloadArtRow.ConnectNotify("active", func() {
		s.config.DownloadArt = downloadArtRow.Active()
		s.config.Save()
	})
	generalGroup.Add(downloadArtRow)

	unzipRow := adw.NewSwitchRow()
	unzipRow.SetTitle("Unzip Downloads")
	unzipRow.SetActive(s.config.UnzipDownloads)
	unzipRow.ConnectNotify("active", func() {
		s.config.UnzipDownloads = unzipRow.Active()
		s.config.Save()
	})
	generalGroup.Add(unzipRow)

	// Appearance Group
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("Appearance")
	page.Add(appearanceGroup)

	showBoxArtRow := adw.NewSwitchRow()
	showBoxArtRow.SetTitle("Show Box Art")
	showBoxArtRow.SetActive(s.config.ShowBoxArt)
	showBoxArtRow.ConnectNotify("active", func() {
		s.config.ShowBoxArt = showBoxArtRow.Active()
		s.config.Save()
	})
	appearanceGroup.Add(showBoxArtRow)

	// More Group
	moreGroup := adw.NewPreferencesGroup()
	moreGroup.SetTitle("More")
	page.Add(moreGroup)

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
