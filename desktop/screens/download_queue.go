package screens

import (
	"fmt"
	"grout/desktop"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DownloadQueueScreen struct {
	router *desktop.Router
}

func NewDownloadQueueScreen(router *desktop.Router) *DownloadQueueScreen {
	return &DownloadQueueScreen{router: router}
}

func (s *DownloadQueueScreen) Build(router *desktop.Router) gtk.Widgetter {
	listBox := gtk.NewListBox()
	listBox.SetSelectionMode(gtk.SelectionNone)
	listBox.AddCSSClass("boxed-list")
	listBox.SetMarginStart(12)
	listBox.SetMarginEnd(12)
	listBox.SetMarginTop(12)
	listBox.SetMarginBottom(12)

	emptyStatus := adw.NewStatusPage()
	emptyStatus.SetTitle("Download Queue")
	emptyStatus.SetDescription("No downloads queued.")
	emptyStatus.SetIconName("folder-download-symbolic")

	stack := gtk.NewStack()
	stack.AddNamed(emptyStatus, "empty")

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listBox)
	scrolled.SetVExpand(true)
	stack.AddNamed(scrolled, "list")

	rebuildList := func() {
		queue := desktop.GetDownloadQueue()
		items := queue.Items()

		for child := listBox.FirstChild(); child != nil; child = listBox.FirstChild() {
			listBox.Remove(child)
		}

		if len(items) == 0 {
			stack.SetVisibleChildName("empty")
			return
		}
		stack.SetVisibleChildName("list")

		for _, item := range items {
			row := adw.NewActionRow()
			row.SetTitle(desktop.EscapeMarkup(item.Game.Name))

			switch item.Status {
			case desktop.DownloadQueued:
				row.SetSubtitle("Queued")
				icon := gtk.NewImageFromIconName("content-loading-symbolic")
				icon.SetPixelSize(16)
				row.AddSuffix(icon)
			case desktop.DownloadInProgress:
				progress := gtk.NewProgressBar()
				progress.SetFraction(item.Progress)
				progress.SetShowText(true)
				progress.SetVAlign(gtk.AlignCenter)
				progress.SetSizeRequest(150, -1)
				row.AddSuffix(progress)
				row.SetSubtitle(fmt.Sprintf("Downloading... %d%%", int(item.Progress*100)))
			case desktop.DownloadComplete:
				row.SetSubtitle("Complete")
				icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
				icon.SetPixelSize(16)
				row.AddSuffix(icon)
			case desktop.DownloadFailed:
				errMsg := "Unknown error"
				if item.Error != nil {
					errMsg = item.Error.Error()
				}
				row.SetSubtitle(fmt.Sprintf("Failed: %s", desktop.EscapeMarkup(errMsg)))
				icon := gtk.NewImageFromIconName("dialog-error-symbolic")
				icon.SetPixelSize(16)
				row.AddSuffix(icon)
			}

			listBox.Append(row)
		}
	}

	rebuildList()

	// Poll for updates while this screen is visible
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(250 * time.Millisecond):
				glib.IdleAdd(rebuildList)
			}
		}
	}()

	header := adw.NewHeaderBar()
	header.SetTitleWidget(adw.NewWindowTitle("Downloads", ""))

	clearBtn := gtk.NewButtonFromIconName("edit-clear-all-symbolic")
	clearBtn.SetTooltipText("Clear completed")
	clearBtn.ConnectClicked(func() {
		desktop.GetDownloadQueue().ClearCompleted()
		rebuildList()
	})
	header.PackEnd(clearBtn)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.Append(header)
	box.Append(stack)

	page := adw.NewNavigationPage(box, "Downloads")

	// Stop polling when the page is hidden
	page.ConnectHidden(func() {
		close(stop)
	})

	return page
}
