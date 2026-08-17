package client

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultSearchHost = "https://freetar.de"
	DefaultTabHost    = "https://freetar.de"
)

func TabPath(tabURL string) string {
	u, err := url.Parse(tabURL)
	path := tabURL
	if err == nil && u.Path != "" {
		path = u.Path
	}
	path = strings.TrimPrefix(path, "/tab/")
	path = strings.TrimPrefix(path, "tab/")
	return strings.TrimPrefix(path, "/")
}

func SearchURL(host, query string, page int) (string, error) {
	if err := rejectUG(host); err != nil {
		return "", err
	}
	u, err := url.Parse(strings.TrimRight(host, "/") + "/search")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("search_term", query)
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func TabURL(host, path string) (string, error) {
	if err := rejectUG(host); err != nil {
		return "", err
	}
	return strings.TrimRight(host, "/") + "/tab/" + TabPath(path), nil
}

func rejectUG(host string) error {
	if strings.Contains(strings.ToLower(host), "ultimate-guitar.com") {
		return errf("refusing ultimate-guitar.com")
	}
	return nil
}
