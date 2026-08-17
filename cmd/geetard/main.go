package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"geetard/internal/client"
	"geetard/internal/config"
	"geetard/internal/tui"
)

func main() {
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
