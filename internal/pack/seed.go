package pack

import (
	"encoding/json"
	"os"
	"sort"
)

type Seed struct {
	Version int    `json:"version"`
	Source  string `json:"source"`
	Count   int    `json:"count"`
	Note    string `json:"note"`
	Songs   []Song `json:"songs"`
}

type Song struct {
	Rank           int    `json:"rank"`
	Title          string `json:"title"`
	Artist         string `json:"artist,omitempty"`
	CoverArtist    string `json:"cover_artist"`
	OriginalArtist string `json:"original_artist"`
}

func LoadSeed(path string) (Seed, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Seed{}, err
	}
	var s Seed
	if err := json.Unmarshal(b, &s); err != nil {
		return Seed{}, err
	}
	if s.Count == 0 {
		s.Count = len(s.Songs)
	}
	return s, nil
}

func SaveSeed(path string, s Seed) error {
	s.Count = len(s.Songs)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func SliceByRank(s Seed, from, to int) Seed {
	out := s
	out.Songs = nil
	for _, song := range s.Songs {
		if from > 0 && song.Rank < from {
			continue
		}
		if to > 0 && song.Rank > to {
			continue
		}
		out.Songs = append(out.Songs, song)
	}
	out.Count = len(out.Songs)
	return out
}

func MergeSeeds(parts []Seed) Seed {
	byRank := map[int]Song{}
	var note, source string
	for _, p := range parts {
		if p.Note != "" {
			note = p.Note
		}
		if p.Source != "" {
			source = p.Source
		}
		for _, song := range p.Songs {
			old, ok := byRank[song.Rank]
			if !ok {
				byRank[song.Rank] = song
				continue
			}
			if old.OriginalArtist == "" && song.OriginalArtist != "" {
				byRank[song.Rank] = song
			}
		}
	}
	ranks := make([]int, 0, len(byRank))
	for r := range byRank {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)
	songs := make([]Song, 0, len(ranks))
	for _, r := range ranks {
		songs = append(songs, byRank[r])
	}
	return Seed{Version: 1, Source: source, Note: note, Count: len(songs), Songs: songs}
}
