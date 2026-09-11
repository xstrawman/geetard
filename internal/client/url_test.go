package client

import (
	"strings"
	"testing"
)

func TestDefaultSearchHosts_doesNotLeadWithDeadFreetar(t *testing.T) {
	if len(DefaultSearchHosts) < 2 {
		t.Fatalf("need fallbacks, got %v", DefaultSearchHosts)
	}
	if DefaultSearchHosts[0] == "https://freetar.de" || DefaultSearchHost == "https://freetar.de" {
		t.Fatal("freetar.de search is dead; do not try it first")
	}
	joined := strings.Join(DefaultSearchHosts, " ")
	if strings.Contains(joined, "ultimate-guitar.com") {
		t.Fatal(joined)
	}
}

func TestSearchURL_usesProxyAndNeverUG(t *testing.T) {
	u, err := SearchURL(DefaultSearchHost, "nelson always", 2)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(u, "ultimate-guitar.com") {
		t.Fatal(u)
	}
	if !strings.HasPrefix(u, DefaultSearchHost+"/search?") {
		t.Fatal(u)
	}
	if !strings.Contains(u, "search_term=") || !strings.Contains(u, "page=2") {
		t.Fatal(u)
	}
}

func TestSearchURL_rejectsUGHost(t *testing.T) {
	_, err := SearchURL("https://www.ultimate-guitar.com", "x", 1)
	if err == nil {
		t.Fatal("expected reject")
	}
}

func TestTabURL_stripsPrefix(t *testing.T) {
	u, err := TabURL(DefaultTabHost, "/tab/willie-nelson/always-on-my-mind-123")
	if err != nil {
		t.Fatal(err)
	}
	if u != DefaultTabHost+"/tab/willie-nelson/always-on-my-mind-123" {
		t.Fatal(u)
	}
}
