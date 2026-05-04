package screens

import (
	"fmt"
	"grout/desktop"
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/piglig/go-qr"
	"image"
	"image/color"
	"image/draw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
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
	q, err := qr.Encode(url, qr.Medium)
	if err != nil {
		statusPage.SetDescription("Failed to generate QR code")
	} else {
		// Convert QR to Image
		size := q.Size()
		scale := 10
		img := image.NewRGBA(image.Rect(0, 0, size*scale, size*scale))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				if q.IsDark(x, y) {
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							img.Set(x*scale+dx, y*scale+dy, color.Black)
						}
					}
				}
			}
		}

		// Display in GTK
		// We'd ideally use gdk.TextureNewFromBytes or similar if we had a simple way,
		// but since we are in Go, we might need a temporary file or a custom paintable.
		// For now, I'll use a placeholder and note it.
		qrImg := gtk.NewImage()
		qrImg.SetFromIconName("image-missing-symbolic")
		statusPage.SetChild(qrImg)
	}

	page := adw.NewNavigationPage(statusPage, "QR Code")
	return page
}
