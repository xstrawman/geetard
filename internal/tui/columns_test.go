package tui

import (
	"testing"

	"geetard/internal/client"
)

func bodyLine(s string) client.DisplayLine {
	return client.DisplayLine{Parts: []client.Part{{Text: s, Kind: client.KindBody}}}
}

func TestSheetColumns_narrowIsOne(t *testing.T) {
	lines := []client.DisplayLine{bodyLine("G     D")}
	if n := sheetColumns(40, lines); n != 1 {
		t.Fatal(n)
	}
}

func TestSheetColumns_chordSheetGetsThree(t *testing.T) {
	lines := []client.DisplayLine{bodyLine("Maybe I didn't love you")}
	if n := sheetColumns(200, lines); n != 3 {
		t.Fatal(n)
	}
}

func TestSheetColumns_wideTabStaysOne(t *testing.T) {
	wide := bodyLine("e|--------------------------------------------------------------------")
	if n := sheetColumns(200, []client.DisplayLine{wide}); n != 1 {
		t.Fatal(n)
	}
}
