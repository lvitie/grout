package screens

import (
	"fmt"
	"grout/cache"
	"grout/desktop"
	"grout/internal"
	"grout/romm"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"go.uber.org/atomic"
)

type RebuildCacheScreen struct {
	router    *desktop.Router
	syncing   *atomic.Bool
	progress  *atomic.Float64
	syncError *atomic.String
}

func NewRebuildCacheScreen(router *desktop.Router) *RebuildCacheScreen {
	return &RebuildCacheScreen{
		router:    router,
		syncing:   atomic.NewBool(false),
		progress:  atomic.NewFloat64(0),
		syncError: atomic.NewString(""),
	}
}

func (s *RebuildCacheScreen) Build(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Rebuilding Cache")
	statusPage.SetDescription("Wiping local cache and re-syncing from server...")
	statusPage.SetIconName("view-refresh-symbolic")

	progress := gtk.NewProgressBar()
	progress.SetShowText(true)
	progress.SetMarginStart(40)
	progress.SetMarginEnd(40)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetVAlign(gtk.AlignCenter)
	box.Append(progress)
	statusPage.SetChild(box)

	if !s.syncing.Load() {
		s.startRebuild(router, statusPage, progress, box)
	}

	page := adw.NewNavigationPage(statusPage, "Rebuild Cache")
	return page
}

func (s *RebuildCacheScreen) startRebuild(router *desktop.Router, statusPage *adw.StatusPage, progress *gtk.ProgressBar, box *gtk.Box) {
	if s.syncing.Swap(true) {
		return
	}

	go func() {
		defer s.syncing.Store(false)

		glib.IdleAdd(func() {
			statusPage.SetDescription("Wiping local cache...")
		})

		err := cache.DeleteCacheFolder()
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Failed to delete cache: %v", err))
			return
		}

		// Re-initialize cache manager
		host := router.State().GetHost()
		if host == nil {
			s.syncError.Store("No host configured")
			return
		}

		cfg, _ := internal.LoadConfig()
		err = cache.InitCacheManager(*host, cfg)
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Failed to re-init cache: %v", err))
			return
		}

		cm := cache.GetCacheManager()
		client := romm.NewClientFromHost(*host)

		glib.IdleAdd(func() {
			statusPage.SetDescription("Fetching platforms...")
		})

		platforms, err := client.GetPlatforms()
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Failed to fetch platforms: %v", err))
			return
		}

		glib.IdleAdd(func() {
			statusPage.SetDescription("Downloading library metadata...")
		})

		// Start progress updater
		stopProgress := make(chan struct{})
		go func() {
			for {
				select {
				case <-stopProgress:
					return
				case <-time.After(100 * time.Millisecond):
					val := s.progress.Load()
					errStr := s.syncError.Load()
					glib.IdleAdd(func() {
						if errStr != "" {
							statusPage.SetTitle("Sync Failed")
							statusPage.SetDescription(errStr)
							statusPage.SetIconName("dialog-error-symbolic")
							return
						}
						progress.SetFraction(val)
					})
				}
			}
		}()
		defer close(stopProgress)

		_, err = cm.PopulateFullCacheWithProgress(platforms, s.progress, true)
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Sync failed: %v", err))
			return
		}

		s.progress.Store(1.0)
		glib.IdleAdd(func() {
			progress.SetFraction(1.0)
			statusPage.SetTitle("Rebuild Complete")
			statusPage.SetDescription("Your library has been successfully re-synced.")
			statusPage.SetIconName("emblem-ok-symbolic")

			doneBtn := gtk.NewButtonWithLabel("Done")
			doneBtn.AddCSSClass("suggested-action")
			doneBtn.SetMarginTop(20)
			doneBtn.ConnectClicked(func() {
				router.Back()
			})
			box.Append(doneBtn)
		})
	}()
}
