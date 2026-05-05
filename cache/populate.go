package cache

import (
	"fmt"
	"grout/internal"
	"grout/romm"
	"sync"
	"time"

	"go.uber.org/atomic"
)

const (
	DefaultRomPageSize           = 500
	MaxConcurrentPlatformFetches = 10
)

type SyncStats struct {
	Platforms         int
	GamesUpdated      int
	CollectionsSynced int
}

func (cm *Manager) populateCache(platforms []romm.Platform, progress *atomic.Float64, force bool) (SyncStats, error) {
	logger := internal.GetLogger()
	stats := SyncStats{Platforms: len(platforms)}

	// Create a single HTTP client for all requests
	client := romm.NewClientFromHost(cm.host, cm.config.GetApiTimeout())

	// Get the last refresh time to use for incremental updates
	var updatedAfter string
	hasCache := cm.HasCache()
	isBulkLoad := !hasCache || force

	logger.Debug("Sync decision inputs", "force", force, "hasCache", hasCache)

	// If we have very few games, we should probably do a full sync anyway
	// to recover from partial or failed previous syncs.
	if !isBulkLoad {
		gameCount := 0
		cm.db.QueryRow("SELECT COUNT(*) FROM games").Scan(&gameCount)
		logger.Debug("Checking game count for recovery", "gameCount", gameCount)
		if gameCount < 500 {
			logger.Info("Cache has very few games, forcing bulk load to ensure complete library", "count", gameCount)
			isBulkLoad = true
		}
	}

	if isBulkLoad {
		updatedAfter = ""
		logger.Info("Sync mode: FULL REFRESH", "force", force, "empty_cache", !hasCache, "is_bulk", isBulkLoad)
	} else {
		if lastRefresh, err := cm.GetLastRefreshTime(MetaKeyGamesRefreshedAt); err == nil {
			updatedAfter = lastRefresh.Format(time.RFC3339)
			logger.Info("Sync mode: INCREMENTAL", "since", updatedAfter)
		} else {
			logger.Info("Sync mode: FULL REFRESH (No timestamp found)")
			isBulkLoad = true
			updatedAfter = ""
		}
	}

	if isBulkLoad {
		// Bulk load optimizations for fresh cache
		cm.enableBulkLoadMode()
		defer cm.disableBulkLoadMode()

		// Save all platforms on first run / empty cache
		logger.Info("Bulk load: fetching all platforms from API")
		allPlatforms, err := client.GetPlatforms()
		if err != nil {
			logger.Error("Failed to fetch all platforms", "error", err)
			if len(platforms) > 0 {
				romm.DisambiguatePlatformNames(platforms)
				if err := cm.SavePlatforms(platforms); err != nil {
					return stats, err
				}
			}
		} else {
			romm.DisambiguatePlatformNames(allPlatforms)
			if err := cm.SavePlatforms(allPlatforms); err != nil {
				return stats, err
			}
			// Use the full list for syncing if our input was empty or we want a full refresh
			if len(platforms) == 0 || isBulkLoad {
				platforms = allPlatforms
			}
		}
		cm.RecordRefreshTime(MetaKeyPlatformsRefreshedAt)
	} else if len(platforms) == 0 {
		// If not bulk load but no platforms provided, fetch them all to be safe
		logger.Info("Incremental sync: fetching platforms to check for updates")
		allPlatforms, err := client.GetPlatforms()
		if err == nil {
			romm.DisambiguatePlatformNames(allPlatforms)

			// Always update platform metadata in cache if we successfully fetched them
			if err := cm.SavePlatforms(allPlatforms); err != nil {
				logger.Warn("Failed to update platform metadata in cache", "error", err)
			}

			platforms = cm.GetPlatformsNeedingSync(allPlatforms)

			// Recovery: if we have platforms but NO games in the DB, sync everything
			var gameCount int
			cm.db.QueryRow("SELECT COUNT(*) FROM games").Scan(&gameCount)
			if gameCount == 0 && len(allPlatforms) > 0 {
				logger.Info("Games table is empty, forcing full sync of all platforms for recovery")
				platforms = allPlatforms
				isBulkLoad = true // Force full sync (no UpdatedAfter)
			}
		}
	}

	totalExpectedGames := int64(0)
	for _, p := range platforms {
		totalExpectedGames += int64(p.ROMCount)
	}

	if totalExpectedGames == 0 && len(platforms) > 0 {
		totalExpectedGames = int64(len(platforms))
	}

	logger.Info("Sync planned", "total_platforms", len(platforms), "total_expected_games", totalExpectedGames)

	// Progress: games 0-85%, collections 85-98%, done 100%
	gamesFetched := &atomic.Int64{}
	updateProgress := func(count int) {
		if progress != nil {
			fetched := gamesFetched.Add(int64(count))
			if totalExpectedGames <= 0 {
				progress.Store(0.85)
				return
			}
			pct := float64(fetched) / float64(totalExpectedGames) * 0.85
			if pct > 0.85 {
				pct = 0.85
			}
			logger.Info("Sync progress", "fetched_games", fetched, "total_expected", totalExpectedGames, "percent", fmt.Sprintf("%.1f%%", pct*100))
			progress.Store(pct)
		}
	}

	var firstErr error
	logger.Info("Starting cache population", "platform_count", len(platforms), "is_bulk_load", isBulkLoad)

	if len(platforms) == 0 {
		logger.Warn("No platforms identified for sync.")
		if progress != nil {
			progress.Store(1.0)
		}
		return stats, nil
	}

	// Fetch all games in one paginated stream (no per-platform round-trips)
	fetchUpdatedAfter := updatedAfter
	if isBulkLoad {
		fetchUpdatedAfter = ""
	}

	offset := 0
	expectedTotal := 0
	gamesByPlatform := make(map[int][]romm.Rom)

	for {
		q := romm.GetRomsQuery{
			Offset:       offset,
			Limit:        DefaultRomPageSize,
			UpdatedAfter: fetchUpdatedAfter,
		}

		res, err := client.GetRoms(q)
		if err != nil {
			logger.Error("Failed to fetch games", "offset", offset, "error", err)
			firstErr = err
			break
		}

		if offset == 0 {
			expectedTotal = res.Total
			logger.Info("Fetching all games", "total", expectedTotal)
		}

		for i := range res.Items {
			pid := res.Items[i].PlatformID
			gamesByPlatform[pid] = append(gamesByPlatform[pid], res.Items[i])
		}

		updateProgress(len(res.Items))

		if len(res.Items) == 0 || offset+len(res.Items) >= expectedTotal {
			break
		}
		offset += len(res.Items)
	}

	// Save grouped by platform (one DB transaction per platform)
	for _, p := range platforms {
		games := gamesByPlatform[p.ID]
		if len(games) == 0 {
			continue
		}
		if err := cm.SavePlatformGames(p.ID, games); err != nil {
			logger.Error("Failed to save platform games", "platformID", p.ID, "name", p.Name, "error", err)
			cm.RecordPlatformSyncFailure(p.ID)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			cm.RecordPlatformSyncSuccess(p.ID, len(games))
		}
	}

	if firstErr == nil {
		cm.RecordRefreshTime(MetaKeyGamesRefreshedAt)
	}

	// Collections (85-98%)
	stats.CollectionsSynced = cm.fetchAndCacheCollectionsWithProgress(progress, 0.85, 0.98)
	cm.RecordRefreshTime(MetaKeyCollectionsRefreshedAt)

	// Purge items deleted from the server (only during incremental updates)
	if !isBulkLoad {
		cm.purgeDeletedItems(client)
	}

	if progress != nil {
		progress.Store(1.0)
	}

	stats.GamesUpdated = int(gamesFetched.Load())
	logger.Info("Cache population completed", "platforms", len(platforms), "games", stats.GamesUpdated)
	return stats, firstErr
}

