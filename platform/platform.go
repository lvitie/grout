package platform

// Platform provides OS-specific paths and behaviors.
type Platform interface {
	Name() string
	RomDirectory() string
	BaseSavePath() string
	GetSaveDirectory(fsSlug string) string
	EmulatorFolderMap() map[string][]string
	ConfigDir() string // XDG_CONFIG_HOME/grout
	CacheDir() string  // XDG_CACHE_HOME/grout
	DataDir() string   // XDG_DATA_HOME/grout
	ScanRoms(config RomScanConfig) LocalRomScan

	GetArtDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtPreviewDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtSplashDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtMarqueeDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtVideoDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtThumbnailDirectory(romDir string, platformFSSlug, platformName string) string
	GetArtBezelDirectory(romDir string, platformFSSlug, platformName string) string
	GetManualDirectory(romDir string, platformFSSlug, platformName string) string
	GetFanartDirectory(romDir string, platformFSSlug, platformName string) string
	GetBoxbackDirectory(romDir string, platformFSSlug, platformName string) string
}

// RomScanConfig provides configuration needed for ROM scanning.
// Implemented by internal.Config to avoid circular imports.
type RomScanConfig interface {
	GetDirectoryMapping(fsSlug string) (relativePath string, ok bool)
	ResolveRommFSSlug(cfwKey string) string
}

type LocalRomFile struct {
	RomID    int
	RomName  string
	FSSlug   string
	FileName string
	FilePath string
}

type LocalRomScan map[string][]LocalRomFile

var current Platform

func GetCurrent() Platform {
	return current
}

func SetCurrent(p Platform) {
	current = p
}
