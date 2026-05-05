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

	collectionsRow := adw.NewActionRow()
	collectionsRow.SetTitle("Collections")
	collectionsRow.SetSubtitle("Configure regular, smart, and virtual collections visibility")
	collectionsRow.SetActivatable(true)
	collectionsRow.ConnectActivated(func() {
		router.Navigate(NewCollectionsSettingsScreen(router))
	})
	generalGroup.Add(collectionsRow)

	// More Group
	moreGroup := adw.NewPreferencesGroup()
	moreGroup.SetTitle("More")
	page.Add(moreGroup)

	romPathRow := adw.NewActionRow()
	romPathRow.SetTitle("ROM Path")
	romPathRow.SetSubtitle("Configure where ROMs are downloaded and folder structure")
	romPathRow.SetActivatable(true)
	romPathRow.ConnectActivated(func() {
		router.Navigate(NewRomPathSettingsScreen(router))
	})
	moreGroup.Add(romPathRow)

	savePathRow := adw.NewActionRow()
	savePathRow.SetTitle("Save Path")
	savePathRow.SetSubtitle("Configure save file storage location and folder structure")
	savePathRow.SetActivatable(true)
	savePathRow.ConnectActivated(func() {
		router.Navigate(NewSavePathSettingsScreen(router))
	})
	moreGroup.Add(savePathRow)

	saveSyncRow := adw.NewActionRow()
	saveSyncRow.SetTitle("Save Sync")
	saveSyncRow.SetSubtitle("Synchronize game saves with RomM")
	saveSyncRow.SetActivatable(true)
	saveSyncRow.ConnectActivated(func() {
		router.Navigate(NewSyncMenuScreen(router))
	})
	moreGroup.Add(saveSyncRow)

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
