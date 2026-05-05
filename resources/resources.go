package resources

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"grout/internal/i18n"
)

//go:embed locales/*.toml splash.png app.romm.Grout.png platforms/*
var embeddedFiles embed.FS

type LocaleFile struct {
	Name string
	Path string
}

var localeFiles = []LocaleFile{
	{Name: "active.en.toml", Path: "locales/active.en.toml"},
	{Name: "active.es.toml", Path: "locales/active.es.toml"},
	{Name: "active.fr.toml", Path: "locales/active.fr.toml"},
	{Name: "active.de.toml", Path: "locales/active.de.toml"},
	{Name: "active.it.toml", Path: "locales/active.it.toml"},
	{Name: "active.pt.toml", Path: "locales/active.pt.toml"},
	{Name: "active.ja.toml", Path: "locales/active.ja.toml"},
	{Name: "active.ru.toml", Path: "locales/active.ru.toml"},
}

func GetLocaleMessageFiles() ([]i18n.MessageFile, error) {
	var messageFiles []i18n.MessageFile

	for _, localeFile := range localeFiles {
		content, err := embeddedFiles.ReadFile(localeFile.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded locale file %s: %w", localeFile.Path, err)
		}

		messageFiles = append(messageFiles, i18n.MessageFile{
			Name:    localeFile.Name,
			Content: content,
		})
	}

	return messageFiles, nil
}

func GetSplashImageBytes() ([]byte, error) {
	data, err := embeddedFiles.ReadFile("splash.png")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded splash image: %w", err)
	}
	return data, nil
}

func GetAppIconBytes() ([]byte, error) {
	data, err := embeddedFiles.ReadFile("app.romm.Grout.png")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded app icon: %w", err)
	}
	return data, nil
}

var (
	platformIconCache map[string]string
	platformIconOnce  sync.Once
)

type IconCandidate struct {
	Data []byte
	Mime string
}

func GetPlatformIconCandidates(slugs ...string) []IconCandidate {
	platformIconOnce.Do(func() {
		platformIconCache = make(map[string]string)
		entries, err := embeddedFiles.ReadDir("platforms")
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					platformIconCache[strings.ToLower(entry.Name())] = entry.Name()
				}
			}
		}
	})

	exts := []struct {
		ext  string
		mime string
	}{
		{".svg", "image/svg+xml"},
		{".png", "image/png"},
		{".ico", "image/x-icon"},
	}

	var candidates []IconCandidate
	seen := make(map[string]bool)

	for _, slug := range slugs {
		if slug == "" {
			continue
		}

		for _, e := range exts {
			target := strings.ToLower(slug + e.ext)
			if actualName, ok := platformIconCache[target]; ok {
				if seen[actualName] {
					continue
				}
				data, err := embeddedFiles.ReadFile("platforms/" + actualName)
				if err == nil {
					candidates = append(candidates, IconCandidate{
						Data: data,
						Mime: e.mime,
					})
					seen[actualName] = true
				}
			}
		}
	}

	// If no icons were found, use default as fallback
	if len(candidates) == 0 {
		if data, err := embeddedFiles.ReadFile("platforms/default.ico"); err == nil {
			candidates = append(candidates, IconCandidate{
				Data: data,
				Mime: "image/x-icon",
			})
		}
	}

	return candidates
}

// Deprecated: Use GetPlatformIconCandidates instead.
func GetPlatformIconBytes(slugs ...string) ([]byte, string, error) {
	candidates := GetPlatformIconCandidates(slugs...)
	if len(candidates) > 0 {
		return candidates[0].Data, candidates[0].Mime, nil
	}
	return nil, "", fmt.Errorf("icon not found for slugs: %v", slugs)
}
