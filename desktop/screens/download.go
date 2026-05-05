package screens

import (
	"fmt"
	"grout/desktop"
	"grout/romm"
	"io"
	"log/slog"
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
	box.SetVAlign(gtk.AlignCenter)
	box.Append(progress)

	statusPage.SetChild(box)

	// Start download in a goroutine
	go func() {
		state := router.State()
		host := state.GetHost()
		if host == nil {
			return
		}

		slog.Info("Starting download", "gameID", s.game.ID, "gameName", s.game.Name, "host", host.URL())

		client := romm.NewClientFromHost(*host)
		resp, err := client.DownloadRomStream(s.game)
		if err != nil {
			glib.IdleAdd(func() {
				statusPage.SetDescription(fmt.Sprintf("Error: %v", err))
				statusPage.SetIconName("dialog-error-symbolic")
			})
			return
		}
		defer resp.Body.Close()

		totalSize := resp.ContentLength

		// Determine target path
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

		out, err := os.Create(targetPath)
		if err != nil {
			glib.IdleAdd(func() {
				statusPage.SetDescription(fmt.Sprintf("Failed to create file: %v", err))
			})
			return
		}
		defer out.Close()

		// Proxy the body with progress updates
		var downloaded int64
		buffer := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				if _, werr := out.Write(buffer[:n]); werr != nil {
					glib.IdleAdd(func() {
						statusPage.SetDescription(fmt.Sprintf("Failed to write file: %v", werr))
					})
					return
				}
				downloaded += int64(n)
				if totalSize > 0 {
					frac := float64(downloaded) / float64(totalSize)
					glib.IdleAdd(func() {
						progress.SetFraction(frac)
					})
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				glib.IdleAdd(func() {
					statusPage.SetDescription(fmt.Sprintf("Download failed: %v", err))
				})
				return
			}
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
