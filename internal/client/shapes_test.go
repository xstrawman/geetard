package client

import (
	"strings"
	"testing"
)

func TestShapes_gFromFixture(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	view := store["store"].(map[string]any)["page"].(map[string]any)["data"].(map[string]any)["tab_view"].(map[string]any)
	shapes := Shapes(view)
	if len(shapes) != 1 || shapes[0].Name != "G" {
		t.Fatalf("%+v", shapes)
	}
	joined := strings.Join(shapes[0].Lines, "\n")
	want := strings.Join([]string{
		"  e |-3-",
		"  B |-0-",
		"  G |-0-",
		"  D |-0-",
		"  A |-2-",
		"  E |-3-",
	}, "\n")
	if joined != want {
		t.Fatalf("got\n%s\nwant\n%s", joined, want)
	}
}

func TestShapes_missingIsEmpty(t *testing.T) {
	store, err := ExtractStore(testdata(t, "tab_no_applicature.html"))
	if err != nil {
		t.Fatal(err)
	}
	view := store["store"].(map[string]any)["page"].(map[string]any)["data"].(map[string]any)["tab_view"].(map[string]any)
	if len(Shapes(view)) != 0 {
		t.Fatal(Shapes(view))
	}
}

func TestShapes_muteIsX(t *testing.T) {
	view := map[string]any{
		"applicature": map[string]any{
			"Em": []any{map[string]any{"frets": []any{0.0, 2.0, 2.0, 0.0, 0.0, -1.0}}},
		},
	}
	shapes := Shapes(view)
	if len(shapes) != 1 || !strings.Contains(strings.Join(shapes[0].Lines, "\n"), "  e |-x-") {
		t.Fatalf("%+v", shapes)
	}
}
