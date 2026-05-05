package widgets

import (
	"grout/desktop"
	"grout/romm"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CollectionGridCell struct {
	*gtk.Box
	collection romm.Collection
	img        *gtk.Image
}

func NewCollectionGridCell(c romm.Collection) *CollectionGridCell {
	img := gtk.NewImage()
	img.SetPixelSize(128)
	img.SetFromIconName("image-missing-symbolic")
	img.SetHExpand(false)
	img.SetVExpand(false)

	label := gtk.NewLabel(desktop.EscapeMarkup(c.Name))
	label.SetWrap(true)
	label.SetMaxWidthChars(15)
	label.SetJustify(gtk.JustifyCenter)
	label.SetHExpand(false)
	label.SetVExpand(false)

	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.SetSizeRequest(150, 200)
	box.SetHomogeneous(false)
	box.SetHAlign(gtk.AlignCenter)
	box.SetVAlign(gtk.AlignStart)
	box.SetHExpand(false)
	box.SetVExpand(false)

	img.SetVAlign(gtk.AlignStart)
	img.SetHAlign(gtk.AlignCenter)

	// Determine icon based on collection type
	if c.IsVirtual {
		img.SetFromIconName("folder-open-symbolic")
	} else if c.IsSmart {
		img.SetFromIconName("system-search-symbolic")
	} else {
		img.SetFromIconName("folder-documents-symbolic")
	}

	box.Append(img)
	box.Append(label)

	return &CollectionGridCell{
		Box:        box,
		collection: c,
		img:        img,
	}
}

func NewCollectionGridCellWithCover(c romm.Collection, host *romm.Host) *CollectionGridCell {
	cell := NewCollectionGridCell(c)

	if host != nil {
		if coverURL := c.GetCoverLargeURL(*host); coverURL != "" {
			desktop.LoadImageAsync(cell.img, coverURL, 128)
		}
	}

	return cell
}

func (c *CollectionGridCell) GetCollection() romm.Collection {
	return c.collection
}
