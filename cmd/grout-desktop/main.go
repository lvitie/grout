package main

import (
	"grout/desktop"
	"grout/internal"
	"grout/platform"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	// Initialize platform
	platform.SetCurrent(platform.NewLinuxDesktop())

	// Initialize logger
	internal.InitLogger(0) // Info level

	app := adw.NewApplication("app.romm.Grout", 0)
	app.ConnectActivate(func() {
		activate(app)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *adw.Application) {
	window := adw.NewApplicationWindow(app)
	window.SetTitle("Grout")
	window.SetDefaultSize(1280, 720)

	// Create router and show first screen
	router := desktop.NewRouter(window)
	router.ShowFirstScreen()

	window.Present()
}
