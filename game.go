package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Game constants
const (
	MaxHands      = 4
	MaxDiscards   = 3
	InitialCards  = 7
	MaxAntes      = 8
	StartingMoney = 4
)

// Money rewards for completing blinds
const (
	SmallBlindReward    = 4
	BigBlindReward      = 5
	BossBlindReward     = 6
	UnusedHandReward    = 1
	UnusedDiscardReward = 1
)

// BlindType represents the type of blind being played
type BlindType int

const (
	SmallBlind BlindType = iota
	BigBlind
	BossBlind
)

func (bt BlindType) String() string {
	switch bt {
	case SmallBlind:
		return "Small Blind"
	case BigBlind:
		return "Big Blind"
	case BossBlind:
		return "Boss Blind"
	default:
		return "Unknown"
	}
}

// Blind represents a single blind challenge
type Blind struct {
	Type        BlindType
	TargetScore int
	Name        string
	Description string
}

// SortMode represents how cards should be displayed
type SortMode int

const (
	SortByRank SortMode = iota
	SortBySuit
)

// Game represents the current game state
type Game struct {
	totalScore   int
	handsPlayed  int
	discardsUsed int
	deck         []Card
	deckIndex    int
	playerCards  []Card
	// scanner removed - UI handles input now
	displayToOriginal map[string]int // maps display character to original position
	sortMode          SortMode
	currentAnte       int
	currentBlind      BlindType
	currentTarget     int
	money             int
	jokers            []Joker
	rerollCost        int
}

// NewGame creates a new game instance
func NewGame() *Game {
	deck := NewDeck()
	ShuffleDeck(deck)

	game := &Game{
		totalScore:        0,
		handsPlayed:       0,
		discardsUsed:      0,
		deck:              deck,
		deckIndex:         0,
		displayToOriginal: make(map[string]int),
		sortMode:          SortByRank,
		currentAnte:       1,
		currentBlind:      SmallBlind,
		money:             StartingMoney,
		jokers:            []Joker{},
		rerollCost:        5, // Initial reroll cost
	}

	// Initialize random seed once
	rand.Seed(time.Now().UnixNano())

	// Load configuration
	if err := LoadConfig(); err != nil {
		// Config loading failed, but we have fallback defaults
		fmt.Printf("Warning: %v\n", err)
	}

	// Load joker configurations
	if err := LoadJokerConfigs(); err != nil {
		// Joker config loading failed, but we have fallback defaults
		fmt.Printf("Warning: %v\n", err)
	}

	// Set initial target
	game.currentTarget = GetAnteRequirement(game.currentAnte, game.currentBlind)

	// Deal initial hand
	game.playerCards = make([]Card, InitialCards)
	copy(game.playerCards, deck[game.deckIndex:game.deckIndex+InitialCards])
	game.deckIndex += InitialCards

	return game
}

// Run starts the main game loop
// Run method removed - UI handles the main game loop now

// showGameStatus displays the current game state
// showGameStatus method removed - UI handles display now

// showCards method removed - UI handles display now

// getPlayerInput reads and parses player input
// getPlayerInput method removed - UI handles input now

