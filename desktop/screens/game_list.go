package screens

import (
	"grout/cache"
	"grout/desktop"
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

	for _, g := range s.games {
		row := widgets.NewGameRow(g)
		listBox.Append(row)
	}

	listBox.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
		text := strings.ToLower(searchBar.Text())
		if text == "" {
			return true
		}
		
		// We need to retrieve the game from the row.
		// Since we used NewGameRow, we can cast it if we exported the game field,
		// or just use the title/subtitle of the ActionRow.
		if actionRow, ok := row.Child().(*adw.ActionRow); ok {
			title := strings.ToLower(actionRow.Title())
			subtitle := strings.ToLower(actionRow.Subtitle())
			return strings.Contains(title, text) || strings.Contains(subtitle, text)
		}
		return true
	})

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		// Search might change the index if we use a filtered list model,
		// but for a simple ListBox we can just look up the row data.
		// For now, keep it simple.
		router.Navigate(NewGameDetailsScreen(router, s.games[idx]))
	})

	searchBar := gtk.NewSearchEntry()
	searchBar.SetPlaceholderText("Search games...")
	searchBar.ConnectChanged(func() {
		listBox.InvalidateFilter()
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle(s.platform.Name, ""))
	header.PackStart(searchBar)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(boxWithSearch(searchBar, scrolled))

	page := adw.NewNavigationPage(box, s.platform.Name)
	return page
}

func boxWithSearch(search *gtk.SearchEntry, child gtk.Widgetter) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(search)
	box.Append(child)
	return box
}
