package game_test

import (
	"testing"

	game "balatno/internal/game"
)

func TestCardString(t *testing.T) {
	tests := []struct {
		card     game.Card
		expected string
	}{
		{game.Card{Rank: game.Ace, Suit: game.Hearts}, "A♥"},
		{game.Card{Rank: game.King, Suit: game.Spades}, "K♠"},
		{game.Card{Rank: game.Ten, Suit: game.Diamonds}, "10♦"},
		{game.Card{Rank: game.Two, Suit: game.Clubs}, "2♣"},
	}

	for _, tt := range tests {
		if got := tt.card.String(); got != tt.expected {
			t.Errorf("game.Card.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestCardValue(t *testing.T) {
	tests := []struct {
		rank     game.Rank
		expected int
	}{
		{game.Ace, 11},
		{game.King, 10},
		{game.Queen, 10},
		{game.Jack, 10},
		{game.Ten, 10},
		{game.Nine, 9},
		{game.Two, 2},
	}

	for _, tt := range tests {
		if got := tt.rank.Value(); got != tt.expected {
			t.Errorf("game.Rank(%v).Value() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestNewDeck(t *testing.T) {
	deck := game.NewDeck()

	// Should have 52 cards
	if len(deck) != 52 {
		t.Errorf("game.NewDeck() length = %v, want 52", len(deck))
	}

	// Should have all unique cards
	seen := make(map[game.Card]bool)
	for _, card := range deck {
		if seen[card] {
			t.Errorf("Duplicate card found: %v", card)
		}
		seen[card] = true
	}

	// Should have 13 cards of each suit
	suitCounts := make(map[game.Suit]int)
	for _, card := range deck {
		suitCounts[card.Suit]++
	}

	for suit := game.Hearts; suit <= game.Spades; suit++ {
		if suitCounts[suit] != 13 {
			t.Errorf("game.Suit %v count = %v, want 13", suit, suitCounts[suit])
		}
	}

	// Should have 4 cards of each rank
	rankCounts := make(map[game.Rank]int)
	for _, card := range deck {
		rankCounts[card.Rank]++
	}

	for rank := game.Ace; rank <= game.King; rank++ {
		if rankCounts[rank] != 4 {
			t.Errorf("game.Rank %v count = %v, want 4", rank, rankCounts[rank])
		}
	}
}

func TestShuffleDeck(t *testing.T) {
	// Test with fixed seed for reproducibility
	game.SetSeed(42)
	deck1 := game.NewDeck()
	original := make([]game.Card, len(deck1))
	copy(original, deck1)

	game.ShuffleDeck(deck1)

	// Should still have same cards
	if len(deck1) != 52 {
		t.Errorf("Shuffled deck length = %v, want 52", len(deck1))
	}

	// Should be in different order (very unlikely to be same with good shuffle)
	same := true
	for i := range deck1 {
		if deck1[i] != original[i] {
			same = false
			break
		}
	}
	if same {
		t.Errorf("Shuffled deck is identical to original (very unlikely)")
	}

	// Should have same cards (different order)
	deck1Map := make(map[game.Card]int)
	originalMap := make(map[game.Card]int)
	for _, card := range deck1 {
		deck1Map[card]++
	}
	for _, card := range original {
		originalMap[card]++
	}

	for card, count := range originalMap {
		if deck1Map[card] != count {
			t.Errorf("game.Card %v count after shuffle = %v, want %v", card, deck1Map[card], count)
		}
	}
}

func TestEvaluateHandHighCard(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Two, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.Nine, Suit: game.Diamonds},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(high card) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("game.EvaluateHand(high card) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 18 { // 2 + 7 + 9
		t.Errorf("game.EvaluateHand(high card) cardValues = %v, want 18", cardValues)
	}
	if score != 23 { // (5 + 18) * 1
		t.Errorf("game.EvaluateHand(high card) score = %v, want 23", score)
	}
}

func TestEvaluateHandPair(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.King, Suit: game.Diamonds},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Pair" {
		t.Errorf("game.EvaluateHand(pair) handType = %v, want Pair", evaluator.Name())
	}
	if baseScore != 10 {
		t.Errorf("game.EvaluateHand(pair) baseScore = %v, want 10", baseScore)
	}
	if cardValues != 24 { // 7 + 7 + 10
		t.Errorf("game.EvaluateHand(pair) cardValues = %v, want 24", cardValues)
	}
	if score != 68 { // (10 + 24) * 2
		t.Errorf("game.EvaluateHand(pair) score = %v, want 68", score)
	}
}

func TestEvaluateHandTwoPair(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.King, Suit: game.Diamonds},
		{Rank: game.King, Suit: game.Spades},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Two Pair" {
		t.Errorf("game.EvaluateHand(two pair) handType = %v, want Two Pair", evaluator.Name())
	}
	if baseScore != 20 {
		t.Errorf("game.EvaluateHand(two pair) baseScore = %v, want 20", baseScore)
	}
	if cardValues != 34 { // 7 + 7 + 10 + 10
		t.Errorf("game.EvaluateHand(two pair) cardValues = %v, want 34", cardValues)
	}
	if score != 108 { // (20 + 34) * 2
		t.Errorf("game.EvaluateHand(two pair) score = %v, want 108", score)
	}
}

func TestEvaluateHandTwoPairLevel2(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.King, Suit: game.Diamonds},
		{Rank: game.King, Suit: game.Spades},
	}}

	levels := map[string]int{"Two Pair": 2}
	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, levels)

	if evaluator.Name() != "Two Pair" {
		t.Errorf("game.EvaluateHand(two pair lvl2) handType = %v, want Two Pair", evaluator.Name())
	}
	if baseScore != 25 {
		t.Errorf("game.EvaluateHand(two pair lvl2) baseScore = %v, want 25", baseScore)
	}
	if cardValues != 34 { // 7 + 7 + 10 + 10
		t.Errorf("game.EvaluateHand(two pair lvl2) cardValues = %v, want 34", cardValues)
	}
	if score != 118 { // (25 + 34) * 2
		t.Errorf("game.EvaluateHand(two pair lvl2) score = %v, want 118", score)
	}
}

