package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
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
	currentBoss       Boss
	money             int
	jokers            []Joker
	rerollCost        int
	eventEmitter      *SimpleEventEmitter
}

// handSize returns the number of cards the player should hold based on jokers
func (g *Game) handSize() int {
	size := InitialCards
	for _, j := range g.jokers {
		if j.Effect == AddHandSize {
			size += j.EffectMagnitude
		}
	}
	return size
}

// maxDiscards returns allowed discards based on jokers
func (g *Game) maxDiscards() int {
	max := MaxDiscards
	for _, j := range g.jokers {
		if j.Effect == AddDiscards {
			max += j.EffectMagnitude
		}
	}
	return max
}

type PrintMode int

const (
	PrintModeLogger PrintMode = iota
	PrintModeTUI
)

// NewGame creates a new game instance
func NewGame(eventHandler EventHandler) *Game {
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
		eventEmitter: NewEventEmitter(),
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

	// Load boss configurations
	if err := LoadBossConfigs(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Set initial target
	game.currentTarget = GetAnteRequirement(game.currentAnte, game.currentBlind)

	// Deal initial hand
	initial := game.handSize()
	game.playerCards = make([]Card, initial)
	copy(game.playerCards, deck[game.deckIndex:game.deckIndex+initial])
	game.deckIndex += initial

	// Initialize display-to-original mapping
	game.displayToOriginal = make([]int, len(game.playerCards))
	for i := range game.playerCards {
		game.displayToOriginal[i] = i
	}

	// Set the event handler
	game.eventEmitter.SetEventHandler(eventHandler)

	return game
}

// Run starts the main game loop
func (g *Game) Run() {
	g.eventEmitter.EmitGameStarted()

	gameRunning := true
	shouldSave := false
	for gameRunning && g.currentAnte <= MaxAntes {
		for g.handsPlayed < MaxHands && g.totalScore < g.currentTarget {
			// Update display mapping and emit current state
			g.updateDisplayToOriginalMapping()
			g.eventEmitter.EmitGameState(g.currentAnte, g.currentBlind, g.currentTarget, g.totalScore,
				MaxHands-g.handsPlayed, g.maxDiscards()-g.discardsUsed, g.money, g.jokers)
			g.eventEmitter.EmitCardsDealt(g.playerCards, g.displayToOriginal, g.sortMode)

			action, params, quit := g.eventEmitter.handler.GetPlayerAction(g.discardsUsed < g.maxDiscards())
			if quit {
				g.eventEmitter.EmitInfo("Thanks for playing!")
				shouldSave = true
				gameRunning = false
				break
			}

			if action == PlayerActionPlay {
				g.handlePlayAction(params)
			} else if action == PlayerActionDiscard {
				g.handleDiscardAction(params)
			} else if action == PlayerActionResort {
				g.handleResortAction()
			} else if action == PlayerActionMoveJoker {
				g.handleMoveJokerAction(params)
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
			g.eventEmitter.EmitEvent(GameOverEvent{
				FinalScore: g.totalScore,
				Target:     g.currentTarget,
				Ante:       g.currentAnte,
			})
			break
		}
	}

	if gameRunning && g.currentAnte > MaxAntes {
		g.eventEmitter.EmitEvent(VictoryEvent{})
	}

	if shouldSave {
		if filename, err := g.Save(); err != nil {
			g.eventEmitter.EmitError(fmt.Sprintf("Failed to save game: %v", err))
		} else {
			g.eventEmitter.EmitInfo(fmt.Sprintf("Game saved to %s", filename))
		}
	}

	g.eventEmitter.handler.Close()
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
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "play",
			Reason: "Please specify cards to play: 'play 1 2 3'",
		})
		return
	}

	if len(params) > 5 {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "play",
			Reason: "You can only play up to 5 cards!",
		})
		return
	}

	selectedCards, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedCards) == 0 {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "play",
			Reason: "Please select at least one card!",
		})
		return
	}

	// Apply replay effects for matching cards
	cardsForJokers, extraCardValue := ApplyReplayCardEffects(g.jokers, selectedCards)

	// Evaluate the hand
	hand := Hand{Cards: selectedCards}
	evaluator, _, cardValues, baseScore := EvaluateHand(hand)
	cardValues += extraCardValue

	// Calculate joker bonuses using cards including replays
	jokerChips, jokerMult := CalculateJokerHandBonus(g.jokers, evaluator.Name(), cardsForJokers)

	// Apply joker bonuses to final score
	finalBaseScore := baseScore + jokerChips
	finalMult := evaluator.Multiplier() + jokerMult
	finalScore := (finalBaseScore + cardValues) * finalMult

	// Emit hand played event with all the details
	g.eventEmitter.EmitEvent(HandPlayedEvent{
		SelectedCards: selectedCards,
		HandType:      evaluator.Name(),
		BaseScore:     baseScore,
		CardValues:    cardValues,
		Multiplier:    evaluator.Multiplier(),
		JokerChips:    jokerChips,
		JokerMult:     jokerMult,
		FinalScore:    finalScore,
		NewTotalScore: g.totalScore + finalScore,
	})

	// Update game state
	g.totalScore += finalScore
	g.handsPlayed++

	// Remove played cards and deal new ones
	g.removeAndDealCards(selectedIndices)
}

