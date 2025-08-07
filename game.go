package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
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
	rerollCost        int
	printMode         PrintMode
}

type PrintMode int

const (
	PrintModeLogger PrintMode = iota
	PrintModeTUI
)

// NewGame creates a new game instance
func NewGame(pm PrintMode) *Game {
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
		rerollCost:   5, // Initial reroll cost
		printMode:    pm,
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

func (g *Game) Printf(format string, args ...any) {
	if g.printMode == PrintModeLogger {
		fmt.Printf(format, args...)
	}
}

func (g *Game) Println(a ...any) {
	if g.printMode == PrintModeLogger {
		fmt.Println(a...)
	}
}

// Run starts the main game loop
func (g *Game) Run() {
	g.Println("🃏 Welcome to Balatro CLI! 🃏")
	g.Println("🎯 CHALLENGE: Progress through 8 Antes, each with 3 Blinds!")
	g.Println("Each Ante: Small Blind → Big Blind → Boss Blind")
	g.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	g.Println()

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

	g.Printf("%s Ante %d - %s\n", blindEmoji, g.currentAnte, g.currentBlind)
	g.Printf("🎯 Target: %d | Score: %d %s (%.1f%%)\n",
		g.currentTarget, g.totalScore, progressBar, progress*100)
	g.Printf("🎴 Hands Left: %d | 🗑️  Discards Left: %d | 💰 Money: $%d\n",
		MaxHands-g.handsPlayed, MaxDiscards-g.discardsUsed, g.money)
	g.Println()
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
	g.Printf("Your cards (sorted by %s):\n", sortModeStr)
	for i, ic := range indexed {
		g.Printf("%d: %s\n", i+1, ic.card)
	}
	g.Println()
}

// getPlayerInput reads and parses player input
func (g *Game) getPlayerInput() (string, []string, bool) {
	if g.discardsUsed >= MaxDiscards {
		g.Printf("(p)lay <cards>, (r)esort, or (q)uit: ")
	} else {
		g.Printf("(p)lay <cards>, (d)iscard <cards>, (r)esort, or (q)uit: ")
	}

	if !g.scanner.Scan() {
		if err := g.scanner.Err(); err != nil {
			g.Println("Error reading input:", err)
		}
		return "", nil, true
	}

	input := strings.TrimSpace(g.scanner.Text())

	if strings.ToLower(input) == "quit" {
		return "", nil, true
	}

	if input == "" {
		g.Println("Please enter an action")
		return "", nil, false
	}

	parts := strings.Fields(input)
	if len(parts) < 1 {
		g.Println("Please enter 'play <cards>' or 'discard <cards>'")
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
		g.Println("Please specify cards to play: 'play 1 2 3'")
		return
	}

	if len(params) > 5 {
		g.Println("You can only play up to 5 cards!")
		return
	}

	selectedCards, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedCards) == 0 {
		g.Println("Please select at least one card!")
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

	g.Println()
	g.Printf("Your hand: %s\n", hand)
	g.Printf("Hand type: %s\n", evaluator.Name())

	if jokerChips > 0 || jokerMult > 0 {
		g.Printf("Base Score: %d", baseScore)
		if jokerChips > 0 {
			g.Printf(" + %d Joker Chips", jokerChips)
		}
		g.Printf(" | Card Values: %d | Mult: %dx", cardValues, evaluator.Multiplier())
		if jokerMult > 0 {
			g.Printf(" + %d Joker Mult", jokerMult)
		}
		g.Printf("\n")
		g.Printf("Final Score: (%d + %d) × %d = %d points\n", finalBaseScore, cardValues, finalMult, finalScore)
	} else {
		g.Printf("Base Score: %d | Card Values: %d | Mult: %dx\n", baseScore, cardValues, evaluator.Multiplier())
		g.Printf("Final Score: (%d + %d) × %d = %d points\n", baseScore, cardValues, evaluator.Multiplier(), finalScore)
	}

	// Use the joker-modified score
	score = finalScore

	g.totalScore += score
	g.handsPlayed++

	g.Printf("💰 Total Score: %d/%d\n", g.totalScore, g.currentTarget)
	g.Println(strings.Repeat("-", 50))
	g.Println()

	// Remove played cards and deal new ones
	g.removeAndDealCards(selectedIndices)
}

// handleDiscardAction processes a discard action
func (g *Game) handleDiscardAction(params []string) {
	if g.discardsUsed >= MaxDiscards {
		g.Println("No discards remaining!")
		return
	}

	if len(params) < 1 {
		g.Println("Please specify cards to discard: 'discard 1 2'")
		return
	}

	_, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedIndices) == 0 {
		g.Println("Please select at least one card!")
		return
	}

	g.Printf("Discarded %d card(s)\n", len(selectedIndices))
	g.discardsUsed++

	// Remove discarded cards and deal new ones
	g.removeAndDealCards(selectedIndices)

	g.Println("New cards dealt!")
	g.Println()
}

