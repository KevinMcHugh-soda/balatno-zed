package main

import (
	"testing"
)

func TestCardString(t *testing.T) {
	tests := []struct {
		card     Card
		expected string
	}{
		{Card{Rank: Ace, Suit: Hearts}, "A♥"},
		{Card{Rank: King, Suit: Spades}, "K♠"},
		{Card{Rank: Ten, Suit: Diamonds}, "10♦"},
		{Card{Rank: Two, Suit: Clubs}, "2♣"},
	}

	for _, tt := range tests {
		if got := tt.card.String(); got != tt.expected {
			t.Errorf("Card.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestCardValue(t *testing.T) {
	tests := []struct {
		rank     Rank
		expected int
	}{
		{Ace, 11},
		{King, 10},
		{Queen, 10},
		{Jack, 10},
		{Ten, 10},
		{Nine, 9},
		{Two, 2},
	}

	for _, tt := range tests {
		if got := tt.rank.Value(); got != tt.expected {
			t.Errorf("Rank(%v).Value() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	// Should have 52 cards
	if len(deck) != 52 {
		t.Errorf("NewDeck() length = %v, want 52", len(deck))
	}

	// Should have all unique cards
	seen := make(map[Card]bool)
	for _, card := range deck {
		if seen[card] {
			t.Errorf("Duplicate card found: %v", card)
		}
		seen[card] = true
	}

	// Should have 13 cards of each suit
	suitCounts := make(map[Suit]int)
	for _, card := range deck {
		suitCounts[card.Suit]++
	}

	for suit := Hearts; suit <= Spades; suit++ {
		if suitCounts[suit] != 13 {
			t.Errorf("Suit %v count = %v, want 13", suit, suitCounts[suit])
		}
	}

	// Should have 4 cards of each rank
	rankCounts := make(map[Rank]int)
	for _, card := range deck {
		rankCounts[card.Rank]++
	}

	for rank := Ace; rank <= King; rank++ {
		if rankCounts[rank] != 4 {
			t.Errorf("Rank %v count = %v, want 4", rank, rankCounts[rank])
		}
	}
}

func TestShuffleDeck(t *testing.T) {
	// Test with fixed seed for reproducibility
	SetSeed(42)
	deck1 := NewDeck()
	original := make([]Card, len(deck1))
	copy(original, deck1)

	ShuffleDeck(deck1)

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
	deck1Map := make(map[Card]int)
	originalMap := make(map[Card]int)
	for _, card := range deck1 {
		deck1Map[card]++
	}
	for _, card := range original {
		originalMap[card]++
	}

	for card, count := range originalMap {
		if deck1Map[card] != count {
			t.Errorf("Card %v count after shuffle = %v, want %v", card, deck1Map[card], count)
		}
	}
}

func TestEvaluateHandHighCard(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Two, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: Nine, Suit: Diamonds},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(high card) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("EvaluateHand(high card) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 18 { // 2 + 7 + 9
		t.Errorf("EvaluateHand(high card) cardValues = %v, want 18", cardValues)
	}
	if score != 23 { // (5 + 18) * 1
		t.Errorf("EvaluateHand(high card) score = %v, want 23", score)
	}
}

func TestEvaluateHandPair(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Seven, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: King, Suit: Diamonds},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Pair" {
		t.Errorf("EvaluateHand(pair) handType = %v, want Pair", evaluator.Name())
	}
	if baseScore != 10 {
		t.Errorf("EvaluateHand(pair) baseScore = %v, want 10", baseScore)
	}
	if cardValues != 24 { // 7 + 7 + 10
		t.Errorf("EvaluateHand(pair) cardValues = %v, want 24", cardValues)
	}
	if score != 68 { // (10 + 24) * 2
		t.Errorf("EvaluateHand(pair) score = %v, want 68", score)
	}
}

func TestEvaluateHandTwoPair(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Seven, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: King, Suit: Diamonds},
		{Rank: King, Suit: Spades},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Two Pair" {
		t.Errorf("EvaluateHand(two pair) handType = %v, want Two Pair", evaluator.Name())
	}
	if baseScore != 20 {
		t.Errorf("EvaluateHand(two pair) baseScore = %v, want 20", baseScore)
	}
	if cardValues != 34 { // 7 + 7 + 10 + 10
		t.Errorf("EvaluateHand(two pair) cardValues = %v, want 34", cardValues)
	}
	if score != 108 { // (20 + 34) * 2
		t.Errorf("EvaluateHand(two pair) score = %v, want 108", score)
	}
}

func TestEvaluateHandThreeOfAKind(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Seven, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: Seven, Suit: Diamonds},
		{Rank: King, Suit: Spades},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Three of a Kind" {
		t.Errorf("EvaluateHand(three of a kind) handType = %v, want Three of a Kind", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("EvaluateHand(three of a kind) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 31 { // 7 + 7 + 7 + 10
		t.Errorf("EvaluateHand(three of a kind) cardValues = %v, want 31", cardValues)
	}
	if score != 183 { // (30 + 31) * 3
		t.Errorf("EvaluateHand(three of a kind) score = %v, want 183", score)
	}
}

func TestEvaluateHandStraight(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Five, Suit: Hearts},
		{Rank: Six, Suit: Clubs},
		{Rank: Seven, Suit: Diamonds},
		{Rank: Eight, Suit: Spades},
		{Rank: Nine, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Straight" {
		t.Errorf("EvaluateHand(straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("EvaluateHand(straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 35 { // 5 + 6 + 7 + 8 + 9
		t.Errorf("EvaluateHand(straight) cardValues = %v, want 35", cardValues)
	}
	if score != 260 { // (30 + 35) * 4
		t.Errorf("EvaluateHand(straight) score = %v, want 260", score)
	}
}

func TestEvaluateHandFlush(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Two, Suit: Hearts},
		{Rank: Seven, Suit: Hearts},
		{Rank: Nine, Suit: Hearts},
		{Rank: Jack, Suit: Hearts},
		{Rank: King, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Flush" {
		t.Errorf("EvaluateHand(flush) handType = %v, want Flush", evaluator.Name())
	}
	if baseScore != 35 {
		t.Errorf("EvaluateHand(flush) baseScore = %v, want 35", baseScore)
	}
	if cardValues != 38 { // 2 + 7 + 9 + 10 + 10
		t.Errorf("EvaluateHand(flush) cardValues = %v, want 38", cardValues)
	}
	if score != 292 { // (35 + 38) * 4
		t.Errorf("EvaluateHand(flush) score = %v, want 292", score)
	}
}

func TestEvaluateHandFullHouse(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Seven, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: Seven, Suit: Diamonds},
		{Rank: King, Suit: Spades},
		{Rank: King, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Full House" {
		t.Errorf("EvaluateHand(full house) handType = %v, want Full House", evaluator.Name())
	}
	if baseScore != 40 {
		t.Errorf("EvaluateHand(full house) baseScore = %v, want 40", baseScore)
	}
	if cardValues != 41 { // 7 + 7 + 7 + 10 + 10
		t.Errorf("EvaluateHand(full house) cardValues = %v, want 41", cardValues)
	}
	if score != 324 { // (40 + 41) * 4
		t.Errorf("EvaluateHand(full house) score = %v, want 324", score)
	}
}

func TestEvaluateHandFourOfAKind(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Seven, Suit: Hearts},
		{Rank: Seven, Suit: Clubs},
		{Rank: Seven, Suit: Diamonds},
		{Rank: Seven, Suit: Spades},
		{Rank: King, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Four of a Kind" {
		t.Errorf("EvaluateHand(four of a kind) handType = %v, want Four of a Kind", evaluator.Name())
	}
	if baseScore != 60 {
		t.Errorf("EvaluateHand(four of a kind) baseScore = %v, want 60", baseScore)
	}
	if cardValues != 38 { // 7 + 7 + 7 + 7 + 10
		t.Errorf("EvaluateHand(four of a kind) cardValues = %v, want 38", cardValues)
	}
	if score != 686 { // (60 + 38) * 7
		t.Errorf("EvaluateHand(four of a kind) score = %v, want 686", score)
	}
}