// handleDiscardAction processes a discard action
func (g *Game) handleDiscardAction(params []string) {
	if g.discardsUsed >= g.maxDiscards() {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "discard",
			Reason: "No discards remaining!",
		})
		return
	}

	if len(params) < 1 {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "discard",
			Reason: "Please specify cards to discard: 'discard 1 2'",
		})
		return
	}

	selectedCards, selectedIndices, valid := g.parseCardSelection(params)
	if !valid {
		return
	}

	if len(selectedIndices) == 0 {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "discard",
			Reason: "Please select at least one card!",
		})
		return
	}

	// Update discard count
	g.discardsUsed++

	// Emit discard event before removing cards
	g.eventEmitter.EmitEvent(CardsDiscardedEvent{
		DiscardedCards: selectedCards,
		NumCards:       len(selectedCards),
		DiscardsLeft:   g.maxDiscards() - g.discardsUsed,
	})

	// Remove discarded cards and deal new ones
	g.removeAndDealCards(selectedIndices)
}

// parseCardSelection parses card selection from string parameters
func (g *Game) parseCardSelection(params []string) ([]Card, []int, bool) {
	var selectedCards []Card
	var selectedIndices []int
	seen := make(map[int]bool)

	for _, param := range params {
		displayIndex, err := strconv.Atoi(param)
		if err != nil || displayIndex < 1 || displayIndex > len(g.playerCards) {
			g.eventEmitter.EmitEvent(InvalidActionEvent{
				Action: "card_selection",
				Reason: fmt.Sprintf("Invalid card number: %s", param),
			})
			return nil, nil, false
		}

		if seen[displayIndex] {
			g.eventEmitter.EmitWarning(fmt.Sprintf("Duplicate card number: %d ignored", displayIndex))
			continue
		}
		seen[displayIndex] = true

		// Since playerCards is already sorted by updateDisplayToOriginalMapping,
		// display position directly corresponds to current array position
		arrayIndex := displayIndex - 1 // Convert 1-based to 0-based
		selectedCards = append(selectedCards, g.playerCards[arrayIndex])
		selectedIndices = append(selectedIndices, arrayIndex)
	}

	return selectedCards, selectedIndices, true
}

// removeAndDealCards removes selected cards and deals new ones
func (g *Game) removeAndDealCards(selectedIndices []int) {
	g.playerCards = RemoveCards(g.playerCards, selectedIndices)
	newCardsNeeded := len(selectedIndices)

	// Deal new cards if available
	if g.deckIndex+newCardsNeeded <= len(g.deck) {
		for i := 0; i < newCardsNeeded; i++ {
			g.playerCards = append(g.playerCards, g.deck[g.deckIndex])
			g.deckIndex++
		}
	}

	// Update display mapping after dealing new cards to maintain sort order
	g.updateDisplayToOriginalMapping()
}

