package pack

import (
	"encoding/json"
	"os"
	"strings"
)

func LoadOverride(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		out[Title(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

func ApplyOverride(song Song, m map[string]string) string {
	if m == nil {
		return song.OriginalArtist
	}
	if got := m[Title(song.Title)]; got != "" {
		return got
	}
	return song.OriginalArtist
}
