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

type PlatformSelectionScreen struct {
	router    *desktop.Router
	platforms []romm.Platform
	syncing   *atomic.Bool
	progress  *atomic.Float64
	syncError *atomic.String
	statusText *atomic.String
	
	// UI elements that need updating
	stack             *adw.ViewStack
	progressBar       *gtk.ProgressBar
	listBox           *gtk.ListBox
	collectionsScreen *CollectionsScreen
}

func NewPlatformSelectionScreen(router *desktop.Router) *PlatformSelectionScreen {
	cm := cache.GetCacheManager()
	platforms, _ := cm.GetPlatforms()
	return &PlatformSelectionScreen{
		router:    router,
		platforms: platforms,
		syncing:   atomic.NewBool(false),
		progress:  atomic.NewFloat64(0),
		syncError: atomic.NewString(""),
		statusText: atomic.NewString(""),
	}
}

func (s *PlatformSelectionScreen) Build(router *desktop.Router) gtk.Widgetter {
	s.stack = adw.NewViewStack()

	// Platform List View
	listView := s.buildListView(router)
	s.stack.AddTitled(listView, "platforms", "Platforms")

	// Collections View
	s.collectionsScreen = NewCollectionsScreen(router)
	collectionsView := s.collectionsScreen.Build(router)
	s.stack.AddTitled(collectionsView, "collections", "Collections")

	// Sync View
	syncView := s.buildSyncView(router)
	s.stack.AddNamed(syncView, "sync")

	if len(s.platforms) == 0 || s.syncing.Load() {
		s.stack.SetVisibleChildName("sync")
	} else {
		s.stack.SetVisibleChildName("platforms")
	}

	// View Switcher Title
	switcherTitle := adw.NewViewSwitcherTitle()
	switcherTitle.SetStack(s.stack)
	switcherTitle.SetTitle("Grout")

	header := adw.NewHeaderBar()
	header.SetTitleWidget(switcherTitle)

	syncBtn := gtk.NewButtonFromIconName("view-refresh-symbolic")
	syncBtn.ConnectClicked(func() {
		s.stack.SetVisibleChildName("sync")
		s.startSync(router)
	})
	header.PackEnd(syncBtn)

	settingsBtn := gtk.NewButtonFromIconName("emblem-system-symbolic")
	settingsBtn.ConnectClicked(func() {
		router.Navigate(NewSettingsScreen(router))
	})
	header.PackEnd(settingsBtn)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(s.stack)

	page := adw.NewNavigationPage(box, "Grout")
	return page
}

func (s *PlatformSelectionScreen) buildListView(router *desktop.Router) gtk.Widgetter {
	s.listBox = gtk.NewListBox()
	s.listBox.SetSelectionMode(gtk.SelectionSingle)
	s.listBox.AddCSSClass("navigation-sidebar")

	s.refreshPlatformList()

	s.listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.platforms) {
			router.Navigate(NewGameListScreen(router, s.platforms[idx]))
		}
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(s.listBox)
	scrolled.SetVExpand(true)
	scrolled.SetHExpand(true)

	s.progressBar = gtk.NewProgressBar()
	s.progressBar.SetVisible(false)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(s.progressBar)
	box.Append(scrolled)

	return box
}

func (s *PlatformSelectionScreen) buildSyncView(router *desktop.Router) gtk.Widgetter {
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("Syncing Library")
	statusPage.SetDescription("Initializing...")
	statusPage.SetIconName("view-refresh-symbolic")

	progress := gtk.NewProgressBar()
	progress.SetShowText(true)
	progress.SetMarginStart(40)
	progress.SetMarginEnd(40)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetVAlign(gtk.AlignCenter)
	box.Append(progress)
	statusPage.SetChild(box)

	// Start progress monitoring
	s.startProgressMonitor(statusPage, progress)

	cm := cache.GetCacheManager()
	if !s.syncing.Load() && !cm.HasPlatforms() {
		s.startSync(router)
	}

	return statusPage
}

