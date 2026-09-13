package pack

import "fmt"

func FillOriginals(seed Seed, wiki *WikiClient, over map[string]string, logfn func(string), save func(Seed) error) (Seed, error) {
	if logfn == nil {
		logfn = func(string) {}
	}
	if save == nil {
		save = func(Seed) error { return nil }
	}
	filled, skipped, failed := 0, 0, 0
	for i := range seed.Songs {
		song := seed.Songs[i]
		if song.OriginalArtist != "" {
			skipped++
			continue
		}
		if got := ApplyOverride(song, over); got != "" {
			seed.Songs[i].OriginalArtist = got
			filled++
			logfn(fmt.Sprintf("ok %d %q → %s (override)", song.Rank, song.Title, got))
			if err := save(seed); err != nil {
				return seed, err
			}
			continue
		}
		if wiki == nil {
			failed++
			continue
		}
		got, err := wiki.Resolve(song.Title, song.CoverArtist)
		if err != nil {
			failed++
			logfn(fmt.Sprintf("fail %d %q: %v", song.Rank, song.Title, err))
			if err := save(seed); err != nil {
				return seed, err
			}
			continue
		}
		if got == "" {
			failed++
			logfn(fmt.Sprintf("miss %d %q (cover %s)", song.Rank, song.Title, song.CoverArtist))
			continue
		}
		seed.Songs[i].OriginalArtist = got
		filled++
		logfn(fmt.Sprintf("ok %d %q → %s", song.Rank, song.Title, got))
		if err := save(seed); err != nil {
			return seed, err
		}
	}
	logfn(fmt.Sprintf("done filled=%d skipped=%d unresolved=%d total=%d", filled, skipped, failed, len(seed.Songs)))
	return seed, save(seed)
}
