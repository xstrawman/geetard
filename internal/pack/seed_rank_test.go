package pack

import "testing"

func TestSliceByRank(t *testing.T) {
	s := Seed{Songs: []Song{
		{Rank: 1, Title: "Hurt"},
		{Rank: 2, Title: "Watchtower"},
		{Rank: 3, Title: "Hallelujah"},
	}}
	got := SliceByRank(s, 2, 3)
	if len(got.Songs) != 2 || got.Songs[0].Rank != 2 || got.Songs[1].Rank != 3 {
		t.Fatalf("%+v", got.Songs)
	}
}

func TestMergeSeeds_prefersFilledOriginal(t *testing.T) {
	a := Seed{Songs: []Song{{Rank: 1, Title: "Hurt", OriginalArtist: ""}}}
	b := Seed{Songs: []Song{{Rank: 1, Title: "Hurt", OriginalArtist: "Nine Inch Nails"}, {Rank: 2, Title: "Watchtower", OriginalArtist: "Bob Dylan"}}}
	got := MergeSeeds([]Seed{a, b})
	if len(got.Songs) != 2 {
		t.Fatal(len(got.Songs))
	}
	if got.Songs[0].OriginalArtist != "Nine Inch Nails" {
		t.Fatal(got.Songs[0])
	}
}
