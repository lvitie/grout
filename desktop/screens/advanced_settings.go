package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
	"grout/internal"
)

type AdvancedSettingsScreen struct {
	router *desktop.Router
	config *internal.Config
}

func NewAdvancedSettingsScreen(router *desktop.Router) *AdvancedSettingsScreen {
	config, _ := internal.LoadConfig()
	return &AdvancedSettingsScreen{
		router: router,
		config: config,
	}
}

func (s *AdvancedSettingsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle("Advanced")

	// Network Group
	networkGroup := adw.NewPreferencesGroup()
	networkGroup.SetTitle("Network")
	page.Add(networkGroup)

	apiTimeoutRow := adw.NewActionRow()
	apiTimeoutRow.SetTitle("API Timeout")
	apiTimeoutRow.SetSubtitle("Seconds")
	// Use an adw.SpinRow or similar if available, or just a button for now
	networkGroup.Add(apiTimeoutRow)

	// Logging Group
	loggingGroup := adw.NewPreferencesGroup()
	loggingGroup.SetTitle("Logging")
	page.Add(loggingGroup)

	logLevelRow := adw.NewComboRow()
	logLevelRow.SetTitle("Log Level")
	// Populate with LogLevel enum values
	loggingGroup.Add(logLevelRow)

	navPage := adw.NewNavigationPage(page, "Advanced")
	return navPage
}
