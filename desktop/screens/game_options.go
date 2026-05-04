package screens

import (
	"grout/desktop"
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameOptionsScreen struct {
	router *desktop.Router
	game   romm.Rom
}

func NewGameOptionsScreen(router *desktop.Router, game romm.Rom) *GameOptionsScreen {
	return &GameOptionsScreen{
		router: router,
		game:   game,
	}
}

func (s *GameOptionsScreen) Build(router *desktop.Router) gtk.Widgetter {
	page := adw.NewPreferencesPage()
	page.SetTitle(s.game.Name)

	group := adw.NewPreferencesGroup()
	group.SetTitle("Options")
	page.Add(group)

	qrRow := adw.NewActionRow()
	qrRow.SetTitle("Show QR Code")
	qrRow.SetSubtitle("Transfer game to another device")
	qrRow.SetActivatable(true)
	qrRow.ConnectActivated(func() {
		router.Navigate(NewGameQRScreen(router, s.game))
	})
	group.Add(qrRow)

	saveRow := adw.NewActionRow()
	saveRow.SetTitle("Manage Saves")
	saveRow.SetActivatable(true)
	group.Add(saveRow)

	navPage := adw.NewNavigationPage(page, "Options")
	return navPage
}
