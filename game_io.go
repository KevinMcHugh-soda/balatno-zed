package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// GameIO interface abstracts all input/output operations from the game logic
type GameIO interface {
	// Welcome and setup
	ShowWelcome()

	// Game state display
	ShowGameStatus(ante int, blind BlindType, target int, score int, hands int, discards int, money int, jokers []Joker)
	ShowCards(cards []Card, displayToOriginal []int)
	ShowMessage(message string)

	// Player input
	GetPlayerAction(canDiscard bool) (action string, params []string, quit bool)

	// Shop functionality
	ShowShop(items []ShopItem, money int, rerollCost int)
	GetShopAction() (action string, quit bool)

	// Game completion
	ShowGameResults(finalScore int, target int, ante int)
	ShowVictoryResults()

	// Cleanup
	Close()
}

// LoggerIO implements GameIO for console/logger mode
type LoggerIO struct {
	scanner *bufio.Scanner
	logger  *log.Logger
}

// NewLoggerIO creates a new LoggerIO instance
func NewLoggerIO(logger *log.Logger) *LoggerIO {
	return &LoggerIO{
		scanner: bufio.NewScanner(os.Stdin),
		logger:  logger,
	}
}

func (io *LoggerIO) ShowWelcome() {
	fmt.Println("🃏 Welcome to Balatro CLI! 🃏")
	fmt.Println("🎯 CHALLENGE: Progress through 8 Antes, each with 3 Blinds!")
	fmt.Println("Each Ante: Small Blind → Big Blind → Boss Blind")
	fmt.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	fmt.Println()
}

func (io *LoggerIO) ShowGameStatus(ante int, blind BlindType, target int, score int, hands int, discards int, money int, jokers []Joker) {
	fmt.Printf("🎯 Ante %d - %s | Target: %d | Current Score: %d\n", ante, blind, target, score)
	fmt.Printf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d\n", hands, discards, money)

	if len(jokers) > 0 {
		fmt.Print("🃏 Jokers: ")
		for i, joker := range jokers {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s (%s)", joker.Name, joker.Description)
		}
		fmt.Println()
	}
	fmt.Println()
}

