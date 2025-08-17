package game

import "testing"

// TestCalculateJokerRewards verifies AddMoney joker rewards at blind end.
func TestCalculateJokerRewards(t *testing.T) {
	j := Joker{Effects: []JokerEffectConfig{{Effect: AddMoney, EffectMagnitude: 4}}}
	if got := CalculateJokerRewards([]Joker{j}); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

// TestCalculateJokerHandBonus verifies chip and multiplier bonuses from jokers.
func TestCalculateJokerHandBonus(t *testing.T) {
	chipCfg := JokerConfig{Name: "Chip", Effects: []JokerEffectConfig{{Effect: AddChips, EffectMagnitude: 30, HandMatchingRule: ContainsPair}}}
	chipJoker := createJokerFromConfig(chipCfg)

	multCfg := JokerConfig{Name: "Mult", Effects: []JokerEffectConfig{{Effect: AddMult, EffectMagnitude: 5, HandMatchingRule: ContainsPair}}}
	multJoker := createJokerFromConfig(multCfg)

	chips, mult, factor := CalculateJokerHandBonus([]Joker{chipJoker}, "Pair", []Card{})
	if chips != 30 || mult != 0 || factor != 1 {
		t.Fatalf("expected 30 chips bonus, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}

	chips, mult, factor = CalculateJokerHandBonus([]Joker{multJoker}, "Pair", []Card{})
	if chips != 0 || mult != 5 || factor != 1 {
		t.Fatalf("expected mult bonus 5, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}

	// Non-matching hand should yield no bonus
	chips, mult, factor = CalculateJokerHandBonus([]Joker{chipJoker}, "High Card", []Card{})
	if chips != 0 || mult != 0 || factor != 1 {
		t.Fatalf("expected no bonus for non-matching hand, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}
}

// TestCardMatchingRule verifies bonuses based on individual card matches.
func TestCardMatchingRule(t *testing.T) {
	cfg := JokerConfig{Name: "Ace Bonus", Effects: []JokerEffectConfig{{Effect: AddChips, EffectMagnitude: 10, CardMatchingRule: CardIsAce}}}
	joker := createJokerFromConfig(cfg)

	hand := []Card{{Rank: Ace, Suit: Hearts}, {Rank: Ace, Suit: Spades}, {Rank: Two, Suit: Clubs}}
	chips, mult, factor := CalculateJokerHandBonus([]Joker{joker}, "High Card", hand)
	if chips != 20 || mult != 0 || factor != 1 {
		t.Fatalf("expected 20 chips bonus, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}

	hand = []Card{{Rank: Two, Suit: Clubs}}
	chips, mult, factor = CalculateJokerHandBonus([]Joker{joker}, "High Card", hand)
	if chips != 0 || mult != 0 || factor != 1 {
		t.Fatalf("expected no bonus without matching cards, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}
}

// TestReplayFaceCards verifies that ReplayCard jokers process matching cards twice.
func TestReplayFaceCards(t *testing.T) {
	replayCfg := JokerConfig{Name: "Face Dancer", Effects: []JokerEffectConfig{{Effect: ReplayCard, CardMatchingRule: CardIsFace}}}
	replayJoker := createJokerFromConfig(replayCfg)
	bonusCfg := JokerConfig{Name: "Face Bonus", Effects: []JokerEffectConfig{{Effect: AddChips, EffectMagnitude: 10, CardMatchingRule: CardIsFace}}}
	bonusJoker := createJokerFromConfig(bonusCfg)

	cards := []Card{{Rank: Jack, Suit: Hearts}, {Rank: Five, Suit: Clubs}}
	hand := Hand{Cards: cards}
	evaluator, _, cardValues, baseScore, baseMult := EvaluateHand(hand, nil)

	cardsForJokers, extraValue := ApplyReplayCardEffects([]Joker{replayJoker, bonusJoker}, cards)
	cardValues += extraValue

	chips, mult, factor := CalculateJokerHandBonus([]Joker{replayJoker, bonusJoker}, evaluator.Name(), cardsForJokers)
	finalBase := baseScore + chips
	finalMult := (evaluator.Multiplier() + mult) * factor
	finalScore := (finalBase + cardValues) * finalMult

	if finalScore != 50 {
		t.Fatalf("expected final score 50, got %d", finalScore)
	}
}

// TestCompositeJoker verifies that multiple effects on a single joker stack.
func TestCompositeJoker(t *testing.T) {
	cfg := JokerConfig{
		Name: "Combo",
		Effects: []JokerEffectConfig{
			{Effect: AddChips, EffectMagnitude: 10, HandMatchingRule: ContainsPair},
			{Effect: AddMult, EffectMagnitude: 2, HandMatchingRule: ContainsPair},
		},
	}
	joker := createJokerFromConfig(cfg)
	chips, mult, factor := CalculateJokerHandBonus([]Joker{joker}, "Pair", []Card{})
	if chips != 10 || mult != 2 || factor != 1 {
		t.Fatalf("expected chips=10 mult=2 factor=1, got chips=%d mult=%d factor=%d", chips, mult, factor)
	}
}

// TestMultiplyMult verifies jokers that multiply the multiplier.
func TestMultiplyMult(t *testing.T) {
	cfg := JokerConfig{Name: "Doubler", Effects: []JokerEffectConfig{{Effect: MultiplyMult, EffectMagnitude: 2, HandMatchingRule: ContainsPair}}}
	joker := createJokerFromConfig(cfg)

	_, mult, factor := CalculateJokerHandBonus([]Joker{joker}, "Pair", []Card{})
	if mult != 0 || factor != 2 {
		t.Fatalf("expected multiplier factor=2, got mult=%d factor=%d", mult, factor)
	}

	// Non-matching hand should not multiply
	_, mult, factor = CalculateJokerHandBonus([]Joker{joker}, "High Card", []Card{})
	if mult != 0 || factor != 1 {
		t.Fatalf("expected no effect for non-matching hand, got mult=%d factor=%d", mult, factor)
	}
}
