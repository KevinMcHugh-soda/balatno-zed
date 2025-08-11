package game

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoggerEventHandler handles events for console/logger mode
type LoggerEventHandler struct {
	scanner *bufio.Scanner
}

// NewLoggerEventHandler creates a new LoggerEventHandler using stdin for input
func NewLoggerEventHandler() *LoggerEventHandler {
	return NewLoggerEventHandlerFromReader(os.Stdin)
}

// NewLoggerEventHandlerFromReader creates a new LoggerEventHandler that reads
// input from the provided reader. Useful for tests where a custom input source
// is needed.
func NewLoggerEventHandlerFromReader(r io.Reader) *LoggerEventHandler {
	return &LoggerEventHandler{scanner: bufio.NewScanner(r)}
}

// HandleEvent processes game events and presents them to the console
func (h *LoggerEventHandler) HandleEvent(event Event) {
	switch e := event.(type) {
	case GameStartedEvent:
		h.handleGameStarted()
	case GameStateChangedEvent:
		h.handleGameStateChanged(e)
	case CardsDealtEvent:
		h.handleCardsDealt(e)
	case HandPlayedEvent:
		h.handleHandPlayed(e)
	case CardsDiscardedEvent:
		h.handleCardsDiscarded(e)
	case CardsResortedEvent:
		h.handleCardsResorted(e)
	case BlindDefeatedEvent:
		h.handleBlindDefeated(e)
	case AnteCompletedEvent:
		h.handleAnteCompleted(e)
	case NewBlindStartedEvent:
		h.handleNewBlindStarted(e)
	case ShopOpenedEvent:
		h.handleShopOpened(e)
	case ShopItemPurchasedEvent:
		h.handleShopItemPurchased(e)
	case ShopRerolledEvent:
		h.handleShopRerolled(e)
	case ShopClosedEvent:
		h.handleShopClosed()
	case InvalidActionEvent:
		h.handleInvalidAction(e)
	case MessageEvent:
		h.handleMessage(e)
	case GameOverEvent:
		h.handleGameOver(e)
	case VictoryEvent:
		h.handleVictory()
	}
}

func (h *LoggerEventHandler) handleGameStarted() {
	fmt.Println("🃏 Welcome to Balatro CLI! 🃏")
	fmt.Println("🎯 CHALLENGE: Progress through 8 Antes, each with 3 Blinds!")
	fmt.Println("Each Ante: Small Blind → Big Blind → Boss Blind")
	fmt.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	fmt.Println()
}

func (h *LoggerEventHandler) handleGameStateChanged(e GameStateChangedEvent) {
	fmt.Printf("🎯 Ante %d - %s | Target: %d | Current Score: %d\n", e.Ante, e.Blind, e.Target, e.Score)
	fmt.Printf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d\n", e.Hands, e.Discards, e.Money)

	if len(e.Jokers) > 0 {
		fmt.Print("🃏 Jokers: ")
		for i, joker := range e.Jokers {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s (%s)", joker.Name, joker.Description)
		}
		fmt.Println()
	}
	fmt.Println()
}

