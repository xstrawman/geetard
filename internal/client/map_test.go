package client

import (
	"strings"
	"testing"
)

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

func TestMapTab_readsMetaAndRaw(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := MapTab(store)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Artist != "Willie Nelson" || tab.Song != "Always On My Mind" {
		t.Fatalf("%+v", tab)
	}
	if tab.Version != 3 || tab.Type != "Chords" || tab.Rating != 5 {
		t.Fatalf("meta %+v", tab)
	}
	if tab.Capo != "1" || tab.Tuning != "E A D G B E (Standard)" || tab.Difficulty != "novice" {
		t.Fatalf("extra %+v", tab)
	}
	if !strings.Contains(tab.Raw, "[ch]G[/ch]") {
		t.Fatalf("raw %q", tab.Raw)
	}
}

func TestMapTab_missingApplicatureStillWorks(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_no_applicature.html"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := MapTab(store)
	if err != nil {
		t.Fatal(err)
	}
	if tab.Raw == "" {
		t.Fatal("empty raw")
	}
}
