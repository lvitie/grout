package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
	"grout/internal"
)

type RomPathSettingsScreen struct {
	router *desktop.Router
	config *internal.Config
}

func NewRomPathSettingsScreen(router *desktop.Router) *RomPathSettingsScreen {
	config, _ := internal.LoadConfig()
	return &RomPathSettingsScreen{
		router: router,
		config: config,
	}
}

func (s *RomPathSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("ROM Path")

	layoutGroup := adw.NewPreferencesGroup()
	layoutGroup.SetTitle("ROM Layout")
	layoutGroup.SetDescription("Choose the folder structure for downloaded ROMs")
	page.Add(layoutGroup)

	layoutRow := adw.NewComboRow()
	layoutRow.SetTitle("Layout")
	layoutRow.SetModel(gtk.NewStringList([]string{
		"RomM (Default)",
		"EmuDeck",
		"RetroDeck",
		"RetroArch",
	}))

	layouts := []internal.SaveLayout{
		internal.SaveLayoutRomM,
		internal.SaveLayoutEmuDeck,
		internal.SaveLayoutRetroDeck,
		internal.SaveLayoutRetroArch,
	}

	currentIdx := 0
	for i, l := range layouts {
		if l == s.config.RomLayout {
			currentIdx = i
			break
		}
	}
	layoutRow.SetSelected(uint(currentIdx))
	layoutGroup.Add(layoutRow)

	basePathGroup := adw.NewPreferencesGroup()
	basePathGroup.SetTitle("Base Path")
	basePathGroup.SetDescription("Root directory where ROMs are downloaded")
	page.Add(basePathGroup)

	basePathRow := adw.NewEntryRow()
	basePathRow.SetTitle("ROM Directory")
	basePathRow.SetText(s.config.GetRomBasePath())
	basePathGroup.Add(basePathRow)

	suppressPathSave := false

	layoutRow.Connect("notify::selected", func() {
		newLayout := layouts[layoutRow.Selected()]
		s.config.RomLayout = newLayout
		s.config.RomBasePath = ""

		suppressPathSave = true
		basePathRow.SetText(internal.GetDefaultRomBasePath(newLayout))
		suppressPathSave = false

		internal.SaveConfig(s.config)
	})

	basePathRow.ConnectChanged(func() {
		if suppressPathSave {
			return
		}
		text := basePathRow.Text()
		defaultPath := internal.GetDefaultRomBasePath(s.config.RomLayout)
		if text == defaultPath {
			s.config.RomBasePath = ""
		} else {
			s.config.RomBasePath = text
		}
		internal.SaveConfig(s.config)
	})

	navPage := adw.NewNavigationPage(page, "ROM Path")
	return navPage
}
