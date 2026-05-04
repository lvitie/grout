package screens

import (
	"grout/desktop"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type LoginScreen struct {
	router *desktop.Router
}

func NewLoginScreen(router *desktop.Router) *LoginScreen {
	return &LoginScreen{router: router}
}

func (s *LoginScreen) Build(router *desktop.Router) gtk.Widgetter {
	group := adw.NewPreferencesGroup()
	group.SetTitle("RomM Login")

	urlRow := adw.NewEntryRow()
	urlRow.SetTitle("Server URL")
	group.Add(urlRow)

	userRow := adw.NewEntryRow()
	userRow.SetTitle("Username")
	group.Add(userRow)

	passRow := adw.NewPasswordEntryRow()
	passRow.SetTitle("Password")
	group.Add(passRow)

	loginBtn := gtk.NewButtonWithLabel("Login")
	loginBtn.SetMarginTop(20)
	loginBtn.AddCSSClass("suggested-action")

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)
	box.SetMarginTop(20)
	box.Append(group)
	box.Append(loginBtn)

	clamp := adw.NewClamp()
	clamp.SetChild(box)

	page := adw.NewNavigationPage(clamp, "Login")
	return page
}
