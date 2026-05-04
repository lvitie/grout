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
	generalGroup.Add(downloadArtRow)

	unzipRow := adw.NewSwitchRow()
	unzipRow.SetTitle("Unzip Downloads")
	unzipRow.SetActive(s.config.UnzipDownloads)
	generalGroup.Add(unzipRow)

	// Appearance Group
	appearanceGroup := adw.NewPreferencesGroup()
	appearanceGroup.SetTitle("Appearance")
	page.Add(appearanceGroup)

	showBoxArtRow := adw.NewSwitchRow()
	showBoxArtRow.SetTitle("Show Box Art")
	showBoxArtRow.SetActive(s.config.ShowBoxArt)
	appearanceGroup.Add(showBoxArtRow)

	// Add more groups as needed...

	navPage := adw.NewNavigationPage(page, "Settings")
	return navPage
}
