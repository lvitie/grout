package main

import (
	"grout/desktop"
	"grout/desktop/controller"
	"grout/desktop/screens"
	"grout/internal"
	"grout/platform"
	"os"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	// Initialize platform
	platform.SetCurrent(platform.NewLinuxDesktop())

	// Initialize logger
	internal.InitLogger(0) // Info level

	// Initialize state
	state := desktop.NewAppState()

	app := adw.NewApplication("app.romm.Grout", 0)
	app.ConnectActivate(func() {
		activate(app, state)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *adw.Application, state *desktop.AppState) {
	window := adw.NewApplicationWindow(app)
	window.SetTitle("Grout")
	window.SetDefaultSize(1280, 720)

	handler := controller.NewHandler()
	handler.Start(func(action controller.Action) {
		// Handle global actions here if needed
		fmt.Printf("Action: %v\n", action)
	})

	// Create router and show first screen
	router := desktop.NewRouter(window, state)
	router.Navigate(screens.NewLoginScreen(router))

	window.Present()
}
