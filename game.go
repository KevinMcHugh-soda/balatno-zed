package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
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

// GetBlindRequirement calculates the score requirement for a blind
func GetBlindRequirement(ante int, blindType BlindType) int {
	base := 300

	// Increase base requirement each ante
	requirement := base + (ante-1)*75

	// Adjust based on blind type
	switch blindType {
	case SmallBlind:
		return requirement
	case BigBlind:
		return int(float64(requirement) * 1.5)
	case BossBlind:
		return requirement * 2
	default:
		return requirement
	}
}

// SortMode represents how cards should be displayed
type SortMode int

const (
	SortByRank SortMode = iota
	SortBySuit
)

// Game represents the current game state
type Game struct {
	totalScore        int
	handsPlayed       int
	discardsUsed      int
	deck              []Card
	deckIndex         int
	playerCards       []Card
	scanner           *bufio.Scanner
	displayToOriginal []int // maps display position (0-based) to original position
	sortMode          SortMode
	currentAnte       int
	currentBlind      BlindType
	currentTarget     int
	money             int
	jokers            []Joker
}

// NewGame creates a new game instance
func NewGame() *Game {
	deck := NewDeck()
	ShuffleDeck(deck)

	game := &Game{
		totalScore:   0,
		handsPlayed:  0,
		discardsUsed: 0,
		deck:         deck,
		deckIndex:    0,
		scanner:      bufio.NewScanner(os.Stdin),
		sortMode:     SortByRank,
		currentAnte:  1,
		currentBlind: SmallBlind,
		money:        StartingMoney,
		jokers:       []Joker{},
	}

	// Set initial target
	game.currentTarget = GetBlindRequirement(game.currentAnte, game.currentBlind)

	// Deal initial hand
	game.playerCards = make([]Card, InitialCards)
	copy(game.playerCards, deck[game.deckIndex:game.deckIndex+InitialCards])
	game.deckIndex += InitialCards

	return game
}

// Run starts the main game loop
func (g *Game) Run() {
	fmt.Println("🃏 Welcome to Balatro CLI! 🃏")
	fmt.Println("🎯 CHALLENGE: Progress through 8 Antes, each with 3 Blinds!")
	fmt.Println("Each Ante: Small Blind → Big Blind → Boss Blind")
	fmt.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	fmt.Println()

	gameRunning := true
	for gameRunning && g.currentAnte <= MaxAntes {
		for g.handsPlayed < MaxHands && g.totalScore < g.currentTarget {
			g.showGameStatus()
			g.showCards()

			action, params, quit := g.getPlayerInput()
			if quit {
				fmt.Println("Thanks for playing!")
				gameRunning = false
				break
			}

			if action == "play" {
				g.handlePlayAction(params)
			} else if action == "discard" {
				g.handleDiscardAction(params)
			} else if action == "resort" {
				g.handleResortAction()
			}
		}

		if !gameRunning {
			break
		}

		// Check if blind was completed
		if g.totalScore >= g.currentTarget {
			g.handleBlindCompletion()
		} else {
			// Failed to beat the blind
			g.showGameResults()
			break
		}
	}

	if gameRunning && g.currentAnte > MaxAntes {
		g.showVictoryResults()
	}
}