func TestEvaluateHandThreeOfAKind(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.Seven, Suit: game.Diamonds},
		{Rank: game.King, Suit: game.Spades},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Three of a Kind" {
		t.Errorf("game.EvaluateHand(three of a kind) handType = %v, want Three of a Kind", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("game.EvaluateHand(three of a kind) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 31 { // 7 + 7 + 7 + 10
		t.Errorf("game.EvaluateHand(three of a kind) cardValues = %v, want 31", cardValues)
	}
	if score != 183 { // (30 + 31) * 3
		t.Errorf("game.EvaluateHand(three of a kind) score = %v, want 183", score)
	}
}

func TestEvaluateHandStraight(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Five, Suit: game.Hearts},
		{Rank: game.Six, Suit: game.Clubs},
		{Rank: game.Seven, Suit: game.Diamonds},
		{Rank: game.Eight, Suit: game.Spades},
		{Rank: game.Nine, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Straight" {
		t.Errorf("game.EvaluateHand(straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("game.EvaluateHand(straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 35 { // 5 + 6 + 7 + 8 + 9
		t.Errorf("game.EvaluateHand(straight) cardValues = %v, want 35", cardValues)
	}
	if score != 260 { // (30 + 35) * 4
		t.Errorf("game.EvaluateHand(straight) score = %v, want 260", score)
	}
}

func TestEvaluateHandFlush(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Two, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Nine, Suit: game.Hearts},
		{Rank: game.Jack, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Flush" {
		t.Errorf("game.EvaluateHand(flush) handType = %v, want Flush", evaluator.Name())
	}
	if baseScore != 35 {
		t.Errorf("game.EvaluateHand(flush) baseScore = %v, want 35", baseScore)
	}
	if cardValues != 38 { // 2 + 7 + 9 + 10 + 10
		t.Errorf("game.EvaluateHand(flush) cardValues = %v, want 38", cardValues)
	}
	if score != 292 { // (35 + 38) * 4
		t.Errorf("game.EvaluateHand(flush) score = %v, want 292", score)
	}
}

func TestEvaluateHandFullHouse(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.Seven, Suit: game.Diamonds},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.King, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Full House" {
		t.Errorf("game.EvaluateHand(full house) handType = %v, want Full House", evaluator.Name())
	}
	if baseScore != 40 {
		t.Errorf("game.EvaluateHand(full house) baseScore = %v, want 40", baseScore)
	}
	if cardValues != 41 { // 7 + 7 + 7 + 10 + 10
		t.Errorf("game.EvaluateHand(full house) cardValues = %v, want 41", cardValues)
	}
	if score != 324 { // (40 + 41) * 4
		t.Errorf("game.EvaluateHand(full house) score = %v, want 324", score)
	}
}

