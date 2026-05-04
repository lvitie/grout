# Grout Desktop Migration Plan: GTK4 + Adwaita

## Goal

Strip all custom-firmware GUI code and platform abstractions from Grout. Keep the backend packages intact. Build a new Linux desktop GUI targeting:

- **Primary:** Linux desktop (Flatpak distribution)
- **Stretch:** Steam Deck (Gaming Mode via gamescope, Desktop Mode natively)

---

## Technology Choice: GTK4 + libadwaita

| Requirement | Solution |
|---|---|
| **Native Linux UI** | GTK4 via [`gotk4`](https://github.com/diamondburned/gotk4) — auto-generated Go bindings |
| **Modern look** | libadwaita via [`gotk4-adwaita`](https://github.com/diamondburned/gotk4-adwaita) — adaptive layouts, dark mode, GNOME HIG |
| **Flatpak** | GTK4/Adwaita are the native Flatpak runtime libs — zero bundling overhead |
| **Controller input** | [`go-evdev`](https://github.com/holoplot/go-evdev) (already an indirect dep) — pure Go, reads `/dev/input/event*` |
| **Keyboard/mouse** | GTK4 `EventControllerKey` + standard widget focus navigation |
| **Steam Deck** | Just Linux — GTK4 works in Desktop Mode; Gaming Mode needs `--device=input` Flatpak permission |

### Why not Fyne?

The previous plan proposed Fyne. Fyne is wrong for this project:

- **Not GTK** — renders via OpenGL, won't look native, won't respect system themes
- **No controller support** — no joystick/gamepad API exists
- **More C deps than SDL2** — needs OpenGL + X11 + GLFW headers (despite claiming "pure Go")
- **Poor Flatpak fit** — non-native widgets, no Adwaita integration

### Key gotk4 facts

- Requires **CGo** + C compiler + GTK4/libadwaita dev headers
- Latest GTK4 uses `GtkListView`/`GtkColumnView` (not the deprecated `GtkTreeView`)
- Event handling uses `GtkEventController` pattern (not old-style signal handlers)
- Occasional memory leak/crash edge cases — mature but not perfect

---

## Phase 0: Inventory — What Stays, What Goes

### ✅ Keep (backend — no gabagool/cfw dependencies)

| Package | Purpose | Gabagool-free? | CFW-free? |
|---|---|---|---|
| `romm/` | RomM API client | ✅ | ✅ |
| `internal/` | Config, imageutil, fileutil, etc. | ✅ | ✅ |
| `cache/` | SQLite cache + background sync | ✅ | ✅ |
| `bios/` | BIOS file handling | ✅ | ✅ |
| `resources/` | Embedded assets (locale files, images) | ✅ | ✅ |
| `version/` | Version info | ✅ | ✅ |
| `update/` | Auto-update logic | ✅ | ✅ |
| `vendored/` | Vendored dependencies | ✅ | ✅ |

### ⚠️ Keep with modifications

| Package | Issue | Action needed |
|---|---|---|
| `sync/` | Imports `cfw` for save paths, ROM scanning, emulator folder maps | Replace `cfw.*` calls with a new `platform` interface (see Phase 1) |

### 🗑️ Remove entirely

| Package/Dir | Reason |
|---|---|
| `ui/` (39 files, ~200KB) | All screens use gabagool APIs (`gaba.List()`, `gaba.DetailScreen()`, etc.) |
| `app/` (7 files, ~56KB) | Router, transitions (28KB), setup — all tightly coupled to gabagool |
| `cfw/` (6 files + 11 subdirs) | muOS, Knulli, Spruce, Allium, Onion, MinUI, ROCKNIX, NextUI, Batocera, TrimUI, Koriki |
| `cmd/grout-gui/` | Empty — was the old entry point |
| `gui/` | Empty |

### 🗑️ Remove from go.mod

```
github.com/BrandonKowalski/gabagool/v2  (+ transitive: go-sdl2, go-evdev via gabagool)
```

### ➕ Add to go.mod

```
github.com/diamondburned/gotk4           # GTK4 bindings
github.com/diamondburned/gotk4-adwaita   # libadwaita bindings
github.com/holoplot/go-evdev             # Controller input (promote from indirect)
```

---

## Phase 1: Platform Abstraction Layer ✅ DONE

Replace `cfw/` with a `platform/` package that provides the same interfaces but for Linux desktop.

### 1.1 New `platform/` package

```go
// platform/platform.go
package platform

// Platform provides OS-specific paths and behaviors.
// On CFW this was cfw.CFW with device-specific impls.
// For desktop Linux, there's a single implementation.
type Platform interface {
    Name() string
    RomDirectory() string
    BaseSavePath() string
    GetSaveDirectory(fsSlug string) string
    EmulatorFolderMap() map[string]string
    ConfigDir() string   // XDG_CONFIG_HOME/grout
    CacheDir() string    // XDG_CACHE_HOME/grout
    DataDir() string     // XDG_DATA_HOME/grout
}
```

### 1.2 Linux desktop implementation

```go
// platform/linux.go
package platform

import "os"
import "path/filepath"

type LinuxDesktop struct {
    configDir string
    cacheDir  string
    dataDir   string
}

func NewLinuxDesktop() *LinuxDesktop {
    home := os.Getenv("HOME")
    return &LinuxDesktop{
        configDir: envOrDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config", "grout")),
        cacheDir:  envOrDefault("XDG_CACHE_HOME", filepath.Join(home, ".cache", "grout")),
        dataDir:   envOrDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share", "grout")),
    }
}
```

### 1.3 Update `sync/` to use interface

Replace direct `cfw.GetCFW()`, `cfw.BaseSavePath()`, `cfw.ScanRoms()` calls with injected `platform.Platform`. This is ~8 call sites in `sync/flow.go` and `sync/roms.go`.

---

## Phase 2: New GUI Architecture (Skeleton ✅)

### 2.1 Project structure

```
cmd/
  grout-desktop/
    main.go                    # GTK4 Application entry point

desktop/
  app.go                      # adw.Application setup, theme, window
  router.go                   # Screen navigation (stack-based, async)
  state.go                    # Reactive app state
  controller/
    controller.go             # go-evdev gamepad polling
    mapping.go                # Button→action mapping
  screens/
    platform_selection.go     # GtkListBox of platforms
    game_list.go              # GtkListView with artwork
    game_details.go           # Detail view with cover art + metadata
    game_options.go           # Options menu
    settings.go               # Settings hub
    general_settings.go
    advanced_settings.go
    collections_settings.go
    tools_settings.go
    login.go                  # Server URL + credentials
    search.go                 # GtkSearchEntry
    download.go               # Progress UI
    save_sync.go              # Sync progress + status
    save_conflict.go          # Conflict resolution dialog
    sync_menu.go
    synced_games.go
    sync_history.go
    save_mapping.go
    collection_selection.go
    collection_platform.go
    game_qr.go
    game_filters.go
    platform_mapping.go
    bios_download.go
    artwork_sync.go
    rebuild_cache.go
    update.go
    info.go
    server_address.go
  widgets/
    game_row.go               # Custom list row with artwork
    progress_overlay.go       # Download/sync progress overlay
    status_indicator.go       # Connection/sync status
  dialogs/
    confirmation.go           # adw.MessageDialog wrapper
    error.go                  # Error display
```

### 2.2 Paradigm shift: synchronous → async

This is the hardest part of the migration.

**Current model (gabagool):** Each screen's `Draw()` method **blocks** until the user selects something, then returns an output struct. The router calls the next screen synchronously. The transition function is a giant switch statement mapping `(screen, output) → (nextScreen, input)`.

**New model (GTK4):** GTK runs an event loop. Screens are built as widgets, user actions trigger **callbacks**, and navigation happens by swapping the content of the main window.

**Migration strategy:**

```go
// desktop/router.go
type Router struct {
    window  *adw.ApplicationWindow
    nav     *adw.NavigationView    // Adwaita's built-in navigation stack
    state   *AppState
}

// Navigate pushes a new screen onto the navigation stack
func (r *Router) Navigate(screen Screen) {
    page := screen.Build(r)   // Screen returns an adw.NavigationPage
    r.nav.Push(page)
}

// Back pops the current screen
func (r *Router) Back() {
    r.nav.Pop()
}
```

Each screen becomes a builder that returns an `adw.NavigationPage`:

```go
// desktop/screens/platform_selection.go
type PlatformSelectionScreen struct {
    router    *Router
    platforms []romm.Platform
}

func (s *PlatformSelectionScreen) Build(router *Router) *adw.NavigationPage {
    listBox := gtk.NewListBox()
    listBox.SetSelectionMode(gtk.SelectionSingle)

    for _, p := range s.platforms {
        row := gtk.NewListBoxRow()
        label := gtk.NewLabel(p.Name)
        row.SetChild(label)
        listBox.Append(row)
    }

    listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
        idx := row.Index()
        router.Navigate(&GameListScreen{
            platform: s.platforms[idx],
        })
    })

    scrolled := gtk.NewScrolledWindow()
    scrolled.SetChild(listBox)

    page := adw.NewNavigationPage(scrolled, "Grout")
    return page
}
```

### 2.3 Application entry point

```go
// cmd/grout-desktop/main.go
package main

import (
    "os"
    "github.com/diamondburned/gotk4-adwaita/pkg/adw"
)

func main() {
    app := adw.NewApplication("app.romm.Grout", 0)
    app.ConnectActivate(func() {
        // Load config, init cache, create router, show first screen
    })
    os.Exit(app.Run(os.Args))
}
```

### 2.4 Controller support

```go
// desktop/controller/controller.go
package controller

import (
    "github.com/holoplot/go-evdev"
)

type Action int
const (
    ActionUp Action = iota
    ActionDown
    ActionLeft
    ActionRight
    ActionConfirm   // A / Cross
    ActionBack      // B / Circle
    ActionMenu      // X / Square
    ActionAlt       // Y / Triangle
    ActionL1
    ActionR1
)

type Handler struct {
    device   *evdev.InputDevice
    actionCh chan Action
    stopCh   chan struct{}
}

// Start polls the evdev device in a goroutine and sends Actions
func (h *Handler) Start() {
    go func() {
        for {
            select {
            case <-h.stopCh:
                return
            default:
                event, err := h.device.ReadOne()
                if err != nil { continue }
                if action, ok := h.mapEvent(event); ok {
                    h.actionCh <- action
                }
            }
        }
    }()
}
```

The controller goroutine sends actions to a channel. The GTK main loop consumes them via `glib.IdleAdd()` to safely dispatch from the main thread:

```go
// In router setup:
go func() {
    for action := range controller.Actions() {
        glib.IdleAdd(func() {
            router.HandleControllerAction(action)
        })
    }
}()
```

---

## Phase 3: Screen-by-Screen Migration (In Progress ⚠️)

### 3.1 Migration order (by dependency)

Each screen is independent — migrate in order of user flow:

| Priority | Screen | GTK4 widget | Notes |
|---|---|---|---|
| 1 | Login | `adw.EntryRow` + `adw.PasswordEntryRow` | ✅ DONE |
| 2 | Platform Selection | `gtk.ListBox` + `adw.ActionRow` | ✅ DONE |
| 3 | Game List | `gtk.ListView` + factory | ✅ DONE |
| 4 | Game Details | `adw.Clamp` + cover image + metadata | ✅ DONE |
| 5 | Settings | `adw.PreferencesPage` | Natural fit for Adwaita |
| 6 | Download progress | `adw.StatusPage` + `gtk.ProgressBar` | Overlay pattern |
| 7 | Search | `gtk.SearchEntry` + filter | Integrate into game list header |
| 8 | Save sync screens | `adw.StatusPage` + progress | 5 related screens |
| 9 | Collections | `gtk.ListBox` | Same pattern as platforms |
| 10 | Remaining settings | `adw.PreferencesPage` | General, Advanced, Tools, Collections |
| 11 | Misc | Various | QR, BIOS, artwork sync, info, update |

### 3.2 gabagool → GTK4 widget mapping

| gabagool API | GTK4/Adwaita equivalent |
|---|---|
| `gaba.List()` | `gtk.ListBox` or `gtk.ListView` (for large lists) |
| `gaba.DetailScreen()` | `adw.Clamp` + `gtk.Box` layout |
| `gaba.OptionsList()` | `adw.PreferencesGroup` + `adw.ComboRow` / `adw.SwitchRow` |
| `gaba.DownloadManager()` | Custom widget with `gtk.ProgressBar` |
| `gaba.ProcessMessage()` | `adw.StatusPage` + spinner |
| `gaba.ConfirmationMessage()` | `adw.MessageDialog` |
| `gaba.SelectionMessage()` | `adw.MessageDialog` with custom content |
| `gaba.Keyboard()` | Native keyboard input (it's a desktop!) |
| `gaba.StatusBar()` | `adw.HeaderBar` with status indicators |
| `gaba.GetLogger()` | `log/slog` (stdlib) — replace throughout |

---

## Phase 4: Flatpak Packaging

### 4.1 Flatpak manifest

```yaml
app-id: app.romm.Grout
runtime: org.gnome.Platform
runtime-version: '47'
sdk: org.gnome.Sdk
command: grout-desktop

finish-args:
  - --share=network             # RomM API access
  - --share=ipc                 # X11 shared memory
  - --socket=fallback-x11       # X11 fallback
  - --socket=wayland            # Wayland native
  - --device=dri                # GPU acceleration
  - --device=input              # Controller/gamepad access (/dev/input)
  - --filesystem=xdg-data/grout:create    # Save data
  - --filesystem=xdg-cache/grout:create   # Cache/DB

modules:
  - name: grout
    buildsystem: simple
    build-commands:
      - go build -o /app/bin/grout-desktop ./cmd/grout-desktop
    sources:
      - type: dir
        path: .
```

### 4.2 Steam Deck notes

- **Desktop Mode:** Works natively — GTK4 + Adwaita render fine under KDE
- **Gaming Mode:** Add as non-Steam game. Needs `--device=input` for controller. Gamescope may need env vars for proper scaling
- The `go-evdev` controller handler reads `/dev/input/event*` directly, which works in both modes

---

## Phase 5: Build System Updates

### 5.1 flake.nix changes

```nix
# Package build
buildInputs = with pkgs; [
  gtk4
  libadwaita
  pkg-config
  gobject-introspection
];

# Dev shell
buildInputs = with pkgs; [
  go_1_25
  go-task
  pkg-config
  gtk4
  libadwaita
  gobject-introspection
  # Remove: SDL2, SDL2_image, SDL2_ttf, SDL2_gfx
];
```

### 5.2 go.mod changes

```diff
 require (
-    github.com/BrandonKowalski/gabagool/v2 v2.19.0
+    github.com/diamondburned/gotk4 v0.3.1
+    github.com/diamondburned/gotk4-adwaita v0.0.0-20241212014936-...
+    github.com/holoplot/go-evdev v0.0.0-20250804134636-ab1d56a1fe83
     github.com/nicksnyder/go-i18n/v2 v2.6.1
     # ... rest stays
 )
```

```diff
 require (
-    github.com/veandco/go-sdl2 v0.4.40 // indirect
-    github.com/srwiley/oksvg v0.0.0-... // indirect  (gabagool dep)
-    github.com/srwiley/rasterx v0.0.0-... // indirect (gabagool dep)
     # ... rest stays
 )
```

---

## Phase 6: Logger Migration (In Progress ⚠️)

`gaba.GetLogger()` is used throughout — not just in UI code but referenced in AGENTS.md as the project logger. Replace with stdlib `log/slog`:

```go
// internal/logger.go
package internal

import (
    "log/slog"
    "os"
)

var logger *slog.Logger

func InitLogger(level slog.Level) {
    logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
    slog.SetDefault(logger)
}

func GetLogger() *slog.Logger {
    if logger == nil {
        InitLogger(slog.LevelInfo)
    }
    return logger
}
```

Then find-and-replace `gaba.GetLogger()` → `internal.GetLogger()` across the codebase.

---

## Phase 7: Cleanup

1. Delete `ui/`, `app/`, `cfw/`, `gui/`, `cmd/grout-gui/`
2. Run `go mod tidy` to drop gabagool + SDL2 + transitive deps
3. Update `AGENTS.md`:
   - Remove gabagool references
   - Update build commands (no SDL2)
   - Document new `desktop/` package structure
   - Update logger guidance
   - Remove CFW references
4. Update `flake.nix` — swap SDL2 for GTK4/Adwaita
5. Update CI/CD for Flatpak builds

---

## Migration Timeline

| Phase | Duration | Notes |
|---|---|---|
| Phase 0: Inventory | Done | This document |
| Phase 1: Platform abstraction | 1-2 days | `platform/` package, update `sync/` |
| Phase 2: GTK4 skeleton | 2-3 days | App, router, controller, window |
| Phase 3: Screen migration | 2-3 weeks | 30+ screens, largest effort |
| Phase 4: Flatpak packaging | 2-3 days | Manifest, testing |
| Phase 5: Build system | 1 day | flake.nix, go.mod |
| Phase 6: Logger migration | 1 day | Find-and-replace |
| Phase 7: Cleanup | 1 day | Delete old code, update docs |

**Total: ~4-5 weeks**

---

## Risk Assessment

| Risk | Impact | Mitigation |
|---|---|---|
| gotk4 memory leaks/crashes | Medium | Test heavily; gotk4 is mature for common widgets |
| Async paradigm shift breaks flow logic | High | Migrate screens incrementally; keep old code as reference |
| Controller input edge cases | Medium | go-evdev is simple; test with Xbox/PS/Steam Deck controllers |
| Large artwork lists performance | Medium | Use `GtkListView` (virtualized) instead of `GtkListBox` for game lists |
| Steam Deck Gaming Mode rendering | Low | GTK4 works under gamescope; may need scaling tweaks |
| Flatpak sandbox blocks controller | Low | `--device=input` permission handles this |
