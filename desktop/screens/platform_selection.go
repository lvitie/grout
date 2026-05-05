package screens

import (
	"fmt"
	"grout/cache"
	"grout/desktop"
	"grout/internal"
	"grout/romm"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"go.uber.org/atomic"
	"grout/resources"
)

type PlatformSelectionScreen struct {
	router     *desktop.Router
	platforms  []romm.Platform
	syncing    *atomic.Bool
	progress   *atomic.Float64
	syncError  *atomic.String
	statusText *atomic.String

	// UI elements that need updating
	stack              *adw.ViewStack
	progressBar        *gtk.ProgressBar
	listBox            *gtk.ListBox
	platformListStack  *gtk.Stack
	collectionsScreen  *CollectionsScreen
}

func NewPlatformSelectionScreen(router *desktop.Router) *PlatformSelectionScreen {
	cm := cache.GetCacheManager()
	platforms, _ := cm.GetPlatforms()
	return &PlatformSelectionScreen{
		router:     router,
		platforms:  platforms,
		syncing:    atomic.NewBool(false),
		progress:   atomic.NewFloat64(0),
		syncError:  atomic.NewString(""),
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

	// Search bar
	searchBar := gtk.NewSearchEntry()
	searchBar.SetPlaceholderText("Search...")

	s.listBox.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
		text := strings.ToLower(searchBar.Text())
		if text == "" {
			return true
		}
		if actionRow, ok := row.Child().(*adw.ActionRow); ok {
			title := strings.ToLower(actionRow.Title())
			return strings.Contains(title, text)
		}
		return true
	})

	searchBar.ConnectChanged(func() {
		s.listBox.InvalidateFilter()
	})

	// Toggle view button
	isGridView := router.State().GetViewMode() == desktop.ViewModeGrid
	iconName := "view-grid-symbolic"
	if isGridView {
		iconName = "view-list-symbolic"
	}

	toggleBtn := gtk.NewButtonFromIconName(iconName)
	toggleBtn.SetTooltipText("Toggle grid/list view")

	toggleBtn.ConnectClicked(func() {
		isGridView = !isGridView
		if isGridView {
			s.platformListStack.SetVisibleChildName("grid")
			if s.collectionsScreen != nil {
				s.collectionsScreen.ShowGridView()
			}
			router.State().SetViewMode(desktop.ViewModeGrid)
			toggleBtn.SetIconName("view-list-symbolic")
		} else {
			s.platformListStack.SetVisibleChildName("list")
			if s.collectionsScreen != nil {
				s.collectionsScreen.ShowListView()
			}
			router.State().SetViewMode(desktop.ViewModeList)
			toggleBtn.SetIconName("view-grid-symbolic")
		}
	})

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.SetMarginStart(6)
	headerBox.SetMarginEnd(6)
	headerBox.Append(searchBar)
	headerBox.SetHExpand(true)
	headerBox.Append(toggleBtn)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(switcherTitle)
	header.PackStart(headerBox)

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

	listScrolled := gtk.NewScrolledWindow()
	listScrolled.SetChild(s.listBox)
	listScrolled.SetVExpand(true)
	listScrolled.SetHExpand(true)

	// Grid view for platforms
	flowBox := gtk.NewFlowBox()
	flowBox.SetSelectionMode(gtk.SelectionSingle)
	flowBox.SetMinChildrenPerLine(2)
	flowBox.SetMaxChildrenPerLine(10)
	flowBox.SetRowSpacing(12)
	flowBox.SetColumnSpacing(12)
	flowBox.SetMarginStart(12)
	flowBox.SetMarginEnd(12)
	flowBox.SetMarginTop(12)
	flowBox.SetMarginBottom(12)

	for _, p := range s.platforms {
		cell := gtk.NewBox(gtk.OrientationVertical, 6)
		cell.SetSizeRequest(150, -1)

		label := gtk.NewLabel(desktop.EscapeMarkup(p.Name))
		label.SetWrap(true)
		label.SetMaxWidthChars(15)
		label.SetJustify(gtk.JustifyCenter)
		cell.Append(label)

		flowBox.Append(cell)
	}

	flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		for idx, p := range s.platforms {
			if flowBox.ChildAtIndex(idx) == child {
				router.Navigate(NewGameListScreen(router, p))
				break
			}
		}
	})

	gridScrolled := gtk.NewScrolledWindow()
	gridScrolled.SetChild(flowBox)
	gridScrolled.SetVExpand(true)
	gridScrolled.SetHExpand(true)

	// Stack to hold list and grid views
	s.platformListStack = gtk.NewStack()
	s.platformListStack.AddNamed(listScrolled, "list")
	s.platformListStack.AddNamed(gridScrolled, "grid")

	// Set initial view
	if router.State().GetViewMode() == desktop.ViewModeGrid {
		s.platformListStack.SetVisibleChildName("grid")
	} else {
		s.platformListStack.SetVisibleChildName("list")
	}

	s.progressBar = gtk.NewProgressBar()
	s.progressBar.SetVisible(false)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(s.progressBar)
	box.Append(s.platformListStack)

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

	// If the list is empty, don't filter out 0-game platforms (helps debugging/initial sync)
	s.platforms = make([]romm.Platform, 0)
	for _, p := range allPlatforms {
		if p.ROMCount > 0 || len(allPlatforms) < 10 { // Show all if we have very few, or if they have games
			s.platforms = append(s.platforms, p)
		}
	}

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

		// Add icon prefix
		iconImg := gtk.NewImageFromIconName("application-x-executable-symbolic")
		iconImg.SetPixelSize(32)
		iconImg.SetMarginEnd(12)
		row.AddPrefix(iconImg)

		// Try to load icon from embedded resources
		logoFilename := ""
		if p.LogoPath != "" {
			logoFilename = strings.TrimSuffix(filepath.Base(p.LogoPath), filepath.Ext(p.LogoPath))
		}

		// Helper to normalize strings for better matching (e.g. "Atari 2600" -> "atari2600")
		normalize := func(s string) string {
			return strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
					return r
				}
				if r >= 'A' && r <= 'Z' {
					return r + ('a' - 'A')
				}
				return -1
			}, s)
		}

		// Helper to slugify (e.g. "Atari Jaguar CD" -> "atari-jaguar-cd")
		slugify := func(s string) string {
			return strings.ToLower(strings.Join(strings.Fields(s), "-"))
		}

		// Common platform aliases lookup table
		type aliasRule struct {
			contains string
			aliases  []string
		}
		rules := []aliasRule{
			// Sony
			{"playstation 5", []string{"ps5"}},
			{"playstation 4", []string{"ps4"}},
			{"playstation 3", []string{"ps3"}},
			{"playstation 2", []string{"ps2"}},
			{"playstation portable", []string{"psp"}},
			{"playstation vita", []string{"psvita"}},
			{"playstation", []string{"psx", "ps1", "ps"}},

			// Nintendo
			{"nintendo 3ds", []string{"3ds"}},
			{"nintendo ds", []string{"nds"}},
			{"nintendo 64", []string{"n64"}},
			{"game boy advance", []string{"gba"}},
			{"game boy color", []string{"gbc"}},
			{"game boy", []string{"gb"}},
			{"gamecube", []string{"ngc", "gc"}},
			{"wii u", []string{"wiiu"}},
			{"wii", []string{"wii"}},
			{"switch", []string{"switch"}},
			{"virtual boy", []string{"vb", "virtualboy"}},
			{"super nintendo", []string{"snes"}},
			{"snes", []string{"snes"}},
			{"famicom disk system", []string{"fds"}},
			{"famicom", []string{"famicom", "nes", "sfam"}},
			{"super famicom", []string{"sfam", "sfc"}},
			{"nintendo entertainment system", []string{"nes"}},

			// Sega
			{"dreamcast", []string{"dc", "dreamcast"}},
			{"saturn", []string{"saturn"}},
			{"mega drive", []string{"genesis", "megadrive", "md"}},
			{"genesis", []string{"genesis", "megadrive", "md"}},
			{"master system", []string{"sms", "ms"}},
			{"game gear", []string{"gamegear", "gg"}},
			{"sega cd", []string{"segacd", "scd"}},
			{"sega 32x", []string{"32x"}},
			{"sg-1000", []string{"sg1000"}},

			// Microsoft
			{"xbox series x", []string{"series-x"}},
			{"xbox series s", []string{"series-s"}},
			{"xbox one", []string{"xboxone"}},
			{"xbox 360", []string{"xbox360"}},
			{"xbox", []string{"xbox"}},
			{"ms-dos", []string{"dos"}},
			{"dos", []string{"dos"}},

			// Atari
			{"atari 2600", []string{"atari2600", "atari-2600"}},
			{"atari 5200", []string{"atari5200", "atari-5200"}},
			{"atari 7800", []string{"atari7800", "atari-7800"}},
			{"atari 800", []string{"atari800", "atari-800"}},
			{"atari lynx", []string{"lynx", "atarilynx"}},
			{"atari jaguar cd", []string{"atari-jaguar-cd"}},
			{"atari jaguar", []string{"jaguar", "atarijaguar"}},

			// Commodore
			{"commodore 64", []string{"c64", "commodore-64"}},
			{"amiga cd32", []string{"amiga-cd32"}},
			{"amiga", []string{"amiga"}},

			// Others
			{"arcade", []string{"arcade", "fbneo", "mame"}},
			{"3do", []string{"3do"}},
			{"neo geo aes", []string{"neogeoaes"}},
			{"neo geo mvs", []string{"neogeomvs"}},
			{"neo geo cd", []string{"neo-geo-cd", "neocd"}},
			{"neo geo pocket color", []string{"neo-geo-pocket-color"}},
			{"neo geo pocket", []string{"neo-geo-pocket"}},
			{"neo geo x", []string{"neo-geo-x"}},
			{"neo geo", []string{"neogeoaes", "neogeomvs", "neo-geo-x", "neogeo"}},
			{"pc engine", []string{"pce", "pcengine"}},
			{"turbografx", []string{"tg16", "turbografx"}},
			{"wonderswan color", []string{"wonderswan-color"}},
			{"wonderswan", []string{"wonderswan"}},
			{"colecovision", []string{"colecovision"}},
			{"intellivision", []string{"intellivision"}},
			{"supergrafx", []string{"supergrafx", "sgfx"}},
		}

		aliases := []string{}
		lowerName := strings.ToLower(p.Name)
		for _, rule := range rules {
			if strings.Contains(lowerName, rule.contains) {
				aliases = append(aliases, rule.aliases...)
				break // Use the first matching rule (most specific should be first)
			}
		}

		// Build comprehensive candidate list with many variations
		candidateSet := make(map[string]bool) // Deduplicate
		addCandidate := func(s string) {
			if s != "" {
				candidateSet[s] = true
			}
		}

		// Primary candidates
		addCandidate(logoFilename)
		addCandidate(p.ShortName)
		addCandidate(p.Slug)
		addCandidate(p.FSSlug)
		addCandidate(p.Name)
		addCandidate(p.ApiName)

		// Normalized variations
		addCandidate(normalize(p.Name))
		addCandidate(normalize(p.ApiName))
		addCandidate(normalize(p.ShortName))
		addCandidate(normalize(p.Slug))
		addCandidate(normalize(p.FSSlug))

		// Slugified variations
		addCandidate(slugify(p.Name))
		addCandidate(slugify(p.ApiName))
		addCandidate(slugify(p.ShortName))
		addCandidate(slugify(p.Slug))
		addCandidate(slugify(p.FSSlug))

		// Space-to-hyphen variations
		addCandidate(strings.ReplaceAll(p.Name, " ", "-"))
		addCandidate(strings.ReplaceAll(p.ApiName, " ", "-"))
		addCandidate(strings.ReplaceAll(p.ShortName, " ", "-"))

		// Space-to-nothing variations
		addCandidate(strings.ReplaceAll(p.Name, " ", ""))
		addCandidate(strings.ReplaceAll(p.ApiName, " ", ""))
		addCandidate(strings.ReplaceAll(p.ShortName, " ", ""))

		// Lowercase direct
		addCandidate(strings.ToLower(p.Name))
		addCandidate(strings.ToLower(p.ApiName))
		addCandidate(strings.ToLower(p.ShortName))

		// Add aliases from rules
		for _, a := range aliases {
			addCandidate(a)
		}

		// Convert set to slice and add default
		var candidatesArgs []string
		for s := range candidateSet {
			candidatesArgs = append(candidatesArgs, s)
		}
		candidatesArgs = append(candidatesArgs, "default")

		candidates := resources.GetPlatformIconCandidates(candidatesArgs...)

		// Helper to load and scale a pixbuf candidate
		tryLoadIcon := func(c resources.IconCandidate) bool {
			loader, err := gdkpixbuf.NewPixbufLoaderWithMIMEType(c.Mime)
			if err != nil {
				loader = gdkpixbuf.NewPixbufLoader()
			}

			if err := loader.Write(c.Data); err != nil {
				loader.Close()
				return false
			}

			if err := loader.Close(); err != nil {
				return false
			}

			pixbuf := loader.Pixbuf()
			if pixbuf == nil {
				return false
			}

			scaled := pixbuf.ScaleSimple(32, 32, gdkpixbuf.InterpBilinear)
			if scaled != nil {
				iconImg.SetFromPixbuf(scaled)
				return true
			}
			return false
		}

		iconLoaded := false
		for _, c := range candidates {
			if tryLoadIcon(c) {
				iconLoaded = true
				break
			}
		}

		// Network fallback: try to load from RomM server if embedded fails
		if !iconLoaded {
			logoPath := p.LogoPath
			if logoPath == "" && p.Slug != "" {
				// Guess lowercase slugified path
				logoPath = "/assets/platforms/" + strings.ToLower(p.Slug) + ".png"
			}

			if logoPath != "" {
				host := s.router.State().GetHost()
				if host != nil {
					go func(p romm.Platform, img *gtk.Image, path string) {
						client := romm.NewClientFromHost(*host)
						data, err := client.GetAsset(path)
						if err == nil {
							glib.IdleAdd(func() {
								loader := gdkpixbuf.NewPixbufLoader()
								if err := loader.Write(data); err == nil {
									loader.Close()
									if pixbuf := loader.Pixbuf(); pixbuf != nil {
										scaled := pixbuf.ScaleSimple(32, 32, gdkpixbuf.InterpBilinear)
										if scaled != nil {
											img.SetFromPixbuf(scaled)
										}
									}
								}
							})
						} else {
							// Try one more guess if the first one failed (.svg instead of .png)
							if strings.HasSuffix(path, ".png") {
								svgPath := strings.TrimSuffix(path, ".png") + ".svg"
								data, err := client.GetAsset(svgPath)
								if err == nil {
									glib.IdleAdd(func() {
										loader := gdkpixbuf.NewPixbufLoader()
										if err := loader.Write(data); err == nil {
											loader.Close()
											if pixbuf := loader.Pixbuf(); pixbuf != nil {
												scaled := pixbuf.ScaleSimple(32, 32, gdkpixbuf.InterpBilinear)
												if scaled != nil {
													img.SetFromPixbuf(scaled)
												}
											}
										}
									})
								}
							}
						}
					}(p, iconImg, logoPath)
				}
			}
		}

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
