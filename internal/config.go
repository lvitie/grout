package internal

import (
	"encoding/json"
	"fmt"
	"grout/internal/artutil"
	"grout/platform"
	"grout/romm"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var kidModeEnabled atomic.Bool

type SaveLayout string

const (
	SaveLayoutRomM      SaveLayout = "" // default, current behavior
	SaveLayoutEmuDeck   SaveLayout = "emudeck"
	SaveLayoutRetroDeck SaveLayout = "retrodeck"
	SaveLayoutRetroArch SaveLayout = "retroarch"
)

type AdditionalDownloads struct {
	Marquee   artutil.ArtKind `json:"marquee,omitempty"`
	Video     bool            `json:"video,omitempty"`
	Thumbnail artutil.ArtKind `json:"thumbnail,omitempty"`
	Bezel     bool            `json:"bezel,omitempty"`
	Manual    bool            `json:"manual,omitempty"`
	BoxBack   bool            `json:"box_back,omitempty"`
	Fanart    bool            `json:"fanart,omitempty"`
}

// DurationSeconds is a time.Duration that marshals to/from JSON as whole seconds.
// Existing configs with nanosecond values are handled by detecting large values on unmarshal.
type DurationSeconds time.Duration

func (d DurationSeconds) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(time.Duration(d).Seconds()))
}

func (d *DurationSeconds) UnmarshalJSON(b []byte) error {
	var raw int64
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	// Values over 1,000,000 are nanoseconds from old configs (e.g. 1800000000000 = 30min).
	// Convert them to the equivalent duration directly.
	if raw > 1_000_000 {
		*d = DurationSeconds(time.Duration(raw))
	} else {
		*d = DurationSeconds(time.Duration(raw) * time.Second)
	}
	return nil
}

func (d DurationSeconds) Duration() time.Duration {
	return time.Duration(d)
}

type Config struct {
	Hosts                        []romm.Host                 `json:"hosts,omitempty"`
	DirectoryMappings            map[string]DirectoryMapping `json:"directory_mappings,omitempty"`
	DownloadArt                  bool                        `json:"download_art,omitempty"`
	ShowBoxArt                   bool                        `json:"show_box_art,omitempty"`
	UnzipDownloads               bool                        `json:"unzip_downloads,omitempty"`
	ShowRegularCollections       bool                        `json:"show_collections"`
	ShowSmartCollections         bool                        `json:"show_smart_collections"`
	ShowVirtualCollections       bool                        `json:"show_virtual_collections"`
	DownloadedGames              DownloadedGamesMode         `json:"downloaded_games,omitempty"`
	ApiTimeout                   DurationSeconds             `json:"api_timeout"`
	DownloadTimeout              DurationSeconds             `json:"download_timeout"`
	LogLevel                     LogLevel                    `json:"log_level,omitempty"`
	Language                     string                      `json:"language,omitempty"`
	CollectionView               CollectionView              `json:"collection_view,omitempty"`
	KidMode                      bool                        `json:"kid_mode,omitempty"`
	ReleaseChannel               ReleaseChannel              `json:"release_channel,omitempty"`
	ArtKind                      artutil.ArtKind             `json:"art_kind,omitempty"`
	DownloadArtScreenshotPreview bool                        `json:"download_art_screenshot_preview,omitempty"`
	DownloadSplashArt            artutil.ArtKind             `json:"download_splash_art,omitempty"`
	AdditionalDownloads          AdditionalDownloads         `json:"additional_downloads,omitempty"`

	SwapFaceButtons       bool              `json:"swap_face_buttons,omitempty"`
	PlatformOrder         []string          `json:"platform_order,omitempty"`
	SaveDirectoryMappings map[string]string `json:"save_directory_mappings,omitempty"`
	SaveLayout            SaveLayout        `json:"save_layout,omitempty"`
	SaveBasePath          string            `json:"save_base_path,omitempty"`
	RomLayout             SaveLayout        `json:"rom_layout,omitempty"`
	RomBasePath           string            `json:"rom_base_path,omitempty"`
	SlotPreferences       map[string]string `json:"-"`                           // Stored in save_slots.json, not config.json
	SaveBackupLimit       int               `json:"save_backup_limit,omitempty"` // 0 = no limit, 5/10/15 = keep N most recent per game

	PlatformsBinding map[string]string `json:"-"`

	CloseToTray *bool `json:"close_to_tray,omitempty"`
}

