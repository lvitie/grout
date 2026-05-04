# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Build/Lint/Test Commands

- `nix develop` - Enter development shell with all dependencies (required before running any Go commands)
- `nix build` - Build the grout binary (local build)
- `nix run` - Run grout directly
- `task code:lint` - Run go fmt, go vet, staticcheck
- `go test ./...` - Run all tests
- `go test ./sync -v` - Run tests in sync package with verbose output
- `go test -run TestName ./...` - Run a specific test by name
- `task build:arm64` - Cross-compile for ARM64 (requires Docker)
- `task package:muos` - Package for muOS platform

## Code Style

- Uses Go 1.25+ with standard practices
- Project-specific utilities:
  - Use `internal.LoadConfig()` to read user settings
  - Use `romm.NewClientFromHost()` for API client creation
  - Use `cache.GetCacheManager()` for cache access
  - Use `sync.StartSync()` for save synchronization
- All core packages are goroutine-safe
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer
- Use `gaba.GetLogger()` for logging (gabagool logger)
- Configuration files are JSON with specific structures in `internal/config.go`

## Custom Utilities & Patterns

- **Configuration Management**: `internal/config.go` handles JSON config loading/saving with default values and validation
- **Cache System**: SQLite-backed cache in `cache/` package with background sync capabilities
- **Save Synchronization**: `sync/` package handles bidirectional save syncing with conflict resolution
- **Platform Abstraction**: `cfw/` package handles device-specific paths and behaviors for different firmware targets
- **Artwork Handling**: `internal/imageutil/` for downloading and converting artwork between formats
- **RomM API Client**: `romm/` package provides typed structs for API interactions
- **Background Operations**: `cache/` and `sync/` packages support background operations without blocking UI

## Non-standard Directory Structure

- Core business logic is completely independent of UI (`romm/`, `internal/`, `cache/`, `sync/`)
- UI layer is separate (`gui/`, `cmd/grout-gui/`) and will be replaced with Flatpak-compatible framework
- Platform-specific code in `cfw/` targets custom firmware (muOS, Knulli, Spruce, etc.)
- Configuration files are loaded from current working directory (not hardcoded paths)
- Cache directory is `.cache/grout.db` in current working directory

## Project-Specific Conventions

- All core packages are goroutine-safe (use `sync.Mutex`, `atomic` operations where needed)
- The required RomM version matches the first three components of Grout's version
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine; UI must not block it
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer
- Grout aggressively tracks new RomM features - older servers may still work but support is not guaranteed
- File operations may fail gracefully (missing artwork, ROM files); handle these without crashing
- Network errors should have user-friendly fallbacks; consider offline mode via the cache
- Localization uses `go-i18n` for translations
- All API errors are typed (`romm.ErrUnauthorized`, etc.) - check the API client for error constants

## Critical Gotchas

- Configuration files are loaded from the current working directory, not hardcoded paths
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