package client

import (
	"strconv"
	"strings"
)

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

func MapTab(store map[string]any) (TabDetail, error) {
	data, err := storeData(store)
	if err != nil {
		return TabDetail{}, err
	}
	tab, _ := data["tab"].(map[string]any)
	view, _ := data["tab_view"].(map[string]any)
	if tab == nil || view == nil {
		return TabDetail{}, errf("could not parse page")
	}
	wiki, _ := view["wiki_tab"].(map[string]any)
	raw, _ := wiki["content"].(string)
	artist, _ := tab["artist_name"].(string)
	song, _ := tab["song_name"].(string)
	typ, _ := tab["type"].(string)
	diff, _ := view["ug_difficulty"].(string)
	tuning := ""
	if meta, ok := view["meta"].(map[string]any); ok {
		if tun, ok := meta["tuning"].(map[string]any); ok {
			val, _ := tun["value"].(string)
			name, _ := tun["name"].(string)
			switch {
			case val != "" && name != "":
				tuning = val + " (" + name + ")"
			case val != "":
				tuning = val
			default:
				tuning = name
			}
		}
	}
	if artist == "" {
		artist = "Unknown artist"
	}
	if song == "" {
		song = "Unknown song"
	}
	return TabDetail{
		Artist:     artist,
		Song:       song,
		Version:    AsInt(tab["version"], 1),
		Type:       typ,
		Rating:     AsFloat(tab["rating"], 0),
		Capo:       formatCapo(view),
		Tuning:     tuning,
		Difficulty: diff,
		Raw:        raw,
		Shapes:     Shapes(view),
	}, nil
}

func formatCapo(view map[string]any) string {
	meta, ok := view["meta"].(map[string]any)
	if !ok {
		return ""
	}
	v, exists := meta["capo"]
	if !exists || v == nil {
		return ""
	}
	if n, ok := v.(string); ok {
		return n
	}
	return strconv.Itoa(AsInt(v, 0))
}
