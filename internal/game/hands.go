package game

import (
	"sort"
	"strings"
)

// HandEvaluator interface defines the behavior of different poker hand types
type HandEvaluator interface {
	// Name returns the display name of the hand type
	Name() string

	// Matches returns true if the given cards match this hand type
	Matches(cards []Card) bool

	// Priority returns the priority of this hand type (higher = better)
	Priority() int
}

// Hand represents a poker hand
type Hand struct {
	Cards []Card
}

func (h Hand) String() string {
	var cards []string
	for _, card := range h.Cards {
		cards = append(cards, card.String())
	}
	return strings.Join(cards, " ")
}

// All hand evaluators in priority order (highest to lowest)
var handEvaluators = []HandEvaluator{
	&RoyalFlushEvaluator{},
	&StraightFlushEvaluator{},
	&FourOfAKindEvaluator{},
	&FullHouseEvaluator{},
	&FlushEvaluator{},
	&StraightEvaluator{},
	&ThreeOfAKindEvaluator{},
	&TwoPairEvaluator{},
	&PairEvaluator{},
	&HighCardEvaluator{},
}

// EvaluateHand determines the best hand type for the given cards using hand levels
func EvaluateHand(hand Hand, levels map[string]int) (HandEvaluator, int, int, int, int) {
	if len(hand.Cards) == 0 {
		return &HighCardEvaluator{}, 0, 0, 0, 1
	}

	// Try each evaluator in priority order
	for _, evaluator := range handEvaluators {
		if evaluator.Matches(hand.Cards) {
			// Calculate total card value
			totalValue := 0
			for _, card := range hand.Cards {
				totalValue += card.Rank.Value()
			}

			level := 1
			if levels != nil {
				if l, ok := levels[evaluator.Name()]; ok && l > 0 {
					level = l
				}
			}

			baseScore, mult := GetHandScore(evaluator.Name(), level)
			finalScore := (baseScore + totalValue) * mult

			return evaluator, finalScore, totalValue, baseScore, mult
		}
	}

	// Should never reach here, but fallback to high card
	evaluator := &HighCardEvaluator{}
	totalValue := 0
	for _, card := range hand.Cards {
		totalValue += card.Rank.Value()
	}
	level := 1
	if levels != nil {
		if l, ok := levels[evaluator.Name()]; ok && l > 0 {
			level = l
		}
	}
	baseScore, mult := GetHandScore(evaluator.Name(), level)
	finalScore := (baseScore + totalValue) * mult

	return evaluator, finalScore, totalValue, baseScore, mult
}

// Helper functions for hand evaluation
func sortCardsByRank(cards []Card) []Card {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank < sorted[j].Rank
	})
	return sorted
}

func getRankCounts(cards []Card) map[Rank]int {
	rankCounts := make(map[Rank]int)
	for _, card := range cards {
		rankCounts[card.Rank]++
	}
	return rankCounts
}

func getSuitCounts(cards []Card) map[Suit]int {
	suitCounts := make(map[Suit]int)
	for _, card := range cards {
		suitCounts[card.Suit]++
	}
	return suitCounts
}

func isFlush(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}
	suitCounts := getSuitCounts(cards)
	return len(suitCounts) == 1
}

func isStraight(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}

	sorted := sortCardsByRank(cards)

	// Check regular straight
	isStraight := true
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Rank != sorted[i-1].Rank+1 {
			isStraight = false
			break
		}
	}

	if isStraight {
		return true
	}

	// Check wheel straight (A-2-3-4-5)
	if sorted[0].Rank == Ace && sorted[1].Rank == Two &&
		sorted[2].Rank == Three && sorted[3].Rank == Four && sorted[4].Rank == Five {
		return true
	}

	// Check high ace straight (A-10-J-Q-K)
	if sorted[0].Rank == Ace && sorted[1].Rank == Ten &&
		sorted[2].Rank == Jack && sorted[3].Rank == Queen && sorted[4].Rank == King {
		return true
	}

	return false
}

// Concrete hand evaluator implementations