func TestEvaluateHandFourOfAKind(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Clubs},
		{Rank: game.Seven, Suit: game.Diamonds},
		{Rank: game.Seven, Suit: game.Spades},
		{Rank: game.King, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Four of a Kind" {
		t.Errorf("game.EvaluateHand(four of a kind) handType = %v, want Four of a Kind", evaluator.Name())
	}
	if baseScore != 60 {
		t.Errorf("game.EvaluateHand(four of a kind) baseScore = %v, want 60", baseScore)
	}
	if cardValues != 38 { // 7 + 7 + 7 + 7 + 10
		t.Errorf("game.EvaluateHand(four of a kind) cardValues = %v, want 38", cardValues)
	}
	if score != 686 { // (60 + 38) * 7
		t.Errorf("game.EvaluateHand(four of a kind) score = %v, want 686", score)
	}
}

func TestEvaluateHandStraightFlush(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Five, Suit: game.Hearts},
		{Rank: game.Six, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Eight, Suit: game.Hearts},
		{Rank: game.Nine, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Straight Flush" {
		t.Errorf("game.EvaluateHand(straight flush) handType = %v, want Straight Flush", evaluator.Name())
	}
	if baseScore != 100 {
		t.Errorf("game.EvaluateHand(straight flush) baseScore = %v, want 100", baseScore)
	}
	if cardValues != 35 { // 5 + 6 + 7 + 8 + 9
		t.Errorf("game.EvaluateHand(straight flush) cardValues = %v, want 35", cardValues)
	}
	if score != 1080 { // (100 + 35) * 8
		t.Errorf("game.EvaluateHand(straight flush) score = %v, want 1080", score)
	}
}

func TestEvaluateHandRoyalFlush(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ten, Suit: game.Hearts},
		{Rank: game.Jack, Suit: game.Hearts},
		{Rank: game.Queen, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Hearts},
		{Rank: game.Ace, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Royal Flush" {
		t.Errorf("game.EvaluateHand(royal flush) handType = %v, want Royal Flush", evaluator.Name())
	}
	if baseScore != 100 {
		t.Errorf("game.EvaluateHand(royal flush) baseScore = %v, want 100", baseScore)
	}
	if cardValues != 51 { // 11 + 10 + 10 + 10 + 10
		t.Errorf("game.EvaluateHand(royal flush) cardValues = %v, want 51", cardValues)
	}
	if score != 1208 { // (100 + 51) * 8
		t.Errorf("game.EvaluateHand(royal flush) score = %v, want 1208", score)
	}
}

func TestEvaluateHandSingleCard(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(single card) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("game.EvaluateHand(single card) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 11 { // game.Ace = 11
		t.Errorf("game.EvaluateHand(single card) cardValues = %v, want 11", cardValues)
	}
	if score != 16 { // (5 + 11) * 1
		t.Errorf("game.EvaluateHand(single card) score = %v, want 16", score)
	}
}

