package screens

import (
	"fmt"
	"grout/desktop"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	qrcode "github.com/piglig/go-qr"
)

type GameQRScreen struct {
	router *desktop.Router
	game   romm.Rom
}

func NewGameQRScreen(router *desktop.Router, game romm.Rom) *GameQRScreen {
	return &GameQRScreen{
		router: router,
		game:   game,
	}
}

func (s *GameQRScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Game QR Code")
	statusPage.SetDescription(fmt.Sprintf("Scan this code to download %s", s.game.Name))

	// Generate QR code
	url := s.game.GetGamePage(*router.State().GetHost())
	_, err := qrcode.EncodeText(url, qrcode.Medium)
	if err != nil {
		statusPage.SetDescription("Failed to generate QR code")
	} else {
		// TODO: Convert QR code to GdkTexture and display in a gtk.Picture
		qrImg := gtk.NewImage()
		qrImg.SetFromIconName("image-missing-symbolic")
		qrImg.SetPixelSize(256)
		statusPage.SetChild(qrImg)
	}

	page := adw.NewNavigationPage(statusPage, "QR Code")
	return page
}
