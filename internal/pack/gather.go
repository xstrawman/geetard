package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"geetard/internal/client"
	"geetard/internal/library"
)

type GatherOpts struct {
	OutDir   string
	Limit    int
	Delay    time.Duration
	Log      func(string)
	MaxFetch int
	Wiki     *WikiClient
	Override map[string]string
}

func Gather(c *client.Client, seed Seed, opt GatherOpts) error {
	if opt.Delay <= 0 {
		opt.Delay = 400 * time.Millisecond
	}
	if opt.MaxFetch <= 0 {
		opt.MaxFetch = 5
	}
	if opt.Log == nil {
		opt.Log = func(string) {}
	}
	if err := os.MkdirAll(filepath.Join(opt.OutDir, "tabs"), 0o755); err != nil {
		return err
	}
	n := 0
	for _, song := range seed.Songs {
		if opt.Limit > 0 && n >= opt.Limit {
			break
		}
		if already(opt.OutDir, song, opt.Override) {
			opt.Log(fmt.Sprintf("skip %d %s (have)", song.Rank, song.Title))
			n++
			continue
		}
		if err := gatherOne(c, song, opt); err != nil {
			opt.Log(fmt.Sprintf("fail %d %s: %v", song.Rank, song.Title, err))
			time.Sleep(opt.Delay)
			continue
		}
		n++
		time.Sleep(opt.Delay)
	}
	return nil
}

func already(dir string, song Song, override map[string]string) bool {
	orig := ApplyOverride(song, override)
	if orig != "" {
		name := library.CanonicalFile(orig, song.Title)
		if name != "" {
			if _, err := os.Stat(filepath.Join(dir, "tabs", name)); err == nil {
				return true
			}
		}
		return false
	}
	for _, a := range []string{song.OriginalArtist, song.CoverArtist} {
		if a == "" {
			continue
		}
		name := library.CanonicalFile(a, song.Title)
		if name == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "tabs", name)); err == nil {
			return true
		}
	}
	return false
}

func gatherOne(c *client.Client, song Song, opt GatherOpts) error {
	original := ApplyOverride(song, opt.Override)
	if original == "" {
		original = song.OriginalArtist
	}
	if original == "" && opt.Wiki != nil {
		got, err := opt.Wiki.Resolve(song.Title, song.CoverArtist)
		if err != nil {
			opt.Log(fmt.Sprintf("  wiki %s: %v", song.Title, err))
		} else {
			original = got
		}
	}
	q := song.Title
	if original != "" {
		q = original + " " + song.Title
	}
	page, err := c.Search(q, 1)
	if err != nil {
		return err
	}
	if original == "" {
		q = song.CoverArtist + " " + song.Title
		page, err = c.Search(q, 1)
		if err != nil {
			return err
		}
	}
	filtered := FilterResults(page.Results, original, song.CoverArtist)
	if len(filtered) == 0 {
		return fmt.Errorf("no chords")
	}
	if len(filtered) > opt.MaxFetch {
		filtered = filtered[:opt.MaxFetch]
	}
	var tabs []client.TabDetail
	for _, r := range filtered {
		time.Sleep(opt.Delay)
		tab, err := c.Tab(r.Path)
		if err != nil {
			opt.Log(fmt.Sprintf("  tab miss %s: %v", r.Path, err))
			continue
		}
		if tab.Type == "" {
			tab.Type = r.Type
		}
		if tab.Rating == 0 {
			tab.Rating = r.Rating
		}
		tabs = append(tabs, tab)
	}
	winner, ok := Pick(tabs)
	if !ok {
		return fmt.Errorf("no usable sheet")
	}
	label := original
	if label == "" {
		label = song.CoverArtist
	}
	rec := library.FromTab(winner)
	rec.Artist = label
	rec.Song = song.Title
	rec.OriginalArtist = original
	rec.CoverArtist = song.CoverArtist
	rec.SheetArtist = winner.Artist
	rec.SourceRank = song.Rank
	if rec.Type == "" {
		rec.Type = "Chords"
	}
	opt.Log(fmt.Sprintf("ok %d orig=%q sheet=%q %s", song.Rank, rec.OriginalArtist, rec.SheetArtist, rec.Song))
	return library.SaveCanonical(opt.OutDir, rec)
}