type fetchOpts struct {
	client        *romm.Client    // Reusable HTTP client
	onProgress    func(count int) // Called with count of games fetched (for batch progress)
	onPctProgress *atomic.Float64 // Set with percentage 0.0-1.0 (for UI progress bars)
	updatedAfter  string
	force         bool
}

func (cm *Manager) fetchPlatformGames(platform romm.Platform, opts *fetchOpts) (int, error) {
	if opts == nil {
		opts = &fetchOpts{}
	}

	logger := internal.GetLogger()
	client := opts.client
	if client == nil {
		client = romm.NewClientFromHost(cm.host, cm.config.GetApiTimeout())
	}

	var allGames []romm.Rom
	offset := 0
	expectedTotal := 0

	fetchUpdatedAfter := opts.updatedAfter
	if opts.force {
		fetchUpdatedAfter = ""
	}

	for {
		q := romm.GetRomsQuery{
			PlatformIDs:  []int{platform.ID},
			Offset:       offset,
			Limit:        DefaultRomPageSize,
			UpdatedAfter: fetchUpdatedAfter,
		}

		logger.Debug("Fetching games page", "platform", platform.Name, "offset", offset, "limit", DefaultRomPageSize, "updated_after", fetchUpdatedAfter)
		res, err := client.GetRoms(q)
		if err != nil {
			logger.Error("Failed to fetch games",
				"platform", platform.Name,
				"offset", offset,
				"error", err)
			return 0, err
		}

		logger.Debug("Fetched games page", "platform", platform.Name, "count", len(res.Items), "total_available", res.Total)

		if offset == 0 {
			expectedTotal = res.Total
			// Pre-allocate the exact slice capacity to prevent memory spikes
			allGames = make([]romm.Rom, 0, expectedTotal)
		}

		allGames = append(allGames, res.Items...)

		if opts.onProgress != nil && len(res.Items) > 0 {
			opts.onProgress(len(res.Items))
		}
		if opts.onPctProgress != nil && expectedTotal > 0 {
			pct := float64(len(allGames)) / float64(expectedTotal)
			if pct > 1.0 {
				pct = 1.0
			}
			opts.onPctProgress.Store(pct)
		}

		if len(allGames) >= expectedTotal || len(res.Items) == 0 || len(res.Items) < DefaultRomPageSize {
			break
		}

		offset += len(res.Items)
	}

	if opts.updatedAfter != "" {
		logger.Debug("Fetched updated platform games",
			"platform", platform.Name,
			"count", len(allGames),
			"updated_after", opts.updatedAfter)
	} else {
		logger.Debug("Cached platform games",
			"platform", platform.Name,
			"count", len(allGames))
	}

	return len(allGames), cm.SavePlatformGames(platform.ID, allGames)
}

