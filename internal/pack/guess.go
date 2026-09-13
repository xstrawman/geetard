package pack

import "geetard/internal/client"

// GuessOriginal picks the best Chords artist on the search page who is not the listed cover act.
func GuessOriginal(title, coverArtist string, page client.SearchPage) string {
	_ = title
	var best client.SearchResult
	found := false
	for _, r := range page.Results {
		if !KeepResult(r) {
			continue
		}
		if MatchArtist(r.Artist, coverArtist) {
			continue
		}
		if !found || r.Rating > best.Rating || (r.Rating == best.Rating && r.Votes > best.Votes) {
			best = r
			found = true
		}
	}
	if !found {
		return ""
	}
	return rArtist(best)
}

func rArtist(r client.SearchResult) string {
	return r.Artist
}