type DirectoryMapping struct {
	RomMSlug     string `json:"slug"`
	RelativePath string `json:"relative_path"`
}

func (c Config) ShouldCloseToTray() bool {
	if c.CloseToTray == nil {
		return true // default: enabled (opt-out)
	}
	return *c.CloseToTray
}

func (c Config) ToLoggable() any {
	safeHosts := make([]map[string]any, len(c.Hosts))
	for i, host := range c.Hosts {
		safeHosts[i] = host.ToLoggable()
	}

	return map[string]any{
		"hosts":                   safeHosts,
		"directory_mappings":      c.DirectoryMappings,
		"api_timeout":             c.ApiTimeout,
		"download_timeout":        c.DownloadTimeout,
		"unzip_downloads":         c.UnzipDownloads,
		"download_art":            c.DownloadArt,
		"art_kind":                c.ArtKind,
		"show_box_art":            c.ShowBoxArt,
		"collections":             c.ShowRegularCollections,
		"smart_collections":       c.ShowSmartCollections,
		"virtual_collections":     c.ShowVirtualCollections,
		"downloaded_games_action": c.DownloadedGames,
		"log_level":               c.LogLevel,
	}
}

func LoadConfig() (*Config, error) {
	path := filepath.Join(platform.GetCurrent().ConfigDir(), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if config.ApiTimeout == 0 {
		config.ApiTimeout = DurationSeconds(30 * time.Second)
	}

	if config.DownloadTimeout == 0 {
		config.DownloadTimeout = DurationSeconds(60 * time.Minute)
	}

	// Clamp API timeout to valid picker range (15s–300s)
	if config.ApiTimeout.Duration() > 300*time.Second {
		config.ApiTimeout = DurationSeconds(30 * time.Second)
	}

	if config.Language == "" {
		config.Language = "en"
	}

	if config.DownloadedGames == "" {
		config.DownloadedGames = DownloadedGamesModeDoNothing
	}

	if config.CollectionView == "" {
		config.CollectionView = CollectionViewPlatform
	}

	if config.ArtKind == "" {
		config.ArtKind = artutil.ArtKindDefault
	}

	if config.AdditionalDownloads.Thumbnail == "" {
		config.AdditionalDownloads.Thumbnail = artutil.ArtKindNone
	}

	if config.AdditionalDownloads.Marquee == "" {
		config.AdditionalDownloads.Marquee = artutil.ArtKindNone
	}

	// Load slot preferences from dedicated file
	config.SlotPreferences = LoadSlotPreferences()

	return &config, nil
}

func SaveConfig(config *Config) error {
	if config.LogLevel == "" {
		config.LogLevel = LogLevelError
	}

	if config.Language == "" {
		config.Language = "en"
	}

	if config.DownloadedGames == "" {
		config.DownloadedGames = DownloadedGamesModeDoNothing
	}

	if config.CollectionView == "" {
		config.CollectionView = CollectionViewPlatform
	}

	if config.ReleaseChannel == "" {
		config.ReleaseChannel = ReleaseChannelMatchRomM
	}

	if config.ArtKind == "" {
		config.ArtKind = artutil.ArtKindDefault
	}

	if config.AdditionalDownloads.Thumbnail == "" {
		config.AdditionalDownloads.Thumbnail = artutil.ArtKindNone
	}

	if config.AdditionalDownloads.Marquee == "" {
		config.AdditionalDownloads.Marquee = artutil.ArtKindNone
	}

	// For now, I'll just ignore it or implement it in internal/logger.go later.

	/*
		if err := i18n.SetWithCode(config.Language); err != nil {
			GetLogger().Error("Failed to set language", "error", err, "language", config.Language)
		}
	*/
	// For now, we don't have a SetWithCode in our minimal wrapper.
	// We will implement this in Phase 7.

	pretty, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		GetLogger().Error("Failed to marshal config to JSON", "error", err)
		return err
	}

	path := filepath.Join(platform.GetCurrent().ConfigDir(), "config.json")
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		GetLogger().Error("Failed to create config directory", "error", err, "path", filepath.Dir(path))
		return err
	}

	if err := os.WriteFile(path, pretty, 0644); err != nil {
		GetLogger().Error("Failed to write config file", "error", err, "path", path)
		return err
	}

	return nil
}

