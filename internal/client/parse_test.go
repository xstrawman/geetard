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