// handleBlindCompletion handles completing a blind and advancing to the next
func (g *Game) handleBlindCompletion() {

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
	unusedDiscards := g.maxDiscards() - g.discardsUsed
	bonusReward := unusedHands*UnusedHandReward + unusedDiscards*UnusedDiscardReward
	jokerReward := CalculateJokerRewards(g.jokers)
	totalReward := baseReward + bonusReward + jokerReward

	g.money += totalReward

	// Emit blind defeated event with all reward information
	g.eventEmitter.EmitEvent(BlindDefeatedEvent{
		BlindType:      g.currentBlind,
		Score:          g.totalScore,
		Target:         g.currentTarget,
		BaseReward:     baseReward,
		BonusReward:    bonusReward,
		JokerReward:    jokerReward,
		TotalReward:    totalReward,
		NewMoney:       g.money,
		UnusedHands:    unusedHands,
		UnusedDiscards: unusedDiscards,
	})

	// Advance to next blind
	if g.currentBlind == SmallBlind {
		g.currentBlind = BigBlind
	} else if g.currentBlind == BigBlind {
		g.currentBlind = BossBlind
	} else {
		// Completed Boss Blind, advance to next ante
		oldAnte := g.currentAnte
		g.currentAnte++
		g.currentBlind = SmallBlind
		if g.currentAnte <= MaxAntes {
			g.eventEmitter.EmitEvent(AnteCompletedEvent{
				CompletedAnte: oldAnte,
				NewAnte:       g.currentAnte,
			})
		}
	}

	if g.currentAnte <= MaxAntes {
		// Reset for next blind
		g.totalScore = 0
		g.handsPlayed = 0
		g.discardsUsed = 0
		g.rerollCost = 5 // Reset reroll cost for new blind
		g.currentTarget = GetAnteRequirement(g.currentAnte, g.currentBlind)

		if g.currentBlind == BossBlind {
			g.currentBoss = GetBossForAnte(g.currentAnte)
			g.applyBossEffect()
		} else {
			g.currentBoss = Boss{}
		}

		// Shuffle and deal new hand
		g.deckIndex = 0
		ShuffleDeck(g.deck)
		handSize := g.handSize()
		g.playerCards = make([]Card, handSize)
		copy(g.playerCards, g.deck[g.deckIndex:g.deckIndex+handSize])
		g.deckIndex += handSize

		// Show next blind info
		var boss *Boss
		if g.currentBlind == BossBlind {
			boss = &g.currentBoss
		}
		g.eventEmitter.EmitEvent(NewBlindStartedEvent{
			Ante:     g.currentAnte,
			Blind:    g.currentBlind,
			Target:   g.currentTarget,
			NewCards: g.playerCards,
			Boss:     boss,
		})

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
		g.eventEmitter.EmitMessage("All available jokers already owned!", "info")
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

	// Convert jokers to shop item data
	var items []ShopItemData
	for _, joker := range shopItems {
		if joker.Name != "" {
			items = append(items, NewShopItemData(joker, g.money))
		} else {
			// Preserve empty slots so indices remain stable
			items = append(items, ShopItemData{})
		}
	}

	// Emit shop opened event
	g.eventEmitter.EmitEvent(ShopOpenedEvent{
		Money:      g.money,
		RerollCost: g.rerollCost,
		Items:      items,
	})

	for {
		action, params, quit := g.eventEmitter.handler.GetShopAction()
		if quit {
			return
		}

		if action == PlayerActionExitShop {
			g.eventEmitter.EmitEvent(ShopClosedEvent{})
			return
		} else if action == PlayerActionReroll {
			if g.money >= g.rerollCost {
				oldCost := g.rerollCost
				g.money -= g.rerollCost
				g.rerollCost += 2

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

				// Convert new items and emit reroll event
				var newItems []ShopItemData
				for _, joker := range shopItems {
					if joker.Name != "" {
						newItems = append(newItems, NewShopItemData(joker, g.money))
					} else {
						newItems = append(newItems, ShopItemData{})
					}
				}

				g.eventEmitter.EmitEvent(ShopRerolledEvent{
					Cost:           oldCost,
					NewRerollCost:  g.rerollCost,
					RemainingMoney: g.money,
					NewItems:       newItems,
				})

				// Recursively call shop with new items
				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.eventEmitter.EmitEvent(InvalidActionEvent{
					Action: "reroll",
					Reason: fmt.Sprintf("Not enough money to reroll! Need $%d more.", g.rerollCost-g.money),
				})
			}
		} else if action == PlayerActionMoveJoker {
			g.handleMoveJokerAction(params)
		} else if action == PlayerActionBuy {
			choice := params[0]
			if choice, err := strconv.Atoi(choice); err == nil && choice >= 1 && choice <= len(shopItems) {
				selectedJoker := shopItems[choice-1]
				if selectedJoker.Name == "" {
					g.eventEmitter.EmitEvent(InvalidActionEvent{
						Action: "buy",
						Reason: "That slot is empty!",
					})
					continue
				}

				if g.money >= selectedJoker.Price {
					g.money -= selectedJoker.Price
					g.jokers = append(g.jokers, selectedJoker)

					g.eventEmitter.EmitEvent(ShopItemPurchasedEvent{
						Item:           NewShopItemData(selectedJoker, g.money+selectedJoker.Price),
						RemainingMoney: g.money,
					})

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
					g.eventEmitter.EmitEvent(InvalidActionEvent{
						Action: "buy",
						Reason: fmt.Sprintf("Not enough money! Need $%d more.", selectedJoker.Price-g.money),
					})
				}
			} else {
				g.eventEmitter.EmitEvent(InvalidActionEvent{
					Action: "buy",
					Reason: fmt.Sprintf("Invalid item number (given %v).", params),
				})
			}
		} else {
			g.eventEmitter.EmitEvent(InvalidActionEvent{
				Action: "unknown",
				Reason: fmt.Sprintf("Invalid action (given '%s'). Use 'buy <number>', 'reroll', or 'exit'.", action),
			})
		}
	}
}

// showShopWithItems is a helper function to continue shop with specific items
func (g *Game) showShopWithItems(availableJokers []Joker, shopItems []Joker) {
	// Convert jokers to shop item data for display
	var items []ShopItemData
	for _, joker := range shopItems {
		if joker.Name != "" {
			items = append(items, NewShopItemData(joker, g.money))
		} else {
			// Keep placeholder to maintain indexing
			items = append(items, ShopItemData{})
		}
	}

	// Check if shop is empty
	if len(items) == 0 {
		g.eventEmitter.EmitMessage("Shop sold out!", "info")
		g.eventEmitter.handler.GetShopAction() // Wait for user to continue
		return
	}

	// Emit shop opened event with current items
	g.eventEmitter.EmitEvent(ShopOpenedEvent{
		Money:      g.money,
		RerollCost: g.rerollCost,
		Items:      items,
	})

	for {
		action, params, quit := g.eventEmitter.handler.GetShopAction()
		if quit {
			return
		}

		if action == PlayerActionExitShop {
			g.eventEmitter.EmitEvent(ShopClosedEvent{})
			return
		} else if action == PlayerActionReroll {
			if g.money >= g.rerollCost {
				oldCost := g.rerollCost
				g.money -= g.rerollCost
				g.rerollCost += 2

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

				// Convert new items and emit reroll event
				var newItems []ShopItemData
				for _, joker := range shopItems {
					if joker.Name != "" {
						newItems = append(newItems, NewShopItemData(joker, g.money))
					} else {
						newItems = append(newItems, ShopItemData{})
					}
				}

				g.eventEmitter.EmitEvent(ShopRerolledEvent{
					Cost:           oldCost,
					NewRerollCost:  g.rerollCost,
					RemainingMoney: g.money,
					NewItems:       newItems,
				})

				g.showShopWithItems(availableJokers, shopItems)
				return
			} else {
				g.eventEmitter.EmitEvent(InvalidActionEvent{
					Action: "reroll",
					Reason: fmt.Sprintf("Not enough money to reroll! Need $%d more.", g.rerollCost-g.money),
				})
			}
		} else if action == PlayerActionMoveJoker {
			g.handleMoveJokerAction(params)
		} else if action == PlayerActionBuy {
			choice := params[0]
			if choice, err := strconv.Atoi(choice); err == nil && choice >= 1 && choice <= len(shopItems) {
				selectedJoker := shopItems[choice-1]
				if selectedJoker.Name == "" {
					g.eventEmitter.EmitEvent(InvalidActionEvent{
						Action: "buy",
						Reason: "That slot is empty!",
					})
					continue
				}

				if g.money >= selectedJoker.Price {
					g.money -= selectedJoker.Price
					g.jokers = append(g.jokers, selectedJoker)

					g.eventEmitter.EmitEvent(ShopItemPurchasedEvent{
						Item:           NewShopItemData(selectedJoker, g.money+selectedJoker.Price),
						RemainingMoney: g.money,
					})

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
					g.eventEmitter.EmitEvent(InvalidActionEvent{
						Action: "buy",
						Reason: fmt.Sprintf("Not enough money! Need $%d more.", selectedJoker.Price-g.money),
					})
				}
			} else {
				g.eventEmitter.EmitEvent(InvalidActionEvent{
					Action: "buy",
					Reason: fmt.Sprintf("Invalid item number (given %v).", params),
				})
			}
		} else {
			g.eventEmitter.EmitEvent(InvalidActionEvent{
				Action: "unknown",
				Reason: fmt.Sprintf("Invalid action (given '%s'). Use 'buy <number>', 'reroll', or 'exit'.", action),
			})
		}
	}
}

// removeCards removes cards at specified indices and returns the new slice
func RemoveCards(cards []Card, indices []int) []Card {
	// Use a set to ensure each index is only removed once
	uniqueSet := make(map[int]struct{})
	for _, idx := range indices {
		uniqueSet[idx] = struct{}{}
	}

	// Collect unique indices and sort in descending order to remove from end first
	uniqueIndices := make([]int, 0, len(uniqueSet))
	for idx := range uniqueSet {
		uniqueIndices = append(uniqueIndices, idx)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(uniqueIndices)))

	result := make([]Card, len(cards))
	copy(result, cards)

	for _, index := range uniqueIndices {
		if index >= 0 && index < len(result) {
			result = append(result[:index], result[index+1:]...)
		}
	}

	return result
}