// parseCardSelection parses card selection from string parameters
func (g *Game) parseCardSelection(params []string) ([]Card, []int, bool) {
	var selectedCards []Card
	var selectedIndices []int

	for _, param := range params {
		displayIndex, err := strconv.Atoi(param)
		if err != nil || displayIndex < 1 || displayIndex > len(g.playerCards) {
			g.Printf("Invalid card number: %s\n", param)
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
		g.Println(strings.Repeat("=", 60))
		g.Println("🔸 SMALL BLIND DEFEATED! 🔸")
		g.Printf("    ✨ Score: %d/%d ✨\n", g.totalScore, g.currentTarget)
		g.Println("   🎯 Advancing to Big Blind...")
		g.Println(strings.Repeat("=", 60))
	case BigBlind:
		g.Println(strings.Repeat("=", 60))
		g.Println("🔶 BIG BLIND CRUSHED! 🔶")
		g.Printf("    ⚡ Score: %d/%d ⚡\n", g.totalScore, g.currentTarget)
		g.Println("   💀 Prepare for the Boss Blind...")
		g.Println(strings.Repeat("=", 60))
	case BossBlind:
		g.Println(strings.Repeat("🎆", 15))
		g.Println("💀 BOSS BLIND ANNIHILATED! 💀")
		g.Printf("    🔥 EPIC SCORE: %d/%d 🔥\n", g.totalScore, g.currentTarget)
		if g.currentAnte < MaxAntes {
			g.Printf("🎊 ANTE %d CONQUERED! 🎊\n", g.currentAnte)
		}
		g.Println(strings.Repeat("🎆", 15))
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

	g.Printf("💰 REWARD BREAKDOWN:\n")
	g.Printf("   Base: $%d", baseReward)
	if bonusReward > 0 {
		g.Printf(" + Unused: $%d (%d hands + %d discards)", bonusReward, unusedHands, unusedDiscards)
	}
	if jokerReward > 0 {
		g.Printf(" + Jokers: $%d", jokerReward)
	}
	g.Printf("\n   💰 Total Earned: $%d | Your Money: $%d\n", totalReward, g.money)
	g.Println()

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
			g.Println("🌟 ANTE PROGRESSION 🌟")
			g.Printf("   Ante %d → Ante %d\n", g.currentAnte-1, g.currentAnte)
			g.Println("🔄 NEW CHALLENGE AWAITS!")
			g.Println()
		}
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

		g.Printf("%s NOW ENTERING: %s (Ante %d) %s\n",
			blindEmoji, g.currentBlind, g.currentAnte, blindEmoji)
		g.Printf("🎯 NEW TARGET: %d points\n", g.currentTarget)
		g.Println("🃏 Fresh hand dealt!")
		g.Println(strings.Repeat("-", 40))
		g.Println()

		// Show shop between blinds
		g.showShop()
	}
}

