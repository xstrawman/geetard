package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSearch_emptyQueryDoesNotHitNetwork(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	page, err := c.Search("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 0 || len(page.Results) != 0 {
		t.Fatalf("hits=%d page=%+v", hits, page)
	}
}

func TestClientSearch_parsesFixture(t *testing.T) {
	html := testdata(t, "search_ok.html")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search.php" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	page, err := c.Search("nelson", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Results) != 1 || page.Results[0].Artist != "Willie Nelson" {
		t.Fatalf("%+v", page)
	}
}

func TestClientTab_parsesFixture(t *testing.T) {
	html := testdata(t, "tab_ok.html")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tab/willie-nelson/always-on-my-mind-123" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	tab, err := c.Tab("willie-nelson/always-on-my-mind-123")
	if err != nil {
		t.Fatal(err)
	}
	if tab.Song != "Always On My Mind" || len(tab.Shapes) != 1 {
		t.Fatalf("%+v", tab)
	}
}

func TestClientSearch_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()
	c := New(srv.URL, srv.URL)
	c.HTTP.Timeout = time.Second
	_, err := c.Search("x", 1)
	if err == nil || err.Error() != "proxy down (503)" {
		t.Fatalf("got %v", err)
	}
}
