package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"geetard/internal/client"
	"geetard/internal/pack"
)

func main() {
	seedPath := flag.String("seed", "testdata/cover-songs-seed.json", "seed JSON")
	overridePath := flag.String("override", "testdata/originals-override.json", "title → original artist")
	outDir := flag.String("out", "corpus", "library output directory")
	limit := flag.Int("limit", 0, "max songs (0 = all)")
	delayMS := flag.Int("delay-ms", 400, "pause between HTTP calls")
	flag.Parse()

	seed, err := pack.LoadSeed(*seedPath)
	if err != nil {
		log.Fatal(err)
	}
	over, err := pack.LoadOverride(*overridePath)
	if err != nil {
		log.Fatal(err)
	}
	opt := pack.GatherOpts{
		OutDir:   *outDir,
		Limit:    *limit,
		Delay:    time.Duration(*delayMS) * time.Millisecond,
		Log:      func(s string) { fmt.Fprintln(os.Stderr, s) },
		Wiki:     &pack.WikiClient{Sleep: 2 * time.Second},
		Override: over,
	}
	c := client.FromEnv()
	fmt.Fprintf(os.Stderr, "packing %d songs from %s → %s\n", seed.Count, *seedPath, *outDir)
	if err := pack.Gather(c, seed, opt); err != nil {
		log.Fatal(err)
	}
}
