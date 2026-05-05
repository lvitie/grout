package widgets

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ProgressOverlay struct {
	*gtk.Overlay
	progress *gtk.ProgressBar
}

func NewProgressOverlay(child gtk.Widgetter) *ProgressOverlay {
	overlay := gtk.NewOverlay()
	overlay.SetChild(child)

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetVAlign(gtk.AlignEnd)
	box.SetHAlign(gtk.AlignCenter)
	box.SetMarginBottom(20)
	box.AddCSSClass("card") // Adwaita card style

	progress := gtk.NewProgressBar()
	progress.SetSizeRequest(300, -1)
	box.Append(progress)

	overlay.AddOverlay(box)

	return &ProgressOverlay{
		Overlay:  overlay,
		progress: progress,
	}
}

func (p *ProgressOverlay) SetFraction(f float64) {
	p.progress.SetFraction(f)
}
