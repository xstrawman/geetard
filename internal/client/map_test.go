package client

import "testing"

func TestMapSearch_dropsProOfficialAndKeepsChords(t *testing.T) {
	store, err := ExtractStore(testdata(t, "search_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	page, err := MapSearch(store, "nelson", 1)
	if err != nil {
		t.Fatal(err)
	}
	if page.Query != "nelson" || page.Page != 1 || page.TotalPages != 3 {
		t.Fatalf("%+v", page)
	}
	if len(page.Results) != 1 {
		t.Fatalf("want 1 result, got %+v", page.Results)
	}
	r := page.Results[0]
	if r.Artist != "Willie Nelson" || r.Song != "Always On My Mind" || r.Type != "Chords" {
		t.Fatalf("%+v", r)
	}
	if r.Version != 3 || r.Votes != 12 || r.Rating != 4.8 {
		t.Fatalf("numbers %+v", r)
	}
	if r.Path != "willie-nelson/always-on-my-mind-123" {
		t.Fatalf("path %q", r.Path)
	}
}
