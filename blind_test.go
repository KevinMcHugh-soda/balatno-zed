package main

import (
	"testing"
)

func TestGetBlindRequirement(t *testing.T) {
	tests := []struct {
		ante      int
		blindType BlindType
		expected  int
		name      string
	}{
		// Ante 1 tests
		{1, SmallBlind, 300, "Ante 1 Small Blind"},
		{1, BigBlind, 450, "Ante 1 Big Blind"},
		{1, BossBlind, 600, "Ante 1 Boss Blind"},

		// Ante 2 tests
		{2, SmallBlind, 375, "Ante 2 Small Blind"},
		{2, BigBlind, 562, "Ante 2 Big Blind"},
		{2, BossBlind, 750, "Ante 2 Boss Blind"},

		// Ante 3 tests
		{3, SmallBlind, 450, "Ante 3 Small Blind"},
		{3, BigBlind, 675, "Ante 3 Big Blind"},
		{3, BossBlind, 900, "Ante 3 Boss Blind"},

		// Ante 5 tests
		{5, SmallBlind, 600, "Ante 5 Small Blind"},
		{5, BigBlind, 900, "Ante 5 Big Blind"},
		{5, BossBlind, 1200, "Ante 5 Boss Blind"},

		// Ante 8 tests (final ante)
		{8, SmallBlind, 825, "Ante 8 Small Blind"},
		{8, BigBlind, 1237, "Ante 8 Big Blind"},
		{8, BossBlind, 1650, "Ante 8 Boss Blind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBlindRequirement(tt.ante, tt.blindType)
			if result != tt.expected {
				t.Errorf("GetBlindRequirement(%d, %v) = %d, want %d",
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
		actualBase := GetBlindRequirement(ante, SmallBlind)

		if actualBase != expectedBase {
			t.Errorf("Ante %d Small Blind = %d, want %d", ante, actualBase, expectedBase)
		}
	}
}

func TestBlindTypeMultipliers(t *testing.T) {
	// Test that Big Blind is 1.5x Small Blind and Boss Blind is 2x Small Blind
	for ante := 1; ante <= 8; ante++ {
		smallBlind := GetBlindRequirement(ante, SmallBlind)
		bigBlind := GetBlindRequirement(ante, BigBlind)
		bossBlind := GetBlindRequirement(ante, BossBlind)

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
		blindType BlindType
		expected  string
	}{
		{SmallBlind, "Small Blind"},
		{BigBlind, "Big Blind"},
		{BossBlind, "Boss Blind"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.blindType.String()
			if result != tt.expected {
				t.Errorf("BlindType(%d).String() = %s, want %s",
					int(tt.blindType), result, tt.expected)
			}
		})
	}
}

func BenchmarkGetBlindRequirement(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetBlindRequirement(5, BossBlind)
	}
}