// handleResortAction toggles the card sort mode and updates the display mapping
func (g *Game) handleResortAction() {
	if g.sortMode == SortByRank {
		g.sortMode = SortBySuit
		g.eventEmitter.EmitEvent(CardsResortedEvent{
			NewSortMode: "suit",
		})
	} else {
		g.sortMode = SortByRank
		g.eventEmitter.EmitEvent(CardsResortedEvent{
			NewSortMode: "rank",
		})
	}
	// Update the display mapping with new sort order
	g.updateDisplayToOriginalMapping()
}

// handleMoveJokerAction moves a joker up or down in the player's joker list
func (g *Game) handleMoveJokerAction(params []string) {
	if len(params) != 2 {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "move_joker",
			Reason: "Usage: move_joker <index> <up|down>",
		})
		return
	}

	idx, err := strconv.Atoi(params[0])
	if err != nil || idx < 1 || idx > len(g.jokers) {
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "move_joker",
			Reason: fmt.Sprintf("Invalid joker number: %s", params[0]),
		})
		return
	}

	direction := params[1]
	i := idx - 1
	switch direction {
	case "up":
		if i == 0 {
			g.eventEmitter.EmitEvent(InvalidActionEvent{
				Action: "move_joker",
				Reason: "Joker already at top",
			})
			return
		}
		g.jokers[i-1], g.jokers[i] = g.jokers[i], g.jokers[i-1]
	case "down":
		if i >= len(g.jokers)-1 {
			g.eventEmitter.EmitEvent(InvalidActionEvent{
				Action: "move_joker",
				Reason: "Joker already at bottom",
			})
			return
		}
		g.jokers[i], g.jokers[i+1] = g.jokers[i+1], g.jokers[i]
	default:
		g.eventEmitter.EmitEvent(InvalidActionEvent{
			Action: "move_joker",
			Reason: fmt.Sprintf("Unknown direction: %s", direction),
		})
		return
	}

	g.eventEmitter.EmitGameState(g.currentAnte, g.currentBlind, g.currentTarget, g.totalScore,
		MaxHands-g.handsPlayed, g.maxDiscards()-g.discardsUsed, g.money, g.jokers)
}

// applyBossEffect modifies game state based on the current boss's effect
func (g *Game) applyBossEffect() {
	switch g.currentBoss.Effect {
	case DoubleChips:
		g.currentTarget *= 2
	case HalveMoney:
		g.money /= 2
	}
}
