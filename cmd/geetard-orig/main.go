package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"geetard/internal/pack"
)

func main() {
	seedPath := flag.String("seed", "cover-songs-seed.json", "input seed JSON")
	outPath := flag.String("out", "originals.json", "output JSON (also used as resume if it exists)")
	overridePath := flag.String("override", "originals-override.json", "title → original artist")
	from := flag.Int("from", 0, "first rank inclusive (0 = start)")
	to := flag.Int("to", 0, "last rank inclusive (0 = end)")
	delayMS := flag.Int("delay-ms", 1200, "min milliseconds between Wikipedia requests")
	merge := flag.Bool("merge", false, "merge shard JSON files listed after flags into -out")
	flag.Parse()

	if *merge {
		if flag.NArg() < 1 {
			log.Fatal("usage: geetard-orig -merge -out originals.json shard-a.json shard-b.json")
		}
		var parts []pack.Seed
		for _, p := range flag.Args() {
			s, err := pack.LoadSeed(p)
			if err != nil {
				log.Fatal(p, err)
			}
			parts = append(parts, s)
		}
		out := pack.MergeSeeds(parts)
		if err := pack.SaveSeed(*outPath, out); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stderr, "merged %d songs → %s\n", out.Count, *outPath)
		return
	}

	seed, err := pack.LoadSeed(*seedPath)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(*outPath); err == nil {
		prev, err := pack.LoadSeed(*outPath)
		if err != nil {
			log.Fatal(err)
		}
		seed = pack.MergeSeeds([]pack.Seed{seed, prev})
		fmt.Fprintf(os.Stderr, "resuming from %s\n", *outPath)
	}
	over, err := pack.LoadOverride(*overridePath)
	if err != nil {
		log.Fatal(err)
	}

	work := pack.SliceByRank(seed, *from, *to)
	fmt.Fprintf(os.Stderr, "resolving %d songs (ranks %d–%d) via Wikipedia\n", work.Count, *from, *to)
	fmt.Fprintf(os.Stderr, "stay off VPN. Ctrl+C is safe; rerun the same command to resume.\n")

	wiki := &pack.WikiClient{Sleep: time.Duration(*delayMS) * time.Millisecond}
	save := func(s pack.Seed) error { return pack.SaveSeed(*outPath, s) }
	work, err = pack.FillOriginals(work, wiki, over, func(s string) { fmt.Fprintln(os.Stderr, s) }, save)
	if err != nil {
		log.Fatal(err)
	}
	n := 0
	for _, song := range work.Songs {
		if song.OriginalArtist != "" {
			n++
		}
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d/%d have original_artist)\n", *outPath, n, work.Count)
}
