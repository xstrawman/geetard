package pack

import "testing"

func TestOriginalFromPageTitle(t *testing.T) {
	got := OriginalFromPageTitle("Hallelujah (Leonard Cohen song)", "Jeff Buckley")
	if got != "Leonard Cohen" {
		t.Fatal(got)
	}
	if OriginalFromPageTitle("Hurt (Nine Inch Nails song)", "Johnny Cash") != "Nine Inch Nails" {
		t.Fatal(OriginalFromPageTitle("Hurt (Nine Inch Nails song)", "Johnny Cash"))
	}
	if OriginalFromPageTitle("American IV: The Man Comes Around", "Johnny Cash") != "" {
		t.Fatal("album page should not count")
	}
	if OriginalFromPageTitle("Hurt (Johnny Cash song)", "Johnny Cash") != "" {
		t.Fatal("cover-artist page should not count")
	}
}

func TestOriginalFromExtract(t *testing.T) {
	ex := `"All Along the Watchtower" is a song by American singer-songwriter Bob Dylan from his eighth studio album, John Wesley Harding (1967).`
	got := OriginalFromExtract(ex, "Jimi Hendrix")
	if got != "Bob Dylan" {
		t.Fatal(got)
	}
	ex = `"Hurt" is a song by American industrial rock band Nine Inch Nails from its 1994 studio album The Downward Spiral`
	if OriginalFromExtract(ex, "Johnny Cash") != "Nine Inch Nails" {
		t.Fatal(OriginalFromExtract(ex, "Johnny Cash"))
	}
	ex = `"Sweet Jane" is a song by American rock band the Velvet Underground. Appearing on their fourth studio album Loaded (1970).`
	if OriginalFromExtract(ex, "Cowboy Junkies") != "Velvet Underground" {
		t.Fatal(OriginalFromExtract(ex, "Cowboy Junkies"))
	}
	ex = `"Respect" is a song originally recorded by American singer-songwriter Otis Redding in 1965.`
	if OriginalFromExtract(ex, "Aretha Franklin") != "Otis Redding" {
		t.Fatal(OriginalFromExtract(ex, "Aretha Franklin"))
	}
}
