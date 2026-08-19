package library

import (
	"testing"

	"geetard/internal/client"
)

func TestSaveLoadList(t *testing.T) {
	dir := t.TempDir()
	tab := client.TabDetail{
		Path:   "willie-nelson/always-on-my-mind-chords-77919",
		Artist: "Willie Nelson",
		Song:   "Always On My Mind",
		Raw:    "[ch]G[/ch]\nhello",
		Type:   "Chords",
	}
	if err := Save(dir, FromTab(tab)); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir, tab.Path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artist != tab.Artist || got.Raw != tab.Raw || got.Path != tab.Path {
		t.Fatalf("%+v", got)
	}
	list, err := List(dir)
	if err != nil || len(list) != 1 || list[0].Song != tab.Song {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestListMissingDir(t *testing.T) {
	list, err := List(t.TempDir())
	if err != nil || len(list) != 0 {
		t.Fatal(list, err)
	}
}

func TestSaveRejectsEmptyPath(t *testing.T) {
	if err := Save(t.TempDir(), Record{Artist: "x"}); err == nil {
		t.Fatal("expected error")
	}
}
