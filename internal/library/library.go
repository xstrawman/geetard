package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"geetard/internal/client"
)

type Record struct {
	Path           string              `json:"path"`
	Artist         string              `json:"artist"`
	Song           string              `json:"song"`
	Version        int                 `json:"version"`
	Type           string              `json:"type"`
	Rating         float64             `json:"rating"`
	Capo           string              `json:"capo"`
	Tuning         string              `json:"tuning"`
	Difficulty     string              `json:"difficulty"`
	Raw            string              `json:"raw"`
	Shapes         []client.ChordShape `json:"shapes"`
	FetchedAt      time.Time           `json:"fetched_at"`
	OriginalArtist string              `json:"original_artist,omitempty"`
	CoverArtist    string              `json:"cover_artist,omitempty"`
	SheetArtist    string              `json:"sheet_artist,omitempty"`
	SourceRank     int                 `json:"source_rank,omitempty"`
}

func FromTab(t client.TabDetail) Record {
	return Record{
		Path:       t.Path,
		Artist:     t.Artist,
		Song:       t.Song,
		Version:    t.Version,
		Type:       t.Type,
		Rating:     t.Rating,
		Capo:       t.Capo,
		Tuning:     t.Tuning,
		Difficulty: t.Difficulty,
		Raw:        t.Raw,
		Shapes:     t.Shapes,
		FetchedAt:  time.Now().UTC(),
	}
}

func (r Record) Tab() client.TabDetail {
	return client.TabDetail{
		Path:       r.Path,
		Artist:     r.Artist,
		Song:       r.Song,
		Version:    r.Version,
		Type:       r.Type,
		Rating:     r.Rating,
		Capo:       r.Capo,
		Tuning:     r.Tuning,
		Difficulty: r.Difficulty,
		Raw:        r.Raw,
		Shapes:     r.Shapes,
	}
}

func fileName(path string) string {
	p := client.TabPath(path)
	p = strings.ReplaceAll(p, "/", "__")
	p = strings.ReplaceAll(p, string(filepath.Separator), "__")
	if p == "" || strings.Contains(p, "..") {
		return ""
	}
	return p + ".json"
}

func Save(dir string, r Record) error {
	name := fileName(r.Path)
	if name == "" {
		return client.ClientError{Msg: "missing tab path"}
	}
	return writeTab(dir, name, r)
}

var (
	verParenFile  = regexp.MustCompile(`(?i)\s*\(ver(?:sion)?\s*\d+\)`)
	liveParenFile = regexp.MustCompile(`(?i)\s*\([^)]*live[^)]*\)`)
	thePrefixFile = regexp.MustCompile(`(?i)^the\s+`)
)

func CanonicalFile(artist, song string) string {
	return canonicalFile(artist, song)
}

func canonicalFile(artist, song string) string {
	a := strings.ToLower(strings.TrimSpace(artist))
	a = thePrefixFile.ReplaceAllString(a, "")
	s := strings.ToLower(strings.TrimSpace(song))
	s = verParenFile.ReplaceAllString(s, "")
	s = liveParenFile.ReplaceAllString(s, "")
	a = strings.TrimSpace(a)
	s = strings.TrimSpace(s)
	if a == "" || s == "" {
		return ""
	}
	a = strings.ReplaceAll(a, "/", "-")
	s = strings.ReplaceAll(s, "/", "-")
	return a + "__" + s + ".json"
}

func SaveCanonical(dir string, r Record) error {
	name := canonicalFile(r.Artist, r.Song)
	if name == "" {
		return client.ClientError{Msg: "missing artist or title"}
	}
	return writeTab(dir, name, r)
}

func writeTab(dir, name string, r Record) error {
	tabs := filepath.Join(dir, "tabs")
	if err := os.MkdirAll(tabs, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tabs, name), append(b, '\n'), 0o644)
}

func Load(dir, path string) (Record, error) {
	name := fileName(path)
	if name == "" {
		return Record{}, client.ClientError{Msg: "missing tab path"}
	}
	b, err := os.ReadFile(filepath.Join(dir, "tabs", name))
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := json.Unmarshal(b, &r); err != nil {
		return Record{}, err
	}
	return r, nil
}

func List(dir string) ([]Record, error) {
	ents, err := os.ReadDir(filepath.Join(dir, "tabs"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Record
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, "tabs", e.Name()))
		if err != nil {
			continue
		}
		var r Record
		if json.Unmarshal(b, &r) != nil || r.Path == "" {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].FetchedAt.Equal(out[j].FetchedAt) {
			return out[i].FetchedAt.After(out[j].FetchedAt)
		}
		if out[i].Artist != out[j].Artist {
			return out[i].Artist < out[j].Artist
		}
		return out[i].Song < out[j].Song
	})
	return out, nil
}