func TestEvaluateHandEmptyHand(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(empty hand) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 0 {
		t.Errorf("game.EvaluateHand(empty hand) baseScore = %v, want 0", baseScore)
	}
	if cardValues != 0 {
		t.Errorf("game.EvaluateHand(empty hand) cardValues = %v, want 0", cardValues)
	}
	if score != 0 {
		t.Errorf("game.EvaluateHand(empty hand) score = %v, want 0", score)
	}
}

func TestEvaluateHandWheelStraight(t *testing.T) {
	// Test A-2-3-4-5 straight (wheel)
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.Two, Suit: game.Clubs},
		{Rank: game.Three, Suit: game.Diamonds},
		{Rank: game.Four, Suit: game.Spades},
		{Rank: game.Five, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Straight" {
		t.Errorf("game.EvaluateHand(wheel straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("game.EvaluateHand(wheel straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 25 { // 11 + 2 + 3 + 4 + 5
		t.Errorf("game.EvaluateHand(wheel straight) cardValues = %v, want 25", cardValues)
	}
	if score != 220 { // (30 + 25) * 4
		t.Errorf("game.EvaluateHand(wheel straight) score = %v, want 220", score)
	}
}

func TestHandEvaluatorNames(t *testing.T) {
	tests := []struct {
		evaluator game.HandEvaluator
		expected  string
	}{
		{&game.HighCardEvaluator{}, "High Card"},
		{&game.PairEvaluator{}, "Pair"},
		{&game.TwoPairEvaluator{}, "Two Pair"},
		{&game.ThreeOfAKindEvaluator{}, "Three of a Kind"},
		{&game.StraightEvaluator{}, "Straight"},
		{&game.FlushEvaluator{}, "Flush"},
		{&game.FullHouseEvaluator{}, "Full House"},
		{&game.FourOfAKindEvaluator{}, "Four of a Kind"},
		{&game.StraightFlushEvaluator{}, "Straight Flush"},
		{&game.RoyalFlushEvaluator{}, "Royal Flush"},
	}

	for _, tt := range tests {
		if got := tt.evaluator.Name(); got != tt.expected {
			t.Errorf("game.HandEvaluator.Name() = %v, want %v", got, tt.expected)
		}
	}
}

func TestSetSeed(t *testing.T) {
	// Test that setting seed produces reproducible results
	game.SetSeed(123)
	deck1 := game.NewDeck()
	game.ShuffleDeck(deck1)

	game.SetSeed(123)
	deck2 := game.NewDeck()
	game.ShuffleDeck(deck2)

	// Both decks should be identical after shuffling with same seed
	for i := range deck1 {
		if deck1[i] != deck2[i] {
			t.Errorf("Decks with same seed differ at position %v: %v vs %v", i, deck1[i], deck2[i])
		}
	}

	// Different seed should produce different result
	game.SetSeed(456)
	deck3 := game.NewDeck()
	game.ShuffleDeck(deck3)

	same := true
	for i := range deck1 {
		if deck1[i] != deck3[i] {
			same = false
			break
		}
	}
	if same {
		t.Errorf("Decks with different seeds are identical (very unlikely)")
	}
}

func TestHandString(t *testing.T) {
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Queen, Suit: game.Diamonds},
	}}

	expected := "A♥ K♠ Q♦"
	if got := hand.String(); got != expected {
		t.Errorf("game.Hand.String() = %v, want %v", got, expected)
	}

	// Test empty hand
	emptyHand := game.Hand{Cards: []game.Card{}}
	if got := emptyHand.String(); got != "" {
		t.Errorf("Empty game.Hand.String() = %v, want empty string", got)
	}
}

func TestSuitString(t *testing.T) {
	tests := []struct {
		suit     game.Suit
		expected string
	}{
		{game.Hearts, "♥"},
		{game.Diamonds, "♦"},
		{game.Clubs, "♣"},
		{game.Spades, "♠"},
		{game.Suit(99), "?"}, // Invalid suit
	}

	for _, tt := range tests {
		if got := tt.suit.String(); got != tt.expected {
			t.Errorf("game.Suit(%v).String() = %v, want %v", tt.suit, got, tt.expected)
		}
	}
}

