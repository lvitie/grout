package i18n

type MessageFile struct {
	Name    string
	Content []byte
}

// Placeholder for now, will be expanded during Phase 3/7
func Localize(id string, defaultMsg string) string {
	return defaultMsg
}
