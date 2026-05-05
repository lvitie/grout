package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/cache"
	"grout/desktop"
	"grout/desktop/widgets"
	"grout/internal"
	"grout/romm"
	"strings"
)

type GameListScreen struct {
	router *desktop.Router
	title  string
	games  []romm.Rom
}

func NewGameListScreen(router *desktop.Router, platform romm.Platform) *GameListScreen {
	cm := cache.GetCacheManager()
	games, _ := cm.GetPlatformGames(platform.ID)
	return &GameListScreen{
		router: router,
		title:  platform.Name,
		games:  games,
	}
}

func NewCollectionGameListScreen(router *desktop.Router, collection romm.Collection) *GameListScreen {
	cm := cache.GetCacheManager()
	games, _ := cm.GetCollectionGames(collection)
	return &GameListScreen{
		router: router,
		title:  collection.Name,
		games:  games,
	}
}

func (s *GameListScreen) Build(router *desktop.Router) gtk.Widgetter {
	host := router.State().GetHost()
	allGames := s.games
	s.games = make([]romm.Rom, 0)
	for _, g := range allGames {
		s.games = append(s.games, g)
	}

	searchBar := gtk.NewSearchEntry()
	searchBar.SetPlaceholderText("Search games...")

	// Build list view
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	for _, g := range s.games {
		row := widgets.NewGameRowWithArt(g, host)
		row.SetActivatable(true)
		listBox.Append(row)
	}

	listBox.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
		text := strings.ToLower(searchBar.Text())
		if text == "" {
			return true
		}

		if actionRow, ok := row.Child().(*adw.ActionRow); ok {
			title := strings.ToLower(actionRow.Title())
			subtitle := strings.ToLower(actionRow.Subtitle())
			return strings.Contains(title, text) || strings.Contains(subtitle, text)
		}
		return true
	})

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.games) {
			game := s.games[idx]
			internal.GetLogger().Info("Opening game details", "id", game.ID, "name", game.Name)
			router.Navigate(NewGameDetailsScreen(router, game))
		}
	})

	listScrolled := gtk.NewScrolledWindow()
	listScrolled.SetChild(listBox)
	listScrolled.SetVExpand(true)
	listScrolled.SetHExpand(true)

	// Build grid view
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

	for _, g := range s.games {
		cell := widgets.NewGameGridCellWithArt(g, host)
		flowBox.Append(cell)
	}

	flowBox.SetFilterFunc(func(child *gtk.FlowBoxChild) bool {
		text := strings.ToLower(searchBar.Text())
		if text == "" {
			return true
		}

		if cell, ok := child.Child().(*widgets.GameGridCell); ok {
			gameName := strings.ToLower(cell.GetGame().Name)
			return strings.Contains(gameName, text)
		}
		return false
	})

	flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		if cell, ok := child.Child().(*widgets.GameGridCell); ok {
			game := cell.GetGame()
			internal.GetLogger().Info("Opening game details", "id", game.ID, "name", game.Name)
			router.Navigate(NewGameDetailsScreen(router, game))
		}
	})

	gridScrolled := gtk.NewScrolledWindow()
	gridScrolled.SetChild(flowBox)
	gridScrolled.SetVExpand(true)
	gridScrolled.SetHExpand(true)

	// Create stack to hold both views
	stack := gtk.NewStack()
	stack.AddNamed(listScrolled, "list")
	stack.AddNamed(gridScrolled, "grid")

	// Set initial view from state
	currentMode := router.State().GetViewMode()
	if currentMode == desktop.ViewModeGrid {
		stack.SetVisibleChildName("grid")
	} else {
		stack.SetVisibleChildName("list")
	}

	// Wire search to both views
	searchBar.ConnectChanged(func() {
		listBox.InvalidateFilter()
		flowBox.InvalidateFilter()
	})

	// Create toggle button
	isGridView := currentMode == desktop.ViewModeGrid
	iconName := "view-grid-symbolic"
	if isGridView {
		iconName = "view-list-symbolic"
	}

	toggleBtn := gtk.NewButtonFromIconName(iconName)
	toggleBtn.SetTooltipText("Toggle grid/list view")

	toggleBtn.ConnectClicked(func() {
		isGridView = !isGridView
		if isGridView {
			stack.SetVisibleChildName("grid")
			router.State().SetViewMode(desktop.ViewModeGrid)
			toggleBtn.SetIconName("view-list-symbolic")
		} else {
			stack.SetVisibleChildName("list")
			router.State().SetViewMode(desktop.ViewModeList)
			toggleBtn.SetIconName("view-grid-symbolic")
		}
	})

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle(desktop.EscapeMarkup(s.title), ""))
	header.PackStart(searchBar)
	header.PackEnd(toggleBtn)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(stack)

	page := adw.NewNavigationPage(box, s.title)
	return page
}

func boxWithSearch(search *gtk.SearchEntry, child gtk.Widgetter) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(search)
	box.Append(child)
	return box
}
