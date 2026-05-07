package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/desktop/dialogs"
	"grout/internal"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
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

	loginBtn.ConnectClicked(func() {
		baseURL := urlRow.Text()
		username := userRow.Text()
		password := passRow.Text()

		host := romm.Host{
			RootURI:  baseURL,
			Username: username,
			Password: password,
		}

		client := romm.NewClientFromHost(host)
		if err := client.ValidateConnection(); err != nil {
			dialogs.ShowError(router.Window(), "Connection Failed", err.Error())
			return
		}

		if err := client.Login(username, password); err != nil {
			dialogs.ShowError(router.Window(), "Login Failed", err.Error())
			return
		}

		// Save to state and config
		router.State().SetHost(&host)
		cfg := router.State().GetConfig()
		cfg.Hosts = []romm.Host{host}
		internal.SaveConfig(cfg)
		cache.InitCacheManager(host, cfg)

		router.Navigate(NewPlatformSelectionScreen(router))
	})

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)
	box.SetMarginTop(20)
	box.Append(group)
	box.Append(loginBtn)

	clamp := adw.NewClamp()
	clamp.SetChild(box)

	page := adw.NewNavigationPage(clamp, "Login")
	glib.IdleAdd(func() {
		urlRow.GrabFocus()
	})
	return page
}
