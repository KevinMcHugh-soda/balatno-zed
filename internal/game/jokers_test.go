package game

import "testing"

// TestCalculateJokerRewards verifies AddMoney joker rewards at blind end.
func TestCalculateJokerRewards(t *testing.T) {
	j := Joker{OnBlindEnd: func() int { return 4 }}
	if got := CalculateJokerRewards([]Joker{j}); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

// TestCalculateJokerHandBonus verifies chip and multiplier bonuses from jokers.
func TestCalculateJokerHandBonus(t *testing.T) {
	chipCfg := JokerConfig{Name: "Chip", Effect: AddChips, EffectMagnitude: 30, HandMatchingRule: ContainsPair}
	chipJoker := createJokerFromConfig(chipCfg)

	multCfg := JokerConfig{Name: "Mult", Effect: AddMult, EffectMagnitude: 5, HandMatchingRule: ContainsPair}
	multJoker := createJokerFromConfig(multCfg)

	chips, mult := CalculateJokerHandBonus([]Joker{chipJoker}, "Pair", []Card{})
	if chips != 30 || mult != 0 {
		t.Fatalf("expected 30 chips bonus, got chips=%d mult=%d", chips, mult)
	}

	chips, mult = CalculateJokerHandBonus([]Joker{multJoker}, "Pair", []Card{})
	if chips != 0 || mult != 5 {
		t.Fatalf("expected mult bonus 5, got chips=%d mult=%d", chips, mult)
	}

	// Non-matching hand should yield no bonus
	chips, mult = CalculateJokerHandBonus([]Joker{chipJoker}, "High Card", []Card{})
	if chips != 0 || mult != 0 {
		t.Fatalf("expected no bonus for non-matching hand, got chips=%d mult=%d", chips, mult)
	}
}

// TestCardMatchingRule verifies bonuses based on individual card matches.
func TestCardMatchingRule(t *testing.T) {
	cfg := JokerConfig{Name: "Ace Bonus", Effect: AddChips, EffectMagnitude: 10, CardMatchingRule: CardIsAce}
	joker := createJokerFromConfig(cfg)

	hand := []Card{{Rank: Ace, Suit: Hearts}, {Rank: Ace, Suit: Spades}, {Rank: Two, Suit: Clubs}}
	chips, mult := CalculateJokerHandBonus([]Joker{joker}, "High Card", hand)
	if chips != 20 || mult != 0 {
		t.Fatalf("expected 20 chips bonus, got chips=%d mult=%d", chips, mult)
	}

	hand = []Card{{Rank: Two, Suit: Clubs}}
	chips, mult = CalculateJokerHandBonus([]Joker{joker}, "High Card", hand)
	if chips != 0 || mult != 0 {
		t.Fatalf("expected no bonus without matching cards, got chips=%d mult=%d", chips, mult)
	}
}

// TestReplayFaceCards verifies that ReplayCard jokers process matching cards twice.
func TestReplayFaceCards(t *testing.T) {
	replayCfg := JokerConfig{Name: "Face Dancer", Effect: ReplayCard, CardMatchingRule: CardIsFace}
	replayJoker := createJokerFromConfig(replayCfg)
	bonusCfg := JokerConfig{Name: "Face Bonus", Effect: AddChips, EffectMagnitude: 10, CardMatchingRule: CardIsFace}
	bonusJoker := createJokerFromConfig(bonusCfg)

	cards := []Card{{Rank: Jack, Suit: Hearts}, {Rank: Five, Suit: Clubs}}
	hand := Hand{Cards: cards}
	evaluator, _, cardValues, baseScore := EvaluateHand(hand)

	cardsForJokers, extraValue := ApplyReplayCardEffects([]Joker{replayJoker, bonusJoker}, cards)
	cardValues += extraValue
	chips, mult := CalculateJokerHandBonus([]Joker{replayJoker, bonusJoker}, evaluator.Name(), cardsForJokers)
	finalBase := baseScore + chips
	finalMult := evaluator.Multiplier() + mult
	finalScore := (finalBase + cardValues) * finalMult

	if finalScore != 50 {
		t.Fatalf("expected final score 50, got %d", finalScore)
	}
}
