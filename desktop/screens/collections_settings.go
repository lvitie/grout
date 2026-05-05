package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/cache"
	"grout/desktop"
	"grout/internal"
)

type CollectionsSettingsScreen struct {
	router *desktop.Router
	config *internal.Config
}

func NewCollectionsSettingsScreen(router *desktop.Router) *CollectionsSettingsScreen {
	config, _ := internal.LoadConfig()
	return &CollectionsSettingsScreen{
		router: router,
		config: config,
	}
}

func (s *CollectionsSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Collections Visibility")

	group := adw.NewPreferencesGroup()
	group.SetTitle("Show Collections")
	page.Add(group)

	regularRow := adw.NewSwitchRow()
	regularRow.SetTitle("Regular Collections")
	regularRow.SetActive(s.config.ShowRegularCollections)
	regularRow.Connect("notify::active", func() {
		s.config.ShowRegularCollections = regularRow.Active()
		internal.SaveConfig(s.config)
		if cm := cache.GetCacheManager(); cm != nil {
			cm.UpdateConfig(s.config)
		}
	})
	group.Add(regularRow)

	smartRow := adw.NewSwitchRow()
	smartRow.SetTitle("Smart Collections")
	smartRow.SetActive(s.config.ShowSmartCollections)
	smartRow.Connect("notify::active", func() {
		s.config.ShowSmartCollections = smartRow.Active()
		internal.SaveConfig(s.config)
		if cm := cache.GetCacheManager(); cm != nil {
			cm.UpdateConfig(s.config)
		}
	})
	group.Add(smartRow)

	virtualRow := adw.NewSwitchRow()
	virtualRow.SetTitle("Virtual Collections")
	virtualRow.SetActive(s.config.ShowVirtualCollections)
	virtualRow.Connect("notify::active", func() {
		s.config.ShowVirtualCollections = virtualRow.Active()
		internal.SaveConfig(s.config)
		if cm := cache.GetCacheManager(); cm != nil {
			cm.UpdateConfig(s.config)
		}
	})
	group.Add(virtualRow)

	navPage := adw.NewNavigationPage(page, "Collections")
	return navPage
}
