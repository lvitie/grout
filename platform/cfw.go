package platform

import (
	"grout/cfw"
)

type CFWPlatform struct{}

func NewCFWPlatform() *CFWPlatform {
	return &CFWPlatform{}
}

func (p *CFWPlatform) Name() string {
	return string(cfw.GetCFW())
}

func (p *CFWPlatform) RomDirectory() string {
	return cfw.GetRomDirectory()
}

func (p *CFWPlatform) BaseSavePath() string {
	return cfw.BaseSavePath()
}

func (p *CFWPlatform) GetSaveDirectory(fsSlug string) string {
	return cfw.GetSaveDirectory(fsSlug)
}

func (p *CFWPlatform) EmulatorFolderMap() map[string][]string {
	return cfw.EmulatorFolderMap(cfw.GetCFW())
}

func (p *CFWPlatform) ConfigDir() string {
	return "."
}

func (p *CFWPlatform) CacheDir() string {
	return ".cache"
}

func (p *CFWPlatform) DataDir() string {
	return "."
}

func (p *CFWPlatform) ScanRoms(config RomScanConfig) LocalRomScan {
	return cfw.ScanRoms(config)
}

func (p *CFWPlatform) GetArtDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtPreviewDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtPreviewDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtSplashDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtSplashDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtMarqueeDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtMarqueeDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtVideoDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtVideoDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtThumbnailDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtThumbnailDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetArtBezelDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetArtBezelDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetManualDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetManualDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetFanartDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetFanartDirectory(romDir, platformFSSlug, platformName)
}

func (p *CFWPlatform) GetBoxbackDirectory(romDir string, platformFSSlug, platformName string) string {
	return cfw.GetBoxbackDirectory(romDir, platformFSSlug, platformName)
}
