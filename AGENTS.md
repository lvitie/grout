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

## Code Style

- Uses Go 1.25+ with standard practices
- Project-specific utilities:
  - Use `internal.LoadConfig()` to read user settings
  - Use `romm.NewClientFromHost()` for API client creation
  - Use `cache.GetCacheManager()` for cache access
  - Use `sync.StartSync()` for save synchronization
  - Use `internal.GetLogger()` for logging (slog-based)
  - Use `platform.GetCurrent()` for platform-specific paths
- All core packages are goroutine-safe
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer
- Configuration files are JSON with specific structures in `internal/config.go`

## Custom Utilities & Patterns

- **Configuration Management**: `internal/config.go` handles JSON config loading/saving with default values and validation
- **Cache System**: SQLite-backed cache in `cache/` package with background sync capabilities
- **Save Synchronization**: `sync/` package handles bidirectional save syncing with conflict resolution
- **Platform Abstraction**: `platform/` package provides a unified interface for Linux desktop and handheld environments
- **Artwork Handling**: `internal/imageutil/` for processing artwork
- **RomM API Client**: `romm/` package provides typed structs for API interactions
- **Background Operations**: `cache/` and `sync/` packages support background operations without blocking UI

## Directory Structure

- **`cmd/grout-desktop`**: Entry point for the GTK4/Adwaita application
- **`desktop/`**: GTK4 application logic, screens, and navigation
- **`platform/`**: Platform abstraction layer (Linux desktop, Handhelds)
- **`internal/`**: Core utilities, config, and logging
- **`cache/`**: SQLite-backed persistent cache
- **`sync/`**: Save synchronization and conflict resolution
- **`romm/`**: RomM API client

## Project-Specific Conventions

- All core packages are goroutine-safe (use `sync.Mutex`, `atomic` operations where needed)
- The required RomM version matches the first three components of Grout's version
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine; UI must not block it
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer
- Grout aggressively tracks new RomM features - older servers may still work but support is not guaranteed
- File operations may fail gracefully (missing artwork, ROM files); handle these without crashing
- Network errors should have user-friendly fallbacks; consider offline mode via the cache
- Localization uses `internal/i18n` wrapper for translations
- All API errors are typed (`romm.ErrUnauthorized`, etc.) - check the API client for error constants

## Critical Gotchas

- Configuration files are loaded via `platform.GetCurrent().ConfigDir()`
- Save sync requires `Host.DeviceID` for client identification
- Cache directory is `platform.GetCurrent().CacheDir()`
- GTK4 UI uses `adw.NavigationView` for navigation; ensure all screens are `adw.NavigationPage`
- The `sync/` package handles directory-based saves (e.g., PPSSPP) differently from file-based saves
- `internal/config.go` has special handling for DurationSeconds that marshals to/from JSON as whole seconds
- `cache/manager.go` has special bulk load mode optimizations for SQLite performance