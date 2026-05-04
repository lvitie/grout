package dialogs

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func ShowConfirmation(parent *gtk.Window, title, body, confirmLabel string, onConfirm func()) {
	dialog := adw.NewMessageDialog(parent, title, body)
	dialog.AddResponse("cancel", "Cancel")
	dialog.AddResponse("confirm", confirmLabel)
	dialog.SetResponseAppearance("confirm", adw.ResponseSuggested)
	dialog.SetDefaultResponse("confirm")
	dialog.SetCloseResponse("cancel")
	
	dialog.ConnectResponse(func(response string) {
		if response == "confirm" {
			onConfirm()
		}
		dialog.Destroy()
	})
	
	dialog.Present()
}
