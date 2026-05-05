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
	listBox     *gtk.ListBox
	stack       *gtk.Stack
	flowBox     *gtk.FlowBox
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
	s.listBox = gtk.NewListBox()
	s.listBox.SetSelectionMode(gtk.SelectionSingle)
	s.listBox.AddCSSClass("navigation-sidebar")

	s.listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(s.collections) {
			router.Navigate(NewCollectionGameListScreen(router, s.collections[idx]))
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

	s.flowBox.ConnectChildActivated(func(child *gtk.FlowBoxChild) {
		for idx, c := range s.collections {
			if s.flowBox.ChildAtIndex(idx) == child {
				router.Navigate(NewCollectionGameListScreen(router, c))
				break
			}
		}
	})

	gridScrolled := gtk.NewScrolledWindow()
	gridScrolled.SetChild(s.flowBox)
	gridScrolled.SetVExpand(true)
	s.stack.AddNamed(gridScrolled, "grid")

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
		s.listBox.Append(row)
	}

	// Add cells to grid
	if s.flowBox != nil {
		for _, c := range s.collections {
			cell := gtk.NewBox(gtk.OrientationVertical, 6)
			cell.SetSizeRequest(150, -1)

			label := gtk.NewLabel(desktop.EscapeMarkup(c.Name))
			label.SetWrap(true)
			label.SetMaxWidthChars(15)
			label.SetJustify(gtk.JustifyCenter)
			cell.Append(label)

			s.flowBox.Append(cell)
		}
	}

	s.stack.SetVisibleChildName("list")
}

func (s *CollectionsScreen) ShowGridView() {
	s.stack.SetVisibleChildName("grid")
}

func (s *CollectionsScreen) ShowListView() {
	s.stack.SetVisibleChildName("list")
}
