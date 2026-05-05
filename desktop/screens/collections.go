package screens

import (
	"fmt"
	"grout/cache"
	"grout/desktop"
	"grout/internal"
	"grout/romm"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CollectionsScreen struct {
	router      *desktop.Router
	collections []romm.Collection
	ListBox     *gtk.ListBox
	stack       *gtk.Stack
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
	s.stack = gtk.NewStack()
	s.stack.SetTransitionType(gtk.StackTransitionTypeSlideLeftRight)

	// Empty State
	statusPage := adw.NewStatusPage()
	statusPage.SetTitle("No Collections Found")
	statusPage.SetDescription("Make sure collections are enabled in Settings and perform a full library sync.")
	statusPage.SetIconName("folder-saved-search-symbolic")
	s.stack.AddNamed(statusPage, "empty")

	// List View
	s.ListBox = gtk.NewListBox()
	s.ListBox.SetSelectionMode(gtk.SelectionSingle)
	s.ListBox.AddCSSClass("navigation-sidebar")

	s.ListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.collections) {
			router.Navigate(NewCollectionGameListScreen(router, s.collections[idx]))
		}
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(s.ListBox)
	scrolled.SetVExpand(true)
	s.stack.AddNamed(scrolled, "list")

	s.Refresh()

	return s.stack
}

func (s *CollectionsScreen) Refresh() {
	cm := cache.GetCacheManager()
	s.collections, _ = cm.GetCollections()

	if len(s.collections) == 0 {
		config, _ := internal.LoadConfig()
		msg := "Make sure collections are enabled in Settings and perform a full library sync."
		if config != nil && !config.ShowRegularCollections && !config.ShowSmartCollections && !config.ShowVirtualCollections {
			msg = "Collections are currently disabled in Settings."
		}

		if child := s.stack.ChildByName("empty"); child != nil {
			if statusPage, ok := child.(*adw.StatusPage); ok {
				statusPage.SetDescription(msg)
			}
		}
		s.stack.SetVisibleChildName("empty")
		return
	}

	s.stack.SetVisibleChildName("list")

	// Clear list
	for {
		child := s.ListBox.FirstChild()
		if child == nil {
			break
		}
		s.ListBox.Remove(child)
	}

	// Add rows
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
		s.ListBox.Append(row)
	}
}
