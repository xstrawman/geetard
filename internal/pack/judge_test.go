package pack

import (
	"testing"

	"geetard/internal/client"
)

func TestKeepResult_chordsOnly(t *testing.T) {
	if !KeepResult(client.SearchResult{Type: "Chords"}) {
		t.Fatal("chords")
	}
	for _, typ := range []string{"Ukulele", "Bass", "Official", "Pro", "Tab", "Guitar Pro", "Tabs"} {
		if KeepResult(client.SearchResult{Type: typ}) {
			t.Fatal(typ)
		}
	}
}

func TestPick_prefersHalfStepCOverConcertPitchB(t *testing.T) {
	good := client.TabDetail{
		Artist: "Collective Soul",
		Song:   "Run",
		Type:   "Chords",
		Tuning: "1/2 step down",
		Rating: 4.6,
		Raw:    "[ch]C[/ch]  [ch]G[/ch]  [ch]Am[/ch]  [ch]F[/ch]\nverse in C shapes",
	}
	bad := client.TabDetail{
		Artist: "Collective Soul",
		Song:   "Run",
		Type:   "Chords",
		Tuning: "Standard",
		Rating: 4.9,
		Raw:    "[ch]B[/ch]  [ch]F#[/ch]  [ch]G#m[/ch]  [ch]E[/ch]  [ch]B[/ch]  [ch]B[/ch]\nconcert pitch B",
	}
	got, ok := Pick([]client.TabDetail{bad, good})
	if !ok {
		t.Fatal("no pick")
	}
	if got.Tuning != good.Tuning || got.Raw != good.Raw {
		t.Fatalf("picked concert-pitch sheet: tuning=%q", got.Tuning)
	}
}

func TestPick_skipsNonChords(t *testing.T) {
	_, ok := Pick([]client.TabDetail{{Type: "Tab", Raw: "e|--", Rating: 5}})
	if ok {
		t.Fatal("should skip tab")
	}
}
