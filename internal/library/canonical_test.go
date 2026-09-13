package library

import (
	"testing"

	"geetard/internal/client"
)

func TestSaveCanonical_oneSheetPerSong(t *testing.T) {
	dir := t.TempDir()
	a := FromTab(client.TabDetail{
		Path:   "oasis/wonderwall-chords-1",
		Artist: "Oasis",
		Song:   "Wonderwall",
		Type:   "Chords",
		Raw:    "ver1",
	})
	b := FromTab(client.TabDetail{
		Path:   "oasis/wonderwall-chords-99",
		Artist: "Oasis",
		Song:   "Wonderwall (ver 14)",
		Type:   "Chords",
		Raw:    "ver14",
	})
	if err := SaveCanonical(dir, a); err != nil {
		t.Fatal(err)
	}
	if err := SaveCanonical(dir, b); err != nil {
		t.Fatal(err)
	}
	list, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 song, got %d", len(list))
	}
	if list[0].Raw != "ver14" {
		t.Fatalf("expected overwrite with later save, got %+v", list[0])
	}
}