func (h *LoggerEventHandler) handleCardsDealt(e CardsDealtEvent) {
	fmt.Printf("🃏 Your Hand (%d cards - sorted by %s):\n", len(e.Cards), e.SortMode)

	for i, card := range e.Cards {
		fmt.Printf("%d: %s ", i+1, card.String())
		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func (h *LoggerEventHandler) handleHandPlayed(e HandPlayedEvent) {
	// Build hand string
	var handStr []string
	for _, card := range e.SelectedCards {
		handStr = append(handStr, card.String())
	}

	fmt.Println()
	fmt.Printf("Your hand: %s\n", strings.Join(handStr, " "))
	fmt.Printf("Hand type: %s\n", e.HandType)

	if e.JokerChips > 0 || e.JokerMult > 0 {
		fmt.Printf("Base Score: %d", e.BaseScore)
		if e.JokerChips > 0 {
			fmt.Printf(" + %d Joker Chips", e.JokerChips)
		}
		fmt.Printf(" | Card Values: %d | Mult: %dx", e.CardValues, e.Multiplier)
		if e.JokerMult > 0 {
			fmt.Printf(" + %d Joker Mult", e.JokerMult)
		}
		fmt.Println()
		fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", e.BaseScore+e.JokerChips, e.CardValues, e.Multiplier+e.JokerMult, e.FinalScore)
	} else {
		fmt.Printf("Base Score: %d | Card Values: %d | Mult: %dx\n", e.BaseScore, e.CardValues, e.Multiplier)
		fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", e.BaseScore, e.CardValues, e.Multiplier, e.FinalScore)
	}

	fmt.Printf("💰 Total Score: %d\n", e.NewTotalScore)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()
}

func (h *LoggerEventHandler) handleCardsDiscarded(e CardsDiscardedEvent) {
	var cardNames []string
	for _, card := range e.DiscardedCards {
		cardNames = append(cardNames, card.String())
	}

	discardedStr := strings.Join(cardNames, ", ")
	if len(discardedStr) > 40 {
		discardedStr = fmt.Sprintf("%d cards", e.NumCards)
	}

	fmt.Printf("🗑️ Discarded %s\n", discardedStr)
	if e.DiscardsLeft > 0 {
		fmt.Printf("💫 New cards dealt! %d discards remaining\n", e.DiscardsLeft)
	} else {
		fmt.Println("💫 New cards dealt! No more discards available!")
	}
	fmt.Println()
}

func (h *LoggerEventHandler) handleCardsResorted(e CardsResortedEvent) {
	fmt.Printf("🔄 Cards now sorted by %s\n", e.NewSortMode)
	fmt.Println()
}

func (h *LoggerEventHandler) handleBlindDefeated(e BlindDefeatedEvent) {
	// Different celebrations for different blind types
	switch e.BlindType {
	case SmallBlind:
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🔸 SMALL BLIND DEFEATED! 🔸")
		fmt.Printf("    ✨ Score: %d/%d ✨\n", e.Score, e.Target)
		fmt.Println("   🎯 Advancing to Big Blind...")
		fmt.Println(strings.Repeat("=", 60))
	case BigBlind:
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🔶 BIG BLIND CRUSHED! 🔶")
		fmt.Printf("    ⚡ Score: %d/%d ⚡\n", e.Score, e.Target)
		fmt.Println("   💀 Prepare for the Boss Blind...")
		fmt.Println(strings.Repeat("=", 60))
	case BossBlind:
		fmt.Println(strings.Repeat("🎆", 15))
		fmt.Println("💀 BOSS BLIND ANNIHILATED! 💀")
		fmt.Printf("    🔥 EPIC SCORE: %d/%d 🔥\n", e.Score, e.Target)
		fmt.Println(strings.Repeat("🎆", 15))
	}

	fmt.Println("💰 REWARD BREAKDOWN:")
	fmt.Printf("   Base: $%d", e.BaseReward)
	if e.BonusReward > 0 {
		fmt.Printf(" + Unused: $%d (%d hands + %d discards)", e.BonusReward, e.UnusedHands, e.UnusedDiscards)
	}
	if e.JokerReward > 0 {
		fmt.Printf(" + Jokers: $%d", e.JokerReward)
	}
	fmt.Printf("\n   💰 Total Earned: $%d | Your Money: $%d\n", e.TotalReward, e.NewMoney)
	fmt.Println()
}

func (h *LoggerEventHandler) handleAnteCompleted(e AnteCompletedEvent) {
	fmt.Println("🎊 ANTE COMPLETED! 🎊")
	fmt.Printf("🌟 ANTE PROGRESSION 🌟\n")
	fmt.Printf("   Ante %d → Ante %d\n", e.CompletedAnte, e.NewAnte)
	fmt.Println("🔄 NEW CHALLENGE AWAITS!")
	fmt.Println()
}

func (h *LoggerEventHandler) handleNewBlindStarted(e NewBlindStartedEvent) {
	blindEmoji := ""
	switch e.Blind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	fmt.Printf("%s NOW ENTERING: %s (Ante %d) %s\n", blindEmoji, e.Blind, e.Ante, blindEmoji)
	fmt.Printf("🎯 NEW TARGET: %d points\n", e.Target)
	fmt.Println("🃏 Fresh hand dealt!")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()
}

func (h *LoggerEventHandler) handleShopOpened(e ShopOpenedEvent) {
	fmt.Println("🛍️  Welcome to the Shop!")
	fmt.Printf("💰 You have $%d\n", e.Money)
	fmt.Printf("🎲 Reroll costs $%d\n", e.RerollCost)
	fmt.Println()

	fmt.Println("Available items:")
	for i, item := range e.Items {
		affordText := ""
		if !item.CanAfford {
			affordText = " (can't afford)"
		}
		fmt.Printf("%d. %s - $%d%s\n", i+1, item.Name, item.Cost, affordText)
		fmt.Printf("   %s\n", item.Description)
		fmt.Println()
	}

	fmt.Println("Commands:")
	fmt.Println("• buy <number> - Purchase an item")
	fmt.Println("• reroll - Reroll the shop items")
	fmt.Println("• exit/q - Leave the shop")
}

func (h *LoggerEventHandler) handleShopItemPurchased(e ShopItemPurchasedEvent) {
	fmt.Printf("✨ Purchased %s! ✨\n", e.Item.Name)
	fmt.Printf("💰 Remaining money: $%d\n", e.RemainingMoney)
	fmt.Println()
}

func (h *LoggerEventHandler) handleShopRerolled(e ShopRerolledEvent) {
	fmt.Printf("💫 Rerolled for $%d! Next reroll: $%d\n", e.Cost, e.NewRerollCost)
	fmt.Printf("💰 Remaining money: $%d\n", e.RemainingMoney)
	fmt.Println()

	fmt.Println("New items:")
	for i, item := range e.NewItems {
		affordText := ""
		if !item.CanAfford {
			affordText = " (can't afford)"
		}
		fmt.Printf("%d. %s - $%d%s\n", i+1, item.Name, item.Cost, affordText)
		fmt.Printf("   %s\n", item.Description)
		fmt.Println()
	}
}

func (h *LoggerEventHandler) handleShopClosed() {
	fmt.Println("👋 Left the shop.")
	fmt.Println()
}

func (h *LoggerEventHandler) handleInvalidAction(e InvalidActionEvent) {
	fmt.Printf("❌ %s\n", e.Reason)
}

func (h *LoggerEventHandler) handleMessage(e MessageEvent) {
	switch e.Type {
	case "error":
		fmt.Printf("❌ %s\n", e.Message)
	case "warning":
		fmt.Printf("⚠️  %s\n", e.Message)
	case "success":
		fmt.Printf("✅ %s\n", e.Message)
	case "info":
		fmt.Printf("ℹ️  %s\n", e.Message)
	default:
		fmt.Println(e.Message)
	}
}

func (h *LoggerEventHandler) handleGameOver(e GameOverEvent) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("💀 DEFEAT! You failed to beat the blind.")
	fmt.Printf("Final Score: %d/%d (Ante %d)\n", e.FinalScore, e.Target, e.Ante)
	fmt.Println("Better luck next time!")
	fmt.Println(strings.Repeat("=", 50))
}

func (h *LoggerEventHandler) handleVictory() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🏆 ULTIMATE VICTORY! 🏆")
	fmt.Println("🎉 You have conquered all 8 Antes! 🎉")
	fmt.Println("You are a true Balatro master!")
	fmt.Println(strings.Repeat("=", 60))
}

