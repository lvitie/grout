# Project Coding Rules (Non-Obvious Only)

- Always use `internal.LoadConfig()` from `internal/config.go` to read user settings instead of direct file I/O
- Use `romm.NewClientFromHost()` for API client creation instead of `romm.NewClient()` directly
- Use `cache.GetCacheManager()` for cache access instead of direct database operations
- Use `sync.StartSync()` for save synchronization instead of direct sync operations
- All core packages (`romm/`, `internal/`, `cache/`, `sync/`) are goroutine-safe - use `sync.Mutex` or `atomic` operations where needed
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine; UI must not block it
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer
- Use `gaba.GetLogger()` for logging (gabagool logger) instead of standard log package
- Configuration files are JSON with specific structures in `internal/config.go`
- The `sync/` package handles directory-based saves (e.g., PPSSPP) differently from file-based saves
- When replacing the GUI, only modify `gui/` and `cmd/grout-gui/` (or create a new equivalent)
- The `cache/` package uses SQLite with WAL journal mode for better concurrency
- The `sync/` package uses a complex conflict resolution system for save files
- `internal/config.go` has special handling for DurationSeconds that marshals to/from JSON as whole seconds
- `cache/manager.go` has special bulk load mode optimizations for SQLite performance
- **Important**: Run `nix develop` before executing any Go development commands