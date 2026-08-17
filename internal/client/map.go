package client

import "strings"

func MapSearch(store map[string]any, query string, page int) (SearchPage, error) {
	data, err := storeData(store)
	if err != nil {
		return SearchPage{}, err
	}
	raw, _ := data["results"].([]any)
	var results []SearchResult
	for _, item := range raw {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		typ, _ := m["type"].(string)
		if strings.EqualFold(typ, "Pro") || strings.EqualFold(typ, "Official") {
			continue
		}
		tabURL, _ := m["tab_url"].(string)
		path := TabPath(tabURL)
		if path == "" {
			continue
		}
		artist, _ := m["artist_name"].(string)
		song, _ := m["song_name"].(string)
		if artist == "" {
			artist = "Unknown artist"
		}
		if song == "" {
			song = "Unknown song"
		}
		results = append(results, SearchResult{
			Artist:  artist,
			Song:    song,
			Type:    typ,
			Version: AsInt(m["version"], 1),
			Rating:  AsFloat(m["rating"], 0),
			Votes:   AsInt(m["votes"], 0),
			Path:    path,
		})
	}
	pag, _ := data["pagination"].(map[string]any)
	total := 1
	cur := page
	if pag != nil {
		total = AsInt(pag["total"], 1)
		cur = AsInt(pag["current"], page)
	}
	return SearchPage{Query: query, Page: cur, TotalPages: total, Results: results}, nil
}
