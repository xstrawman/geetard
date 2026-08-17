package client

import (
	"strings"
	"testing"
)

func TestParseFreetarSearch(t *testing.T) {
	page, err := ParseFreetarSearch(testdata(t, "freetar_search.html"), "nelson", 1)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalPages != 10 || len(page.Results) != 2 {
		t.Fatalf("%+v", page)
	}
	if page.Results[1].Artist != "Willie Nelson" || page.Results[1].Path != "willie-nelson/on-the-road-again-chords-1" {
		t.Fatalf("%+v", page.Results[1])
	}
	if page.Results[0].Version != 1 {
		t.Fatal(page.Results[0].Version)
	}
}

func TestParseFreetarTab(t *testing.T) {
	tab, err := ParseFreetarTab(testdata(t, "freetar_tab.html"))
	if err != nil {
		t.Fatal(err)
	}
	if tab.Artist != "Nelson" || !strings.Contains(tab.Song, "Cant Live") {
		t.Fatalf("%+v", tab)
	}
	if tab.Capo != "no capo" || !strings.Contains(tab.Tuning, "Standard") || tab.Difficulty != "beginner" {
		t.Fatalf("meta %+v", tab)
	}
	if !strings.Contains(tab.Raw, "[ch]G[/ch]") || !strings.Contains(tab.Raw, "Maybe I didn't love you") {
		t.Fatalf("raw %q", tab.Raw)
	}
}

func TestParseFreetarTab_oops(t *testing.T) {
	_, err := ParseFreetarTab(`<div>Oops: Could not parse chord</div>`)
	if err == nil || err.Error() != "not found" {
		t.Fatalf("got %v", err)
	}
}
