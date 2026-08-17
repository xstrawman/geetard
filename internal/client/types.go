package client

import "fmt"

type ClientError struct{ Msg string }

func (e ClientError) Error() string { return e.Msg }

func errf(msg string, args ...any) error {
	if len(args) == 0 {
		return ClientError{Msg: msg}
	}
	return ClientError{Msg: fmt.Sprintf(msg, args...)}
}

type SearchPage struct {
	Query      string
	Page       int
	TotalPages int
	Results    []SearchResult
}

type SearchResult struct {
	Artist  string
	Song    string
	Type    string
	Version int
	Rating  float64
	Votes   int
	Path    string
}

type TabDetail struct {
	Artist     string
	Song       string
	Version    int
	Type       string
	Rating     float64
	Capo       string
	Tuning     string
	Difficulty string
	Raw        string
	Shapes     []ChordShape
}

type ChordShape struct {
	Name  string
	Lines []string
}

type Kind int

const (
	KindBody Kind = iota
	KindChord
)

type Part struct {
	Text string
	Kind Kind
}

type DisplayLine struct {
	Parts []Part
}