func TestRankString(t *testing.T) {
	tests := []struct {
		rank     game.Rank
		expected string
	}{
		{game.Ace, "A"},
		{game.Jack, "J"},
		{game.Queen, "Q"},
		{game.King, "K"},
		{game.Ten, "10"},
		{game.Nine, "9"},
		{game.Two, "2"},
	}

	for _, tt := range tests {
		if got := tt.rank.String(); got != tt.expected {
			t.Errorf("game.Rank(%v).String() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestEvaluateHandTwoCards(t *testing.T) {
	// Test with just two cards
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Spades},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(two cards) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("game.EvaluateHand(two cards) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 21 { // 11 + 10
		t.Errorf("game.EvaluateHand(two cards) cardValues = %v, want 21", cardValues)
	}
	if score != 26 { // (5 + 21) * 1
		t.Errorf("game.EvaluateHand(two cards) score = %v, want 26", score)
	}
}

func TestEvaluateHandThreeCards(t *testing.T) {
	// Test with three cards - no pair
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Queen, Suit: game.Diamonds},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(three cards) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("game.EvaluateHand(three cards) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 31 { // 11 + 10 + 10
		t.Errorf("game.EvaluateHand(three cards) cardValues = %v, want 31", cardValues)
	}
	if score != 36 { // (5 + 31) * 1
		t.Errorf("game.EvaluateHand(three cards) score = %v, want 36", score)
	}
}

func TestEvaluateHandFourCards(t *testing.T) {
	// Test with four cards - flush but less than 5 cards
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Two, Suit: game.Hearts},
		{Rank: game.Seven, Suit: game.Hearts},
		{Rank: game.Nine, Suit: game.Hearts},
		{Rank: game.Jack, Suit: game.Hearts},
	}}

	evaluator, _, _, _, _ := game.EvaluateHand(hand, nil)

	// Should not be a flush since we need exactly 5 cards
	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(four cards same suit) handType = %v, want High Card", evaluator.Name())
	}
}