func InitKidMode(config *Config) {
	kidModeEnabled.Store(config.KidMode)
}

func IsKidModeEnabled() bool {
	return kidModeEnabled.Load()
}

func SetKidMode(enabled bool) {
	kidModeEnabled.Store(enabled)
}

// LoadPlatformsBinding fetches the PLATFORMS_BINDING from the RomM server
// and stores it in the config for use in CFW lookups.
// This requires the pointer receiver!
//
//goland:noinspection ALL
func (c *Config) LoadPlatformsBinding(host romm.Host, timeout ...time.Duration) error {
	client := romm.NewClientFromHost(host, timeout...)

	rommConfig, err := client.GetConfig()
	if err != nil {
		// Non-fatal - older RomM versions may not have this endpoint
		return err
	}

	c.PlatformsBinding = rommConfig.PlatformsBinding
	return nil
}

func (c Config) GetDirectoryMapping(fsSlug string) (string, bool) {
	if mapping, ok := c.DirectoryMappings[fsSlug]; ok {
		return mapping.RelativePath, true
	}
	return "", false
}

func LoadSlotPreferences() map[string]string {
	path := filepath.Join(platform.GetCurrent().ConfigDir(), "save_slots.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var prefs map[string]string
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil
	}
	return prefs
}

func SaveSlotPreferences(config *Config) error {
	path := filepath.Join(platform.GetCurrent().ConfigDir(), "save_slots.json")
	if len(config.SlotPreferences) == 0 {
		os.Remove(path)
		return nil
	}
	pretty, err := json.MarshalIndent(config.SlotPreferences, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, pretty, 0644)
}

func (c Config) GetSlotPreference(romID int) string {
	if c.SlotPreferences != nil {
		if slot, ok := c.SlotPreferences[fmt.Sprintf("%d", romID)]; ok {
			return slot
		}
	}
	return "default"
}

func (c *Config) SetSlotPreference(romID int, slot string) {
	if c.SlotPreferences == nil {
		c.SlotPreferences = make(map[string]string)
	}
	key := fmt.Sprintf("%d", romID)
	if slot == "default" {
		delete(c.SlotPreferences, key)
	} else {
		c.SlotPreferences[key] = slot
	}
}

func (c Config) GetApiTimeout() time.Duration    { return c.ApiTimeout.Duration() }
func (c Config) GetShowCollections() bool        { return c.ShowRegularCollections }
func (c Config) GetShowSmartCollections() bool   { return c.ShowSmartCollections }
func (c Config) GetShowVirtualCollections() bool { return c.ShowVirtualCollections }

// ResolveFSSlug returns the effective fs_slug for CFW lookups.
// If the fs_slug has a binding in PlatformsBinding, the bound value is returned.
// Otherwise, the original fs_slug is returned.
// Example: PlatformsBinding {"ms": "sms"} means RomM "ms" -> CFW "sms"
// So ResolveFSSlug("ms") returns "sms"
func (c Config) ResolveFSSlug(fsSlug string) string {
	if c.PlatformsBinding != nil {
		if bound, ok := c.PlatformsBinding[fsSlug]; ok {
			GetLogger().Debug("Using platform binding for CFW lookup",
				"fsSlug", fsSlug, "boundTo", bound)
			return bound
		}
	}
	return fsSlug
}

// ResolveRommFSSlug returns the RomM fs_slug for a given CFW platform key.
// This is the inverse of ResolveFSSlug - it finds which RomM fs_slug maps TO the given CFW key.
// Example: PlatformsBinding {"ms": "sms"} means RomM "ms" -> CFW "sms"
// So ResolveRommFSSlug("sms") returns "ms"
func (c Config) ResolveRommFSSlug(cfwKey string) string {
	if c.PlatformsBinding != nil {
		for rommSlug, cfwSlug := range c.PlatformsBinding {
			if cfwSlug == cfwKey {
				GetLogger().Debug("Using inverse platform binding",
					"cfwKey", cfwKey, "rommFSSlug", rommSlug)
				return rommSlug
			}
		}
	}
	return cfwKey
}

