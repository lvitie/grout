# Grout GUI Migration Plan: From gabagool/SDL2 to Fyne

## Executive Summary

This plan outlines the migration of Grout's GUI from the gabagool/SDL2-based framework to **Fyne**, a cross-platform Go GUI toolkit. Fyne is the ideal choice because:

- **Native Linux/Flatpak support** - Uses GTK under the hood
- **Built-in controller support** - Via evdev
- **Single binary distribution** - No external dependencies
- **Cross-platform** - Works on Linux, macOS, Windows
- **Go-native** - Pure Go, no C bindings needed

---

## Phase 1: Analysis & Planning

### 1.1 Current Architecture Analysis

**Dependencies to Remove:**
- `github.com/BrandonKowalski/gabagool/v2` (v2.19.0)
- `github.com/veandco/go-sdl2` (v0.4.40) - indirect
- All gabagool-specific APIs:
  - `gaba.List()` - list menus
  - `gaba.DetailScreen()` - info screens
  - `gaba.OptionsList()` - options menus
  - `gaba.DownloadManager()` - file downloads
  - `gaba.ProcessMessage()` - progress dialogs
  - `gaba.ConfirmationMessage()` - confirm dialogs
  - `gaba.SelectionMessage()` - selection dialogs
  - `gaba.Keyboard()` - virtual keyboard
  - `gaba.StatusBar()` - status bar
  - `gaba.DynamicStatusBarIcon()` - dynamic icons

**Controller Input:**
- Currently handled via gabagool's evdev integration
- Fyne provides equivalent support via `fyne/app/keyboard` and joystick detection

### 1.2 Fyne Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Fyne Application                  │
├─────────────────────────────────────────────────────┤
│  Window Manager                                      │
│  ├─ Window (GTK backend)                             │
│  ├─ Theme (Dark theme matching Grout)                │
│  └─ AppMetadata                                      │
├─────────────────────────────────────────────────────┤
│  Router (Screen Navigation)                          │
│  ├─ ScreenStack                                      │
│  ├─ ScreenRegistry                                   │
│  └─ TransitionHandler                                │
├─────────────────────────────────────────────────────┤
│  Screen Components                                   │
│  ├─ PlatformSelectionScreen                          │
│  ├─ GameListScreen                                   │
│  ├─ GameDetailsScreen                                │
│  ├─ SettingsScreen                                   │
│  └─ ...                                              │
├─────────────────────────────────────────────────────┤
│  Shared Components                                   │
│  ├─ DownloadManager                                  │
│  ├─ ProgressIndicator                                │
│  ├─ DialogManager                                    │
│  └─ ControllerHandler                                │
└─────────────────────────────────────────────────────┘
```

### 1.3 Screen Mapping

| Current gabagool Screen | Fyne Equivalent |
|-------------------------|-----------------|
| `gaba.List()` | `fyne.NewList()` / `fyne.NewTable()` |
| `gaba.DetailScreen()` | `fyne.NewForm()` / `fyne.NewObject()` |
| `gaba.OptionsList()` | `fyne.NewForm()` with `fyne.NewSelect()` |
| `gaba.DownloadManager()` | Custom `fyne.NewObject()` with progress |
| `gaba.ProcessMessage()` | `fyne.NewProgressBar()` + `fyne.NewLabel()` |
| `gaba.ConfirmationMessage()` | `fyne.NewDialog()` |
| `gaba.SelectionMessage()` | `fyne.NewDialog()` with `fyne.NewSelect()` |
| `gaba.Keyboard()` | `fyne.NewForm()` with `fyne.NewTextField()` |
| `gaba.StatusBar()` | `fyne.NewStatusBar()` (or custom widget) |

---

## Phase 2: New GUI Framework Setup (Fyne)

### 2.1 Project Structure

```
gui/fyne/
├── app/
│   ├── app.go              # Main application entry
│   ├── window.go           # Window management
│   └── theme.go            # Dark theme implementation
├── components/
│   ├── button.go           # Custom buttons
│   ├── list.go             # List views
│   ├── dropdown.go         # Dropdown menus
│   ├── dialog.go           # Dialog management
│   ├── progress.go         # Progress indicators
│   └── artwork.go          # Artwork display
├── screens/
│   ├── platform_selection.go
│   ├── game_list.go
│   ├── game_details.go
│   ├── game_options.go
│   ├── game_qr.go
│   ├── search.go
│   ├── collection_selection.go
│   ├── collection_platform_selection.go
│   ├── settings.go
│   ├── general_settings.go
│   ├── collections_settings.go
│   ├── advanced_settings.go
│   ├── tools_settings.go
│   ├── platform_mapping.go
│   ├── info.go
│   ├── logout_confirmation.go
│   ├── rebuild_cache.go
│   ├── bios_download.go
│   ├── artwork_sync.go
│   ├── update.go
│   ├── game_filters.go
│   ├── save_sync.go
│   ├── save_conflict.go
│   ├── sync_menu.go
│   ├── synced_games.go
│   ├── sync_history.go
│   ├── save_mapping.go
│   ├── server_address.go
│   └── input_mapping.go
├── router/
│   ├── router.go           # Screen router
│   ├── stack.go            # Screen stack
│   └── transitions.go      # Transition handlers
├── download/
│   ├── manager.go          # Download manager
│   └── artwork.go          # Artwork download
├── dialogs/
│   ├── confirmation.go
│   ├── selection.go
│   └── keyboard.go
└── statusbar/
    └── statusbar.go        # Status bar implementation
