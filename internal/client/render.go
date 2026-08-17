package client

import (
	"regexp"
	"strings"
)

var chordRE = regexp.MustCompile(`\[ch\](.*?)\[/ch\]`)

func Render(raw string) []DisplayLine {
	s := strings.ReplaceAll(raw, "[tab]", "")
	s = strings.ReplaceAll(s, "[/tab]", "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	var lines []DisplayLine
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, DisplayLine{Parts: splitParts(line)})
	}
	if len(lines) == 1 && len(lines[0].Parts) == 0 {
		return nil
	}
	return lines
}

func splitParts(s string) []Part {
	var parts []Part
	last := 0
	for _, m := range chordRE.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > last {
			parts = append(parts, Part{Text: s[last:m[0]], Kind: KindBody})
		}
		inner := s[m[2]:m[3]]
		if inner != "" {
			parts = append(parts, Part{Text: inner, Kind: KindChord})
		}
		last = m[1]
	}
	if last < len(s) {
		parts = append(parts, Part{Text: s[last:], Kind: KindBody})
	}
	return parts
}