// GetPlayerAction gets input for player actions
func (h *LoggerEventHandler) GetPlayerAction(canDiscard bool) (PlayerAction, []string, bool) {
	if canDiscard {
		fmt.Print("(p)lay <cards>, (d)iscard <cards>, (r)esort, or (q)uit: ")
	} else {
		fmt.Print("(p)lay <cards>, (r)esort, or (q)uit: ")
	}

	if !h.scanner.Scan() {
		if err := h.scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
		return PlayerActionNone, nil, true
	}

	input := strings.TrimSpace(h.scanner.Text())

	if strings.ToLower(input) == "quit" || strings.ToLower(input) == "q" {
		return PlayerActionNone, nil, true
	}

	if input == "" {
		fmt.Println("Please enter an action")
		return PlayerActionNone, nil, false
	}

	parts := strings.Fields(input)
	if len(parts) < 1 {
		fmt.Println("Please enter 'play <cards>' or 'discard <cards>'")
		return PlayerActionNone, nil, false
	}

	actionChar := strings.ToLower(parts[0])
	var selectedAction PlayerAction
	// Support abbreviated commands
	if actionChar == "p" {
		selectedAction = PlayerActionPlay
	} else if actionChar == "d" {
		selectedAction = PlayerActionDiscard
	} else if actionChar == "r" {
		selectedAction = PlayerActionResort
	} else if actionChar == "q" {
		return PlayerActionNone, nil, true
	}

	var params []string
	if len(parts) > 1 {
		params = parts[1:]
	}

	return selectedAction, params, false
}

// GetShopAction gets input for shop actions
func (h *LoggerEventHandler) GetShopAction() (PlayerAction, []string, bool) {
	fmt.Print("Shop action (buy <number>, reroll, exit/q): ")

	if !h.scanner.Scan() {
		if err := h.scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
		return PlayerActionNone, nil, true
	}

	input := strings.TrimSpace(h.scanner.Text())
	if strings.ToLower(input) == "exit" || strings.ToLower(input) == "q" || input == "" {
		return PlayerActionExitShop, nil, false
	}

	parts := strings.Fields(input)
	var params []string
	if len(parts) > 1 {
		params = parts[1:]
	}

	var action PlayerAction
	switch strings.ToLower(parts[0]) {
	case "b", "buy":
		action = PlayerActionBuy
	case "r", "reroll":
		action = PlayerActionReroll
	default:
		fmt.Println("No action recognized", input)
		action = PlayerActionNone
	}

	return action, params, false
}

// Close cleans up resources
func (h *LoggerEventHandler) Close() {
	// Nothing to clean up for console mode
}
