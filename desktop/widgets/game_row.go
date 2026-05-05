package widgets

import (
	"grout/desktop"
	"grout/internal/artutil"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameRow struct {
	*adw.ActionRow
	game romm.Rom
	img  *gtk.Image
}

func NewGameRow(game romm.Rom) *GameRow {
	row := adw.NewActionRow()
	row.SetTitle(desktop.EscapeMarkup(game.Name))
	row.SetSubtitle(desktop.EscapeMarkup(game.FsName))

	img := gtk.NewImage()
	img.SetPixelSize(48)
	img.SetFromIconName("image-missing-symbolic")
	row.AddPrefix(img)

	return &GameRow{
		ActionRow: row,
		game:      game,
		img:       img,
	}
}

// NewGameRowWithArt creates a game row and kicks off async artwork loading.
func NewGameRowWithArt(game romm.Rom, host *romm.Host) *GameRow {
	row := NewGameRow(game)

	if host != nil {
		artURL := game.GetArtworkURL(artutil.ArtKindDefault, *host)
		if artURL != "" {
			desktop.LoadImageAsync(row.img, artURL, 48)
		}
	}

	return row
}
