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
	tui := flag.Bool("tui", false, "Run in TUI mode instead of console mode")
	flag.Parse()

	// Set seed if provided
	if *seed != 0 {
		game.SetSeed(*seed)
		fmt.Printf("Using seed: %d\n", *seed)
	}

	// Run in TUI mode or console mode
	if *tui {
		if err := ui.RunTUI(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	} else {
		// Create event handler for console mode
		eventHandler := game.NewLoggerEventHandler()

		// Create and run the game
		g := game.NewGame(eventHandler)
		g.Run()
	}
}
