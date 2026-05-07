package main

import (
	"grout/cache"
	"grout/desktop"
	"grout/desktop/input"
	"grout/desktop/screens"
	"grout/internal"
	"grout/platform"
	"grout/resources"
	"os"

	"fyne.io/systray"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	// Initialize platform
	platform.SetCurrent(platform.NewLinuxDesktop())

	// Initialize logger
	internal.InitLogger(-4) // Debug level

	// Initialize state
	state := desktop.NewAppState()

	app := adw.NewApplication("app.romm.Grout", 0)

	var window *adw.ApplicationWindow

	startTray, stopTray := systray.RunWithExternalLoop(func() {
		iconBytes, err := resources.GetAppIconBytes()
		if err == nil {
			systray.SetIcon(iconBytes)
		}
		systray.SetTitle("Grout")
		systray.SetTooltip("Grout — RomM Client")

		mShow := systray.AddMenuItem("Show", "Show Grout window")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit Grout")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					glib.IdleAdd(func() {
						if window != nil {
							window.SetVisible(true)
							window.Present()
						}
					})
				case <-mQuit.ClickedCh:
					glib.IdleAdd(func() {
						app.Quit()
					})
				}
			}
		}()
	}, func() {})

	app.ConnectActivate(func() {
		window = activateWindow(app, state)
		startTray()

		window.ConnectCloseRequest(func() bool {
			config := state.GetConfig()
			if config != nil && config.ShouldCloseToTray() {
				window.SetVisible(false)
				return true
			}
			return false
		})
	})

	app.ConnectShutdown(func() {
		stopTray()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activateWindow(app *adw.Application, state *desktop.AppState) *adw.ApplicationWindow {
	window := adw.NewApplicationWindow(&app.Application)
	window.SetTitle("Grout")
	window.SetDefaultSize(1280, 720)
	window.SetIconName("app.romm.Grout")

	// Add resources to icon theme for development (go run)
	display := gdk.DisplayGetDefault()
	if display != nil {
		theme := gtk.IconThemeGetForDisplay(display)
		theme.AddSearchPath("resources")
	}

	// Create router and show first screen
	router := desktop.NewRouter(window, state)

	// Initialize input system (gamepad + keyboard)
	inputMgr := input.NewManager(window, input.Callbacks{
		Back: router.Back,
		TabLeft: func() {
			if router.TabLeftFn != nil {
				router.TabLeftFn()
			}
		},
		TabRight: func() {
			if router.TabRightFn != nil {
				router.TabRightFn()
			}
		},
		ToggleView: func() {
			if router.ToggleViewFn != nil {
				router.ToggleViewFn()
			}
		},
		QuickMenu: func() {
			if router.QuickMenuFn != nil {
				router.QuickMenuFn()
			}
		},
		ExitSearch: func() bool {
			if router.IsSearchActive == nil || !router.IsSearchActive() {
				return false
			}
			router.ClearSearch()
			return true
		},
		FocusContent: func() {
			if router.FocusContent != nil {
				router.FocusContent()
			}
		},
	})
	inputMgr.Start()

	window.ConnectDestroy(func() {
		inputMgr.Stop()
	})

	host := state.GetHost()
	if host != nil {
		cache.InitCacheManager(*host, state.GetConfig())
		router.Navigate(screens.NewPlatformSelectionScreen(router))
	} else {
		router.Navigate(screens.NewLoginScreen(router))
	}

	window.Present()
	return window
}