// showShop displays the shop interface between blinds
func (g *Game) showShop() {
	// Get all jokers player doesn't own
	allJokers := GetAvailableJokers()
	var availableJokers []Joker
	for _, joker := range allJokers {
		if !PlayerHasJoker(g.jokers, joker.Name) {
			availableJokers = append(availableJokers, joker)
		}
	}

	// If no jokers available, skip shop
	if len(availableJokers) == 0 {
		g.Println("🏪 SHOP 🏪")
		g.Printf("💰 Your Money: $%d\n", g.money)
		g.Println()
		g.Println("All available jokers already owned!")
		g.Println("Press enter to continue...")
		g.scanner.Scan()
		return
	}

	// Randomly select up to 2 jokers for shop
	shopItems := make([]Joker, 0, 2)
	if len(availableJokers) >= 2 {
		// Create a copy and shuffle it
		shuffled := make([]Joker, len(availableJokers))
		copy(shuffled, availableJokers)

		// Fisher-Yates shuffle
		for i := len(shuffled) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		shopItems = shuffled[:2]
	} else {
		shopItems = availableJokers
	}

	// Display shop once and handle single input
	g.Println("🏪 SHOP 🏪")
	g.Printf("💰 Your Money: $%d\n", g.money)
	g.Println()

	// Show shop items
	availableSlots := 0
	for i, joker := range shopItems {
		if joker.Name != "" { // Item still available
			canAfford := g.money >= joker.Price
			affordText := ""
			if !canAfford {
				affordText = " (can't afford)"
			}
			g.Printf("%d. %s - $%d%s\n", i+1, joker.Name, joker.Price, affordText)
			g.Printf("   %s\n", joker.Description)
			g.Println()
			availableSlots++
		}
	}

	// Show current jokers
	if len(g.jokers) > 0 {
		g.Printf("🃏 Your Jokers: %s\n", FormatJokersList(g.jokers))
		g.Println()
	}

	// If no items left, exit shop
	if availableSlots == 0 {
		g.Println("Shop sold out!")
		g.Println("Press enter to continue...")
		g.scanner.Scan()
		return
	}

	// Show options and handle single input
	g.Printf("Buy item (1-%d), (r)eroll ($%d), or (s)kip shop: ", len(shopItems), g.rerollCost)

	if g.scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(g.scanner.Text()))

		if input == "s" || input == "skip" {
			g.Println("Skipped shop.")
			return
		} else if input == "r" || input == "reroll" {
			if g.money >= g.rerollCost {
				g.money -= g.rerollCost
				g.Printf("💫 Rerolled for $%d!\n", g.rerollCost)
				g.rerollCost++ // Increase cost for next reroll

				// Generate new shop items
				if len(availableJokers) >= 2 {
					// Re-shuffle for new items
					shuffled := make([]Joker, len(availableJokers))
					copy(shuffled, availableJokers)

					// Fisher-Yates shuffle for reroll
					for i := len(shuffled) - 1; i > 0; i-- {
						j := rand.Intn(i + 1)
						shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
					}

					shopItems = shuffled[:2]
				} else {
					shopItems = availableJokers
				}

				g.Printf("💰 Remaining money: $%d (Next reroll: $%d)\n", g.money, g.rerollCost)
				g.Println()

				// Recursively call shop to show new items
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.Printf("Not enough money to reroll! Need $%d more.\n", g.rerollCost-g.money)
				return
			}
		} else if choice, err := strconv.Atoi(input); err == nil && choice >= 1 && choice <= len(shopItems) {
			selectedJoker := shopItems[choice-1]
			if selectedJoker.Name == "" {
				g.Println("That slot is empty!")
				return
			}

			if g.money >= selectedJoker.Price {
				g.money -= selectedJoker.Price
				g.jokers = append(g.jokers, selectedJoker)
				g.Printf("✨ Purchased %s! ✨\n", selectedJoker.Name)
				g.Printf("💰 Remaining money: $%d\n", g.money)

				// Remove purchased item from shop
				shopItems[choice-1] = Joker{} // Empty slot

				// Update available jokers list
				for i, joker := range availableJokers {
					if joker.Name == selectedJoker.Name {
						availableJokers = append(availableJokers[:i], availableJokers[i+1:]...)
						break
					}
				}

				g.Println()

				// Recursively call shop to allow more purchases
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.Printf("Not enough money! Need $%d more.\n", selectedJoker.Price-g.money)
				return
			}
		} else {
			g.Println("Invalid choice.")
			return
		}
	}
}

