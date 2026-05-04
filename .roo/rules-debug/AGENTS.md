# Project Debug Rules (Non-Obvious Only)

- Webview dev tools accessed via Command Palette > "Developer: Open Webview Developer Tools" (not F12)
- IPC messages fail silently if not wrapped in try/catch in packages/ipc/src/
- Production builds require NODE_ENV=production or certain features break without error
- Database migrations must run from packages/evals/ directory, not root
- Extension logs only visible in "Extension Host" output channel, not Debug Console
- Cache directory is `.cache/grout.db` in current working directory
- Configuration files are loaded from the current working directory, not hardcoded paths
- The `config.json` and `save_slots.json` files are in the working directory
- Save sync requires `Host.DeviceID` for client identification
- Platform-specific code in `cfw/` is tailored for custom firmware; will need adaptation for Linux/Flatpak
- The `sync/` package handles directory-based saves (e.g., PPSSPP) differently from file-based saves
- When replacing the GUI, only modify `gui/` and `cmd/grout-gui/` (or create a new equivalent)
- The `cache/` package uses SQLite with WAL journal mode for better concurrency
- The `sync/` package uses a complex conflict resolution system for save files
- `internal/config.go` has special handling for DurationSeconds that marshals to/from JSON as whole seconds
- `cache/manager.go` has special bulk load mode optimizations for SQLite performance