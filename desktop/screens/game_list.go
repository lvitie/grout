package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/internal"
	"grout/romm"
	"strings"
	"grout/desktop/widgets"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type GameListScreen struct {
	router   *desktop.Router
	platform romm.Platform
	games    []romm.Rom
}

func NewGameListScreen(router *desktop.Router, platform romm.Platform) *GameListScreen {
	cm := cache.GetCacheManager()
	games, _ := cm.GetPlatformGames(platform.ID)
	return &GameListScreen{
		router:   router,
		platform: platform,
		games:    games,
	}
}

func (s *GameListScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	// Filter and add rows
	host := router.State().GetHost()
	allGames := s.games
	s.games = make([]romm.Rom, 0)
	for _, g := range allGames {
		s.games = append(s.games, g)
		row := widgets.NewGameRowWithArt(g, host)
		row.SetActivatable(true)
		listBox.Append(row)
	}

	searchBar := gtk.NewSearchEntry()
	searchBar.SetPlaceholderText("Search games...")

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

	searchBar.ConnectChanged(func() {
		listBox.InvalidateFilter()
	})

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.games) {
			game := s.games[idx]
			internal.GetLogger().Info("Opening game details", "id", game.ID, "name", game.Name)
			router.Navigate(NewGameDetailsScreen(router, game))
		}
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)
	scrolled.SetVExpand(true)
	scrolled.SetHExpand(true)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle(s.platform.Name, ""))
	header.PackStart(searchBar)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(scrolled)

	page := adw.NewNavigationPage(box, s.platform.Name)
	return page
}

func boxWithSearch(search *gtk.SearchEntry, child gtk.Widgetter) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(search)
	box.Append(child)
	return box
}
