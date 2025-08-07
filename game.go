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
	totalScore        int
	handsPlayed       int
	discardsUsed      int
	deck              []Card
	deckIndex         int
	playerCards       []Card
	displayToOriginal []int // maps display position (0-based) to original position
	sortMode          SortMode
	currentAnte       int
	currentBlind      BlindType
	currentTarget     int
	money             int
	jokers            []Joker
	rerollCost        int
	io                GameIO
}

type PrintMode int

const (
	PrintModeLogger PrintMode = iota
	PrintModeTUI
)

// NewGame creates a new game instance
func NewGame(gameIO GameIO) *Game {
	deck := NewDeck()
	ShuffleDeck(deck)

	game := &Game{
		totalScore:   0,
		handsPlayed:  0,
		discardsUsed: 0,
		deck:         deck,
		deckIndex:    0,
		sortMode:     SortByRank,
		currentAnte:  1,
		currentBlind: SmallBlind,
		money:        StartingMoney,
		jokers:       []Joker{},
		rerollCost:   5, // Initial reroll cost
		io:           gameIO,
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

	// Initialize display-to-original mapping
	game.displayToOriginal = make([]int, len(game.playerCards))
	for i := range game.playerCards {
		game.displayToOriginal[i] = i
	}

	return game
}

// Run starts the main game loop
func (g *Game) Run() {
	g.io.ShowWelcome()

	gameRunning := true
	for gameRunning && g.currentAnte <= MaxAntes {
		for g.handsPlayed < MaxHands && g.totalScore < g.currentTarget {
			g.io.ShowGameStatus(g.currentAnte, g.currentBlind, g.currentTarget, g.totalScore,
				MaxHands-g.handsPlayed, MaxDiscards-g.discardsUsed, g.money, g.jokers)
			g.io.ShowCards(g.playerCards, g.displayToOriginal)

			action, params, quit := g.io.GetPlayerAction(g.discardsUsed < MaxDiscards)
			if quit {
				g.io.ShowMessage("Thanks for playing!")
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
			g.io.ShowGameResults(g.totalScore, g.currentTarget, g.currentAnte)
			break
		}
	}

	if gameRunning && g.currentAnte > MaxAntes {
		g.io.ShowVictoryResults()
	}

	g.io.Close()
}

// updateDisplayToOriginalMapping sorts cards and updates the display mapping
func (g *Game) updateDisplayToOriginalMapping() {
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

	// Update the display-to-original mapping and reorder playerCards
	g.displayToOriginal = make([]int, len(indexed))
	sortedCards := make([]Card, len(indexed))
	for i, ic := range indexed {
		g.displayToOriginal[i] = ic.index
		sortedCards[i] = ic.card
	}
	g.playerCards = sortedCards
}

// handlePlayAction processes a play action
func (g *Game) handlePlayAction(params []string) {
	if len(params) < 1 {
		g.io.ShowMessage("Please specify cards to play: 'play 1 2 3'")
		return
	}

	if len(params) > 5 {
		g.io.ShowMessage("You can only play up to 5 cards!")
		return
	}

	selectedCards, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedCards) == 0 {
		g.io.ShowMessage("Please select at least one card!")
		return
	}

	// Evaluate the hand
	hand := Hand{Cards: selectedCards}
	evaluator, _, cardValues, baseScore := EvaluateHand(hand)

	// Calculate joker bonuses
	jokerChips, jokerMult := CalculateJokerHandBonus(g.jokers, evaluator.Name())

	// Apply joker bonuses to final score
	finalBaseScore := baseScore + jokerChips
	finalMult := evaluator.Multiplier() + jokerMult
	finalScore := (finalBaseScore + cardValues) * finalMult

	var message strings.Builder
	message.WriteString(fmt.Sprintf("\nYour hand: %s\n", hand))
	message.WriteString(fmt.Sprintf("Hand type: %s\n", evaluator.Name()))

	if jokerChips > 0 || jokerMult > 0 {
		message.WriteString(fmt.Sprintf("Base Score: %d", baseScore))
		if jokerChips > 0 {
			message.WriteString(fmt.Sprintf(" + %d Joker Chips", jokerChips))
		}
		message.WriteString(fmt.Sprintf(" | Card Values: %d | Mult: %dx", cardValues, evaluator.Multiplier()))
		if jokerMult > 0 {
			message.WriteString(fmt.Sprintf(" + %d Joker Mult", jokerMult))
		}
		message.WriteString("\n")
		message.WriteString(fmt.Sprintf("Final Score: (%d + %d) × %d = %d points\n", finalBaseScore, cardValues, finalMult, finalScore))
	} else {
		message.WriteString(fmt.Sprintf("Base Score: %d | Card Values: %d | Mult: %dx\n", baseScore, cardValues, evaluator.Multiplier()))
		message.WriteString(fmt.Sprintf("Final Score: (%d + %d) × %d = %d points\n", baseScore, cardValues, evaluator.Multiplier(), finalScore))
	}

	// Use the joker-modified score
	g.totalScore += finalScore
	g.handsPlayed++

	message.WriteString(fmt.Sprintf("💰 Total Score: %d/%d\n", g.totalScore, g.currentTarget))
	message.WriteString(strings.Repeat("-", 50))
	g.io.ShowMessage(message.String())

	// Remove played cards and deal new ones
	g.removeAndDealCards(selectedIndices)
}

// handleDiscardAction processes a discard action
func (g *Game) handleDiscardAction(params []string) {
	if g.discardsUsed >= MaxDiscards {
		g.io.ShowMessage("No discards remaining!")
		return
	}

	if len(params) < 1 {
		g.io.ShowMessage("Please specify cards to discard: 'discard 1 2'")
		return
	}

	_, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedIndices) == 0 {
		g.io.ShowMessage("Please select at least one card!")
		return
	}

	g.io.ShowMessage(fmt.Sprintf("Discarded %d card(s)", len(selectedIndices)))
	g.discardsUsed++

	// Remove discarded cards and deal new ones
	g.removeAndDealCards(selectedIndices)

	g.io.ShowMessage("New cards dealt!")
}

