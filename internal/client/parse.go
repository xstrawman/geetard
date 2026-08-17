package client

import (
	"encoding/json"
	"fmt"
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

func AsInt(v any, def int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return def
		}
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscanf(n, "%d", &i); err != nil {
			return def
		}
		return i
	default:
		return def
	}
}

func AsFloat(v any, def float64) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return def
		}
		return f
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%f", &f); err != nil {
			return def
		}
		return f
	default:
		return def
	}
}
