package main

import (
	"flag"
	"fmt"
)

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", 0, "Set random seed for reproducible gameplay (0 for random)")
	flag.Parse()

	// Set seed if provided
	if *seed != 0 {
		SetSeed(*seed)
		fmt.Printf("Using seed: %d\n", *seed)
	}

	// Create and run the game
	game := NewGame()
	game.Run()
}
