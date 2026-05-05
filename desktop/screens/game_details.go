package screens

import (
	"grout/desktop"
	"grout/internal/artutil"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameDetailsScreen struct {
	router *desktop.Router
	game   romm.Rom
}

func NewGameDetailsScreen(router *desktop.Router, game romm.Rom) *GameDetailsScreen {
	return &GameDetailsScreen{
		router: router,
		game:   game,
	}
}

func (s *GameDetailsScreen) Build(router *desktop.Router) gtk.Widgetter {
	clamp := adw.NewClamp()
	clamp.SetMaximumSize(800)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)

	title := gtk.NewLabel(s.game.Name)
	title.AddCSSClass("title-1")
	box.Append(title)

	// Artwork — load from server
	artImg := gtk.NewImage()
	artImg.SetPixelSize(300)
	artImg.SetFromIconName("image-missing-symbolic")
	box.Append(artImg)

	host := router.State().GetHost()
	if host != nil {
		artURL := s.game.GetArtworkURL(artutil.ArtKindDefault, *host)
		if artURL != "" {
			desktop.LoadImageAsync(artImg, artURL, 300)
		}
	}

	desc := gtk.NewLabel(s.game.Summary)
	desc.SetWrap(true)
	desc.SetMaxWidthChars(60)
	box.Append(desc)

	downloadBtn := gtk.NewButtonWithLabel("Download")
	downloadBtn.AddCSSClass("suggested-action")
	downloadBtn.ConnectClicked(func() {
		router.Navigate(NewDownloadScreen(router, s.game))
	})
	box.Append(downloadBtn)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(box)

	page := adw.NewNavigationPage(scrolled, s.game.Name)
	return page
}