// showShopWithItems is a helper function to continue shop with specific items
func (g *Game) showShopWithItems(availableJokers []Joker, shopItems []Joker) {
	// Display shop with current items
	g.Println("🏪 SHOP 🏪")
	g.Printf("💰 Your Money: $%d\n", g.money)
	g.Println()

	// Show shop items
	availableSlots := 0
	for i, joker := range shopItems {
		if joker.Name != "" { // Item still available
			canAfford := g.money >= joker.Price
			affordText := ""
			if !canAfford {
				affordText = " (can't afford)"
			}
			g.Printf("%d. %s - $%d%s\n", i+1, joker.Name, joker.Price, affordText)
			g.Printf("   %s\n", joker.Description)
			g.Println()
			availableSlots++
		}
	}

	// Show current jokers
	if len(g.jokers) > 0 {
		g.Printf("🃏 Your Jokers: %s\n", FormatJokersList(g.jokers))
		g.Println()
	}

	// If no items left, exit shop
	if availableSlots == 0 {
		g.Println("Shop sold out!")
		g.Println("Press enter to continue...")
		g.scanner.Scan()
		return
	}

	// Show options and handle single input
	g.Printf("Buy item (1-%d), (r)eroll ($%d), or (s)kip shop: ", len(shopItems), g.rerollCost)

	if g.scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(g.scanner.Text()))

		if input == "s" || input == "skip" {
			g.Println("Skipped shop.")
			return
		} else if input == "r" || input == "reroll" {
			if g.money >= g.rerollCost {
				g.money -= g.rerollCost
				g.Printf("💫 Rerolled for $%d!\n", g.rerollCost)
				g.rerollCost++ // Increase cost for next reroll

				// Generate new shop items
				if len(availableJokers) >= 2 {
					// Re-shuffle for new items
					shuffled := make([]Joker, len(availableJokers))
					copy(shuffled, availableJokers)

					// Fisher-Yates shuffle for reroll
					for i := len(shuffled) - 1; i > 0; i-- {
						j := rand.Intn(i + 1)
						shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
					}

					shopItems = shuffled[:2]
				} else {
					shopItems = availableJokers
				}

				g.Printf("💰 Remaining money: $%d (Next reroll: $%d)\n", g.money, g.rerollCost)
				g.Println()

				// Recursively call shop to show new items
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.Printf("Not enough money to reroll! Need $%d more.\n", g.rerollCost-g.money)
				return
			}
		} else if choice, err := strconv.Atoi(input); err == nil && choice >= 1 && choice <= len(shopItems) {
			selectedJoker := shopItems[choice-1]
			if selectedJoker.Name == "" {
				g.Println("That slot is empty!")
				return
			}

			if g.money >= selectedJoker.Price {
				g.money -= selectedJoker.Price
				g.jokers = append(g.jokers, selectedJoker)
				g.Printf("✨ Purchased %s! ✨\n", selectedJoker.Name)
				g.Printf("💰 Remaining money: $%d\n", g.money)

				// Remove purchased item from shop
				shopItems[choice-1] = Joker{} // Empty slot

				// Update available jokers list
				for i, joker := range availableJokers {
					if joker.Name == selectedJoker.Name {
						availableJokers = append(availableJokers[:i], availableJokers[i+1:]...)
						break
					}
				}

				g.Println()

				// Recursively call shop to allow more purchases
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.Printf("Not enough money! Need $%d more.\n", selectedJoker.Price-g.money)
				return
			}
		} else {
			g.Println("Invalid choice.")
			return
		}
	}
}

// showGameResults displays the final game results for a failed blind
func (g *Game) showGameResults() {
	g.Println(strings.Repeat("=", 50))
	g.Println("💀 DEFEAT! You failed to beat the blind.")
	g.Printf("Final Score: %d/%d (Ante %d - %s)\n", g.totalScore, g.currentTarget, g.currentAnte, g.currentBlind)
	g.Printf("Hands Played: %d/%d\n", g.handsPlayed, MaxHands)
	g.Printf("Discards Used: %d/%d\n", g.discardsUsed, MaxDiscards)
	g.Println(strings.Repeat("=", 50))
}

// showVictoryResults displays the final victory results
func (g *Game) showVictoryResults() {
	g.Println(strings.Repeat("=", 60))
	g.Println("🏆 ULTIMATE VICTORY! 🏆")
	g.Println("🎉 You have conquered all 8 Antes! 🎉")
	g.Printf("Final Ante: %d\n", MaxAntes)
	g.Println("You are a true Balatro master!")
	g.Println(strings.Repeat("=", 60))
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
		g.Println("Cards now sorted by suit")
	} else {
		g.sortMode = SortByRank
		g.Println("Cards now sorted by rank")
	}
	g.Println()
}
