package screens

import (
	"grout/desktop"
	"grout/internal"
	"grout/internal/artutil"
	"grout/romm"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameDetailsScreen struct {
	router *desktop.Router
	game   romm.Rom
}

func NewGameDetailsScreen(router *desktop.Router, game romm.Rom) *GameDetailsScreen {
	return &GameDetailsScreen{
		router: router,
		game:   game,
	}
}

func (s *GameDetailsScreen) uninstallGame(config *internal.Config) {
	platform := romm.Platform{
		ID:     s.game.PlatformID,
		FSSlug: s.game.PlatformFSSlug,
	}
	romDir := config.GetPlatformRomDirectory(platform)

	if s.game.HasMultipleFiles {
		m3uPath := filepath.Join(romDir, s.game.FsNameNoExt+".m3u")
		if err := os.Remove(m3uPath); err != nil && !os.IsNotExist(err) {
			slog.Error("Failed to remove m3u", "path", m3uPath, "error", err)
		}
		dirPath := filepath.Join(romDir, s.game.FsNameNoExt)
		if err := os.RemoveAll(dirPath); err != nil {
			slog.Error("Failed to remove multi-file dir", "path", dirPath, "error", err)
		}
	}

	for _, file := range s.game.Files {
		path := filepath.Join(romDir, file.FileName)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			slog.Error("Failed to remove ROM file", "path", path, "error", err)
		}
	}

	mainPath := filepath.Join(romDir, s.game.FsName)
	if err := os.Remove(mainPath); err != nil && !os.IsNotExist(err) {
		slog.Error("Failed to remove ROM", "path", mainPath, "error", err)
	}

	slog.Info("Game uninstalled", "name", s.game.Name)
}

func (s *GameDetailsScreen) Build(router *desktop.Router) gtk.Widgetter {
	clamp := adw.NewClamp()
	clamp.SetMaximumSize(800)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)

	title := gtk.NewLabel(s.game.Name)
	title.AddCSSClass("title-1")
	box.Append(title)

	// Artwork — load from server
	artImg := gtk.NewImage()
	artImg.SetPixelSize(300)
	artImg.SetFromIconName("image-missing-symbolic")
	box.Append(artImg)

	host := router.State().GetHost()
	if host != nil {
		artURL := s.game.GetArtworkURL(artutil.ArtKindDefault, *host)
		if artURL != "" {
			desktop.LoadImageAsync(artImg, artURL, 300)
		}
	}

	desc := gtk.NewLabel(s.game.Summary)
	desc.SetWrap(true)
	desc.SetMaxWidthChars(60)
	box.Append(desc)

	config := router.State().GetConfig()
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
	btnBox.SetHAlign(gtk.AlignCenter)

	downloadBtn := gtk.NewButtonWithLabel("Download")
	downloadBtn.AddCSSClass("suggested-action")

	uninstallBtn := gtk.NewButtonWithLabel("Uninstall")
	uninstallBtn.AddCSSClass("destructive-action")

	installed := config != nil && s.game.IsDownloaded(config)

	downloadBtn.SetVisible(!installed)
	uninstallBtn.SetVisible(installed)

	downloadBtn.ConnectClicked(func() {
		queue := desktop.GetDownloadQueue()
		queue.SetHostAndConfig(router.State().GetHost(), config)
		queue.Enqueue(s.game)
		router.Navigate(NewDownloadQueueScreen(router))
	})

	uninstallBtn.ConnectClicked(func() {
		dialog := adw.NewAlertDialog("Uninstall Game?", "This will delete the ROM files for "+desktop.EscapeMarkup(s.game.Name)+" from disk.")
		dialog.AddResponse("cancel", "Cancel")
		dialog.AddResponse("uninstall", "Uninstall")
		dialog.SetResponseAppearance("uninstall", adw.ResponseDestructive)
		dialog.SetDefaultResponse("cancel")
		dialog.SetCloseResponse("cancel")
		dialog.ConnectResponse(func(response string) {
			if response != "uninstall" {
				return
			}
			s.uninstallGame(config)
			downloadBtn.SetVisible(true)
			uninstallBtn.SetVisible(false)
			router.State().NotifyInstalledChanged()
		})
		dialog.Present(router.Window())
	})

	btnBox.Append(downloadBtn)
	btnBox.Append(uninstallBtn)
	box.Append(btnBox)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(box)

	page := adw.NewNavigationPage(scrolled, s.game.Name)
	return page
}