// showGameStatus displays the current game state
func (g *Game) showGameStatus() {
	// Create visual progress bar for score
	progress := float64(g.totalScore) / float64(g.currentTarget)
	if progress > 1.0 {
		progress = 1.0
	}
	progressWidth := 20
	filled := int(progress * float64(progressWidth))

	progressBar := "["
	for i := 0; i < progressWidth; i++ {
		if i < filled {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}
	progressBar += "]"

	// Blind type emojis
	blindEmoji := ""
	switch g.currentBlind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	fmt.Printf("%s Ante %d - %s\n", blindEmoji, g.currentAnte, g.currentBlind)
	fmt.Printf("🎯 Target: %d | Score: %d %s (%.1f%%)\n",
		g.currentTarget, g.totalScore, progressBar, progress*100)
	fmt.Printf("🎴 Hands Left: %d | 🗑️  Discards Left: %d | 💰 Money: $%d\n",
		MaxHands-g.handsPlayed, MaxDiscards-g.discardsUsed, g.money)
	fmt.Println()
}

// showCards displays the player's current cards sorted by rank
func (g *Game) showCards() {
	// Create a sorted copy of cards with their original indices
	type indexedCard struct {
		card  Card
		index int
	}

	indexed := make([]indexedCard, len(g.playerCards))
	for i, card := range g.playerCards {
		indexed[i] = indexedCard{card: card, index: i}
	}

	// Sort based on current sort mode
	sort.Slice(indexed, func(i, j int) bool {
		if g.sortMode == SortByRank {
			return indexed[i].card.Rank < indexed[j].card.Rank
		} else {
			// Sort by suit, then by rank within suit
			if indexed[i].card.Suit != indexed[j].card.Suit {
				return indexed[i].card.Suit < indexed[j].card.Suit
			}
			return indexed[i].card.Rank < indexed[j].card.Rank
		}
	})

	// Update the display-to-original mapping
	g.displayToOriginal = make([]int, len(indexed))
	for i, ic := range indexed {
		g.displayToOriginal[i] = ic.index
	}

	sortModeStr := "rank"
	if g.sortMode == SortBySuit {
		sortModeStr = "suit"
	}
	fmt.Printf("Your cards (sorted by %s):\n", sortModeStr)
	for i, ic := range indexed {
		fmt.Printf("%d: %s\n", i+1, ic.card)
	}
	fmt.Println()
}

// getPlayerInput reads and parses player input
func (g *Game) getPlayerInput() (string, []string, bool) {
	if g.discardsUsed >= MaxDiscards {
		fmt.Print("(p)lay <cards>, (r)esort, or (q)uit: ")
	} else {
		fmt.Print("(p)lay <cards>, (d)iscard <cards>, (r)esort, or (q)uit: ")
	}

	if !g.scanner.Scan() {
		if err := g.scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
		return "", nil, true
	}

	input := strings.TrimSpace(g.scanner.Text())

	if strings.ToLower(input) == "quit" {
		return "", nil, true
	}

	if input == "" {
		fmt.Println("Please enter an action")
		return "", nil, false
	}

	parts := strings.Fields(input)
	if len(parts) < 1 {
		fmt.Println("Please enter 'play <cards>' or 'discard <cards>'")
		return "", nil, false
	}

	action := strings.ToLower(parts[0])

	// Support abbreviated commands
	if action == "p" {
		action = "play"
	} else if action == "d" {
		action = "discard"
	} else if action == "r" {
		action = "resort"
	} else if action == "q" {
		return "", nil, true
	}

	var params []string
	if len(parts) > 1 {
		params = parts[1:]
	}

	return action, params, false
}

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

	fmt.Println()
	fmt.Printf("Your hand: %s\n", hand)
	fmt.Printf("Hand type: %s\n", evaluator.Name())
	fmt.Printf("Base Score: %d | Card Values: %d | Mult: %dx\n", baseScore, cardValues, evaluator.Multiplier())
	fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", baseScore, cardValues, evaluator.Multiplier(), score)

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
		displayIndex, err := strconv.Atoi(param)
		if err != nil || displayIndex < 1 || displayIndex > len(g.playerCards) {
			fmt.Printf("Invalid card number: %s\n", param)
			return nil, nil, false
		}
		// Map display position to original position
		originalIndex := g.displayToOriginal[displayIndex-1]
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
	// Different celebrations for different blind types
	switch g.currentBlind {
	case SmallBlind:
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🔸 SMALL BLIND DEFEATED! 🔸")
		fmt.Printf("    ✨ Score: %d/%d ✨\n", g.totalScore, g.currentTarget)
		fmt.Println("   🎯 Advancing to Big Blind...")
		fmt.Println(strings.Repeat("=", 60))
	case BigBlind:
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🔶 BIG BLIND CRUSHED! 🔶")
		fmt.Printf("    ⚡ Score: %d/%d ⚡\n", g.totalScore, g.currentTarget)
		fmt.Println("   💀 Prepare for the Boss Blind...")
		fmt.Println(strings.Repeat("=", 60))
	case BossBlind:
		fmt.Println(strings.Repeat("🎆", 15))
		fmt.Println("💀 BOSS BLIND ANNIHILATED! 💀")
		fmt.Printf("    🔥 EPIC SCORE: %d/%d 🔥\n", g.totalScore, g.currentTarget)
		if g.currentAnte < MaxAntes {
			fmt.Printf("🎊 ANTE %d CONQUERED! 🎊\n", g.currentAnte)
		}
		fmt.Println(strings.Repeat("🎆", 15))
	}

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

	fmt.Printf("💰 REWARD BREAKDOWN:\n")
	fmt.Printf("   Base: $%d", baseReward)
	if bonusReward > 0 {
		fmt.Printf(" + Unused: $%d (%d hands + %d discards)", bonusReward, unusedHands, unusedDiscards)
	}
	if jokerReward > 0 {
		fmt.Printf(" + Jokers: $%d", jokerReward)
	}
	fmt.Printf("\n   💰 Total Earned: $%d | Your Money: $%d\n", totalReward, g.money)
	fmt.Println()

	// Advance to next blind
	if g.currentBlind == SmallBlind {
		g.currentBlind = BigBlind
	} else if g.currentBlind == BigBlind {
		g.currentBlind = BossBlind
	} else {
		// Completed Boss Blind, advance to next ante
		g.currentAnte++
		g.currentBlind = SmallBlind
		if g.currentAnte <= MaxAntes {
			fmt.Println("🌟 ANTE PROGRESSION 🌟")
			fmt.Printf("   Ante %d → Ante %d\n", g.currentAnte-1, g.currentAnte)
			fmt.Println("🔄 NEW CHALLENGE AWAITS!")
			fmt.Println()
		}
	}

	if g.currentAnte <= MaxAntes {
		// Reset for next blind
		g.totalScore = 0
		g.handsPlayed = 0
		g.discardsUsed = 0
		g.currentTarget = GetBlindRequirement(g.currentAnte, g.currentBlind)

		// Shuffle and deal new hand
		g.deckIndex = 0
		ShuffleDeck(g.deck)
		g.playerCards = make([]Card, InitialCards)
		copy(g.playerCards, g.deck[g.deckIndex:g.deckIndex+InitialCards])
		g.deckIndex += InitialCards

		// Show next blind info
		blindEmoji := ""
		switch g.currentBlind {
		case SmallBlind:
			blindEmoji = "🔸"
		case BigBlind:
			blindEmoji = "🔶"
		case BossBlind:
			blindEmoji = "💀"
		}

		fmt.Printf("%s NOW ENTERING: %s (Ante %d) %s\n",
			blindEmoji, g.currentBlind, g.currentAnte, blindEmoji)
		fmt.Printf("🎯 NEW TARGET: %d points\n", g.currentTarget)
		fmt.Println("🃏 Fresh hand dealt!")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println()

		// Show shop between blinds
		g.showShop()
	}
}