func (c Config) GetPlatformRomDirectory(pi romm.Platform) string {
	if mapping, ok := c.DirectoryMappings[pi.FSSlug]; ok && mapping.RelativePath != "" {
		return filepath.Join(c.GetRomBasePath(), mapping.RelativePath)
	}

	subfolder := mapFSSlugToLayout(pi.FSSlug, c.RomLayout)
	return filepath.Join(c.GetRomBasePath(), subfolder)
}

func (c Config) GetArtDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtPreviewDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtPreviewDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtSplashDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtSplashDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtMarqueeDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtMarqueeDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtVideoDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtVideoDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtThumbnailDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtThumbnailDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetArtBezelDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetArtBezelDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetManualDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetManualDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetFanartDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetFanartDirectory(romDir, pi.FSSlug, pi.Name)
}

func (c Config) GetBoxbackDirectory(pi romm.Platform) string {
	romDir := c.GetPlatformRomDirectory(pi)
	return platform.GetCurrent().GetBoxbackDirectory(romDir, pi.FSSlug, pi.Name)
}

// GetRomBasePath returns the base path for ROMs.
func (c Config) GetRomBasePath() string {
	if c.RomBasePath != "" {
		expanded, err := expandTilde(c.RomBasePath)
		if err == nil {
			return expanded
		}
		return c.RomBasePath
	}
	return GetDefaultRomBasePath(c.RomLayout)
}

func GetDefaultRomBasePath(layout SaveLayout) string {
	home, _ := os.UserHomeDir()

	switch layout {
	case SaveLayoutEmuDeck:
		return filepath.Join(home, "Emulation", "roms")
	case SaveLayoutRetroDeck:
		return filepath.Join(home, ".var", "app", "net.retrodeck.retrodeck", "data", "roms")
	case SaveLayoutRetroArch:
		return filepath.Join(home, ".config", "retroarch", "roms")
	default:
		return platform.GetCurrent().RomDirectory()
	}
}

// GetSaveBasePath returns the base path for saves.
// If SaveBasePath is set, it's returned after expanding ~.
// Otherwise, the default for the current layout is returned.
func (c Config) GetSaveBasePath() string {
	if c.SaveBasePath != "" {
		expanded, err := expandTilde(c.SaveBasePath)
		if err == nil {
			return expanded
		}
		return c.SaveBasePath
	}
	return GetDefaultSaveBasePath(c.SaveLayout)
}

// GetDefaultSaveBasePath returns the conventional default base path for a given layout.
func GetDefaultSaveBasePath(layout SaveLayout) string {
	home, _ := os.UserHomeDir()

	switch layout {
	case SaveLayoutEmuDeck:
		return filepath.Join(home, "Emulation", "saves")
	case SaveLayoutRetroDeck:
		return filepath.Join(home, ".var", "app", "net.retrodeck.retrodeck", "data", "saves")
	case SaveLayoutRetroArch:
		return filepath.Join(home, ".config", "retroarch", "saves")
	default:
		// RomM layout or empty: use platform default
		return platform.GetCurrent().BaseSavePath()
	}
}

// expandTilde expands ~ to the user's home directory.
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

// mapFSSlugToLayout maps a RomM fs_slug to the appropriate folder name for the given layout.
// If the slug is not in the map, the slug itself is returned as the folder name.
func mapFSSlugToLayout(fsSlug string, layout SaveLayout) string {
	var folderMap map[string]string

	switch layout {
	case SaveLayoutEmuDeck:
		folderMap = emuDeckFolderMap
	case SaveLayoutRetroArch:
		folderMap = retroArchFolderMap
	case SaveLayoutRetroDeck:
		folderMap = retroDeckFolderMap
	default:
		// RomM layout uses fsSlug directly
		return fsSlug
	}

	if mapped, ok := folderMap[fsSlug]; ok {
		return mapped
	}
	return fsSlug
}

