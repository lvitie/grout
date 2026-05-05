package dialogs

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func ShowError(parent *gtk.Window, title, body string) {
	dialog := adw.NewMessageDialog(parent, title, body)
	dialog.AddResponse("ok", "OK")
	dialog.SetDefaultResponse("ok")
	dialog.SetResponseAppearance("ok", adw.ResponseDestructive)

	dialog.ConnectResponse(func(response string) {
		dialog.Destroy()
	})

	dialog.Present()
}
