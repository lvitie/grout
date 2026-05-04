package screens

import (
	"grout/cache"
	"grout/desktop"
	"grout/romm"
	"strings"
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
		row := adw.NewActionRow()
		row.SetTitle(g.Name)
		row.SetSubtitle(g.FsName)
		row.SetSelectable(true)
		listBox.Append(row)
	}

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
		text := strings.ToLower(searchBar.Text())
		row := listBox.FirstChild()
		i := 0
		for row != nil {
			if lbRow, ok := row.(*gtk.ListBoxRow); ok {
				game := s.games[i]
				visible := text == "" || 
					strings.Contains(strings.ToLower(game.Name), text) ||
					strings.Contains(strings.ToLower(game.FsName), text)
				
				lbRow.SetVisible(visible)
				i++
			}
			row = row.NextSibling()
		}
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
