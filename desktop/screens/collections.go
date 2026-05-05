package screens

import (
	"fmt"
	"grout/cache"
	"grout/desktop"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CollectionsScreen struct {
	router      *desktop.Router
	collections []romm.Collection
}

func NewCollectionsScreen(router *desktop.Router) *CollectionsScreen {
	cm := cache.GetCacheManager()
	collections, _ := cm.GetCollections()
	return &CollectionsScreen{
		router:      router,
		collections: collections,
	}
}

func (s *CollectionsScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.AddCSSClass("navigation-sidebar")

	for _, c := range s.collections {
		row := adw.NewActionRow()
		row.SetTitle(desktop.EscapeMarkup(c.Name))
		
		subtitle := fmt.Sprintf("%d games", c.ROMCount)
		if c.IsSmart {
			subtitle += " (Smart)"
		} else if c.IsVirtual {
			subtitle += " (Virtual)"
		}
		row.SetSubtitle(desktop.EscapeMarkup(subtitle))
		row.SetActivatable(true)
		listBox.Append(row)
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.collections) {
			router.Navigate(NewCollectionGameListScreen(router, s.collections[idx]))
		}
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)
	scrolled.SetVExpand(true)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle("Collections", ""))

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(scrolled)

	page := adw.NewNavigationPage(box, "Collections")
	return page
}