```

### 2.2 Theme Implementation

```go
// gui/fyne/theme/theme.go
package theme

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/theme"
)

type GroutTheme struct {
    primaryColor   fyne.Color
    secondaryColor fyne.Color
    accentColor    fyne.Color
}

func New() *GroutTheme {
    return &GroutTheme{
        primaryColor:   theme.Color(theme.NamePrimary),
        secondaryColor: theme.Color(theme.NameSecondary),
        accentColor:    theme.Color(theme.NameAccent),
    }
}

func (t *GroutTheme) Color(name fyne.ThemeColorName, size fyne.ThemeSize) fyne.Color {
    // Custom color mapping for Grout
    switch name {
    case "Primary":
        return t.primaryColor
    case "Secondary":
        return t.secondaryColor
    case "Accent":
        return t.accentColor
    default:
        return theme.Color(name, size)
    }
}

func (t *GroutTheme) Icon(name fyne.ThemeIconName) fyne.CanvasObject {
    // Custom icon handling
    return theme.Icon(name)
}
```

### 2.3 Controller Support

```go
// gui/fyne/controller/controller.go
package controller

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
)

type ControllerHandler struct {
    app        fyne.App
    joystick   *fyne.Joystick
    keymap     map[string]fyne.KeyName
}

func New() *ControllerHandler {
    app := app.New()
    joystick, _ := app.Device().Joystick()
    
    return &ControllerHandler{
        app:    app,
        joystick: joystick,
        keymap: buildKeymap(),
    }
}

func (h *ControllerHandler) HandleInput(event fyne.KeyEvent) {
    // Map controller input to Fyne events
    if event.Name == fyne.KeyUp {
        h.handleNavigation(-1, 0)
    } else if event.Name == fyne.KeyDown {
        h.handleNavigation(1, 0)
    } else if event.Name == fyne.KeyLeft {
        h.handleNavigation(0, -1)
    } else if event.Name == fyne.KeyRight {
        h.handleNavigation(0, 1)
    }
}

func (h *ControllerHandler) handleNavigation(dx, dy int) {
    // Implement navigation logic
}
```

---

## Phase 3: Screen Implementation (Fyne)

### 3.1 Platform Selection Screen

```go
// gui/fyne/screens/platform_selection.go
package screens

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type PlatformSelectionScreen struct {
    platforms []romm.Platform
    selected  int
    visible   int
}

