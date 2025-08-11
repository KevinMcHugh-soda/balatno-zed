package main

import (
	"flag"
	"fmt"

	game "balatno/internal/game"
	ui "balatno/internal/ui"
)

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", 0, "Set random seed for reproducible gameplay (0 for random)")
	load := flag.String("load", "", "Load game state from JSON file")
	tui := flag.Bool("tui", false, "Run in TUI mode instead of console mode")
	flag.Parse()

	// Run in TUI mode or console mode
	if *tui {
		if *load != "" {
			fmt.Println("Load flag currently only supported in console mode")
		}
		if err := ui.RunTUI(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	} else {
		// Create event handler for console mode
		eventHandler := game.NewLoggerEventHandler()

		var g *game.Game
		var err error

		if *load != "" {
			g, err = game.LoadGameFromFile(*load, eventHandler)
			if err != nil {
				fmt.Printf("Error loading game: %v\n", err)
				return
			}
			if *seed != 0 {
				fmt.Println("Seed flag ignored when loading game")
			}
		} else {
			if *seed != 0 {
				game.SetSeed(*seed)
				fmt.Printf("Using seed: %d\n", *seed)
			}
			g = game.NewGame(eventHandler)
		}

		// Run the game
		g.Run()
	}
}
