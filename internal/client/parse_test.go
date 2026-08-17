package client

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdata(t *testing.T, name string) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	p := filepath.Join(filepath.Dir(file), "..", "..", "testdata", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestExtractStore_unescapesAndReturnsObject(t *testing.T) {
	store, err := ExtractStore(testdata(t, "search_ok.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := store["store"].(map[string]any)["page"].(map[string]any)
	data := page["data"].(map[string]any)
	results := data["results"].([]any)
	if len(results) != 3 {
		t.Fatalf("got %d results", len(results))
	}
}

func TestExtractStore_missingDiv(t *testing.T) {
	_, err := ExtractStore(testdata(t, "no_store.html"))
	if err == nil || err.Error() != "page layout changed" {
		t.Fatalf("got %v", err)
	}
}

func TestAsInt_stringOrNumber(t *testing.T) {
	if AsInt("3", 0) != 3 || AsInt(3.0, 0) != 3 || AsInt(nil, 7) != 7 {
		t.Fatal(AsInt("3", 0), AsInt(3.0, 0), AsInt(nil, 7))
	}
}

func TestAsFloat_stringOrNumber(t *testing.T) {
	if AsFloat("4.8", 0) != 4.8 || AsFloat(5, 0) != 5 || AsFloat(nil, 1) != 1 {
		t.Fatal(AsFloat("4.8", 0), AsFloat(5, 0), AsFloat(nil, 1))
	}
}
