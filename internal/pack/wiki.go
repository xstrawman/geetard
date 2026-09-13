package pack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	pageSongArtist = regexp.MustCompile(`(?i)^(.+?)\s+\((.+?)\s+song\)$`)
	isASongBy      = regexp.MustCompile(`(?i)is a song (?:written and )?(?:recorded )?by (.+?)(?: from|,|—|\.| written)`)
	originallyBy   = regexp.MustCompile(`(?i)originally (?:written(?: and composed)? |recorded |performed )?by ([^.]+?)(?:[.|,]| in )`)
)

var roleJunk = []string{
	"american ", "british ", "english ", "canadian ", "irish ", "australian ",
	"scottish ", "german ", "french ", "swedish ", "welsh ", "new zealand ",
	"industrial rock band ", "rock band ", "pop band ", "folk band ",
	"country musician ", "jazz band ", "metal band ",
	"singer-songwriter ", "singer ", "musician ", "rapper ",
	"band ", "trio ", "duo ", "quartet ", "the ",
}

func OriginalFromPageTitle(pageTitle, cover string) string {
	m := pageSongArtist.FindStringSubmatch(pageTitle)
	if m == nil {
		return ""
	}
	artist := strings.TrimSpace(m[2])
	if artist == "" || MatchArtist(artist, cover) {
		return ""
	}
	return artist
}

func OriginalFromExtract(extract, cover string) string {
	for _, re := range []*regexp.Regexp{originallyBy, isASongBy} {
		m := re.FindStringSubmatch(extract)
		if m == nil {
			continue
		}
		name := stripRoles(strings.TrimSpace(m[1]))
		if name == "" || MatchArtist(name, cover) {
			continue
		}
		if i := strings.Index(name, ". "); i > 0 {
			name = name[:i]
		}
		return strings.TrimSpace(name)
	}
	return ""
}

func stripRoles(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	changed := true
	for changed {
		changed = false
		for _, p := range roleJunk {
			if strings.HasPrefix(lower, p) {
				s = strings.TrimSpace(s[len(p):])
				lower = strings.ToLower(s)
				changed = true
			}
		}
	}
	if i := strings.Index(s, " and "); i > 0 {
		// "Bob Dylan and produced by..." already cut by regex; keep duos
	}
	return strings.TrimSpace(s)
}

type WikiClient struct {
	HTTP  *http.Client
	Sleep time.Duration
	mu    sync.Mutex
	last  time.Time
}

func (w *WikiClient) Resolve(title, cover string) (string, error) {
	if w.HTTP == nil {
		w.HTTP = &http.Client{Timeout: 12 * time.Second}
	}
	if w.Sleep == 0 {
		w.Sleep = time.Second
	}
	q := fmt.Sprintf("%q %q song", title, cover)
	hits, err := w.search(q)
	if err != nil {
		return "", err
	}
	for _, hit := range hits {
		if got := OriginalFromPageTitle(hit, cover); got != "" {
			return got, nil
		}
	}
	for _, hit := range hits {
		ex, err := w.extract(hit)
		if err != nil {
			continue
		}
		if got := OriginalFromExtract(ex, cover); got != "" {
			return got, nil
		}
	}
	return "", nil
}

func (w *WikiClient) search(q string) ([]string, error) {
	u := "https://en.wikipedia.org/w/api.php?action=query&list=search&srlimit=5&format=json&srsearch=" + url.QueryEscape(q)
	b, err := w.get(u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Query struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	var out []string
	for _, s := range resp.Query.Search {
		out = append(out, s.Title)
	}
	return out, nil
}

func (w *WikiClient) extract(pageTitle string) (string, error) {
	u := "https://en.wikipedia.org/w/api.php?action=query&prop=extracts&exintro=1&explaintext=1&format=json&titles=" + url.QueryEscape(pageTitle)
	b, err := w.get(u)
	if err != nil {
		return "", err
	}
	var resp struct {
		Query struct {
			Pages map[string]struct {
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", err
	}
	for _, p := range resp.Query.Pages {
		return p.Extract, nil
	}
	return "", nil
}

func (w *WikiClient) throttle() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.Sleep <= 0 {
		w.Sleep = time.Second
	}
	if !w.last.IsZero() {
		wait := w.Sleep - time.Since(w.last)
		if wait > 0 {
			time.Sleep(wait)
		}
	}
	w.last = time.Now()
}

func (w *WikiClient) get(u string) ([]byte, error) {
	if w.HTTP == nil {
		w.HTTP = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	ua := "geetard-orig/0.1 (guitar corpus original-artist lookup; +https://en.wikipedia.org/wiki/Wikipedia:Bot_policy)"
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Api-User-Agent", ua)
	req.Header.Set("Accept", "application/json")

	backoff := 60 * time.Second
	var last error
	for attempt := 0; attempt < 4; attempt++ {
		w.throttle()
		resp, err := w.HTTP.Do(req)
		if err != nil {
			last = err
			time.Sleep(backoff)
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			last = readErr
			time.Sleep(backoff)
			continue
		}
		if resp.StatusCode == 429 {
			wait := backoff
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if n, err := strconv.Atoi(ra); err == nil && n > 0 {
					wait = time.Duration(n) * time.Second
				}
			}
			last = fmt.Errorf("wikipedia 429")
			time.Sleep(wait)
			backoff *= 2
			if backoff > 5*time.Minute {
				backoff = 5 * time.Minute
			}
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("wikipedia %d", resp.StatusCode)
		}
		return b, nil
	}
	if last == nil {
		last = fmt.Errorf("wikipedia 429")
	}
	return nil, last
}