// handlePlayAction processes a play action
func (g *Game) handlePlayAction(params []string) {
	if len(params) < 1 {
		fmt.Println("Please specify cards to play: 'play 1 2 3'")
		return
	}

	if len(params) > 5 {
		fmt.Println("You can only play up to 5 cards!")
		return
	}

	selectedCards, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedCards) == 0 {
		fmt.Println("Please select at least one card!")
		return
	}

	// Evaluate the hand
	hand := Hand{Cards: selectedCards}
	evaluator, score, cardValues, baseScore := EvaluateHand(hand)

	// Calculate joker bonuses
	jokerChips, jokerMult := CalculateJokerHandBonus(g.jokers, evaluator.Name())

	// Apply joker bonuses to final score
	finalBaseScore := baseScore + jokerChips
	finalMult := evaluator.Multiplier() + jokerMult
	finalScore := (finalBaseScore + cardValues) * finalMult

	fmt.Println()
	fmt.Printf("Your hand: %s\n", hand)
	fmt.Printf("Hand type: %s\n", evaluator.Name())

	if jokerChips > 0 || jokerMult > 0 {
		fmt.Printf("Base Score: %d", baseScore)
		if jokerChips > 0 {
			fmt.Printf(" + %d Joker Chips", jokerChips)
		}
		fmt.Printf(" | Card Values: %d | Mult: %dx", cardValues, evaluator.Multiplier())
		if jokerMult > 0 {
			fmt.Printf(" + %d Joker Mult", jokerMult)
		}
		fmt.Printf("\n")
		fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", finalBaseScore, cardValues, finalMult, finalScore)
	} else {
		fmt.Printf("Base Score: %d | Card Values: %d | Mult: %dx\n", baseScore, cardValues, evaluator.Multiplier())
		fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", baseScore, cardValues, evaluator.Multiplier(), finalScore)
	}

	// Use the joker-modified score
	score = finalScore

	g.totalScore += score
	g.handsPlayed++

	fmt.Printf("💰 Total Score: %d/%d\n", g.totalScore, g.currentTarget)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	// Remove played cards and deal new ones
	g.removeAndDealCards(selectedIndices)
}

// handleDiscardAction processes a discard action
func (g *Game) handleDiscardAction(params []string) {
	if g.discardsUsed >= MaxDiscards {
		fmt.Println("No discards remaining!")
		return
	}

	if len(params) < 1 {
		fmt.Println("Please specify cards to discard: 'discard 1 2'")
		return
	}

	_, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedIndices) == 0 {
		fmt.Println("Please select at least one card!")
		return
	}

	fmt.Printf("Discarded %d card(s)\n", len(selectedIndices))
	g.discardsUsed++

	// Remove discarded cards and deal new ones
	g.removeAndDealCards(selectedIndices)

	fmt.Println("New cards dealt!")
	fmt.Println()
}

// parseCardSelection parses card selection from string parameters
func (g *Game) parseCardSelection(params []string) ([]Card, []int, bool) {
	var selectedCards []Card
	var selectedIndices []int

	for _, param := range params {
		var originalIndex int
		var found bool

		// Try character-based selection first (A, B, C, etc.)
		if len(param) == 1 {
			if idx, exists := g.displayToOriginal[param]; exists {
				originalIndex = idx
				found = true
			}
		}

		// Try numeric selection (1, 2, 3, etc.)
		if !found {
			if displayIndex, err := strconv.Atoi(param); err == nil && displayIndex >= 1 && displayIndex <= len(g.playerCards) {
				originalIndex = displayIndex - 1
				found = true
			}
		}

		if !found {
			return nil, nil, false
		}

		selectedCards = append(selectedCards, g.playerCards[originalIndex])
		selectedIndices = append(selectedIndices, originalIndex)
	}

	return selectedCards, selectedIndices, true
}

// removeAndDealCards removes selected cards and deals new ones
func (g *Game) removeAndDealCards(selectedIndices []int) {
	g.playerCards = removeCards(g.playerCards, selectedIndices)
	newCardsNeeded := len(selectedIndices)

	// Deal new cards if available
	if g.deckIndex+newCardsNeeded <= len(g.deck) {
		for i := 0; i < newCardsNeeded; i++ {
			g.playerCards = append(g.playerCards, g.deck[g.deckIndex])
			g.deckIndex++
		}
	}
}

// calculateBlindReward calculates money earned for completing a blind
func (g *Game) calculateBlindReward() int {
	// Base reward for blind type
	var baseReward int
	switch g.currentBlind {
	case SmallBlind:
		baseReward = SmallBlindReward
	case BigBlind:
		baseReward = BigBlindReward
	case BossBlind:
		baseReward = BossBlindReward
	}

	// Bonus for unused resources
	unusedHands := MaxHands - g.handsPlayed
	unusedDiscards := MaxDiscards - g.discardsUsed
	bonusReward := unusedHands*UnusedHandReward + unusedDiscards*UnusedDiscardReward

	// Joker rewards
	jokerReward := CalculateJokerRewards(g.jokers)

	return baseReward + bonusReward + jokerReward
}