func TestEvaluateHandBroadwayStraight(t *testing.T) {
	// Test 10-J-Q-K-A straight (Broadway)
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Ten, Suit: game.Hearts},
		{Rank: game.Jack, Suit: game.Clubs},
		{Rank: game.Queen, Suit: game.Diamonds},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Ace, Suit: game.Hearts},
	}}

	evaluator, score, cardValues, baseScore, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "Straight" {
		t.Errorf("game.EvaluateHand(broadway straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("game.EvaluateHand(broadway straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 51 { // 10 + 10 + 10 + 10 + 11
		t.Errorf("game.EvaluateHand(broadway straight) cardValues = %v, want 51", cardValues)
	}
	if score != 324 { // (30 + 51) * 4
		t.Errorf("game.EvaluateHand(broadway straight) score = %v, want 324", score)
	}
}

func TestEvaluateHandAlmostStraight(t *testing.T) {
	// Test cards that are almost a straight but missing one
	hand := game.Hand{Cards: []game.Card{
		{Rank: game.Five, Suit: game.Hearts},
		{Rank: game.Six, Suit: game.Clubs},
		{Rank: game.Seven, Suit: game.Diamonds},
		{Rank: game.Nine, Suit: game.Spades}, // Missing 8
		{Rank: game.Ten, Suit: game.Hearts},
	}}

	evaluator, _, _, _, _ := game.EvaluateHand(hand, nil)

	if evaluator.Name() != "High Card" {
		t.Errorf("game.EvaluateHand(almost straight) handType = %v, want High Card", evaluator.Name())
	}
}

func TestHandEvaluatorPriority(t *testing.T) {
	// Test that evaluators have correct priorities
	tests := []struct {
		evaluator game.HandEvaluator
		expected  int
	}{
		{&game.HighCardEvaluator{}, 1},
		{&game.PairEvaluator{}, 2},
		{&game.TwoPairEvaluator{}, 3},
		{&game.ThreeOfAKindEvaluator{}, 4},
		{&game.StraightEvaluator{}, 5},
		{&game.FlushEvaluator{}, 6},
		{&game.FullHouseEvaluator{}, 7},
		{&game.FourOfAKindEvaluator{}, 8},
		{&game.StraightFlushEvaluator{}, 9},
		{&game.RoyalFlushEvaluator{}, 10},
	}

	for _, tt := range tests {
		if got := tt.evaluator.Priority(); got != tt.expected {
			t.Errorf("game.HandEvaluator.Priority() = %v, want %v", got, tt.expected)
		}
	}
}

func TestRemoveCards(t *testing.T) {
	// Test removing single card from middle
	cards := []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Queen, Suit: game.Diamonds},
		{Rank: game.Jack, Suit: game.Clubs},
	}

	result := game.RemoveCards(cards, []int{1}) // Remove game.King
	expected := []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.Queen, Suit: game.Diamonds},
		{Rank: game.Jack, Suit: game.Clubs},
	}

	if len(result) != len(expected) {
		t.Errorf("game.RemoveCards() length = %v, want %v", len(result), len(expected))
	}

	for i, card := range expected {
		if result[i] != card {
			t.Errorf("game.RemoveCards() result[%d] = %v, want %v", i, result[i], card)
		}
	}

	// Test removing multiple cards
	result2 := game.RemoveCards(cards, []int{0, 2}) // Remove game.Ace and game.Queen
	expected2 := []game.Card{
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Jack, Suit: game.Clubs},
	}

	if len(result2) != len(expected2) {
		t.Errorf("game.RemoveCards(multiple) length = %v, want %v", len(result2), len(expected2))
	}

	for i, card := range expected2 {
		if result2[i] != card {
			t.Errorf("game.RemoveCards(multiple) result[%d] = %v, want %v", i, result2[i], card)
		}
	}

	// Test removing from empty slice
	emptyCards := []game.Card{}
	result3 := game.RemoveCards(emptyCards, []int{0})
	if len(result3) != 0 {
		t.Errorf("game.RemoveCards(empty) length = %v, want 0", len(result3))
	}

	// Test removing with invalid indices (should not crash)
	result4 := game.RemoveCards(cards, []int{10, -1})
	if len(result4) != len(cards) {
		t.Errorf("game.RemoveCards(invalid indices) length = %v, want %v", len(result4), len(cards))
	}

	// Test removing all cards
	result5 := game.RemoveCards(cards, []int{0, 1, 2, 3})
	if len(result5) != 0 {
		t.Errorf("game.RemoveCards(all cards) length = %v, want 0", len(result5))
	}

	// Test removing cards from end
	result6 := game.RemoveCards(cards, []int{3}) // Remove game.Jack
	expected6 := []game.Card{
		{Rank: game.Ace, Suit: game.Hearts},
		{Rank: game.King, Suit: game.Spades},
		{Rank: game.Queen, Suit: game.Diamonds},
	}

	if len(result6) != len(expected6) {
		t.Errorf("game.RemoveCards(from end) length = %v, want %v", len(result6), len(expected6))
	}

	for i, card := range expected6 {
		if result6[i] != card {
			t.Errorf("game.RemoveCards(from end) result[%d] = %v, want %v", i, result6[i], card)
		}
	}
}

func TestReproducibleGameplay(t *testing.T) {
	// Test that the same seed produces identical game state
	game.SetSeed(12345)
	deck1 := game.NewDeck()
	game.ShuffleDeck(deck1)
	hand1 := deck1[:7] // First 7 cards

	game.SetSeed(12345)
	deck2 := game.NewDeck()
	game.ShuffleDeck(deck2)
	hand2 := deck2[:7] // First 7 cards

	// Should be identical
	for i := range hand1 {
		if hand1[i] != hand2[i] {
			t.Errorf("Reproducible gameplay failed at card %d: %v vs %v", i, hand1[i], hand2[i])
		}
	}
}
