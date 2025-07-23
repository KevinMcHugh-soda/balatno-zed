package main

import (
	"flag"
	"log"
)

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", 0, "Set random seed for reproducible gameplay (0 for random)")
	flag.Parse()

	// Set seed if provided
	if *seed != 0 {
		SetSeed(*seed)
	}

	// Create the game
	game := NewGame()

	// Create and run the terminal UI
	ui, err := NewTerminalUI(game)
	if err != nil {
		log.Fatalf("Failed to initialize terminal UI: %v", err)
	}

	ui.Run()
}
