package client

import "testing"

func TestRender_splitsChordsAndBody(t *testing.T) {
	lines := Render("[tab][ch]Em[/ch]  [ch]G[/ch]\nverse line[/tab]")
	if len(lines) != 2 {
		t.Fatalf("%+v", lines)
	}
	if len(lines[0].Parts) != 3 ||
		lines[0].Parts[0] != (Part{Text: "Em", Kind: KindChord}) ||
		lines[0].Parts[1] != (Part{Text: "  ", Kind: KindBody}) ||
		lines[0].Parts[2] != (Part{Text: "G", Kind: KindChord}) {
		t.Fatalf("line0 %+v", lines[0].Parts)
	}
	if len(lines[1].Parts) != 1 || lines[1].Parts[0].Text != "verse line" || lines[1].Parts[0].Kind != KindBody {
		t.Fatalf("line1 %+v", lines[1].Parts)
	}
}

func TestRender_empty(t *testing.T) {
	if len(Render("")) != 0 || len(Render("[tab][/tab]")) != 0 {
		t.Fatal(Render(""), Render("[tab][/tab]"))
	}
}

func TestRender_noHTMLEntities(t *testing.T) {
	lines := Render("hello  world")
	if len(lines) != 1 || lines[0].Parts[0].Text != "hello  world" {
		t.Fatalf("%+v", lines)
	}
}