// parseCardSelection parses card selection from string parameters
func (g *Game) parseCardSelection(params []string) ([]Card, []int, bool) {
	var selectedCards []Card
	var selectedIndices []int

	for _, param := range params {
		displayIndex, err := strconv.Atoi(param)
		if err != nil || displayIndex < 1 || displayIndex > len(g.playerCards) {
			g.io.ShowMessage(fmt.Sprintf("Invalid card number: %s", param))
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
	// Show completion message based on blind type
	var completionMsg string
	switch g.currentBlind {
	case SmallBlind:
		completionMsg = fmt.Sprintf("%s\n🔸 SMALL BLIND DEFEATED! 🔸\n    ✨ Score: %d/%d ✨\n   🎯 Advancing to Big Blind...\n%s",
			strings.Repeat("=", 60), g.totalScore, g.currentTarget, strings.Repeat("=", 60))
	case BigBlind:
		completionMsg = fmt.Sprintf("%s\n🔶 BIG BLIND CRUSHED! 🔶\n    ⚡ Score: %d/%d ⚡\n   💀 Prepare for the Boss Blind...\n%s",
			strings.Repeat("=", 60), g.totalScore, g.currentTarget, strings.Repeat("=", 60))
	case BossBlind:
		anteMsg := ""
		if g.currentAnte < MaxAntes {
			anteMsg = fmt.Sprintf("\n🎊 ANTE %d CONQUERED! 🎊", g.currentAnte)
		}
		completionMsg = fmt.Sprintf("%s\n💀 BOSS BLIND ANNIHILATED! 💀\n    🔥 EPIC SCORE: %d/%d 🔥%s\n%s",
			strings.Repeat("🎆", 15), g.totalScore, g.currentTarget, anteMsg, strings.Repeat("🎆", 15))
	}
	g.io.ShowMessage(completionMsg)

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

	var rewardMsg strings.Builder
	rewardMsg.WriteString("💰 REWARD BREAKDOWN:\n")
	rewardMsg.WriteString(fmt.Sprintf("   Base: $%d", baseReward))
	if bonusReward > 0 {
		rewardMsg.WriteString(fmt.Sprintf(" + Unused: $%d (%d hands + %d discards)", bonusReward, unusedHands, unusedDiscards))
	}
	if jokerReward > 0 {
		rewardMsg.WriteString(fmt.Sprintf(" + Jokers: $%d", jokerReward))
	}
	rewardMsg.WriteString(fmt.Sprintf("\n   💰 Total Earned: $%d | Your Money: $%d", totalReward, g.money))
	g.io.ShowMessage(rewardMsg.String())

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
			progressionMsg := fmt.Sprintf("🌟 ANTE PROGRESSION 🌟\n   Ante %d → Ante %d\n🔄 NEW CHALLENGE AWAITS!", g.currentAnte-1, g.currentAnte)
			g.io.ShowMessage(progressionMsg)
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

		nextBlindMsg := fmt.Sprintf("%s NOW ENTERING: %s (Ante %d) %s\n🎯 NEW TARGET: %d points\n🃏 Fresh hand dealt!\n%s",
			blindEmoji, g.currentBlind, g.currentAnte, blindEmoji, g.currentTarget, strings.Repeat("-", 40))
		g.io.ShowMessage(nextBlindMsg)

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
		g.io.ShowMessage(fmt.Sprintf("🏪 SHOP 🏪\n💰 Your Money: $%d\n\nAll available jokers already owned!\nPress enter to continue...", g.money))
		g.io.GetShopAction() // Wait for user to continue
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

	// Convert jokers to shop items
	var items []ShopItem
	for _, joker := range shopItems {
		if joker.Name != "" {
			items = append(items, ShopItem{
				Name:        joker.Name,
				Description: joker.Description,
				Cost:        joker.Price,
				Type:        "joker",
			})
		}
	}

	// Show shop and handle interactions through GameIO
	g.io.ShowShop(items, g.money, g.rerollCost)

	for {
		action, quit := g.io.GetShopAction()
		if quit {
			return
		}

		if action == "exit" {
			g.io.ShowMessage("Left the shop.")
			return
		} else if action == "reroll" {
			if g.money >= g.rerollCost {
				g.money -= g.rerollCost
				g.io.ShowMessage(fmt.Sprintf("💫 Rerolled for $%d! Next reroll: $%d", g.rerollCost, g.rerollCost+1))
				g.rerollCost++

				// Generate new shop items
				if len(availableJokers) >= 2 {
					shuffled := make([]Joker, len(availableJokers))
					copy(shuffled, availableJokers)
					for i := len(shuffled) - 1; i > 0; i-- {
						j := rand.Intn(i + 1)
						shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
					}
					shopItems = shuffled[:2]
				} else {
					shopItems = availableJokers
				}

				// Recursively call shop with new items
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.io.ShowMessage(fmt.Sprintf("Not enough money to reroll! Need $%d more.", g.rerollCost-g.money))
			}
		} else if strings.HasPrefix(action, "buy ") {
			choiceStr := strings.TrimPrefix(action, "buy ")
			if choice, err := strconv.Atoi(choiceStr); err == nil && choice >= 1 && choice <= len(shopItems) {
				selectedJoker := shopItems[choice-1]
				if selectedJoker.Name == "" {
					g.io.ShowMessage("That slot is empty!")
					continue
				}

				if g.money >= selectedJoker.Price {
					g.money -= selectedJoker.Price
					g.jokers = append(g.jokers, selectedJoker)
					g.io.ShowMessage(fmt.Sprintf("✨ Purchased %s! ✨\n💰 Remaining money: $%d", selectedJoker.Name, g.money))

					// Remove purchased item and update available jokers
					shopItems[choice-1] = Joker{}
					for i, joker := range availableJokers {
						if joker.Name == selectedJoker.Name {
							availableJokers = append(availableJokers[:i], availableJokers[i+1:]...)
							break
						}
					}

					g.showShopWithItems(availableJokers, shopItems)
					return
				} else {
					g.io.ShowMessage(fmt.Sprintf("Not enough money! Need $%d more.", selectedJoker.Price-g.money))
				}
			} else {
				g.io.ShowMessage("Invalid item number.")
			}
		} else {
			g.io.ShowMessage("Invalid action. Use 'buy <number>', 'reroll', or 'exit'.")
		}
	}
}

// showShopWithItems is a helper function to continue shop with specific items
func (g *Game) showShopWithItems(availableJokers []Joker, shopItems []Joker) {
	// Convert jokers to shop items for display
	var items []ShopItem
	for _, joker := range shopItems {
		if joker.Name != "" {
			items = append(items, ShopItem{
				Name:        joker.Name,
				Description: joker.Description,
				Cost:        joker.Price,
				Type:        "joker",
			})
		}
	}

	// Check if shop is empty
	if len(items) == 0 {
		g.io.ShowMessage("Shop sold out!")
		g.io.GetShopAction() // Wait for user to continue
		return
	}

	// Show shop through GameIO interface
	g.io.ShowShop(items, g.money, g.rerollCost)

	for {
		action, quit := g.io.GetShopAction()
		if quit {
			return
		}

		if action == "exit" {
			return
		} else if action == "reroll" {
			if g.money >= g.rerollCost {
				g.money -= g.rerollCost
				g.io.ShowMessage(fmt.Sprintf("💫 Rerolled for $%d! Next reroll: $%d", g.rerollCost, g.rerollCost+1))
				g.rerollCost++

				// Generate new shop items
				if len(availableJokers) >= 2 {
					shuffled := make([]Joker, len(availableJokers))
					copy(shuffled, availableJokers)
					for i := len(shuffled) - 1; i > 0; i-- {
						j := rand.Intn(i + 1)
						shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
					}
					shopItems = shuffled[:2]
				} else {
					shopItems = availableJokers
				}

				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.io.ShowMessage(fmt.Sprintf("Not enough money to reroll! Need $%d more.", g.rerollCost-g.money))
			}
		} else if strings.HasPrefix(action, "buy ") {
			choiceStr := strings.TrimPrefix(action, "buy ")
			if choice, err := strconv.Atoi(choiceStr); err == nil && choice >= 1 && choice <= len(shopItems) {
				selectedJoker := shopItems[choice-1]
				if selectedJoker.Name == "" {
					g.io.ShowMessage("That slot is empty!")
					continue
				}

				if g.money >= selectedJoker.Price {
					g.money -= selectedJoker.Price
					g.jokers = append(g.jokers, selectedJoker)
					g.io.ShowMessage(fmt.Sprintf("✨ Purchased %s! ✨\n💰 Remaining money: $%d", selectedJoker.Name, g.money))

					// Remove purchased item
					shopItems[choice-1] = Joker{}
					for i, joker := range availableJokers {
						if joker.Name == selectedJoker.Name {
							availableJokers = append(availableJokers[:i], availableJokers[i+1:]...)
							break
						}
					}

					g.showShopWithItems(availableJokers, shopItems)
					return
				} else {
					g.io.ShowMessage(fmt.Sprintf("Not enough money! Need $%d more.", selectedJoker.Price-g.money))
				}
			} else {
				g.io.ShowMessage("Invalid item number.")
			}
		} else {
			g.io.ShowMessage("Invalid action. Use 'buy <number>', 'reroll', or 'exit'.")
		}
	}
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
// handleResortAction handles resorting cards
func (g *Game) handleResortAction() {
	if g.sortMode == SortByRank {
		g.sortMode = SortBySuit
		g.io.ShowMessage("Cards now sorted by suit")
	} else {
		g.sortMode = SortByRank
		g.io.ShowMessage("Cards now sorted by rank")
	}
	// Update the display mapping with new sort order
	g.updateDisplayToOriginalMapping()
}