func NewPlatformSelectionScreen(platforms []romm.Platform) *PlatformSelectionScreen {
    return &PlatformSelectionScreen{
        platforms: platforms,
    }
}

func (s *PlatformSelectionScreen) Create() fyne.CanvasObject {
    items := make([]widget.ListItem, len(s.platforms))
    for i, p := range s.platforms {
        items[i] = widget.ListItem{
            Text: p.Name,
            Data: p,
        }
    }
    
    list := widget.NewList(func(int, widget.ListItem) {
        // Item selected callback
    }, func(int, *widget.ListItem) {
        // Item focused callback
    }, func(int, widget.ListItem) {
        // Item layout callback
    }, items...)
    
    return container.NewBorder(
        widget.NewLabel("Select Platform"),
        nil, nil, nil,
        list,
    )
}
```

### 3.2 Game List Screen

```go
// gui/fyne/screens/game_list.go
package screens

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "image"
)

type GameListScreen struct {
    games     []romm.Rom
    platform  romm.Platform
    selected  int
    visible   int
}

func NewGameListScreen(games []romm.Rom, platform romm.Platform) *GameListScreen {
    return &GameListScreen{
        games:    games,
        platform: platform,
    }
}

func (s *GameListScreen) Create() fyne.CanvasObject {
    tree := widget.NewTree(func() []widget.TreeItem {
        items := make([]widget.TreeItem, len(s.games))
        for i, g := range s.games {
            items[i] = widget.TreeItem{
                ID:   g.ID,
                Child: s.buildGameRow(g),
            }
        }
        return items
    })
    
    return tree
}

func (s *GameListScreen) buildGameRow(game romm.Rom) *widget.TreeBranch {
    return &widget.TreeBranch{
        Items: []widget.TreeItem{
            {
                ID:   "artwork",
                Data: game.Artwork,
                Content: container.NewStack(
                    container.NewBorder(nil, nil, nil, nil, 
                        canvas.NewImageFromData(game.Artwork)),
                    container.NewBorder(nil, nil, nil, nil,
                        widget.NewLabel(game.Name)),
                ),
            },
        },
    }
}
```

### 3.3 Game Details Screen

```go
// gui/fyne/screens/game_details.go
package screens

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type GameDetailsScreen struct {
    game      romm.Rom
    platform  romm.Platform
    config    *internal.Config
}

func NewGameDetailsScreen(game romm.Rom, platform romm.Platform, config *internal.Config) *GameDetailsScreen {
    return &GameDetailsScreen{
        game:      game,
        platform:  platform,
        config:    config,
    }
}

func (s *GameDetailsScreen) Create() fyne.CanvasObject {
    form := widget.NewForm()
    
    // Cover image
    cover := container.NewBorder(nil, nil, nil, nil,
        canvas.NewImageFromData(s.game.Artwork),
    )
    
    // Metadata
    metadata := container.NewVBox(
        widget.NewLabel(s.game.Name),
        widget.NewLabel(s.game.PlatformDisplayName),
        widget.NewLabel(s.game.Summary),
    )
    
    // Download button
    downloadBtn := widget.NewButtonWithIcon(
        theme.DownloadIcon(),
        "Download",
        func() {
            // Trigger download
        },
    )
    
    return container.NewPadded(
        container.NewVBox(
            cover,
            metadata,
            container.NewHBox(downloadBtn),
        ),
    )
}
```

---

## Phase 4: Integration & Migration

### 4.1 Router Implementation

```go
// gui/fyne/router/router.go
package router

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
)

type Screen interface {
    Create() fyne.CanvasObject
}

type ScreenStack struct {
    screens    map[ScreenName]Screen
    current    Screen
    history    []Screen
}

type Router struct {
    app        fyne.App
    window     fyne.Window
    stack      *ScreenStack
    controller *controller.ControllerHandler
}

