package game_test

import (
	"testing"

	game "balatno/internal/game"
)

func TestGetAnteRequirement(t *testing.T) {
	tests := []struct {
		ante      int
		blindType game.BlindType
		expected  int
		name      string
	}{
		// Ante 1 tests
		{1, game.SmallBlind, 300, "Ante 1 Small Blind"},
		{1, game.BigBlind, 450, "Ante 1 Big Blind"},
		{1, game.BossBlind, 600, "Ante 1 Boss Blind"},

		// Ante 2 tests
		{2, game.SmallBlind, 375, "Ante 2 Small Blind"},
		{2, game.BigBlind, 562, "Ante 2 Big Blind"},
		{2, game.BossBlind, 750, "Ante 2 Boss Blind"},

		// Ante 3 tests
		{3, game.SmallBlind, 450, "Ante 3 Small Blind"},
		{3, game.BigBlind, 675, "Ante 3 Big Blind"},
		{3, game.BossBlind, 900, "Ante 3 Boss Blind"},

		// Ante 5 tests
		{5, game.SmallBlind, 600, "Ante 5 Small Blind"},
		{5, game.BigBlind, 900, "Ante 5 Big Blind"},
		{5, game.BossBlind, 1200, "Ante 5 Boss Blind"},

		// Ante 8 tests (final ante)
		{8, game.SmallBlind, 825, "Ante 8 Small Blind"},
		{8, game.BigBlind, 1237, "Ante 8 Big Blind"},
		{8, game.BossBlind, 1650, "Ante 8 Boss Blind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := game.GetAnteRequirement(tt.ante, tt.blindType)
			if result != tt.expected {
				t.Errorf("game.GetAnteRequirement(%d, %v) = %d, want %d",
					tt.ante, tt.blindType, result, tt.expected)
			}
		})
	}
}

func TestBlindProgression(t *testing.T) {
	// Test that each ante increases the base requirement by 75
	baseRequirement := 300
	increment := 75

	for ante := 1; ante <= 8; ante++ {
		expectedBase := baseRequirement + (ante-1)*increment
		actualBase := game.GetAnteRequirement(ante, game.SmallBlind)

		if actualBase != expectedBase {
			t.Errorf("Ante %d Small Blind = %d, want %d", ante, actualBase, expectedBase)
		}
	}
}

func TestBlindTypeMultipliers(t *testing.T) {
	// Test that Big Blind is 1.5x Small Blind and Boss Blind is 2x Small Blind
	for ante := 1; ante <= 8; ante++ {
		smallBlind := game.GetAnteRequirement(ante, game.SmallBlind)
		bigBlind := game.GetAnteRequirement(ante, game.BigBlind)
		bossBlind := game.GetAnteRequirement(ante, game.BossBlind)

		expectedBigBlind := int(float64(smallBlind) * 1.5)
		expectedBossBlind := smallBlind * 2

		if bigBlind != expectedBigBlind {
			t.Errorf("Ante %d Big Blind = %d, want %d (1.5x Small Blind %d)",
				ante, bigBlind, expectedBigBlind, smallBlind)
		}

		if bossBlind != expectedBossBlind {
			t.Errorf("Ante %d Boss Blind = %d, want %d (2x Small Blind %d)",
				ante, bossBlind, expectedBossBlind, smallBlind)
		}
	}
}

func TestBlindTypeString(t *testing.T) {
	tests := []struct {
		blindType game.BlindType
		expected  string
	}{
		{game.SmallBlind, "Small Blind"},
		{game.BigBlind, "Big Blind"},
		{game.BossBlind, "Boss Blind"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.blindType.String()
			if result != tt.expected {
				t.Errorf("game.BlindType(%d).String() = %s, want %s",
					int(tt.blindType), result, tt.expected)
			}
		})
	}
}

func BenchmarkGetAnteRequirement(b *testing.B) {
	for i := 0; i < b.N; i++ {
		game.GetAnteRequirement(5, game.BossBlind)
	}
}
