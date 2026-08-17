package client

import (
	"html"
	"regexp"
	"strings"
)

var (
	favRE = regexp.MustCompile(
		`data-artist="([^"]*)"\s+data-song="([^"]*)"\s+data-type="([^"]*)"\s+data-rating="([^"]*)"\s+data-url="([^"]*)"`,
	)
	pageRE     = regexp.MustCompile(`[?&]page=(\d+)`)
	tabBodyRE   = regexp.MustCompile(`(?is)<div class="tab font-monospace">(.*?)</div>`)
	chordPartRE = regexp.MustCompile(`(?is)<span class="chord-(?:root|quality|bass)"[^>]*>(.*?)</span>`)
	chordBoxRE  = regexp.MustCompile(`(?is)<span class="chord[^"]*"[^>]*>(.*?)</span>`)
	tagRE       = regexp.MustCompile(`(?is)<[^>]+>`)
	brRE       = regexp.MustCompile(`(?i)<br\s*/?>`)
	h5RE       = regexp.MustCompile(`(?is)<h5>(.*?)</h5>`)
	diffRE     = regexp.MustCompile(`(?i)Difficulty:\s*([^<]+)`)
	capoRE     = regexp.MustCompile(`(?i)Capo:\s*([^<]+)`)
	tunRE      = regexp.MustCompile(`(?i)Tuning:\s*([^<]+)`)
	oopsRE     = regexp.MustCompile(`(?i)Oops:`)
	verRE      = regexp.MustCompile(`(?i)\(ver\s+(\d+)\)`)
)

func ParseFreetarSearch(htmlPage, query string, page int) (SearchPage, error) {
	if oopsRE.MatchString(htmlPage) && !strings.Contains(htmlPage, `id="results"`) {
		return SearchPage{}, errf("could not parse page")
	}
	matches := favRE.FindAllStringSubmatch(htmlPage, -1)
	var results []SearchResult
	seen := map[string]bool{}
	for _, m := range matches {
		path := TabPath(html.UnescapeString(m[5]))
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		typ := html.UnescapeString(m[3])
		if strings.EqualFold(typ, "Pro") || strings.EqualFold(typ, "Official") {
			continue
		}
		song := html.UnescapeString(m[2])
		ver := 1
		if vm := verRE.FindStringSubmatch(song); vm != nil {
			ver = AsInt(vm[1], 1)
		}
		results = append(results, SearchResult{
			Artist:  html.UnescapeString(m[1]),
			Song:    strings.TrimSpace(verRE.ReplaceAllString(song, "")),
			Type:    typ,
			Version: ver,
			Rating:  AsFloat(m[4], 0),
			Path:    path,
		})
	}
	total := 1
	for _, pm := range pageRE.FindAllStringSubmatch(htmlPage, -1) {
		if n := AsInt(pm[1], 1); n > total {
			total = n
		}
	}
	if page < 1 {
		page = 1
	}
	return SearchPage{Query: query, Page: page, TotalPages: total, Results: results}, nil
}

func ParseFreetarTab(htmlPage string) (TabDetail, error) {
	if oopsRE.MatchString(htmlPage) {
		return TabDetail{}, errf("not found")
	}
	body := tabBodyRE.FindStringSubmatch(htmlPage)
	if body == nil {
		return TabDetail{}, errf("could not parse page")
	}
	raw := htmlToRaw(body[1])
	if strings.TrimSpace(raw) == "" {
		return TabDetail{}, errf("could not parse page")
	}
	tab := TabDetail{Raw: raw, Version: 1}
	if h := h5RE.FindStringSubmatch(htmlPage); h != nil {
		plain := strings.TrimSpace(stripTags(h[1]))
		if i := strings.Index(plain, " - "); i >= 0 {
			tab.Artist = strings.TrimSpace(plain[:i])
			rest := strings.TrimSpace(plain[i+3:])
			if vm := verRE.FindStringSubmatch(rest); vm != nil {
				tab.Version = AsInt(vm[1], 1)
				rest = strings.TrimSpace(verRE.ReplaceAllString(rest, ""))
			}
			tab.Song = rest
		}
	}
	if fav := favRE.FindStringSubmatch(htmlPage); fav != nil {
		tab.Artist = html.UnescapeString(fav[1])
		tab.Song = html.UnescapeString(fav[2])
		tab.Type = html.UnescapeString(fav[3])
		tab.Rating = AsFloat(fav[4], 0)
	}
	if m := diffRE.FindStringSubmatch(htmlPage); m != nil {
		tab.Difficulty = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	if m := capoRE.FindStringSubmatch(htmlPage); m != nil {
		tab.Capo = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	if m := tunRE.FindStringSubmatch(htmlPage); m != nil {
		tab.Tuning = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return tab, nil
}

func htmlToRaw(s string) string {
	s = brRE.ReplaceAllString(s, "\n")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&#160;", " ")
	s = chordPartRE.ReplaceAllString(s, "$1")
	s = chordBoxRE.ReplaceAllStringFunc(s, func(in string) string {
		inner := chordBoxRE.FindStringSubmatch(in)
		if inner == nil {
			return ""
		}
		text := strings.TrimSpace(stripTags(inner[1]))
		if text == "" {
			return ""
		}
		return "[ch]" + text + "[/ch]"
	})
	s = stripTags(s)
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return s
}

func stripTags(s string) string {
	return tagRE.ReplaceAllString(s, "")
}
