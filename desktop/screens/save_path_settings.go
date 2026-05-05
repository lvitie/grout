package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
	"grout/internal"
)

type SavePathSettingsScreen struct {
	router *desktop.Router
	config *internal.Config
}

func NewSavePathSettingsScreen(router *desktop.Router) *SavePathSettingsScreen {
	config, _ := internal.LoadConfig()
	return &SavePathSettingsScreen{
		router: router,
		config: config,
	}
}

func (s *SavePathSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Save Path")

	layoutGroup := adw.NewPreferencesGroup()
	layoutGroup.SetTitle("Save Layout")
	layoutGroup.SetDescription("Choose the folder structure for save files")
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
		if l == s.config.SaveLayout {
			currentIdx = i
			break
		}
	}
	layoutRow.SetSelected(uint(currentIdx))
	layoutGroup.Add(layoutRow)

	basePathGroup := adw.NewPreferencesGroup()
	basePathGroup.SetTitle("Base Path")
	basePathGroup.SetDescription("Root directory where save files are stored")
	page.Add(basePathGroup)

	basePathRow := adw.NewEntryRow()
	basePathRow.SetTitle("Save Directory")
	basePathRow.SetText(s.config.GetSaveBasePath())
	basePathGroup.Add(basePathRow)

	suppressPathSave := false

	layoutRow.Connect("notify::selected", func() {
		newLayout := layouts[layoutRow.Selected()]
		s.config.SaveLayout = newLayout
		s.config.SaveBasePath = ""

		suppressPathSave = true
		basePathRow.SetText(internal.GetDefaultSaveBasePath(newLayout))
		suppressPathSave = false

		internal.SaveConfig(s.config)
	})

	basePathRow.ConnectChanged(func() {
		if suppressPathSave {
			return
		}
		text := basePathRow.Text()
		defaultPath := internal.GetDefaultSaveBasePath(s.config.SaveLayout)
		if text == defaultPath {
			s.config.SaveBasePath = ""
		} else {
			s.config.SaveBasePath = text
		}
		internal.SaveConfig(s.config)
	})

	navPage := adw.NewNavigationPage(page, "Save Path")
	return navPage
}
