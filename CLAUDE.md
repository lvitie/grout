# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Grout is a RomM (ROM Manager) client written in Go for retro gaming. The codebase cleanly separates business logic from UI:
- **Core packages** — reusable, UI-agnostic: `romm/` (RomM API client), `internal/` (config, image utilities, EmulationStation parsing), `cache/` (SQLite-backed ROM cache), `sync/` (save synchronization)
- **UI layer** — currently being replaced; old layer uses gabagool/Fyne, new one should use a Flatpak-compatible framework with controller support
- **Platform adaptation** — `cfw/` handles device-specific paths and behaviors; currently targets custom firmware (muOS, Knulli, Spruce, etc.) but will be adapted for Linux

Core responsibilities:
- RomM API communication (fetch metadata, platforms, artwork)
- Game/ROM management (discovery, multi-disk handling, collections)
- Artwork handling (download, format conversion, caching)
- EmulationStation gamelist.xml integration
- Game save synchronization across devices

## Development Setup

### Using Nix (Recommended for Linux)

```bash
# Enter dev environment with all dependencies
nix develop

# Or with cross-compilation tools (for ARM64 builds)
nix develop .#cross

# Or directly build the binary
nix build

# Or run the app directly
nix run
```

The `flake.nix` provides:
- **devShells.default** — Go toolchain, Task runner, SDL2, linting tools
- **devShells.cross** — Same as default, plus Docker and QEMU for cross-compilation
- **packages.default** — Prebuilt grout binary

After updating dependencies in `go.mod`, update the `vendorHash` in `flake.nix`:
```bash
# Set vendorHash to empty or wrong value, then run:
nix build  # Will fail and show correct hash
# Copy hash and update flake.nix
```

### Manual Installation

If not using Nix:

```bash
# Go toolchain (1.25.6+)
go version

# Task runner for build/lint/package tasks
go install github.com/go-task/task/v3/cmd/task@latest

# Code linting
go install honnef.co/go/tools/cmd/staticcheck@latest

# SDL2 dev libraries (if building locally on macOS)
# brew install sdl2 sdl2_image sdl2_ttf sdl2_gfx
```

### Code Quality

```bash
task code:lint              # Run go fmt, go vet, staticcheck
task code:gen-platforms     # Generate platforms.json from docs
```

### Testing

```bash
go test ./...               # Run all tests
go test ./sync -v           # Run tests in sync package with verbose output
go test -run TestName ./... # Run a specific test by name
```

### Building

```bash
# For development/local testing (currently Fyne-based)
go run ./cmd/grout-gui

# Cross-compile for ARM64 (used for device deployment)
task build:arm64            # Requires Docker

# Build and package for a specific platform
task package:muos           # muOS
task package:knulli         # Knulli
task package:next           # NextUI (TrimUI)
task package:spruce         # Spruce
```

## Core Architecture

### Separation of Concerns: Core vs. UI

The **core business logic is completely independent of UI**. When replacing the GUI, only modify `gui/` and `cmd/grout-gui/` (or create a new equivalent). Core packages handle:

- **`romm/`** — RomM API client (platform-agnostic)
- **`internal/`** — Configuration, image utilities, XML parsing, utilities
- **`cache/`** — SQLite-backed ROM cache system
- **`sync/`** — Save file synchronization

### Configuration & Startup

**`internal/config.go`** — Configuration management
- Loads/parses JSON config (RomM hosts, artwork preferences, language, etc.)
- `Config` struct: `Hosts []Host`, `AdditionalDownloads`, `Language`, `ShowCollections`
- Entry point for new UIs: `internal.LoadConfig()` to read user settings
- Settings file location is platform-dependent (handled by new UI)

**`romm/host.go`** — Per-server configuration
- `URL` — RomM server address
- `Username`, `Password` — Authentication credentials
- `DeviceID` — Used for save sync identification

### RomM API Client

**`romm/client.go`** — Main API client (use directly in your UI)
- `romm.NewClientFromHost(host *Host)` — Create authenticated client
- Methods: `GetHeartbeat()`, `GetPlatforms()`, `GetGames(platformID)`, `GetArtwork(gameID, artType)`
- Handles pagination transparently for large collections
- Returns typed structs: `Platform`, `Game`, `Artwork`

**`romm/` structs** — Data models (read-only after fetch)
- `Platform` — ID, Slug, Name, Icon URL, Emulator info, ROM file extensions
- `Game` — ID, Name, PlatformID, FileCount, IsMultiFile, Artwork URLs
- `Artwork` — Type-keyed map of art URLs per game

### Game & ROM Management

**`internal/gamelist/`** — EmulationStation gamelist.xml integration
- Use `gamelist.Load(path)` / `gamelist.Write(path)` to sync metadata with launcher configs
- Handles game collections (organizes ROM subsets)
- Used when integrating with launcher ecosystems

**`cfw/roms.go`, `cfw/saves.go`** — File discovery (platform-specific)
- Currently tailored to custom firmware; will need adaptation for Linux/Flatpak
- `FindROMs(platformID)` discovers ROM files by extension
- `FindSaves(platformID)` locates save files
- May be refactored into a platform abstraction layer during GUI rewrite

### Artwork Handling

**`internal/imageutil/`** — Image download and conversion
- Download artwork from RomM URLs
- Convert between WEBP, PNG, JPG formats
- QR code generation for game info
- Caching to avoid re-downloads
- Use `imageutil.Download()` and format conversion functions directly

**`internal/artutil/`** — Art type constants
- Types: Marquee, Cover, Screenshot, Video, Manual, Bezel, Fanart, BoxBack, BoxFront
- Used as keys when requesting artwork from RomM API

### Background Operations

