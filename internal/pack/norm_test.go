package pack

import "testing"

func TestTitle_stripsVersionAndLive(t *testing.T) {
	got := Title("Wonderwall (ver 14)")
	if got != "wonderwall" {
		t.Fatal(got)
	}
	if Title("Time Of The Season (Live at Red Rocks)") != "time of the season" {
		t.Fatal(Title("Time Of The Season (Live at Red Rocks)"))
	}
}

func TestArtist_stripsTheAndFeat(t *testing.T) {
	if Artist("The Beatles") != "beatles" {
		t.Fatal(Artist("The Beatles"))
	}
	if Artist("Shovels & Rope (feat. Shakey Graves)") != "shovels & rope" {
		t.Fatal(Artist("Shovels & Rope (feat. Shakey Graves)"))
	}
}

func TestKey_collapsesWonderwallVersions(t *testing.T) {
	a := Key("Oasis", "Wonderwall (ver 1)")
	b := Key("oasis", "Wonderwall (ver 14)")
	if a != b || a == "" {
		t.Fatal(a, b)
	}
	if Key("The Beatles", "Let It Be") == Key("Beatles", "Let It Be") {
		// good
	} else {
		t.Fatal(Key("The Beatles", "Let It Be"), Key("Beatles", "Let It Be"))
	}
}
