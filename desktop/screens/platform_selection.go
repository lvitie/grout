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
	stack             *adw.ViewStack
	outerStack        *gtk.Stack
	progressBar       *gtk.ProgressBar
	listBox           *gtk.ListBox
	flowBox           *gtk.FlowBox
	platformListStack *gtk.Stack
	collectionsScreen *CollectionsScreen
	installedScreen   *InstalledScreen
	collectionsBox    *gtk.ListBox
	gamesBox          *gtk.ListBox
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
	platformsPage := s.stack.AddTitled(listView, "platforms", "Platforms")
	platformsPage.SetIconName("application-x-executable-symbolic")

	// Collections View
	s.collectionsScreen = NewCollectionsScreen(router)
	collectionsView := s.collectionsScreen.Build(router)
	collectionsPage := s.stack.AddTitled(collectionsView, "collections", "Collections")
	collectionsPage.SetIconName("folder-documents-symbolic")

	// Installed View
	s.installedScreen = NewInstalledScreen(router)
	installedView := s.installedScreen.Build(router)
	installedPage := s.stack.AddTitled(installedView, "installed", "Installed")
	installedPage.SetIconName("drive-harddisk-symbolic")

	// We'll wire the search bar to collections and installed after creating it

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

	// Setup outer stack for main/search views
	s.outerStack = gtk.NewStack()
	s.outerStack.SetVExpand(true)
	s.outerStack.AddNamed(s.stack, "main")

	// Create search panel (collections + games results)
	s.collectionsBox = gtk.NewListBox()
	s.collectionsBox.SetSelectionMode(gtk.SelectionSingle)
	s.collectionsBox.AddCSSClass("navigation-sidebar")

	s.gamesBox = gtk.NewListBox()
	s.gamesBox.SetSelectionMode(gtk.SelectionSingle)
	s.gamesBox.AddCSSClass("navigation-sidebar")

	searchPanel := s.buildSearchPanel(router)
	s.outerStack.AddNamed(searchPanel, "search")

	// Search bar
	searchBar := gtk.NewSearchEntry()
	searchBar.SetPlaceholderText("Search...")

	// Wire search to collections and installed screens
	s.collectionsScreen.SetSearchQuery(searchBar)
	s.installedScreen.SetSearchQuery(searchBar)

	searchBar.ConnectChanged(func() {
		query := strings.TrimSpace(searchBar.Text())
		activeTab := s.stack.VisibleChildName()

		// Always invalidate collections and installed filters
		s.collectionsScreen.InvalidateCollectionFilters()
		s.installedScreen.InvalidateFilters()

		if query == "" {
			s.outerStack.SetVisibleChildName("main")
			return
		}

		switch activeTab {
		case "platforms":
			s.outerStack.SetVisibleChildName("search")
			go s.runGlobalSearch(query, router)
		case "collections":
			s.outerStack.SetVisibleChildName("main")
		case "installed":
			s.outerStack.SetVisibleChildName("main")
		}
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
			if s.installedScreen != nil {
				s.installedScreen.ShowGridView()
			}
			router.State().SetViewMode(desktop.ViewModeGrid)
			toggleBtn.SetIconName("view-list-symbolic")
		} else {
			s.platformListStack.SetVisibleChildName("list")
			if s.collectionsScreen != nil {
				s.collectionsScreen.ShowListView()
			}
			if s.installedScreen != nil {
				s.installedScreen.ShowListView()
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

	queueBtn := gtk.NewButtonFromIconName("folder-download-symbolic")
	queueBtn.SetTooltipText("Download queue")
	queueBtn.ConnectClicked(func() {
		router.Navigate(NewDownloadQueueScreen(router))
	})
	header.PackEnd(queueBtn)

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

	// Controller hints footer
	footer := gtk.NewBox(gtk.OrientationHorizontal, 24)
	footer.SetHAlign(gtk.AlignCenter)
	footer.SetMarginTop(4)
	footer.SetMarginBottom(4)
	footer.AddCSSClass("dim-label")
	for _, hint := range []string{"LB/RB  Tabs", "Select  View", "Start  Menu", "A  Select", "B  Back"} {
		lbl := gtk.NewLabel(hint)
		lbl.AddCSSClass("caption")
		footer.Append(lbl)
	}

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(s.outerStack)
	box.Append(footer)

	// Wire controller callbacks to the router
	tabNames := []string{"platforms", "collections", "installed"}
	router.TabLeftFn = func() {
		current := s.stack.VisibleChildName()
		for i, name := range tabNames {
			if name == current && i > 0 {
				s.stack.SetVisibleChildName(tabNames[i-1])
				return
			}
		}
	}
	router.TabRightFn = func() {
		current := s.stack.VisibleChildName()
		for i, name := range tabNames {
			if name == current && i < len(tabNames)-1 {
				s.stack.SetVisibleChildName(tabNames[i+1])
				return
			}
		}
	}
	router.ToggleViewFn = func() {
		toggleBtn.Activate()
	}
	router.SearchFn = func() {
		searchBar.GrabFocus()
	}
	router.IsSearchActive = func() bool {
		return searchBar.Text() != ""
	}
	router.ClearSearch = func() {
		searchBar.SetText("")
	}
	router.FocusContent = func() {
		// If search results are showing, focus them
		if s.outerStack.VisibleChildName() == "search" {
			if row := s.collectionsBox.RowAtIndex(0); row != nil {
				row.GrabFocus()
				return
			}
			if row := s.gamesBox.RowAtIndex(0); row != nil {
				row.GrabFocus()
				return
			}
			return
		}

		tab := s.stack.VisibleChildName()
		switch tab {
		case "collections":
			s.collectionsScreen.FocusContent()
		case "installed":
			s.installedScreen.FocusContent()
		default:
			if s.platformListStack.VisibleChildName() == "grid" {
				child := s.flowBox.ChildAtIndex(0)
				if child != nil {
					child.GrabFocus()
				}
			} else {
				s.listBox.GrabFocus()
			}
		}
	}
	router.QuickMenuFn = func() {
		s.showQuickMenu(router, searchBar)
	}

	page := adw.NewNavigationPage(box, "Grout")
	glib.IdleAdd(func() {
		s.listBox.GrabFocus()
	})
	return page
}

func (s *PlatformSelectionScreen) showQuickMenu(router *desktop.Router, searchBar *gtk.SearchEntry) {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("navigation-sidebar")

	type menuItem struct {
		label  string
		action func()
	}

	items := []menuItem{
		{"Search", func() { searchBar.GrabFocus() }},
		{"Platforms", func() { s.stack.SetVisibleChildName("platforms") }},
		{"Collections", func() { s.stack.SetVisibleChildName("collections") }},
		{"Installed", func() { s.stack.SetVisibleChildName("installed") }},
		{"Downloads", func() { router.Navigate(NewDownloadQueueScreen(router)) }},
		{"Sync Library", func() {
			s.stack.SetVisibleChildName("sync")
			s.startSync(router)
		}},
		{"Settings", func() { router.Navigate(NewSettingsScreen(router)) }},
	}

	dialog := adw.NewDialog()
	dialog.SetTitle("Quick Menu")
	dialog.SetContentWidth(300)
	dialog.SetContentHeight(350)

	for _, item := range items {
		item := item
		row := adw.NewActionRow()
		row.SetTitle(item.label)
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			dialog.Close()
			item.action()
		})
		listBox.Append(row)
	}

	toolbarView := adw.NewToolbarView()
	headerBar := adw.NewHeaderBar()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(listBox)

	dialog.SetChild(toolbarView)
	dialog.Present(router.Window())

	glib.IdleAdd(func() {
		if row := listBox.RowAtIndex(0); row != nil {
			row.GrabFocus()
		}
	})
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
	s.flowBox = gtk.NewFlowBox()
	flowBox := s.flowBox
	flowBox.SetSelectionMode(gtk.SelectionSingle)
	flowBox.SetMinChildrenPerLine(2)
	flowBox.SetMaxChildrenPerLine(10)
	flowBox.SetRowSpacing(12)
	flowBox.SetColumnSpacing(12)
	flowBox.SetMarginStart(12)
	flowBox.SetMarginEnd(12)
	flowBox.SetMarginTop(12)
	flowBox.SetMarginBottom(12)
	flowBox.SetHomogeneous(false)
	flowBox.SetVExpand(false)
	flowBox.SetHExpand(true)

	gridPlatforms := s.platforms // Keep reference for click handling
	for _, p := range gridPlatforms {
		cell := gtk.NewBox(gtk.OrientationVertical, 6)
		cell.SetSizeRequest(150, 200)
		cell.SetHAlign(gtk.AlignCenter)
		cell.SetVAlign(gtk.AlignStart)
		cell.SetHExpand(false)
		cell.SetVExpand(false)

		img := gtk.NewImage()
		img.SetPixelSize(128)
		img.SetFromIconName("application-x-executable-symbolic")
		cell.Append(img)

		label := gtk.NewLabel(desktop.EscapeMarkup(p.Name))
		label.SetWrap(true)
		label.SetMaxWidthChars(15)
		label.SetJustify(gtk.JustifyCenter)
		cell.Append(label)

		flowBox.Append(cell)

		// Load actual platform icon asynchronously
		loadPlatformIconAsync(p, img)
	}

	// Handle selection changes
	flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		idx := child.Index()
		if idx >= 0 && idx < len(gridPlatforms) {
			p := gridPlatforms[idx]
			router.Navigate(NewGameListScreen(router, p))
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

	if s.installedScreen != nil {
		s.installedScreen.Refresh()
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

		loadPlatformListIconAsync(p, iconImg, 32, s.router)

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

func loadPlatformIconAsync(p romm.Platform, img *gtk.Image) {
	go func() {
		candidates := resources.GetPlatformIconCandidates(p.Slug, p.ShortName, p.Name)
		if len(candidates) == 0 {
			return
		}

		for _, c := range candidates {
			loader, err := gdkpixbuf.NewPixbufLoaderWithMIMEType(c.Mime)
			if err != nil {
				loader = gdkpixbuf.NewPixbufLoader()
			}

			if err := loader.Write(c.Data); err != nil {
				loader.Close()
				continue
			}

			if err := loader.Close(); err != nil {
				continue
			}

			pixbuf := loader.Pixbuf()
			if pixbuf == nil {
				continue
			}

			scaled := pixbuf.ScaleSimple(128, 128, gdkpixbuf.InterpBilinear)
			if scaled != nil {
				glib.IdleAdd(func() {
					img.SetFromPixbuf(scaled)
				})
				return
			}
		}
	}()
}

func loadPlatformListIconAsync(p romm.Platform, img *gtk.Image, size int, router *desktop.Router) {
	go func() {
		logoFilename := ""
		if p.LogoPath != "" {
			logoFilename = strings.TrimSuffix(filepath.Base(p.LogoPath), filepath.Ext(p.LogoPath))
		}

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

		slugify := func(s string) string {
			return strings.ToLower(strings.Join(strings.Fields(s), "-"))
		}

		type aliasRule struct {
			contains string
			aliases  []string
		}
		rules := []aliasRule{
			{"playstation 5", []string{"ps5"}},
			{"playstation 4", []string{"ps4"}},
			{"playstation 3", []string{"ps3"}},
			{"playstation 2", []string{"ps2"}},
			{"playstation portable", []string{"psp"}},
			{"playstation vita", []string{"psvita"}},
			{"playstation", []string{"psx", "ps1", "ps"}},
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
			{"dreamcast", []string{"dc", "dreamcast"}},
			{"saturn", []string{"saturn"}},
			{"mega drive", []string{"genesis", "megadrive", "md"}},
			{"genesis", []string{"genesis", "megadrive", "md"}},
			{"master system", []string{"sms", "ms"}},
			{"game gear", []string{"gamegear", "gg"}},
			{"sega cd", []string{"segacd", "scd"}},
			{"sega 32x", []string{"32x"}},
			{"sg-1000", []string{"sg1000"}},
			{"xbox series x", []string{"series-x"}},
			{"xbox series s", []string{"series-s"}},
			{"xbox one", []string{"xboxone"}},
			{"xbox 360", []string{"xbox360"}},
			{"xbox", []string{"xbox"}},
			{"ms-dos", []string{"dos"}},
			{"dos", []string{"dos"}},
			{"atari 2600", []string{"atari2600", "atari-2600"}},
			{"atari 5200", []string{"atari5200", "atari-5200"}},
			{"atari 7800", []string{"atari7800", "atari-7800"}},
			{"atari 800", []string{"atari800", "atari-800"}},
			{"atari lynx", []string{"lynx", "atarilynx"}},
			{"atari jaguar cd", []string{"atari-jaguar-cd"}},
			{"atari jaguar", []string{"jaguar", "atarijaguar"}},
			{"commodore 64", []string{"c64", "commodore-64"}},
			{"amiga cd32", []string{"amiga-cd32"}},
			{"amiga", []string{"amiga"}},
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

		var aliases []string
		lowerName := strings.ToLower(p.Name)
		for _, rule := range rules {
			if strings.Contains(lowerName, rule.contains) {
				aliases = append(aliases, rule.aliases...)
				break
			}
		}

		candidateSet := make(map[string]bool)
		addCandidate := func(s string) {
			if s != "" {
				candidateSet[s] = true
			}
		}

		addCandidate(logoFilename)
		addCandidate(p.ShortName)
		addCandidate(p.Slug)
		addCandidate(p.FSSlug)
		addCandidate(p.Name)
		addCandidate(p.ApiName)
		addCandidate(normalize(p.Name))
		addCandidate(normalize(p.ApiName))
		addCandidate(normalize(p.ShortName))
		addCandidate(normalize(p.Slug))
		addCandidate(normalize(p.FSSlug))
		addCandidate(slugify(p.Name))
		addCandidate(slugify(p.ApiName))
		addCandidate(slugify(p.ShortName))
		addCandidate(slugify(p.Slug))
		addCandidate(slugify(p.FSSlug))
		addCandidate(strings.ReplaceAll(p.Name, " ", "-"))
		addCandidate(strings.ReplaceAll(p.ApiName, " ", "-"))
		addCandidate(strings.ReplaceAll(p.ShortName, " ", "-"))
		addCandidate(strings.ReplaceAll(p.Name, " ", ""))
		addCandidate(strings.ReplaceAll(p.ApiName, " ", ""))
		addCandidate(strings.ReplaceAll(p.ShortName, " ", ""))
		addCandidate(strings.ToLower(p.Name))
		addCandidate(strings.ToLower(p.ApiName))
		addCandidate(strings.ToLower(p.ShortName))
		for _, a := range aliases {
			addCandidate(a)
		}

		var candidatesArgs []string
		for s := range candidateSet {
			candidatesArgs = append(candidatesArgs, s)
		}
		candidatesArgs = append(candidatesArgs, "default")

		candidates := resources.GetPlatformIconCandidates(candidatesArgs...)

		for _, c := range candidates {
			loader, err := gdkpixbuf.NewPixbufLoaderWithMIMEType(c.Mime)
			if err != nil {
				loader = gdkpixbuf.NewPixbufLoader()
			}
			if err := loader.Write(c.Data); err != nil {
				loader.Close()
				continue
			}
			if err := loader.Close(); err != nil {
				continue
			}
			pixbuf := loader.Pixbuf()
			if pixbuf == nil {
				continue
			}
			scaled := pixbuf.ScaleSimple(size, size, gdkpixbuf.InterpBilinear)
			if scaled != nil {
				glib.IdleAdd(func() {
					img.SetFromPixbuf(scaled)
				})
				return
			}
		}

		// Network fallback
		logoPath := p.LogoPath
		if logoPath == "" && p.Slug != "" {
			logoPath = "/assets/platforms/" + strings.ToLower(p.Slug) + ".png"
		}
		if logoPath == "" {
			return
		}
		host := router.State().GetHost()
		if host == nil {
			return
		}
		client := romm.NewClientFromHost(*host)
		data, err := client.GetAsset(logoPath)
		if err != nil && strings.HasSuffix(logoPath, ".png") {
			data, err = client.GetAsset(strings.TrimSuffix(logoPath, ".png") + ".svg")
		}
		if err != nil {
			return
		}
		glib.IdleAdd(func() {
			loader := gdkpixbuf.NewPixbufLoader()
			if err := loader.Write(data); err == nil {
				loader.Close()
				if pixbuf := loader.Pixbuf(); pixbuf != nil {
					if scaled := pixbuf.ScaleSimple(size, size, gdkpixbuf.InterpBilinear); scaled != nil {
						img.SetFromPixbuf(scaled)
					}
				}
			}
		})
	}()
}

func (s *PlatformSelectionScreen) buildSearchPanel(router *desktop.Router) gtk.Widgetter {
	collectionsLabel := gtk.NewLabel("Collections")
	collectionsLabel.AddCSSClass("heading")
	collectionsLabel.SetXAlign(0)
	collectionsLabel.SetMarginStart(12)
	collectionsLabel.SetMarginTop(12)
	collectionsLabel.SetMarginBottom(6)

	gamesLabel := gtk.NewLabel("Games")
	gamesLabel.AddCSSClass("heading")
	gamesLabel.SetXAlign(0)
	gamesLabel.SetMarginStart(12)
	gamesLabel.SetMarginTop(12)
	gamesLabel.SetMarginBottom(6)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(collectionsLabel)
	box.Append(s.collectionsBox)
	box.Append(gamesLabel)
	box.Append(s.gamesBox)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(box)
	scrolled.SetVExpand(true)
	return scrolled
}

func (s *PlatformSelectionScreen) runGlobalSearch(query string, router *desktop.Router) {
	cm := cache.GetCacheManager()
	lq := desktop.NormalizeSearch(query)

	// Collections: filter in Go
	allCollections, _ := cm.GetCollections()
	var matchedCollections []romm.Collection
	for _, c := range allCollections {
		if strings.Contains(desktop.NormalizeSearch(c.Name), lq) {
			matchedCollections = append(matchedCollections, c)
			if len(matchedCollections) >= 20 {
				break
			}
		}
	}

	// Games: fetch all and filter in Go for accent-insensitive matching
	allGames, _ := cm.GetFilteredGames(cache.GameFilter{})
	var games []romm.Rom
	for _, g := range allGames {
		if strings.Contains(desktop.NormalizeSearch(g.Name), lq) {
			games = append(games, g)
			if len(games) >= 50 {
				break
			}
		}
	}

	glib.IdleAdd(func() {
		// Clear collections box
		for child := s.collectionsBox.FirstChild(); child != nil; child = s.collectionsBox.FirstChild() {
			s.collectionsBox.Remove(child)
		}
		for _, c := range matchedCollections {
			c := c
			row := adw.NewActionRow()
			row.SetTitle(desktop.EscapeMarkup(c.Name))
			row.SetSubtitle(fmt.Sprintf("%d games", c.ROMCount))
			row.SetActivatable(true)
			row.ConnectActivated(func() {
				router.Navigate(NewCollectionGameListScreen(router, c))
			})
			s.collectionsBox.Append(row)
		}
		if len(matchedCollections) == 0 {
			row := gtk.NewListBoxRow()
			lbl := gtk.NewLabel("No collections found")
			lbl.SetMarginTop(8)
			lbl.SetMarginBottom(8)
			row.SetChild(lbl)
			s.collectionsBox.Append(row)
		}

		// Clear games box
		for child := s.gamesBox.FirstChild(); child != nil; child = s.gamesBox.FirstChild() {
			s.gamesBox.Remove(child)
		}
		for _, g := range games {
			g := g
			row := adw.NewActionRow()
			row.SetTitle(desktop.EscapeMarkup(g.Name))
			row.SetSubtitle(desktop.EscapeMarkup(g.PlatformDisplayName))
			row.SetActivatable(true)
			row.ConnectActivated(func() {
				router.Navigate(NewGameDetailsScreen(router, g))
			})
			s.gamesBox.Append(row)
		}
		if len(games) == 0 {
			row := gtk.NewListBoxRow()
			lbl := gtk.NewLabel("No games found")
			lbl.SetMarginTop(8)
			lbl.SetMarginBottom(8)
			row.SetChild(lbl)
			s.gamesBox.Append(row)
		}
	})
}
