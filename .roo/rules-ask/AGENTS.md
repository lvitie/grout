# Project Documentation Rules (Non-Obvious Only)

- "src/" contains VSCode extension code, not source for web apps (counterintuitive)
- Provider examples in src/api/providers/ are the canonical reference (docs are outdated)
- UI runs in VSCode webview with restrictions (no localStorage, limited APIs)
- Package.json scripts must be run from specific directories, not root
- Locales in root are for extension, webview-ui/src/i18n for UI (two separate systems)
- Configuration files are loaded from current working directory (not hardcoded paths)
- The `config.json` and `save_slots.json` files are in the working directory
- Save sync requires `Host.DeviceID` for client identification
- Cache directory is `.cache/grout.db` in current working directory
- Platform-specific code in `cfw/` is tailored for custom firmware; will need adaptation for Linux/Flatpak
- The `sync/` package handles directory-based saves (e.g., PPSSPP) differently from file-based saves
- When replacing the GUI, only modify `gui/` and `cmd/grout-gui/` (or create a new equivalent)
- The `cache/` package uses SQLite with WAL journal mode for better concurrency
- The `sync/` package uses a complex conflict resolution system for save files
- `internal/config.go` has special handling for DurationSeconds that marshals to/from JSON as whole seconds
- `cache/manager.go` has special bulk load mode optimizations for SQLite performance