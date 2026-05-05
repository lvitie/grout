package screens

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"grout/desktop"
	"grout/romm"
)

type CollectionPlatformScreen struct {
	router     *desktop.Router
	collection romm.Collection
}

func NewCollectionPlatformScreen(router *desktop.Router, col romm.Collection) *CollectionPlatformScreen {
	return &CollectionPlatformScreen{
		router:     router,
		collection: col,
	}
}

func (s *CollectionPlatformScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()

	// Implementation would filter platforms that have games in this collection

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)

	page := adw.NewNavigationPage(scrolled, s.collection.Name)
	return page
}
