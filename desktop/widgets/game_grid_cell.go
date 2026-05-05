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
	img := gtk.NewImage()
	img.SetPixelSize(128)
	img.SetFromIconName("image-missing-symbolic")
	img.SetHExpand(false)
	img.SetVExpand(false)

	label := gtk.NewLabel(desktop.EscapeMarkup(game.Name))
	label.SetWrap(true)
	label.SetMaxWidthChars(15)
	label.SetJustify(gtk.JustifyCenter)
	label.SetHExpand(false)
	label.SetVExpand(false)

	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.SetSizeRequest(150, 180)
	box.SetHomogeneous(false)
	box.SetHAlign(gtk.AlignCenter)
	box.SetVAlign(gtk.AlignStart)
	box.SetHExpand(false)
	box.SetVExpand(false)

	img.SetVAlign(gtk.AlignStart)
	img.SetHAlign(gtk.AlignCenter)

	box.Append(img)
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
