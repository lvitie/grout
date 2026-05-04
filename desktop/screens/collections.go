package screens

import (
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
		row.SetTitle(c.Name)
		row.SetSubtitle(c.Description)
		row.SetSelectable(true)
		listBox.Append(row)
	}

	listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		_ = row.Index() // TODO: navigate to collection game list
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle("Collections", ""))

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(scrolled)

	page := adw.NewNavigationPage(box, "Collections")
	return page
}
