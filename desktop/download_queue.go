package desktop

import (
	"fmt"
	"grout/internal"
	"grout/romm"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type DownloadStatus int

const (
	DownloadQueued DownloadStatus = iota
	DownloadInProgress
	DownloadComplete
	DownloadFailed
)

type DownloadItem struct {
	Game     romm.Rom
	Status   DownloadStatus
	Progress float64
	Error    error
}

type DownloadQueue struct {
	mu         sync.Mutex
	items      []*DownloadItem
	listeners  []func()
	processing bool
	host       *romm.Host
	config     *internal.Config
}

var (
	globalQueue     *DownloadQueue
	globalQueueOnce sync.Once
)

func GetDownloadQueue() *DownloadQueue {
	globalQueueOnce.Do(func() {
		globalQueue = &DownloadQueue{}
	})
	return globalQueue
}

func (q *DownloadQueue) SetHostAndConfig(host *romm.Host, config *internal.Config) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.host = host
	q.config = config
}

func (q *DownloadQueue) AddListener(l func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners = append(q.listeners, l)
}

func (q *DownloadQueue) notify() {
	q.mu.Lock()
	listeners := make([]func(), len(q.listeners))
	copy(listeners, q.listeners)
	q.mu.Unlock()
	for _, l := range listeners {
		l()
	}
}

func (q *DownloadQueue) Enqueue(games ...romm.Rom) {
	q.mu.Lock()
	for _, game := range games {
		already := false
		for _, item := range q.items {
			if item.Game.ID == game.ID && item.Status != DownloadFailed {
				already = true
				break
			}
		}
		if !already {
			q.items = append(q.items, &DownloadItem{
				Game:   game,
				Status: DownloadQueued,
			})
		}
	}
	shouldStart := !q.processing
	q.mu.Unlock()

	q.notify()

	if shouldStart {
		go q.processQueue()
	}
}

func (q *DownloadQueue) Items() []*DownloadItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*DownloadItem, len(q.items))
	copy(out, q.items)
	return out
}

func (q *DownloadQueue) ActiveItem() *DownloadItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, item := range q.items {
		if item.Status == DownloadInProgress {
			return item
		}
	}
	return nil
}

func (q *DownloadQueue) QueuedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	count := 0
	for _, item := range q.items {
		if item.Status == DownloadQueued || item.Status == DownloadInProgress {
			count++
		}
	}
	return count
}

func (q *DownloadQueue) ClearCompleted() {
	q.mu.Lock()
	filtered := make([]*DownloadItem, 0, len(q.items))
	for _, item := range q.items {
		if item.Status != DownloadComplete && item.Status != DownloadFailed {
			filtered = append(filtered, item)
		}
	}
	q.items = filtered
	q.mu.Unlock()
	q.notify()
}

func (q *DownloadQueue) processQueue() {
	q.mu.Lock()
	q.processing = true
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.processing = false
		q.mu.Unlock()
	}()

	for {
		q.mu.Lock()
		var next *DownloadItem
		for _, item := range q.items {
			if item.Status == DownloadQueued {
				next = item
				break
			}
		}
		host := q.host
		config := q.config
		q.mu.Unlock()

		if next == nil {
			return
		}

		if host == nil || config == nil {
			q.mu.Lock()
			next.Status = DownloadFailed
			next.Error = fmt.Errorf("no host or config configured")
			q.mu.Unlock()
			q.notify()
			continue
		}

		q.downloadItem(next, *host, config)
	}
}

func (q *DownloadQueue) downloadItem(item *DownloadItem, host romm.Host, config *internal.Config) {
	q.mu.Lock()
	item.Status = DownloadInProgress
	item.Progress = 0
	q.mu.Unlock()
	q.notify()

	slog.Info("Download queue: starting", "gameID", item.Game.ID, "gameName", item.Game.Name)

	client := romm.NewClientFromHost(host)
	resp, err := client.DownloadRomStream(item.Game)
	if err != nil {
		q.mu.Lock()
		item.Status = DownloadFailed
		item.Error = err
		q.mu.Unlock()
		q.notify()
		return
	}
	defer resp.Body.Close()

	totalSize := resp.ContentLength

	romDir := config.GetPlatformRomDirectory(romm.Platform{
		ID:     item.Game.PlatformID,
		FSSlug: item.Game.PlatformFSSlug,
	})

	targetPath := filepath.Join(romDir, item.Game.FsName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		q.mu.Lock()
		item.Status = DownloadFailed
		item.Error = fmt.Errorf("create directory: %w", err)
		q.mu.Unlock()
		q.notify()
		return
	}

	out, err := os.Create(targetPath)
	if err != nil {
		q.mu.Lock()
		item.Status = DownloadFailed
		item.Error = fmt.Errorf("create file: %w", err)
		q.mu.Unlock()
		q.notify()
		return
	}
	defer out.Close()

	var downloaded int64
	buffer := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, werr := out.Write(buffer[:n]); werr != nil {
				q.mu.Lock()
				item.Status = DownloadFailed
				item.Error = fmt.Errorf("write file: %w", werr)
				q.mu.Unlock()
				q.notify()
				return
			}
			downloaded += int64(n)
			if totalSize > 0 {
				q.mu.Lock()
				item.Progress = float64(downloaded) / float64(totalSize)
				q.mu.Unlock()
				q.notify()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			q.mu.Lock()
			item.Status = DownloadFailed
			item.Error = readErr
			q.mu.Unlock()
			q.notify()
			return
		}
	}

	q.mu.Lock()
	item.Status = DownloadComplete
	item.Progress = 1.0
	q.mu.Unlock()
	q.notify()
	slog.Info("Download queue: complete", "gameID", item.Game.ID, "gameName", item.Game.Name)
}
