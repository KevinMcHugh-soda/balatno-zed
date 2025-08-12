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

	chips, mult := CalculateJokerHandBonus([]Joker{chipJoker}, "Pair")
	if chips != 30 || mult != 0 {
		t.Fatalf("expected 30 chips bonus, got chips=%d mult=%d", chips, mult)
	}

	chips, mult = CalculateJokerHandBonus([]Joker{multJoker}, "Pair")
	if chips != 0 || mult != 5 {
		t.Fatalf("expected mult bonus 5, got chips=%d mult=%d", chips, mult)
	}

	// Non-matching hand should yield no bonus
	chips, mult = CalculateJokerHandBonus([]Joker{chipJoker}, "High Card")
	if chips != 0 || mult != 0 {
		t.Fatalf("expected no bonus for non-matching hand, got chips=%d mult=%d", chips, mult)
	}
}

// TestCompositeJoker verifies that a joker can apply multiple effects.
func TestCompositeJoker(t *testing.T) {
	cfg := JokerConfig{
		Name: "Combo",
		Effects: []JokerEffectConfig{
			{Effect: AddChips, EffectMagnitude: 10, HandMatchingRule: ContainsPair},
			{Effect: AddMult, EffectMagnitude: 3, HandMatchingRule: ContainsPair},
			{Effect: AddMoney, EffectMagnitude: 2, HandMatchingRule: None},
		},
	}
	j := createJokerFromConfig(cfg)

	chips, mult := CalculateJokerHandBonus([]Joker{j}, "Pair")
	if chips != 10 || mult != 3 {
		t.Fatalf("expected chips=10 mult=3, got chips=%d mult=%d", chips, mult)
	}

	reward := CalculateJokerRewards([]Joker{j})
	if reward != 2 {
		t.Fatalf("expected money reward 2, got %d", reward)
	}
}
