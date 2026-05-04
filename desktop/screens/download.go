package screens

import (
	"fmt"
	"grout/desktop"
	"grout/internal/fileutil"
	"grout/platform"
	"grout/romm"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DownloadScreen struct {
	router *desktop.Router
	game   romm.Rom
}

func NewDownloadScreen(router *desktop.Router, game romm.Rom) *DownloadScreen {
	return &DownloadScreen{
		router: router,
		game:   game,
	}
}

func (s *DownloadScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle(s.game.Name)
	statusPage.SetDescription("Downloading game content...")
	statusPage.SetIconName("folder-download-symbolic")

	progress := gtk.NewProgressBar()
	progress.SetShowText(true)
	progress.SetFraction(0.0)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetValign(gtk.AlignCenter)
	box.Append(progress)

	statusPage.SetChild(box)

	// Start download in a goroutine
	go func() {
		state := router.State()
		host := state.GetHost()
		if host == nil {
			return
		}

		client := romm.NewClientFromHost(*host)
		data, err := client.DownloadRom(s.game.ID)
		if err != nil {
			glib.IdleAdd(func() {
				statusPage.SetDescription(fmt.Sprintf("Error: %v", err))
				statusPage.SetIconName("dialog-error-symbolic")
			})
			return
		}

		// Determine target path
		p := platform.GetCurrent()
		config := state.GetConfig()
		romDir := config.GetPlatformRomDirectory(romm.Platform{
			ID:     s.game.PlatformID,
			FSSlug: s.game.PlatformFSSlug,
		})

		targetPath := filepath.Join(romDir, s.game.FsName)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			glib.IdleAdd(func() {
				statusPage.SetDescription(fmt.Sprintf("Failed to create directory: %v", err))
			})
			return
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			glib.IdleAdd(func() {
				statusPage.SetDescription(fmt.Sprintf("Failed to write file: %v", err))
			})
			return
		}

		glib.IdleAdd(func() {
			progress.SetFraction(1.0)
			statusPage.SetDescription("Download complete!")
			statusPage.SetIconName("emblem-ok-symbolic")
		})
	}()

	page := adw.NewNavigationPage(statusPage, "Downloading")
	return page
}
