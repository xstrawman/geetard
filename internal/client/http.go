package client

import (
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	SearchHost string
	TabHost    string
	HTTP       *http.Client
	UA         string
}

func New(searchHost, tabHost string) *Client {
	if searchHost == "" {
		searchHost = DefaultSearchHost
	}
	if tabHost == "" {
		tabHost = DefaultTabHost
	}
	return &Client{
		SearchHost: searchHost,
		TabHost:    tabHost,
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
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
	u, err := SearchURL(c.SearchHost, query, page)
	if err != nil {
		return SearchPage{}, err
	}
	body, err := c.get(u)
	if err != nil {
		return SearchPage{}, err
	}
	if store, err := ExtractStore(body); err == nil {
		return MapSearch(store, query, page)
	}
	return ParseFreetarSearch(body, query, page)
}

func (c *Client) Tab(path string) (TabDetail, error) {
	u, err := TabURL(c.TabHost, path)
	if err != nil {
		return TabDetail{}, err
	}
	body, err := c.get(u)
	if err != nil {
		return TabDetail{}, err
	}
	var tab TabDetail
	var perr error
	if store, err := ExtractStore(body); err == nil {
		tab, perr = MapTab(store)
	} else {
		tab, perr = ParseFreetarTab(body)
	}
	if perr != nil {
		return TabDetail{}, perr
	}
	tab.Path = TabPath(path)
	return tab, nil
}

func (c *Client) get(u string) (string, error) {
	var last error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", c.UA)
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
