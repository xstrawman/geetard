package pack

import (
	"path/filepath"
	"testing"
)

func TestLoadOverride_prine(t *testing.T) {
	p := filepath.Join("..", "..", "testdata", "originals-override.json")
	m, err := LoadOverride(p)
	if err != nil {
		t.Fatal(err)
	}
	if m["angel from montgomery"] != "John Prine" {
		t.Fatalf("%v", m)
	}
}

func TestApplyOverride(t *testing.T) {
	m := map[string]string{"angel from montgomery": "John Prine"}
	s := Song{Title: "Angel From Montgomery", CoverArtist: "Bonnie Raitt"}
	if ApplyOverride(s, m) != "John Prine" {
		t.Fatal(ApplyOverride(s, m))
	}
}