// handleBlindCompletion handles completing a blind and advancing to the next
func (g *Game) handleBlindCompletion() {
	// Blind completed - UI will handle celebration display

	// Calculate and award money with detailed breakdown
	var baseReward int
	switch g.currentBlind {
	case SmallBlind:
		baseReward = SmallBlindReward
	case BigBlind:
		baseReward = BigBlindReward
	case BossBlind:
		baseReward = BossBlindReward
	}

	unusedHands := MaxHands - g.handsPlayed
	unusedDiscards := MaxDiscards - g.discardsUsed
	bonusReward := unusedHands*UnusedHandReward + unusedDiscards*UnusedDiscardReward
	jokerReward := CalculateJokerRewards(g.jokers)
	totalReward := baseReward + bonusReward + jokerReward

	g.money += totalReward

	// Reward calculated and added - UI will handle display

	// Advance to next blind
	if g.currentBlind == SmallBlind {
		g.currentBlind = BigBlind
	} else if g.currentBlind == BigBlind {
		g.currentBlind = BossBlind
	} else {
		// Completed Boss Blind, advance to next ante
		g.currentAnte++
		g.currentBlind = SmallBlind
		// Ante progression - UI will handle display
	}

	if g.currentAnte <= MaxAntes {
		// Reset for next blind
		g.totalScore = 0
		g.handsPlayed = 0
		g.discardsUsed = 0
		g.rerollCost = 5 // Reset reroll cost for new blind
		g.currentTarget = GetAnteRequirement(g.currentAnte, g.currentBlind)

		// Shuffle and deal new hand
		g.deckIndex = 0
		ShuffleDeck(g.deck)
		g.playerCards = make([]Card, InitialCards)
		copy(g.playerCards, g.deck[g.deckIndex:g.deckIndex+InitialCards])
		g.deckIndex += InitialCards

		// New blind setup complete - UI will handle display

		// Show shop between blinds
		g.showShop()
	}
}

// showShop displays the shop interface between blinds
// TODO: Implement terminal UI version
func (g *Game) showShop() {
	// Temporarily skip shop - will implement UI version later
	return
}

// showShopWithItems is a helper function to continue shop with specific items
// TODO: Implement terminal UI version
func (g *Game) showShopWithItems(availableJokers []Joker, shopItems []Joker) {
	// Temporarily skip shop - will implement UI version later
	return
}

// showGameResults displays the final game results for a failed blind
func (g *Game) showGameResults() {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("💀 DEFEAT! You failed to beat the blind.")
	fmt.Printf("Final Score: %d/%d (Ante %d - %s)\n", g.totalScore, g.currentTarget, g.currentAnte, g.currentBlind)
	fmt.Printf("Hands Played: %d/%d\n", g.handsPlayed, MaxHands)
	fmt.Printf("Discards Used: %d/%d\n", g.discardsUsed, MaxDiscards)
	fmt.Println(strings.Repeat("=", 50))
}

// showVictoryResults displays the final victory results
func (g *Game) showVictoryResults() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🏆 ULTIMATE VICTORY! 🏆")
	fmt.Println("🎉 You have conquered all 8 Antes! 🎉")
	fmt.Printf("Final Ante: %d\n", MaxAntes)
	fmt.Println("You are a true Balatro master!")
	fmt.Println(strings.Repeat("=", 60))
}

// removeCards removes cards at specified indices and returns the new slice
func removeCards(cards []Card, indices []int) []Card {
	// Sort indices in descending order to remove from end first
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	result := make([]Card, len(cards))
	copy(result, cards)

	for _, index := range indices {
		if index >= 0 && index < len(result) {
			result = append(result[:index], result[index+1:]...)
		}
	}

	return result
}

// handleResortAction toggles the sort mode and redisplays cards
func (g *Game) handleResortAction() {
	if g.sortMode == SortByRank {
		g.sortMode = SortBySuit
		fmt.Println("Cards now sorted by suit")
	} else {
		g.sortMode = SortByRank
		fmt.Println("Cards now sorted by rank")
	}
	fmt.Println()
}
