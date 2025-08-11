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