func (s *PlatformSelectionScreen) startProgressMonitor(statusPage *adw.StatusPage, progress *gtk.ProgressBar) {
	go func() {
		for {
			isSyncing := s.syncing.Load()
			errStr := s.syncError.Load()
			val := s.progress.Load()
			statusMsg := s.statusText.Load()
			
			glib.IdleAdd(func() {
				if errStr != "" {
					statusPage.SetTitle("Sync Failed")
					statusPage.SetDescription(desktop.EscapeMarkup(errStr))
					statusPage.SetIconName("dialog-error-symbolic")
				} else {
					if isSyncing {
						statusPage.SetTitle("Syncing Library")
						statusPage.SetIconName("view-refresh-symbolic")
					}
					if statusMsg != "" {
						statusPage.SetDescription(desktop.EscapeMarkup(statusMsg))
					}
					progress.SetFraction(val)
				}
				
				if s.progressBar != nil {
					s.progressBar.SetFraction(val)
					s.progressBar.SetVisible(isSyncing)
				}
			})

			if !isSyncing && val >= 1.0 {
				break
			}
			if !isSyncing && errStr != "" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		
		glib.IdleAdd(func() {
			if s.syncError.Load() == "" && s.progress.Load() >= 1.0 {
				progress.SetFraction(1.0)
				statusPage.SetTitle("Sync Complete")
				statusPage.SetDescription("Library updated successfully.")
				statusPage.SetIconName("emblem-ok-symbolic")
				
				s.refreshPlatformList()
				
				// Auto-switch back to list after a short delay if it was an empty start
				time.AfterFunc(1500*time.Millisecond, func() {
					glib.IdleAdd(func() {
						s.stack.SetVisibleChildName("platforms")
					})
				})
			}
			
			if s.progressBar != nil {
				s.progressBar.SetVisible(false)
			}
		})
	}()
}

func (s *PlatformSelectionScreen) refreshPlatformList() {
	if s.listBox == nil {
		return
	}

	cm := cache.GetCacheManager()
	var err error
	allPlatforms, err := cm.GetPlatforms()
	
	logger := internal.GetLogger()
	if err != nil {
		logger.Error("Failed to load platforms from cache", "error", err)
	}

	// Show all platforms from the database
	s.platforms = allPlatforms
	
	logger.Info("Refreshing platform list", 
		"visible_platforms", len(s.platforms), 
		"total_platforms", len(allPlatforms),
		"db_path", cm.GetDBPath())

	if s.collectionsScreen != nil {
		s.collectionsScreen.Refresh()
	}

	// Clear list
	for {
		row := s.listBox.FirstChild()
		if row == nil {
			break
		}
		s.listBox.Remove(row)
	}

	// Add Platform rows
	for _, p := range s.platforms {
		row := adw.NewActionRow()
		row.SetTitle(desktop.EscapeMarkup(p.Name))
		row.SetSubtitle(desktop.EscapeMarkup(fmt.Sprintf("%d games", p.ROMCount)))
		row.SetActivatable(true)
		s.listBox.Append(row)
	}
}

func (s *PlatformSelectionScreen) startSync(router *desktop.Router) {
	if s.syncing.Swap(true) {
		return // Already syncing
	}

	go func() {
		defer s.syncing.Store(false)
		s.progress.Store(0)
		s.syncError.Store("")
		s.statusText.Store("Connecting to RomM...")

		cm := cache.GetCacheManager()
		host := router.State().GetHost()
		if host == nil {
			s.syncError.Store("No RomM host configured")
			return
		}
		
		client := romm.NewClientFromHost(*host)
		
		s.statusText.Store("Fetching platforms...")
		rommPlatforms, err := client.GetPlatforms()
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Failed to fetch platforms: %v", err))
			return
		}
		
		// If we have platforms, disambiguate them
		romm.DisambiguatePlatformNames(rommPlatforms)

		s.statusText.Store("Downloading library metadata...")
		_, err = cm.PopulateFullCacheWithProgress(rommPlatforms, s.progress, false)
		if err != nil {
			s.syncError.Store(fmt.Sprintf("Sync failed: %v", err))
			return
		}

		s.progress.Store(1.0)
		s.statusText.Store("Sync complete!")
	}()
}
