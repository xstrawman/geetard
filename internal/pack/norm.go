package pack

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	verParen  = regexp.MustCompile(`(?i)\s*\(ver(?:sion)?\s*\d+\)`)
	liveParen = regexp.MustCompile(`(?i)\s*\([^)]*live[^)]*\)`)
	featParen = regexp.MustCompile(`(?i)\s*\((?:feat\.?|ft\.?|featuring)[^)]*\)`)
	featTail  = regexp.MustCompile(`(?i)\s+(?:feat\.?|ft\.?|featuring)\s+.+$`)
	thePrefix = regexp.MustCompile(`(?i)^the\s+`)
)

func fold(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "‘", "'")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func Title(s string) string {
	s = fold(s)
	s = verParen.ReplaceAllString(s, "")
	s = liveParen.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func Artist(s string) string {
	s = fold(s)
	s = featParen.ReplaceAllString(s, "")
	s = featTail.ReplaceAllString(s, "")
	s = thePrefix.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func Key(artist, title string) string {
	return Artist(artist) + "|" + Title(title)
}
