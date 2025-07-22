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
	TargetScore  = 300
	MaxHands     = 4
	MaxDiscards  = 3
	InitialCards = 7
)

// Game represents the current game state
type Game struct {
	totalScore   int
	handsPlayed  int
	discardsUsed int
	deck         []Card
	deckIndex    int
	playerCards  []Card
	scanner      *bufio.Scanner
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
	}

	// Deal initial hand
	game.playerCards = make([]Card, InitialCards)
	copy(game.playerCards, deck[game.deckIndex:game.deckIndex+InitialCards])
	game.deckIndex += InitialCards

	return game
}

// Run starts the main game loop
func (g *Game) Run() {
	fmt.Println("🃏 Welcome to Balatro CLI! 🃏")
	fmt.Println("🎯 CHALLENGE: Score 300 points with 4 hands and 3 discards!")
	fmt.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	fmt.Println()

	for g.handsPlayed < MaxHands && g.totalScore < TargetScore {
		g.showGameStatus()
		g.showCards()

		action, params, quit := g.getPlayerInput()
		if quit {
			fmt.Println("Thanks for playing!")
			break
		}

		if action == "play" {
			g.handlePlayAction(params)
		} else if action == "discard" {
			g.handleDiscardAction(params)
		}
	}

	g.showGameResults()
}

// showGameStatus displays the current game state
func (g *Game) showGameStatus() {
	fmt.Printf("🎯 Target: %d | Current Score: %d | Hands Left: %d | Discards Left: %d\n",
		TargetScore, g.totalScore, MaxHands-g.handsPlayed, MaxDiscards-g.discardsUsed)
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

	// Sort by rank
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].card.Rank < indexed[j].card.Rank
	})

	fmt.Println("Your cards:")
	for _, ic := range indexed {
		fmt.Printf("%d: %s\n", ic.index+1, ic.card)
	}
	fmt.Println()
}

// getPlayerInput reads and parses player input
func (g *Game) getPlayerInput() (string, []string, bool) {
	if g.discardsUsed >= MaxDiscards {
		fmt.Print("Choose action: 'play <cards>' (or 'p <cards>') to play hand, or 'quit': ")
	} else {
		fmt.Print("Choose action: 'play <cards>' (or 'p <cards>') to play hand, 'discard <cards>' (or 'd <cards>') to discard, or 'quit': ")
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

	fmt.Printf("💰 Total Score: %d/%d\n", g.totalScore, TargetScore)
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
		index, err := strconv.Atoi(param)
		if err != nil || index < 1 || index > len(g.playerCards) {
			fmt.Printf("Invalid card number: %s\n", param)
			return nil, nil, false
		}
		selectedCards = append(selectedCards, g.playerCards[index-1])
		selectedIndices = append(selectedIndices, index-1)
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

// showGameResults displays the final game results
func (g *Game) showGameResults() {
	fmt.Println(strings.Repeat("=", 50))
	if g.totalScore >= TargetScore {
		fmt.Println("🎉 VICTORY! You reached the target score!")
	} else {
		fmt.Println("💀 DEFEAT! You ran out of hands.")
	}
	fmt.Printf("Final Score: %d/%d\n", g.totalScore, TargetScore)
	fmt.Printf("Hands Played: %d/%d\n", g.handsPlayed, MaxHands)
	fmt.Printf("Discards Used: %d/%d\n", g.discardsUsed, MaxDiscards)
	fmt.Println(strings.Repeat("=", 50))
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