func (cm *Manager) fetchAndCacheCollectionsWithProgress(progress *atomic.Float64, progressStart, progressEnd float64) int {
	logger := internal.GetLogger()

	showRegular := cm.config.GetShowCollections()
	showSmart := cm.config.GetShowSmartCollections()
	showVirtual := cm.config.GetShowVirtualCollections()

	if !showRegular && !showSmart && !showVirtual {
		logger.Debug("Skipping collection sync - no collection types enabled")
		if progress != nil {
			progress.Store(progressEnd)
		}
		return 0
	}

	client := romm.NewClientFromHost(cm.host, cm.config.GetApiTimeout())

	var updatedAfter string
	if lastRefresh, err := cm.GetLastRefreshTime(MetaKeyCollectionsRefreshedAt); err == nil {
		updatedAfter = lastRefresh.Format(time.RFC3339)
		logger.Debug("Using incremental collection update", "updated_after", updatedAfter)
	}

	var query romm.GetCollectionsQuery
	if updatedAfter != "" {
		query = romm.GetCollectionsQuery{UpdatedAfter: updatedAfter}
	}

	var allCollections []romm.Collection
	var mu sync.Mutex
	var wg sync.WaitGroup

	if showRegular {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collections, err := client.GetCollections(query)
			if err != nil {
				logger.Error("Failed to fetch regular collections", "error", err)
				return
			}
			mu.Lock()
			allCollections = append(allCollections, collections...)
			mu.Unlock()
		}()
	}

	if showSmart {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collections, err := client.GetSmartCollections(query)
			if err != nil {
				logger.Error("Failed to fetch smart collections", "error", err)
				return
			}
			for i := range collections {
				collections[i].IsSmart = true
			}
			mu.Lock()
			allCollections = append(allCollections, collections...)
			mu.Unlock()
		}()
	}

	if showVirtual {
		wg.Add(1)
		go func() {
			defer wg.Done()
			virtualCollections, err := client.GetVirtualCollections()
			if err != nil {
				logger.Error("Failed to fetch virtual collections", "error", err)
				return
			}
			mu.Lock()
			for _, vc := range virtualCollections {
				allCollections = append(allCollections, vc.ToCollection())
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if progress != nil {
		progress.Store(progressStart + (progressEnd-progressStart)*0.5)
	}

	if len(allCollections) == 0 {
		if progress != nil {
			progress.Store(progressEnd)
		}
		return 0
	}

	if err := cm.SaveCollections(allCollections); err != nil {
		logger.Error("Failed to save collections", "error", err)
	}

	if err := cm.SaveAllCollectionMappings(allCollections); err != nil {
		logger.Error("Failed to save collection mappings", "error", err)
	}

	if progress != nil {
		progress.Store(progressEnd)
	}

	logger.Debug("Cached collections", "count", len(allCollections))
	return len(allCollections)
}

// purgeDeletedItems fetches identifier lists from the server and removes any
// cached items that no longer exist. This handles server-side deletions that
// incremental (UpdatedAfter) syncing would otherwise miss.
func (cm *Manager) purgeDeletedItems(client *romm.Client) {
	logger := internal.GetLogger()

	var platformIDs, romIDs, collectionIDs []int
	var platformErr, romErr, collectionErr error

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		platformIDs, platformErr = client.GetPlatformIdentifiers()
	}()
	go func() {
		defer wg.Done()
		romIDs, romErr = client.GetRomIdentifiers()
	}()
	go func() {
		defer wg.Done()
		collectionIDs, collectionErr = client.GetCollectionIdentifiers()
	}()

	wg.Wait()

	if platformErr != nil {
		logger.Debug("Failed to fetch platform identifiers for purge", "error", platformErr)
	} else if len(platformIDs) > 0 {
		if _, err := cm.PurgeDeletedPlatforms(platformIDs); err != nil {
			logger.Debug("Failed to purge deleted platforms", "error", err)
		}
	}

	if romErr != nil {
		logger.Debug("Failed to fetch rom identifiers for purge", "error", romErr)
	} else if len(romIDs) > 0 {
		if _, err := cm.PurgeDeletedGames(romIDs); err != nil {
			logger.Debug("Failed to purge deleted games", "error", err)
		}
	}

	if collectionErr != nil {
		logger.Debug("Failed to fetch collection identifiers for purge", "error", collectionErr)
	} else if len(collectionIDs) > 0 {
		if _, err := cm.PurgeDeletedCollections(collectionIDs); err != nil {
			logger.Debug("Failed to purge deleted collections", "error", err)
		}
	}
}

func (cm *Manager) RefreshPlatformGames(platform romm.Platform) error {
	if cm == nil || !cm.initialized {
		return ErrNotInitialized
	}

	_, err := cm.fetchPlatformGames(platform, nil)
	return err
}

func (cm *Manager) RefreshPlatformGamesWithProgress(platform romm.Platform, progress *atomic.Float64) error {
	if cm == nil || !cm.initialized {
		return ErrNotInitialized
	}

	var updatedAfter string
	if lastRefresh, err := cm.GetLastRefreshTime(MetaKeyGamesRefreshedAt); err == nil {
		updatedAfter = lastRefresh.Format(time.RFC3339)
		internal.GetLogger().Debug("Using incremental refresh", "updated_after", updatedAfter)
	}

	_, err := cm.fetchPlatformGames(platform, &fetchOpts{
		onPctProgress: progress,
		updatedAfter:  updatedAfter,
	})

	if progress != nil {
		progress.Store(1.0)
	}

	return err
}
