package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", 0, "Set random seed for reproducible gameplay (0 for random)")
	tui := flag.Bool("tui", false, "Run in TUI mode instead of console mode")
	flag.Parse()

	// Set seed if provided
	if *seed != 0 {
		SetSeed(*seed)
		fmt.Printf("Using seed: %d\n", *seed)
	}

	// Run in TUI mode or console mode
	if *tui {
		if err := RunTUI(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	} else {
		// Create logger for console mode
		logger := log.New(os.Stdout, "", log.LstdFlags)

		// Create GameIO for console mode
		gameIO := NewLoggerIO(logger)

		// Create and run the game
		game := NewGame(gameIO)
		game.Run()
	}
}