// Folder name mappings for different layouts
var emuDeckFolderMap = map[string]string{
	"snes":       "snes",
	"nes":        "nes",
	"gba":        "gba",
	"gbc":        "gbc",
	"gb":         "gb",
	"n64":        "n64",
	"nds":        "nds",
	"3ds":        "3ds",
	"gc":         "gc",
	"wii":        "wii",
	"wiiu":       "wiiu",
	"switch":     "switch",
	"psx":        "psx",
	"ps2":        "ps2",
	"psp":        "psp",
	"genesis":    "genesis",
	"megadrive":  "genesis",
	"sms":        "mastersystem",
	"ms":         "mastersystem",
	"gg":         "gamegear",
	"saturn":     "saturn",
	"dreamcast":  "dreamcast",
	"dc":         "dreamcast",
	"segacd":     "segacd",
	"sega32x":    "sega32x",
	"atari2600":  "atari2600",
	"atari7800":  "atari7800",
	"lynx":       "atarilynx",
	"jaguar":     "atarijaguar",
	"ngp":        "ngp",
	"ngpc":       "ngpc",
	"neogeo":     "neogeo",
	"arcade":     "arcade",
	"mame":       "mame",
	"pcengine":   "pcengine",
	"tg16":       "tg16",
	"wonderswan": "wonderswan",
	"wsc":        "wonderswancolor",
}

var retroArchFolderMap = map[string]string{
	"snes":       "Nintendo - Super Nintendo Entertainment System",
	"nes":        "Nintendo - Nintendo Entertainment System",
	"gba":        "Nintendo - Game Boy Advance",
	"gbc":        "Nintendo - Game Boy Color",
	"gb":         "Nintendo - Game Boy",
	"n64":        "Nintendo - Nintendo 64",
	"nds":        "Nintendo - Nintendo DS",
	"gc":         "Nintendo - GameCube",
	"genesis":    "Sega - Mega Drive - Genesis",
	"megadrive":  "Sega - Mega Drive - Genesis",
	"sms":        "Sega - Master System - Mark III",
	"ms":         "Sega - Master System - Mark III",
	"gg":         "Sega - Game Gear",
	"saturn":     "Sega - Saturn",
	"dreamcast":  "Sega - Dreamcast",
	"dc":         "Sega - Dreamcast",
	"psx":        "Sony - PlayStation",
	"ps2":        "Sony - PlayStation 2",
	"psp":        "Sony - PlayStation Portable",
	"atari2600":  "Atari - 2600",
	"atari7800":  "Atari - 7800",
	"lynx":       "Atari - Lynx",
	"neogeo":     "SNK - Neo Geo",
	"arcade":     "MAME",
	"mame":       "MAME",
	"pcengine":   "NEC - PC Engine - TurboGrafx 16",
	"tg16":       "NEC - PC Engine - TurboGrafx 16",
	"wonderswan": "Bandai - WonderSwan",
	"wsc":        "Bandai - WonderSwan Color",
	"ngp":        "SNK - Neo Geo Pocket",
	"ngpc":       "SNK - Neo Geo Pocket Color",
}

// RetroDeck uses the same as RetroArch system folder names
var retroDeckFolderMap = retroArchFolderMap

// GetEffectiveSaveDirectory returns the effective save directory for a given fs_slug.
// It considers SaveDirectoryMappings (per-platform override), then SaveLayout and SaveBasePath,
// then falls back to the platform default.
func (c Config) GetEffectiveSaveDirectory(fsSlug string) string {
	// Check if there's an explicit mapping for this slug
	if c.SaveDirectoryMappings != nil {
		if mapped, ok := c.SaveDirectoryMappings[fsSlug]; ok && mapped != "" {
			baseSavePath := platform.GetCurrent().BaseSavePath()
			if baseSavePath != "" {
				return filepath.Join(baseSavePath, mapped)
			}
		}
	}

	// Use the layout-based resolution
	basePath := c.GetSaveBasePath()
	subfolder := mapFSSlugToLayout(fsSlug, c.SaveLayout)
	return filepath.Join(basePath, subfolder)
}
