package client

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
)

var storeDiv = regexp.MustCompile(`(?is)<div[^>]*class="[^"]*\bjs-store\b[^"]*"[^>]*data-content="([^"]*)"`)

func ExtractStore(htmlPage string) (map[string]any, error) {
	m := storeDiv.FindStringSubmatch(htmlPage)
	if m == nil {
		return nil, errf("page layout changed")
	}
	raw := html.UnescapeString(m[1])
	if strings.TrimSpace(raw) == "" {
		return nil, errf("page layout changed")
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, errf("could not parse page")
	}
	return out, nil
}

func storeData(store map[string]any) (map[string]any, error) {
	s, _ := store["store"].(map[string]any)
	if s == nil {
		return nil, errf("could not parse page")
	}
	page, _ := s["page"].(map[string]any)
	if page == nil {
		return nil, errf("could not parse page")
	}
	data, _ := page["data"].(map[string]any)
	if data == nil {
		return nil, errf("could not parse page")
	}
	return data, nil
}
