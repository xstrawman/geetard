package pack

import (
	"regexp"
	"strings"

	"geetard/internal/client"
)

var (
	halfDownRE = regexp.MustCompile(`(?i)1\s*/\s*2\s*step\s*down|half[-\s]?step\s*down|semitone\s*down|tune\s*down\s*(a\s*)?half|eb\s*tuning|tuning:\s*eb\b`)
	chRE       = regexp.MustCompile(`(?i)\[ch\]([^[]+)\[/ch\]|<strong>([A-G][#b]?(?:m|maj|min|sus|dim|aug|\d)*)`)
)

func KeepResult(r client.SearchResult) bool {
	t := strings.ToLower(r.Type)
	if t == "" {
		return false
	}
	if strings.Contains(t, "ukulele") || strings.Contains(t, "bass") || strings.Contains(t, "drum") {
		return false
	}
	if strings.Contains(t, "official") || strings.Contains(t, "guitar pro") {
		return false
	}
	if t == "pro" || strings.HasPrefix(t, "pro ") {
		return false
	}
	if strings.Contains(t, "tab") && !strings.Contains(t, "chord") {
		return false
	}
	return strings.Contains(t, "chord")
}

func KeepTab(tab client.TabDetail) bool {
	return KeepResult(client.SearchResult{Type: tab.Type})
}

func HalfDown(tab client.TabDetail) bool {
	blob := tab.Tuning + "\n" + tab.Capo + "\n" + tab.Raw
	return halfDownRE.MatchString(blob)
}

func concertPitchRespell(tab client.TabDetail) bool {
	if HalfDown(tab) {
		return false
	}
	var b, c int
	for _, m := range chRE.FindAllStringSubmatch(tab.Raw, -1) {
		name := m[1]
		if name == "" {
			name = m[2]
		}
		name = strings.ToLower(name)
		if strings.HasPrefix(name, "b") && !strings.HasPrefix(name, "bb") {
			b++
		}
		if strings.HasPrefix(name, "c") {
			c++
		}
	}
	return b >= 3 && b > c
}

func Score(tab client.TabDetail) float64 {
	s := tab.Rating
	if s <= 0 {
		s = 3
	}
	if HalfDown(tab) {
		s += 2
	}
	if concertPitchRespell(tab) {
		s -= 2
	}
	if tab.Version > 8 {
		s -= 0.2
	}
	return s
}

func Pick(tabs []client.TabDetail) (client.TabDetail, bool) {
	var best client.TabDetail
	bestS := -1e9
	found := false
	for _, tab := range tabs {
		if !KeepTab(tab) {
			continue
		}
		if strings.TrimSpace(tab.Raw) == "" {
			continue
		}
		sc := Score(tab)
		if !found || sc > bestS {
			best, bestS, found = tab, sc, true
		}
	}
	return best, found
}

func MatchArtist(got, want string) bool {
	if want == "" {
		return true
	}
	g, w := Artist(got), Artist(want)
	return g == w || strings.Contains(g, w) || strings.Contains(w, g)
}

func FilterResults(results []client.SearchResult, originalArtist, coverArtist string) []client.SearchResult {
	var out []client.SearchResult
	for _, r := range results {
		if !KeepResult(r) {
			continue
		}
		if originalArtist != "" && MatchArtist(r.Artist, originalArtist) {
			out = append(out, r)
			continue
		}
		if originalArtist == "" && coverArtist != "" && MatchArtist(r.Artist, coverArtist) {
			out = append(out, r)
		}
	}
	if len(out) > 0 {
		return out
	}
	// B: cover artist if original pool empty
	for _, r := range results {
		if !KeepResult(r) {
			continue
		}
		if coverArtist != "" && MatchArtist(r.Artist, coverArtist) {
			out = append(out, r)
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, r := range results {
		if KeepResult(r) {
			out = append(out, r)
		}
	}
	return out
}