// showShop displays the shop interface between blinds
func (g *Game) showShop() {
	fmt.Println("🏪 SHOP 🏪")
	fmt.Printf("💰 Your Money: $%d\n", g.money)
	fmt.Println()

	// Show available jokers
	goldenJoker := GetGoldenJoker()
	hasGoldenJoker := PlayerHasJoker(g.jokers, goldenJoker.Name)

	if !hasGoldenJoker {
		fmt.Printf("1. %s - $%d\n", goldenJoker.Name, goldenJoker.Price)
		fmt.Printf("   %s\n", goldenJoker.Description)
		fmt.Println()
	}

	// Show current jokers
	if len(g.jokers) > 0 {
		fmt.Printf("🃏 Your Jokers: %s\n", FormatJokersList(g.jokers))
		fmt.Println()
	}

	// Shop interaction
	if !hasGoldenJoker && g.money >= goldenJoker.Price {
		fmt.Print("Buy (1) The Golden Joker, or (s)kip shop: ")
		if g.scanner.Scan() {
			input := strings.TrimSpace(strings.ToLower(g.scanner.Text()))
			if input == "1" {
				g.money -= goldenJoker.Price
				g.jokers = append(g.jokers, goldenJoker)
				fmt.Printf("✨ Purchased %s! ✨\n", goldenJoker.Name)
				fmt.Printf("💰 Remaining money: $%d\n", g.money)
			} else {
				fmt.Println("Skipped shop.")
			}
		}
	} else if hasGoldenJoker {
		fmt.Println("No new jokers available.")
		fmt.Print("Press enter to continue...")
		g.scanner.Scan()
	} else {
		fmt.Printf("Need $%d to buy The Golden Joker (you have $%d)\n", goldenJoker.Price, g.money)
		fmt.Print("Press enter to continue...")
		g.scanner.Scan()
	}

	fmt.Println()
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