type RoyalFlushEvaluator struct{}

func (e *RoyalFlushEvaluator) Name() string  { return "Royal Flush" }
func (e *RoyalFlushEvaluator) Priority() int { return 10 }

func (e *RoyalFlushEvaluator) Matches(cards []Card) bool {
	if !isFlush(cards) || !isStraight(cards) {
		return false
	}

	sorted := sortCardsByRank(cards)
	return len(cards) == 5 && sorted[0].Rank == Ace && sorted[1].Rank == Ten &&
		sorted[2].Rank == Jack && sorted[3].Rank == Queen && sorted[4].Rank == King
}

type StraightFlushEvaluator struct{}

func (e *StraightFlushEvaluator) Name() string  { return "Straight Flush" }
func (e *StraightFlushEvaluator) Priority() int { return 9 }

func (e *StraightFlushEvaluator) Matches(cards []Card) bool {
	return isFlush(cards) && isStraight(cards)
}

type FourOfAKindEvaluator struct{}

func (e *FourOfAKindEvaluator) Name() string  { return "Four of a Kind" }
func (e *FourOfAKindEvaluator) Priority() int { return 8 }

func (e *FourOfAKindEvaluator) Matches(cards []Card) bool {
	rankCounts := getRankCounts(cards)
	for _, count := range rankCounts {
		if count == 4 {
			return true
		}
	}
	return false
}

type FullHouseEvaluator struct{}

func (e *FullHouseEvaluator) Name() string  { return "Full House" }
func (e *FullHouseEvaluator) Priority() int { return 7 }

func (e *FullHouseEvaluator) Matches(cards []Card) bool {
	rankCounts := getRankCounts(cards)
	hasThree := false
	hasTwo := false

	for _, count := range rankCounts {
		if count == 3 {
			hasThree = true
		} else if count == 2 {
			hasTwo = true
		}
	}

	return hasThree && hasTwo
}

type FlushEvaluator struct{}

func (e *FlushEvaluator) Name() string  { return "Flush" }
func (e *FlushEvaluator) Priority() int { return 6 }

func (e *FlushEvaluator) Matches(cards []Card) bool {
	return isFlush(cards)
}

type StraightEvaluator struct{}

func (e *StraightEvaluator) Name() string  { return "Straight" }
func (e *StraightEvaluator) Priority() int { return 5 }

func (e *StraightEvaluator) Matches(cards []Card) bool {
	return isStraight(cards)
}

type ThreeOfAKindEvaluator struct{}

func (e *ThreeOfAKindEvaluator) Name() string  { return "Three of a Kind" }
func (e *ThreeOfAKindEvaluator) Priority() int { return 4 }

func (e *ThreeOfAKindEvaluator) Matches(cards []Card) bool {
	rankCounts := getRankCounts(cards)
	for _, count := range rankCounts {
		if count == 3 {
			return true
		}
	}
	return false
}

type TwoPairEvaluator struct{}

func (e *TwoPairEvaluator) Name() string  { return "Two Pair" }
func (e *TwoPairEvaluator) Priority() int { return 3 }

func (e *TwoPairEvaluator) Matches(cards []Card) bool {
	rankCounts := getRankCounts(cards)
	pairCount := 0

	for _, count := range rankCounts {
		if count == 2 {
			pairCount++
		}
	}

	return pairCount == 2
}

type PairEvaluator struct{}

func (e *PairEvaluator) Name() string  { return "Pair" }
func (e *PairEvaluator) Priority() int { return 2 }

func (e *PairEvaluator) Matches(cards []Card) bool {
	rankCounts := getRankCounts(cards)
	for _, count := range rankCounts {
		if count == 2 {
			return true
		}
	}
	return false
}

type HighCardEvaluator struct{}

func (e *HighCardEvaluator) Name() string  { return "High Card" }
func (e *HighCardEvaluator) Priority() int { return 1 }

func (e *HighCardEvaluator) Matches(cards []Card) bool {
	// High card always matches as fallback
	return true
}
