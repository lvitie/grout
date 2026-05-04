package gamelist

import (
	"grout/internal/fileutil"
	"os"

	internal "grout/internal"
)

const (
	GroutEntryGameListName = "Grout"
)

func AddGroutEntry(path string, groutEntryPath string) {
	internal.GetLogger().Debug("looking for correct gamelist.xml path", "path", path)
	gl := New()

	if fileutil.FileExists(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			internal.GetLogger().Debug("Error reading gamelist.xml file", "error", err)
		}

		if len(data) > 0 {
			internal.GetLogger().Debug("Found gamelist.xml file", "data", string(data))
			if err := gl.Parse(data); err != nil {
				internal.GetLogger().Debug("gamelist.xml not found or can't be parsed, skipping grout entry check", "path", path, "error", err)
				return
			}
		} else {
			internal.GetLogger().Debug("gamelist.xml file is empty", "path", path)
		}
	}

	if gl.GameContainsElements(GroutEntryGameListName, []string{
		PathElement, DescElement,
		ImageElement, DeveloperElement,
		PlayersElement, GenreElement,
	}) {
		internal.GetLogger().Debug("gamelist.xml already contains Grout entry, skipping addition", "path", path)
		return
	}

	gl.AdddOrUpdateEntry(GroutEntryGameListName, map[string]string{
		NameElement:      GroutEntryGameListName,
		DescElement:      "Download games wirelessly from your RomM instance",
		ImageElement:     "./Grout/logo.png",
		PlayersElement:   "1",
		GenreElement:     "Rom Manager",
		PathElement:      groutEntryPath,
		DeveloperElement: "The RomM Community",
	})

	if err := gl.Save(path); err != nil {
		internal.GetLogger().Debug("Unable to save gamelist.xml file", "error", err)
		return
	}

	internal.GetLogger().Debug("Successfully saved gamelist.xml file with Grout entry", "path", path)

	return
}
