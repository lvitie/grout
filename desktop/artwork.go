package desktop

import (
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// LoadImageAsync fetches an image from a URL in a background goroutine
// and sets it on the given gtk.Image widget when ready.
func LoadImageAsync(img *gtk.Image, url string, size int) {
	if url == "" {
		return
	}

	go func() {
		data, err := fetchURL(url)
		if err != nil {
			slog.Debug("Failed to fetch artwork", "url", url, "error", err)
			return
		}

		glib.IdleAdd(func() {
			loader := gdkpixbuf.NewPixbufLoader()
			if err := loader.Write(data); err != nil {
				slog.Debug("Failed to write pixbuf data", "error", err)
				return
			}
			if err := loader.Close(); err != nil {
				slog.Debug("Failed to close pixbuf loader", "error", err)
				return
			}

			pixbuf := loader.Pixbuf()
			if pixbuf == nil {
				return
			}

			// Scale to requested size while keeping aspect ratio
			w := pixbuf.Width()
			h := pixbuf.Height()
			if w > 0 && h > 0 {
				var newW, newH int
				if w > h {
					newW = size
					newH = size * h / w
				} else {
					newH = size
					newW = size * w / h
				}
				if newW > 0 && newH > 0 {
					pixbuf = pixbuf.ScaleSimple(newW, newH, gdkpixbuf.InterpBilinear)
				}
			}

			if pixbuf != nil {
				img.SetFromPixbuf(pixbuf)
			}
		})
	}()
}

var artHTTPClient *http.Client
var artHTTPOnce sync.Once

func fetchURL(url string) ([]byte, error) {
	artHTTPOnce.Do(func() {
		artHTTPClient = &http.Client{}
	})

	resp, err := artHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