func (io *LoggerIO) ShowCards(cards []Card, displayToOriginal []int) {
	fmt.Printf("🃏 Your Hand (%d cards):\n", len(cards))

	for i, card := range cards {
		fmt.Printf("%d: %s ", i+1, card.String())
		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
	fmt.Println("\n")
}

func (io *LoggerIO) ShowMessage(message string) {
	fmt.Println(message)
}

func (io *LoggerIO) GetPlayerAction(canDiscard bool) (string, []string, bool) {
	if canDiscard {
		fmt.Print("(p)lay <cards>, (d)iscard <cards>, (r)esort, or (q)uit: ")
	} else {
		fmt.Print("(p)lay <cards>, (r)esort, or (q)uit: ")
	}

	if !io.scanner.Scan() {
		if err := io.scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
		return "", nil, true
	}

	input := strings.TrimSpace(io.scanner.Text())

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

func (io *LoggerIO) ShowShop(items []ShopItem, money int, rerollCost int) {
	fmt.Println("🛍️  Welcome to the Shop!")
	fmt.Printf("💰 You have $%d\n", money)
	fmt.Printf("🎲 Reroll costs $%d\n", rerollCost)
	fmt.Println()

	fmt.Println("Available items:")
	for i, item := range items {
		fmt.Printf("%d. %s - $%d\n", i+1, item.Name, item.Cost)
		fmt.Printf("   %s\n", item.Description)
		fmt.Println()
	}

	fmt.Println("Commands:")
	fmt.Println("• buy <number> - Purchase an item")
	fmt.Println("• reroll - Reroll the shop items")
	fmt.Println("• exit - Leave the shop")
}

func (io *LoggerIO) GetShopAction() (string, bool) {
	fmt.Print("Shop action (buy <number>, reroll, exit/q): ")

	if !io.scanner.Scan() {
		if err := io.scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
		return "", true
	}

	input := strings.TrimSpace(io.scanner.Text())
	if strings.ToLower(input) == "exit" || strings.ToLower(input) == "q" || input == "" {
		return "exit", false
	}

	return input, false
}

func (io *LoggerIO) ShowGameResults(finalScore int, target int, ante int) {
	fmt.Println("💀 GAME OVER 💀")
	fmt.Printf("Final Score: %d/%d (Ante %d)\n", finalScore, target, ante)
	fmt.Println("Better luck next time!")
}

func (io *LoggerIO) ShowVictoryResults() {
	fmt.Println("🎉 VICTORY! 🎉")
	fmt.Println("Congratulations! You've conquered all 8 Antes!")
	fmt.Println("You are a true Balatro master!")
}

func (io *LoggerIO) Close() {
	// Nothing to clean up for logger mode
}

// TUIIO implements GameIO for TUI mode - stores state for the TUI to read
type TUIIO struct {
	// Game state for TUI to display
	gameStatus     GameStatus
	cards          []Card
	displayMap     []int
	statusMessage  string
	shopItems      []ShopItem
	shopMoney      int
	shopRerollCost int

	// Action channels for communication between game logic and TUI
	actionChan  chan PlayerAction
	shopChan    chan string
	messageChan chan string
}

// GameStatus holds the current game state for display
type GameStatus struct {
	Ante     int
	Blind    BlindType
	Target   int
	Score    int
	Hands    int
	Discards int
	Money    int
	Jokers   []Joker
}

// PlayerAction represents an action taken by the player
type PlayerAction struct {
	Action string
	Params []string
	Quit   bool
}

// NewTUIIO creates a new TUIIO instance
func NewTUIIO() *TUIIO {
	return &TUIIO{
		actionChan:  make(chan PlayerAction, 1),
		shopChan:    make(chan string, 1),
		messageChan: make(chan string, 10),
	}
}

// Methods for TUI to call to get current state
func (io *TUIIO) GetGameStatus() GameStatus {
	return io.gameStatus
}

func (io *TUIIO) GetCards() ([]Card, []int) {
	return io.cards, io.displayMap
}

func (io *TUIIO) GetStatusMessage() string {
	select {
	case msg := <-io.messageChan:
		return msg
	default:
		return ""
	}
}

func (io *TUIIO) GetShopInfo() ([]ShopItem, int, int) {
	return io.shopItems, io.shopMoney, io.shopRerollCost
}

// Methods for TUI to send actions back to game
func (io *TUIIO) SendPlayerAction(action string, params []string, quit bool) {
	select {
	case io.actionChan <- PlayerAction{Action: action, Params: params, Quit: quit}:
	default:
		// Channel full, ignore
	}
}

func (io *TUIIO) SendShopAction(action string) {
	select {
	case io.shopChan <- action:
	default:
		// Channel full, ignore
	}
}

// GameIO interface implementation
func (io *TUIIO) ShowWelcome() {
	// TUI handles welcome screen in its view
}

func (io *TUIIO) ShowGameStatus(ante int, blind BlindType, target int, score int, hands int, discards int, money int, jokers []Joker) {
	io.gameStatus = GameStatus{
		Ante:     ante,
		Blind:    blind,
		Target:   target,
		Score:    score,
		Hands:    hands,
		Discards: discards,
		Money:    money,
		Jokers:   jokers,
	}
}

func (io *TUIIO) ShowCards(cards []Card, displayToOriginal []int) {
	io.cards = make([]Card, len(cards))
	copy(io.cards, cards)
	io.displayMap = make([]int, len(displayToOriginal))
	copy(io.displayMap, displayToOriginal)
}

func (io *TUIIO) ShowMessage(message string) {
	select {
	case io.messageChan <- message:
	default:
		// Channel full, ignore oldest message
		select {
		case <-io.messageChan:
		default:
		}
		select {
		case io.messageChan <- message:
		default:
		}
	}
}

func (io *TUIIO) GetPlayerAction(canDiscard bool) (string, []string, bool) {
	// Block waiting for action from TUI
	action := <-io.actionChan
	return action.Action, action.Params, action.Quit
}

func (io *TUIIO) ShowShop(items []ShopItem, money int, rerollCost int) {
	io.shopItems = make([]ShopItem, len(items))
	copy(io.shopItems, items)
	io.shopMoney = money
	io.shopRerollCost = rerollCost
}

func (io *TUIIO) GetShopAction() (string, bool) {
	action := <-io.shopChan
	if action == "quit" {
		return "exit", true
	}
	return action, false
}

func (io *TUIIO) ShowGameResults(finalScore int, target int, ante int) {
	io.ShowMessage(fmt.Sprintf("💀 GAME OVER 💀\nFinal Score: %d/%d (Ante %d)\nBetter luck next time!", finalScore, target, ante))
}

func (io *TUIIO) ShowVictoryResults() {
	io.ShowMessage("🎉 VICTORY! 🎉\nCongratulations! You've conquered all 8 Antes!\nYou are a true Balatro master!")
}

func (io *TUIIO) Close() {
	close(io.actionChan)
	close(io.shopChan)
	close(io.messageChan)
}

// ShopItem represents an item in the shop
type ShopItem struct {
	Name        string
	Description string
	Cost        int
	Type        string // "joker", "consumable", etc.
}
