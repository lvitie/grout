package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/desktop/widgets"
	"grout/romm"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type InstalledScreen struct {
	router  *desktop.Router
	games   []romm.Rom
	listBox *gtk.ListBox
	stack   *gtk.Stack
	flowBox *gtk.FlowBox
}

func NewInstalledScreen(router *desktop.Router) *InstalledScreen {
	return &InstalledScreen{
		router: router,
	}
}

func (s *InstalledScreen) Build(router *desktop.Router) gtk.Widgetter {
	s.stack = gtk.NewStack()
	s.stack.SetTransitionType(gtk.StackTransitionTypeSlideLeftRight)

	// Empty State
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("No Installed Games")
	statusPage.SetDescription("Download games to see them here.")
	statusPage.SetIconName("drive-harddisk-symbolic")
	s.stack.AddNamed(statusPage, "empty")

	// List View
	s.listBox = gtk.NewListBox()
	s.listBox.SetSelectionMode(gtk.SelectionSingle)
	s.listBox.AddCSSClass("navigation-sidebar")

	s.listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.games) {
			router.Navigate(NewGameDetailsScreen(router, s.games[idx]))
		}
	})

	listScrolled := gtk.NewScrolledWindow()
	listScrolled.SetChild(s.listBox)
	listScrolled.SetVExpand(true)
	s.stack.AddNamed(listScrolled, "list")

	// Grid View
	s.flowBox = gtk.NewFlowBox()
	s.flowBox.SetSelectionMode(gtk.SelectionSingle)
	s.flowBox.SetMinChildrenPerLine(2)
	s.flowBox.SetMaxChildrenPerLine(10)
	s.flowBox.SetRowSpacing(12)
	s.flowBox.SetColumnSpacing(12)
	s.flowBox.SetMarginStart(12)
	s.flowBox.SetMarginEnd(12)
	s.flowBox.SetMarginTop(12)
	s.flowBox.SetMarginBottom(12)
	s.flowBox.SetHomogeneous(false)
	s.flowBox.SetVExpand(false)
	s.flowBox.SetHExpand(true)

	// Handle selection changes
	s.flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		idx := child.Index()
		if idx >= 0 && idx < len(s.games) {
			game := s.games[idx]
			router.Navigate(NewGameDetailsScreen(router, game))
		}
	})

	gridScrolled := gtk.NewScrolledWindow()
	gridScrolled.SetChild(s.flowBox)
	gridScrolled.SetVExpand(true)
	s.stack.AddNamed(gridScrolled, "grid")

	// Defer initial load to avoid blocking startup
	go func() {
		s.loadGames()
		glib.IdleAdd(func() {
			s.rebuildUI()
		})
	}()

	router.State().AddInstalledListener(func() {
		go func() {
			s.loadGames()
			glib.IdleAdd(func() {
				s.rebuildUI()
			})
		}()
	})

	return s.stack
}

func (s *InstalledScreen) Refresh() {
	s.loadGames()
	s.rebuildUI()
}

func (s *InstalledScreen) loadGames() {
	cm := cache.GetCacheManager()
	allGames, _ := cm.GetFilteredGames(cache.GameFilter{})
	config := s.router.State().GetConfig()

	installed := make([]romm.Rom, 0)
	for _, g := range allGames {
		if g.IsDownloaded(config) {
			installed = append(installed, g)
		}
	}
	s.games = installed
}

func (s *InstalledScreen) rebuildUI() {
	if len(s.games) == 0 {
		s.stack.SetVisibleChildName("empty")
		return
	}

	// Clear list
	for {
		child := s.listBox.FirstChild()
		if child == nil {
			break
		}
		s.listBox.Remove(child)
	}

	// Clear grid
	if s.flowBox != nil {
		for {
			child := s.flowBox.FirstChild()
			if child == nil {
				break
			}
			s.flowBox.Remove(child)
		}
	}

	// Add rows to list
	host := s.router.State().GetHost()
	for _, g := range s.games {
		row := widgets.NewGameRowWithArt(g, host)
		row.SetActivatable(true)
		s.listBox.Append(row)
	}

	// Add cells to grid
	if s.flowBox != nil {
		for _, g := range s.games {
			cell := widgets.NewGameGridCellWithArt(g, host)
			s.flowBox.Append(cell)
		}
	}

	s.stack.SetVisibleChildName("list")
}

func (s *InstalledScreen) ShowGridView() {
	s.stack.SetVisibleChildName("grid")
}

func (s *InstalledScreen) ShowListView() {
	s.stack.SetVisibleChildName("list")
}

func (s *InstalledScreen) SetSearchQuery(searchBar *gtk.SearchEntry) {
	s.listBox.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
		text := desktop.NormalizeSearch(searchBar.Text())
		if text == "" {
			return true
		}
		idx := row.Index()
		if idx >= 0 && idx < len(s.games) {
			name := desktop.NormalizeSearch(s.games[idx].Name)
			return strings.Contains(name, text)
		}
		return true
	})

	if s.flowBox != nil {
		s.flowBox.SetFilterFunc(func(child *gtk.FlowBoxChild) bool {
			text := desktop.NormalizeSearch(searchBar.Text())
			if text == "" {
				return true
			}
			idx := child.Index()
			if idx >= 0 && idx < len(s.games) {
				name := desktop.NormalizeSearch(s.games[idx].Name)
				return strings.Contains(name, text)
			}
			return false
		})
	}
}

func (s *InstalledScreen) InvalidateFilters() {
	s.listBox.InvalidateFilter()
	if s.flowBox != nil {
		s.flowBox.InvalidateFilter()
	}
}