func TestEvaluateHandStraightFlush(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Five, Suit: Hearts},
		{Rank: Six, Suit: Hearts},
		{Rank: Seven, Suit: Hearts},
		{Rank: Eight, Suit: Hearts},
		{Rank: Nine, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Straight Flush" {
		t.Errorf("EvaluateHand(straight flush) handType = %v, want Straight Flush", evaluator.Name())
	}
	if baseScore != 100 {
		t.Errorf("EvaluateHand(straight flush) baseScore = %v, want 100", baseScore)
	}
	if cardValues != 35 { // 5 + 6 + 7 + 8 + 9
		t.Errorf("EvaluateHand(straight flush) cardValues = %v, want 35", cardValues)
	}
	if score != 1080 { // (100 + 35) * 8
		t.Errorf("EvaluateHand(straight flush) score = %v, want 1080", score)
	}
}

func TestEvaluateHandRoyalFlush(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Ten, Suit: Hearts},
		{Rank: Jack, Suit: Hearts},
		{Rank: Queen, Suit: Hearts},
		{Rank: King, Suit: Hearts},
		{Rank: Ace, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Royal Flush" {
		t.Errorf("EvaluateHand(royal flush) handType = %v, want Royal Flush", evaluator.Name())
	}
	if baseScore != 100 {
		t.Errorf("EvaluateHand(royal flush) baseScore = %v, want 100", baseScore)
	}
	if cardValues != 51 { // 11 + 10 + 10 + 10 + 10
		t.Errorf("EvaluateHand(royal flush) cardValues = %v, want 51", cardValues)
	}
	if score != 1208 { // (100 + 51) * 8
		t.Errorf("EvaluateHand(royal flush) score = %v, want 1208", score)
	}
}

