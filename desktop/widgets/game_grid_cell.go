package widgets

import (
	"grout/desktop"
	"grout/internal/artutil"
	"grout/romm"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameGridCell struct {
	*gtk.Box
	game romm.Rom
	img  *gtk.Image
}

func NewGameGridCell(game romm.Rom) *GameGridCell {
	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.SetSizeRequest(150, 200)
	box.SetHAlign(gtk.AlignCenter)
	box.SetVAlign(gtk.AlignStart)

	img := gtk.NewImage()
	img.SetPixelSize(128)
	img.SetFromIconName("image-missing-symbolic")
	box.Append(img)

	label := gtk.NewLabel(desktop.EscapeMarkup(game.Name))
	label.SetWrap(true)
	label.SetMaxWidthChars(15)
	label.SetJustify(gtk.JustifyCenter)
	box.Append(label)

	return &GameGridCell{
		Box:  box,
		game: game,
		img:  img,
	}
}

func NewGameGridCellWithArt(game romm.Rom, host *romm.Host) *GameGridCell {
	cell := NewGameGridCell(game)

	if host != nil {
		artURL := game.GetArtworkURL(artutil.ArtKindDefault, *host)
		if artURL != "" {
			desktop.LoadImageAsync(cell.img, artURL, 128)
		}
	}

	return cell
}

func (c *GameGridCell) GetGame() romm.Rom {
	return c.game
}
