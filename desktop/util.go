package desktop

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// EscapeMarkup escapes text for use in Pango markup (GTK labels).
func EscapeMarkup(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// NormalizeSearch returns a lowercase, diacritic-stripped version of s
// for accent-insensitive search (e.g. "pokémon" → "pokemon").
func NormalizeSearch(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(s)) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
