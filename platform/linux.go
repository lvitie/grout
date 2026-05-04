package platform

import (
	"os"
	"path/filepath"
	"strings"
)

type LinuxDesktop struct {
	configDir string
	cacheDir  string
	dataDir   string
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func NewLinuxDesktop() *LinuxDesktop {
	home, _ := os.UserHomeDir()
	return &LinuxDesktop{
		configDir: envOrDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config", "grout")),
		cacheDir:  envOrDefault("XDG_CACHE_HOME", filepath.Join(home, ".cache", "grout")),
		dataDir:   envOrDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share", "grout")),
	}
}

func (l *LinuxDesktop) Name() string {
	return "Linux Desktop"
}

func (l *LinuxDesktop) RomDirectory() string {
	return filepath.Join(l.dataDir, "roms")
}

func (l *LinuxDesktop) BaseSavePath() string {
	return filepath.Join(l.dataDir, "saves")
}

func (l *LinuxDesktop) GetSaveDirectory(fsSlug string) string {
	return filepath.Join(l.BaseSavePath(), fsSlug)
}

func (l *LinuxDesktop) EmulatorFolderMap() map[string][]string {
	// For Linux desktop, we'll probably want a way to populate this.
	// For now, return an empty map or a basic one.
	return make(map[string][]string)
}

func (l *LinuxDesktop) ConfigDir() string { return l.configDir }
func (l *LinuxDesktop) CacheDir() string  { return l.cacheDir }
func (l *LinuxDesktop) DataDir() string   { return l.dataDir }

func (l *LinuxDesktop) ScanRoms(config RomScanConfig) LocalRomScan {
	result := make(LocalRomScan)
	baseDir := l.RomDirectory()

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fsSlug := entry.Name()
		rommFSSlug := fsSlug
		if config != nil {
			rommFSSlug = config.ResolveRommFSSlug(fsSlug)
		}

		romDir := filepath.Join(baseDir, fsSlug)
		roms, err := os.ReadDir(romDir)
		if err != nil {
			continue
		}

		var files []LocalRomFile
		for _, r := range roms {
			if r.IsDir() || strings.HasPrefix(r.Name(), ".") {
				continue
			}
			files = append(files, LocalRomFile{
				FSSlug:   rommFSSlug,
				FileName: r.Name(),
				FilePath: filepath.Join(romDir, r.Name()),
			})
		}
		if len(files) > 0 {
			result[rommFSSlug] = files
		}
	}

	return result
}

func (l *LinuxDesktop) GetArtDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtPreviewDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtSplashDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtMarqueeDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtVideoDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtThumbnailDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetArtBezelDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetManualDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "manuals")
}

func (l *LinuxDesktop) GetFanartDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}

func (l *LinuxDesktop) GetBoxbackDirectory(romDir string, _, _ string) string {
	return filepath.Join(romDir, "images")
}