func TestEvaluateHandSingleCard(t *testing.T) {
	hand := Hand{Cards: []Card{
		{Rank: Ace, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(single card) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("EvaluateHand(single card) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 11 { // Ace = 11
		t.Errorf("EvaluateHand(single card) cardValues = %v, want 11", cardValues)
	}
	if score != 16 { // (5 + 11) * 1
		t.Errorf("EvaluateHand(single card) score = %v, want 16", score)
	}
}

func TestEvaluateHandEmptyHand(t *testing.T) {
	hand := Hand{Cards: []Card{}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(empty hand) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 0 {
		t.Errorf("EvaluateHand(empty hand) baseScore = %v, want 0", baseScore)
	}
	if cardValues != 0 {
		t.Errorf("EvaluateHand(empty hand) cardValues = %v, want 0", cardValues)
	}
	if score != 0 {
		t.Errorf("EvaluateHand(empty hand) score = %v, want 0", score)
	}
}

func TestEvaluateHandWheelStraight(t *testing.T) {
	// Test A-2-3-4-5 straight (wheel)
	hand := Hand{Cards: []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: Two, Suit: Clubs},
		{Rank: Three, Suit: Diamonds},
		{Rank: Four, Suit: Spades},
		{Rank: Five, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Straight" {
		t.Errorf("EvaluateHand(wheel straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("EvaluateHand(wheel straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 25 { // 11 + 2 + 3 + 4 + 5
		t.Errorf("EvaluateHand(wheel straight) cardValues = %v, want 25", cardValues)
	}
	if score != 220 { // (30 + 25) * 4
		t.Errorf("EvaluateHand(wheel straight) score = %v, want 220", score)
	}
}

func TestHandEvaluatorNames(t *testing.T) {
	tests := []struct {
		evaluator HandEvaluator
		expected  string
	}{
		{&HighCardEvaluator{}, "High Card"},
		{&PairEvaluator{}, "Pair"},
		{&TwoPairEvaluator{}, "Two Pair"},
		{&ThreeOfAKindEvaluator{}, "Three of a Kind"},
		{&StraightEvaluator{}, "Straight"},
		{&FlushEvaluator{}, "Flush"},
		{&FullHouseEvaluator{}, "Full House"},
		{&FourOfAKindEvaluator{}, "Four of a Kind"},
		{&StraightFlushEvaluator{}, "Straight Flush"},
		{&RoyalFlushEvaluator{}, "Royal Flush"},
	}

	for _, tt := range tests {
		if got := tt.evaluator.Name(); got != tt.expected {
			t.Errorf("HandEvaluator.Name() = %v, want %v", got, tt.expected)
		}
	}
}

func TestHandEvaluatorBaseScore(t *testing.T) {
	tests := []struct {
		evaluator HandEvaluator
		expected  int
	}{
		{&HighCardEvaluator{}, 5},
		{&PairEvaluator{}, 10},
		{&TwoPairEvaluator{}, 20},
		{&ThreeOfAKindEvaluator{}, 30},
		{&StraightEvaluator{}, 30},
		{&FlushEvaluator{}, 35},
		{&FullHouseEvaluator{}, 40},
		{&FourOfAKindEvaluator{}, 60},
		{&StraightFlushEvaluator{}, 100},
		{&RoyalFlushEvaluator{}, 100},
	}

	for _, tt := range tests {
		if got := tt.evaluator.BaseScore(); got != tt.expected {
			t.Errorf("HandEvaluator.BaseScore() = %v, want %v", got, tt.expected)
		}
	}
}

func TestHandEvaluatorMultiplier(t *testing.T) {
	tests := []struct {
		evaluator HandEvaluator
		expected  int
	}{
		{&HighCardEvaluator{}, 1},
		{&PairEvaluator{}, 2},
		{&TwoPairEvaluator{}, 2},
		{&ThreeOfAKindEvaluator{}, 3},
		{&StraightEvaluator{}, 4},
		{&FlushEvaluator{}, 4},
		{&FullHouseEvaluator{}, 4},
		{&FourOfAKindEvaluator{}, 7},
		{&StraightFlushEvaluator{}, 8},
		{&RoyalFlushEvaluator{}, 8},
	}

	for _, tt := range tests {
		if got := tt.evaluator.Multiplier(); got != tt.expected {
			t.Errorf("HandEvaluator.Multiplier() = %v, want %v", got, tt.expected)
		}
	}
}

func TestSetSeed(t *testing.T) {
	// Test that setting seed produces reproducible results
	SetSeed(123)
	deck1 := NewDeck()
	ShuffleDeck(deck1)

	SetSeed(123)
	deck2 := NewDeck()
	ShuffleDeck(deck2)

	// Both decks should be identical after shuffling with same seed
	for i := range deck1 {
		if deck1[i] != deck2[i] {
			t.Errorf("Decks with same seed differ at position %v: %v vs %v", i, deck1[i], deck2[i])
		}
	}

	// Different seed should produce different result
	SetSeed(456)
	deck3 := NewDeck()
	ShuffleDeck(deck3)

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
	hand := Hand{Cards: []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: King, Suit: Spades},
		{Rank: Queen, Suit: Diamonds},
	}}

	expected := "A♥ K♠ Q♦"
	if got := hand.String(); got != expected {
		t.Errorf("Hand.String() = %v, want %v", got, expected)
	}

	// Test empty hand
	emptyHand := Hand{Cards: []Card{}}
	if got := emptyHand.String(); got != "" {
		t.Errorf("Empty Hand.String() = %v, want empty string", got)
	}
}

func TestSuitString(t *testing.T) {
	tests := []struct {
		suit     Suit
		expected string
	}{
		{Hearts, "♥"},
		{Diamonds, "♦"},
		{Clubs, "♣"},
		{Spades, "♠"},
		{Suit(99), "?"}, // Invalid suit
	}

	for _, tt := range tests {
		if got := tt.suit.String(); got != tt.expected {
			t.Errorf("Suit(%v).String() = %v, want %v", tt.suit, got, tt.expected)
		}
	}
}

func TestRankString(t *testing.T) {
	tests := []struct {
		rank     Rank
		expected string
	}{
		{Ace, "A"},
		{Jack, "J"},
		{Queen, "Q"},
		{King, "K"},
		{Ten, "10"},
		{Nine, "9"},
		{Two, "2"},
	}

	for _, tt := range tests {
		if got := tt.rank.String(); got != tt.expected {
			t.Errorf("Rank(%v).String() = %v, want %v", tt.rank, got, tt.expected)
		}
	}
}

func TestEvaluateHandTwoCards(t *testing.T) {
	// Test with just two cards
	hand := Hand{Cards: []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: King, Suit: Spades},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(two cards) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("EvaluateHand(two cards) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 21 { // 11 + 10
		t.Errorf("EvaluateHand(two cards) cardValues = %v, want 21", cardValues)
	}
	if score != 26 { // (5 + 21) * 1
		t.Errorf("EvaluateHand(two cards) score = %v, want 26", score)
	}
}

func TestEvaluateHandThreeCards(t *testing.T) {
	// Test with three cards - no pair
	hand := Hand{Cards: []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: King, Suit: Spades},
		{Rank: Queen, Suit: Diamonds},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(three cards) handType = %v, want High Card", evaluator.Name())
	}
	if baseScore != 5 {
		t.Errorf("EvaluateHand(three cards) baseScore = %v, want 5", baseScore)
	}
	if cardValues != 31 { // 11 + 10 + 10
		t.Errorf("EvaluateHand(three cards) cardValues = %v, want 31", cardValues)
	}
	if score != 36 { // (5 + 31) * 1
		t.Errorf("EvaluateHand(three cards) score = %v, want 36", score)
	}
}

func TestEvaluateHandFourCards(t *testing.T) {
	// Test with four cards - flush but less than 5 cards
	hand := Hand{Cards: []Card{
		{Rank: Two, Suit: Hearts},
		{Rank: Seven, Suit: Hearts},
		{Rank: Nine, Suit: Hearts},
		{Rank: Jack, Suit: Hearts},
	}}

	evaluator, _, _, _ := EvaluateHand(hand)

	// Should not be a flush since we need exactly 5 cards
	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(four cards same suit) handType = %v, want High Card", evaluator.Name())
	}
}

func TestEvaluateHandBroadwayStraight(t *testing.T) {
	// Test 10-J-Q-K-A straight (Broadway)
	hand := Hand{Cards: []Card{
		{Rank: Ten, Suit: Hearts},
		{Rank: Jack, Suit: Clubs},
		{Rank: Queen, Suit: Diamonds},
		{Rank: King, Suit: Spades},
		{Rank: Ace, Suit: Hearts},
	}}

	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	if evaluator.Name() != "Straight" {
		t.Errorf("EvaluateHand(broadway straight) handType = %v, want Straight", evaluator.Name())
	}
	if baseScore != 30 {
		t.Errorf("EvaluateHand(broadway straight) baseScore = %v, want 30", baseScore)
	}
	if cardValues != 51 { // 10 + 10 + 10 + 10 + 11
		t.Errorf("EvaluateHand(broadway straight) cardValues = %v, want 51", cardValues)
	}
	if score != 324 { // (30 + 51) * 4
		t.Errorf("EvaluateHand(broadway straight) score = %v, want 324", score)
	}
}

func TestEvaluateHandAlmostStraight(t *testing.T) {
	// Test cards that are almost a straight but missing one
	hand := Hand{Cards: []Card{
		{Rank: Five, Suit: Hearts},
		{Rank: Six, Suit: Clubs},
		{Rank: Seven, Suit: Diamonds},
		{Rank: Nine, Suit: Spades}, // Missing 8
		{Rank: Ten, Suit: Hearts},
	}}

	evaluator, _, _, _ := EvaluateHand(hand)

	if evaluator.Name() != "High Card" {
		t.Errorf("EvaluateHand(almost straight) handType = %v, want High Card", evaluator.Name())
	}
}

func TestHandEvaluatorPriority(t *testing.T) {
	// Test that evaluators have correct priorities
	tests := []struct {
		evaluator HandEvaluator
		expected  int
	}{
		{&HighCardEvaluator{}, 1},
		{&PairEvaluator{}, 2},
		{&TwoPairEvaluator{}, 3},
		{&ThreeOfAKindEvaluator{}, 4},
		{&StraightEvaluator{}, 5},
		{&FlushEvaluator{}, 6},
		{&FullHouseEvaluator{}, 7},
		{&FourOfAKindEvaluator{}, 8},
		{&StraightFlushEvaluator{}, 9},
		{&RoyalFlushEvaluator{}, 10},
	}

	for _, tt := range tests {
		if got := tt.evaluator.Priority(); got != tt.expected {
			t.Errorf("HandEvaluator.Priority() = %v, want %v", got, tt.expected)
		}
	}
}

func TestRemoveCards(t *testing.T) {
	// Test removing single card from middle
	cards := []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: King, Suit: Spades},
		{Rank: Queen, Suit: Diamonds},
		{Rank: Jack, Suit: Clubs},
	}

	result := removeCards(cards, []int{1}) // Remove King
	expected := []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: Queen, Suit: Diamonds},
		{Rank: Jack, Suit: Clubs},
	}

	if len(result) != len(expected) {
		t.Errorf("removeCards() length = %v, want %v", len(result), len(expected))
	}

	for i, card := range expected {
		if result[i] != card {
			t.Errorf("removeCards() result[%d] = %v, want %v", i, result[i], card)
		}
	}

	// Test removing multiple cards
	result2 := removeCards(cards, []int{0, 2}) // Remove Ace and Queen
	expected2 := []Card{
		{Rank: King, Suit: Spades},
		{Rank: Jack, Suit: Clubs},
	}

	if len(result2) != len(expected2) {
		t.Errorf("removeCards(multiple) length = %v, want %v", len(result2), len(expected2))
	}

	for i, card := range expected2 {
		if result2[i] != card {
			t.Errorf("removeCards(multiple) result[%d] = %v, want %v", i, result2[i], card)
		}
	}

	// Test removing from empty slice
	emptyCards := []Card{}
	result3 := removeCards(emptyCards, []int{0})
	if len(result3) != 0 {
		t.Errorf("removeCards(empty) length = %v, want 0", len(result3))
	}

	// Test removing with invalid indices (should not crash)
	result4 := removeCards(cards, []int{10, -1})
	if len(result4) != len(cards) {
		t.Errorf("removeCards(invalid indices) length = %v, want %v", len(result4), len(cards))
	}

	// Test removing all cards
	result5 := removeCards(cards, []int{0, 1, 2, 3})
	if len(result5) != 0 {
		t.Errorf("removeCards(all cards) length = %v, want 0", len(result5))
	}

	// Test removing cards from end
	result6 := removeCards(cards, []int{3}) // Remove Jack
	expected6 := []Card{
		{Rank: Ace, Suit: Hearts},
		{Rank: King, Suit: Spades},
		{Rank: Queen, Suit: Diamonds},
	}

	if len(result6) != len(expected6) {
		t.Errorf("removeCards(from end) length = %v, want %v", len(result6), len(expected6))
	}

	for i, card := range expected6 {
		if result6[i] != card {
			t.Errorf("removeCards(from end) result[%d] = %v, want %v", i, result6[i], card)
		}
	}
}

func TestReproducibleGameplay(t *testing.T) {
	// Test that the same seed produces identical game state
	SetSeed(12345)
	deck1 := NewDeck()
	ShuffleDeck(deck1)
	hand1 := deck1[:7] // First 7 cards

	SetSeed(12345)
	deck2 := NewDeck()
	ShuffleDeck(deck2)
	hand2 := deck2[:7] // First 7 cards

	// Should be identical
	for i := range hand1 {
		if hand1[i] != hand2[i] {
			t.Errorf("Reproducible gameplay failed at card %d: %v vs %v", i, hand1[i], hand2[i])
		}
	}
}
