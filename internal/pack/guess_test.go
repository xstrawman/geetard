package pack

import (
	"testing"

	"geetard/internal/client"
)

func TestGuessOriginal_skipsCoverArtist(t *testing.T) {
	page := client.SearchPage{Results: []client.SearchResult{
		{Artist: "Johnny Cash", Song: "Hurt", Type: "Chords", Rating: 4.9},
		{Artist: "Nine Inch Nails", Song: "Hurt", Type: "Chords", Rating: 4.8},
		{Artist: "Johnny Cash", Song: "Hurt", Type: "Ukulele", Rating: 5},
	}}
	got := GuessOriginal("Hurt", "Johnny Cash", page)
	if got != "Nine Inch Nails" {
		t.Fatal(got)
	}
}

func TestGuessOriginal_emptyIfOnlyCover(t *testing.T) {
	page := client.SearchPage{Results: []client.SearchResult{
		{Artist: "Animotion", Song: "Obsession", Type: "Chords", Rating: 4.7},
	}}
	if GuessOriginal("Obsession", "Animotion", page) != "" {
		t.Fatal("expected empty so we fall back to cover")
	}
}