func New() *Router {
    app := app.New()
    window := app.NewWindow("Grout")
    
    return &Router{
        app:    app,
        window: window,
        stack: &ScreenStack{
            screens: make(map[ScreenName]Screen),
            history: make([]Screen, 0),
        },
        controller: controller.New(),
    }
}

func (r *Router) Navigate(screen ScreenName) {
    if s, ok := r.stack.screens[screen]; ok {
        r.stack.history = append(r.stack.history, r.stack.current)
        r.stack.current = s
        r.window.SetContent(s.Create())
    }
}
```

### 4.2 Download Manager

```go
// gui/fyne/download/manager.go
package download

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

type DownloadManager struct {
    downloads []Download
    progress  *widget.ProgressBar
    completed []string
    failed    []string
}

type Download struct {
    Name    string
    URL     string
    Progress float64
}

func New() *DownloadManager {
    return &DownloadManager{
        downloads: make([]Download, 0),
        progress:  widget.NewProgressBar(),
    }
}

func (m *DownloadManager) Start(downloads []Download) {
    m.downloads = downloads
    m.progress.Show()
    
    // Download in background
    go func() {
        for _, dl := range downloads {
            // Download file
            // Update progress
        }
        m.onComplete()
    }()
}

func (m *DownloadManager) onComplete() {
    m.progress.Hide()
    // Show completion dialog
}
```

---

## Phase 5: Testing & Polish

### 5.1 Test Checklist

- [ ] Platform selection screen works
- [ ] Game list displays artwork correctly
- [ ] Game details show all metadata
- [ ] Download manager handles progress
- [ ] Settings screens save/load correctly
- [ ] Controller navigation works
- [ ] Keyboard navigation works
- [ ] Dialogs display correctly
- [ ] Artwork caching works
- [ ] Performance is acceptable

### 5.2 Performance Targets

- Initial launch: < 2 seconds
- Screen transitions: < 100ms
- Artwork loading: < 500ms (cached)
- Memory usage: < 200MB

---

## Phase 6: Cleanup

### 6.1 Dependency Updates

**Before (go.mod):**
```go
require (
    github.com/BrandonKowalski/gabagool/v2 v2.19.0
    github.com/veandco/go-sdl2 v0.4.40 // indirect
    // ... other dependencies
)
```

**After (go.mod):**
```go
require (
    fyne.io/fyne/v2 v2.4.0
    // ... other dependencies
    // gabagool and SDL2 removed
)
```

### 6.2 Build Script Updates

**Before (flake.nix):**
```nix
buildInputs = with pkgs; [
  SDL2
  SDL2_image
  SDL2_ttf
  SDL2_gfx
];
```

**After (flake.nix):**
```nix
buildInputs = with pkgs; [
  go_1_25
];
```

---

## Migration Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Analysis | 1-2 days | None |
| Phase 2: Framework | 3-5 days | Fyne v2.4+ |
| Phase 3: Screens | 1-2 weeks | Fyne v2.4+ |
| Phase 4: Integration | 5-7 days | Fyne v2.4+ |
| Phase 5: Testing | 3-5 days | None |
| Phase 6: Cleanup | 1-2 days | None |

**Total Estimated Time: 3-4 weeks**

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| gabagool API incompatibility | High | Create adapter layer |
| Fyne learning curve | Medium | Use Fyne examples |
| Artwork rendering differences | Medium | Test thoroughly |
| Controller support gaps | Medium | Use Fyne's built-in support |
| Performance regression | Medium | Profile and optimize |

---

## Conclusion

Fyne is the ideal choice for migrating Grout's GUI. It provides:

1. **Native Linux/Flatpak support** via GTK backend
2. **Built-in controller support** via evdev
3. **Cross-platform compatibility** for future expansion
4. **Single binary distribution** - no external dependencies
5. **Active community** and good documentation

The migration will require significant code changes but will result in a more maintainable and portable codebase.
