package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/tui"
)

func main() {
	if term := os.Getenv("TERM"); term == "" || term == "dumb" {
		_ = os.Setenv("TERM", "xterm-256color")
	}
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		fmt.Fprintln(os.Stderr, "geetard needs a real terminal.")
		fmt.Fprintln(os.Stderr, "On ChromeOS: Settings → Advanced → Developers → Linux, then open the Linux Terminal app and run geetard.")
		os.Exit(1)
	} else {
		_ = tty.Close()
	}

	dir := config.Dir()
	sel, warn, err := config.Load(dir)
	if err != nil {
		log.Fatal(err)
	}
	m := tui.New(client.FromEnv(), sel, dir)
	m.Status = warn
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Println("Unable to run tui:", err)
		os.Exit(1)
	}
}
