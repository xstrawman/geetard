package client

import (
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	SearchHost  string
	TabHost     string
	SearchHosts []string
	TabHosts    []string
	HTTP        *http.Client
	UA          string
}

func New(searchHost, tabHost string) *Client {
	c := &Client{
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if err := rejectUG(req.URL.Host); err != nil {
					return err
				}
				if len(via) >= 10 {
					return errf("proxy timed out")
				}
				return nil
			},
		},
		UA: "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
	}
	if searchHost == "" {
		c.SearchHost = DefaultSearchHost
		c.SearchHosts = append([]string(nil), DefaultSearchHosts...)
	} else {
		c.SearchHost = searchHost
		c.SearchHosts = []string{searchHost}
	}
	if tabHost == "" {
		c.TabHost = DefaultTabHost
		c.TabHosts = append([]string(nil), DefaultTabHosts...)
	} else {
		c.TabHost = tabHost
		c.TabHosts = []string{tabHost}
	}
	return c
}

func FromEnv() *Client {
	return New(os.Getenv("FREETAR_SEARCH_HOST"), os.Getenv("FREETAR_TAB_HOST"))
}

func (c *Client) Search(query string, page int) (SearchPage, error) {
	if query == "" {
		return SearchPage{Query: query, Page: page, TotalPages: 1}, nil
	}
	if page < 1 {
		page = 1
	}
	hosts := c.SearchHosts
	if len(hosts) == 0 {
		hosts = []string{c.SearchHost}
	}
	attempts := 1
	if len(hosts) == 1 {
		attempts = 2
	}
	var last error
	var empty SearchPage
	sawEmpty := false
	for _, host := range hosts {
		u, err := SearchURL(host, query, page)
		if err != nil {
			last = err
			continue
		}
		body, err := c.getAttempts(u, attempts)
		if err != nil {
			last = err
			continue
		}
		got, err := parseSearchBody(body, query, page)
		if err != nil {
			last = err
			continue
		}
		if len(got.Results) > 0 {
			return got, nil
		}
		empty = got
		sawEmpty = true
	}
	if sawEmpty {
		return empty, nil
	}
	if last == nil {
		last = errf("catalog down")
	}
	return SearchPage{}, last
}

func (c *Client) Tab(path string) (TabDetail, error) {
	hosts := c.TabHosts
	if len(hosts) == 0 {
		hosts = []string{c.TabHost}
	}
	attempts := 1
	if len(hosts) == 1 {
		attempts = 2
	}
	var last error
	for _, host := range hosts {
		u, err := TabURL(host, path)
		if err != nil {
			last = err
			continue
		}
		body, err := c.getAttempts(u, attempts)
		if err != nil {
			last = err
			continue
		}
		tab, err := parseTabBody(body)
		if err != nil {
			last = err
			continue
		}
		tab.Path = TabPath(path)
		return tab, nil
	}
	if last == nil {
		last = errf("catalog down")
	}
	return TabDetail{}, last
}

func parseSearchBody(body, query string, page int) (SearchPage, error) {
	if store, err := ExtractStore(body); err == nil {
		return MapSearch(store, query, page)
	}
	return ParseFreetarSearch(body, query, page)
}

func parseTabBody(body string) (TabDetail, error) {
	if store, err := ExtractStore(body); err == nil {
		return MapTab(store)
	}
	return ParseFreetarTab(body)
}

func (c *Client) get(u string) (string, error) {
	return c.getAttempts(u, 2)
}

func (c *Client) getAttempts(u string, attempts int) (string, error) {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", c.UA)
		req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			last = errf("proxy timed out")
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			last = errf("proxy timed out")
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", errf("not found")
		}
		if resp.StatusCode >= 500 {
			last = errf("proxy down (%d)", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return "", errf("proxy down (%d)", resp.StatusCode)
		}
		return string(b), nil
	}
	if last == nil {
		last = errf("proxy timed out")
	}
	return "", last
}
