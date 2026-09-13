package pack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSeed_readsCoverList(t *testing.T) {
	p := filepath.Join("..", "..", "testdata", "cover-songs-seed.json")
	if _, err := os.Stat(p); err != nil {
		p = filepath.Join("testdata", "cover-songs-seed.json")
	}
	seed, err := LoadSeed(p)
	if err != nil {
		t.Fatal(err)
	}
	if seed.Count != 850 || len(seed.Songs) != 850 {
		t.Fatalf("count %d n %d", seed.Count, len(seed.Songs))
	}
	if seed.Songs[0].Rank != 1 || seed.Songs[0].Title != "Hurt" || seed.Songs[0].CoverArtist != "Johnny Cash" {
		t.Fatalf("%+v", seed.Songs[0])
	}
	if seed.Songs[0].OriginalArtist != "" {
		t.Fatal("original should still be empty")
	}
}
