package widgets

import (
	"grout/romm"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameRow struct {
	*adw.ActionRow
	game romm.Rom
}

func NewGameRow(game romm.Rom) *GameRow {
	row := adw.NewActionRow()
	row.SetTitle(game.Name)
	row.SetSubtitle(game.FsName)
	
	// Add a thumbnail placeholder
	img := gtk.NewImage()
	img.SetFromIconName("image-missing-symbolic")
	img.SetPixelSize(48)
	row.AddPrefix(img)
	
	return &GameRow{
		ActionRow: row,
		game:      game,
	}
}
