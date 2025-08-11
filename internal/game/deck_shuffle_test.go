package game

import "testing"

// TestShuffleDeckDeterministic ensures that shuffling with the same seed
// produces the same card order.
func TestShuffleDeckDeterministic(t *testing.T) {
	SetSeed(99)
	d1 := NewDeck()
	ShuffleDeck(d1)

	SetSeed(99)
	d2 := NewDeck()
	ShuffleDeck(d2)

	for i := range d1 {
		if d1[i] != d2[i] {
			t.Fatalf("expected deterministic shuffle, card %d differs", i)
		}
	}
}