**`cache/`** — SQLite cache and periodic sync
- `cache.NewBackgroundSync()` — Manages periodic cache refreshes
- Tracks ROM collections, artwork, metadata locally
- Reduces load on RomM server and enables offline browsing
- Can run in background without UI involvement

**`sync/`** — Game save synchronization
- `sync.StartSync()` — Coordinates save syncing across devices
- Uses `Host.DeviceID` for client identification
- Handles bidirectional sync with conflict resolution

## Important Architecture Notes

### RomM Version Compatibility
Grout aggressively tracks new RomM features. The required RomM version matches the first three components of Grout's version. Older servers may still work but support is not guaranteed. If API calls fail, check the RomM server version first.

### Threading & Concurrency
- Core packages are goroutine-safe (use `sync.Mutex`, `atomic` operations where needed)
- Background cache sync (`cache.NewBackgroundSync()`) runs in its own goroutine; UI must not block it
- Save sync (`sync.StartSync()`) is async; handle errors and completion callbacks in the UI layer

### Localization
Grout uses `go-i18n` for translations. Strings are extracted into `resources/locales/active.en.toml`. If adding new user-facing strings, mark them for translation using the i18n patterns already in the codebase.

### Error Handling
- RomM API errors are typed (`romm.ErrUnauthorized`, etc.) — check the API client for error constants
- File operations may fail gracefully (missing artwork, ROM files); handle these without crashing
- Network errors should have user-friendly fallbacks; consider offline mode via the cache

## Key Data Structures

**`Config`** (internal/config.go)
- `Hosts []Host` — Multiple RomM server configurations
- `AdditionalDownloads` — Which artwork types to download automatically
- `Language`, `ShowCollections` — User preferences
- Custom platform mappings

**`Host`** (romm/host.go)
- `URL` — RomM server address (http/https)
- `Username`, `Password` — Credentials (optional for public servers)
- `DeviceID` — Unique ID for save sync; must persist across sessions

**`Platform`** (romm/platform.go)
- `ID`, `Slug`, `Name` — Identifiers and display name
- `Icon` — URL to platform icon image
- `Emulator` — Emulator metadata (name, core, parameters)
- `RomExtensions` — File extensions associated (e.g., `.nes`, `.rom`)

**`Game`** (romm/game.go)
- `ID`, `Name`, `PlatformID` — Game identifiers
- `FileSizeBytes`, `FileCount` — Multi-file/multi-disk info
- `IsMultiFile` — Boolean flag for multi-disk games
- `Artwork` — Map of art type → URL (use `artutil` constants as keys)

## GUI Replacement: Flatpak + Controller Support

This project is transitioning from a Fyne-based UI (desktop) + gabagool-based UI (retro handhelds) to a new Linux-first, Flatpak-compatible GUI with controller support. The core business logic is ready for reuse; only the UI layer needs replacement.

### What to Keep (Core Packages — These Should NOT Change)

- **`romm/`** — RomM API client. Directly use `romm.Client` in your new UI.
- **`cache/`** — SQLite cache system. Integrate for offline browsing and reduced server load.
- **`sync/`** — Save synchronization. Use `sync.StartSync()` for game save syncing.
- **`internal/config.go`** — Configuration struct and loader. Adapt config file path for Flatpak (`~/.var/app/com.example.grout/config/grout.json` or similar).
- **`internal/imageutil/`** — Image download and conversion. Reuse directly.
- **`internal/artutil/`, `internal/gamelist/`, `internal/emulationstation/`** — Utility packages. Keep as-is.

### What to Replace/Adapt (UI and Platform-Specific Code)

1. **Remove**:
   - `app/` — Gabagool-specific state management and routing
   - `gui/` — Current Fyne/desktop GUI implementation
   - `cmd/grout-gui/` — Current entry point; replace with your Flatpak entry point

2. **Adapt or Create**:
   - **Entry point** — Replace `cmd/grout-gui/main.go` with Flatpak-compatible initialization. Reference `internal.LoadConfig()` and `romm.NewClientFromHost()`.
   - **UI framework** — Choose a Go library with good controller support (e.g., Ebiten, SDL2, or a headless API server + web frontend)
   - **Config file handling** — Update `internal/config.go` path defaults for Linux/Flatpak environment
   - **Controller input** — New UI must map controller buttons for menu navigation, game selection, settings

3. **Example Flow for New UI**:
   ```go
   // 1. Load config
   config, err := internal.LoadConfig()
   
   // 2. Choose a host from config.Hosts
   host := config.Hosts[0]
   client := romm.NewClientFromHost(host)
   
   // 3. Fetch data
   platforms, _ := client.GetPlatforms()
   games, _ := client.GetGames(platformID)
   
   // 4. Handle artwork
   artURL := game.Artwork[artutil.Cover]
   imageutil.Download(artURL, cachePath)
   
   // 5. Sync saves (optional)
   sync.StartSync(host.DeviceID, romPath, syncInterval)
   ```

### Build for Flatpak

Once your new UI code is in place:

```bash
# Build the new binary
go build -o grout ./cmd/grout-gui  # (or your new entry point)

# Package as Flatpak (you'll need a flatpak manifest)
# Use the binary in Flatseal to set up permissions, runtime, etc.
```

### Codebase Organization After Replacement

After the GUI rewrite, the structure should be:

```
grout/
├── romm/                    # [KEEP] RomM API client
├── internal/                # [ADAPT] Core utilities; keep packages, update config paths
├── cache/                   # [KEEP] SQLite cache
├── sync/                    # [KEEP] Save sync
├── cfw/                     # [CONSIDER] Platform-specific code; may shrink for Linux-only
├── cmd/
│   └── grout-gui/           # [REPLACE] New Flatpak/controller-based entry point
├── gui/                     # [REPLACE] New UI implementation
└── ...
```
