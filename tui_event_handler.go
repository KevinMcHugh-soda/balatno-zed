package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TUIEventHandler handles events for TUI mode and coordinates with the TUI
type TUIEventHandler struct {
	// Game state for TUI to display
	gameState     GameStateChangedEvent
	cards         []Card
	displayMap    []int
	sortMode      string
	statusMessage string
	shopInfo      *ShopOpenedEvent

	// Communication channels
	actionChan chan PlayerActionRequest
	shopChan   chan string

	// Synchronization
	mutex sync.RWMutex

	// TUI integration
	tuiModel *TUIModel
}

// PlayerActionRequest represents a request for player input
type PlayerActionRequest struct {
	CanDiscard   bool
	ResponseChan chan PlayerActionResponse
}

// PlayerActionResponse represents the player's response
type PlayerActionResponse struct {
	Action string
	Params []string
	Quit   bool
}

// NewTUIEventHandler creates a new TUIEventHandler
func NewTUIEventHandler() *TUIEventHandler {
	return &TUIEventHandler{
		actionChan: make(chan PlayerActionRequest, 1),
		shopChan:   make(chan string, 1),
	}
}

// SetTUIModel links the event handler to the TUI model
func (h *TUIEventHandler) SetTUIModel(model *TUIModel) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.tuiModel = model
}

// HandleEvent processes game events and updates TUI state
func (h *TUIEventHandler) HandleEvent(event Event) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	switch e := event.(type) {
	case GameStartedEvent:
		h.statusMessage = "🎮 Game started! Select cards with 1-7, play with Enter/P, discard with D"

	case GameStateChangedEvent:
		h.gameState = e

	case CardsDealtEvent:
		h.cards = make([]Card, len(e.Cards))
		copy(h.cards, e.Cards)
		h.displayMap = make([]int, len(e.DisplayMapping))
		copy(h.displayMap, e.DisplayMapping)
		h.sortMode = e.SortMode

	case HandPlayedEvent:
		var msg string
		scoreGained := e.FinalScore

		// Check if this completed the blind
		if e.NewTotalScore >= h.gameState.Target {
			msg = fmt.Sprintf("🎉 %s for +%d points! BLIND DEFEATED!", e.HandType, scoreGained)
		} else {
			handsLeft := h.gameState.Hands - 1
			if handsLeft <= 0 {
				msg = fmt.Sprintf("💀 %s for +%d points, but Game Over! Final: %d/%d", e.HandType, scoreGained, e.NewTotalScore, h.gameState.Target)
			} else {
				progressPercent := float64(e.NewTotalScore) / float64(h.gameState.Target) * 100
				msg = fmt.Sprintf("✅ %s for +%d points! %d/%d (%.0f%%) | %d hands left", e.HandType, scoreGained, e.NewTotalScore, h.gameState.Target, progressPercent, handsLeft)
			}
		}
		h.statusMessage = msg

	case CardsDiscardedEvent:
		var cardNames []string
		for _, card := range e.DiscardedCards {
			cardNames = append(cardNames, card.String())
		}

		discardedStr := strings.Join(cardNames, ", ")
		if len(discardedStr) > 20 {
			discardedStr = fmt.Sprintf("%d cards", e.NumCards)
		}

		if e.DiscardsLeft > 0 {
			h.statusMessage = fmt.Sprintf("🗑️ Discarded %s, dealt new cards | %d discards remaining", discardedStr, e.DiscardsLeft)
		} else {
			h.statusMessage = fmt.Sprintf("🗑️ Discarded %s, dealt new cards | No more discards available!", discardedStr)
		}

	case CardsResortedEvent:
		h.statusMessage = fmt.Sprintf("🔄 Cards now sorted by %s", e.NewSortMode)

	case BlindDefeatedEvent:
		var msg string
		switch e.BlindType {
		case SmallBlind:
			msg = "🔸 SMALL BLIND DEFEATED! Advancing to Big Blind..."
		case BigBlind:
			msg = "🔶 BIG BLIND CRUSHED! Prepare for the Boss Blind..."
		case BossBlind:
			msg = "💀 BOSS BLIND ANNIHILATED! 💀"
		}
		h.statusMessage = msg

	case AnteCompletedEvent:
		h.statusMessage = fmt.Sprintf("🎊 ANTE %d COMPLETE! Starting Ante %d", e.CompletedAnte, e.NewAnte)

	case NewBlindStartedEvent:
		blindEmoji := ""
		switch e.Blind {
		case SmallBlind:
			blindEmoji = "🔸"
		case BigBlind:
			blindEmoji = "🔶"
		case BossBlind:
			blindEmoji = "💀"
		}
		h.statusMessage = fmt.Sprintf("%s NOW ENTERING: %s (Ante %d) | Target: %d points", blindEmoji, e.Blind, e.Ante, e.Target)

	case ShopOpenedEvent:
		shopCopy := e
		h.shopInfo = &shopCopy
		h.statusMessage = "🛍️ Welcome to the Shop! (Shop interface not yet implemented in TUI)"

	case ShopItemPurchasedEvent:
		h.statusMessage = fmt.Sprintf("✨ Purchased %s! Remaining: $%d", e.Item.Name, e.RemainingMoney)

	case ShopRerolledEvent:
		h.statusMessage = fmt.Sprintf("💫 Shop rerolled for $%d! Next reroll: $%d", e.Cost, e.NewRerollCost)

	case ShopClosedEvent:
		h.shopInfo = nil
		h.statusMessage = "👋 Left the shop"

	case InvalidActionEvent:
		h.statusMessage = fmt.Sprintf("❌ %s", e.Reason)

	case MessageEvent:
		switch e.Type {
		case "error":
			h.statusMessage = fmt.Sprintf("❌ %s", e.Message)
		case "warning":
			h.statusMessage = fmt.Sprintf("⚠️ %s", e.Message)
		case "success":
			h.statusMessage = fmt.Sprintf("✅ %s", e.Message)
		case "info":
			h.statusMessage = fmt.Sprintf("ℹ️ %s", e.Message)
		default:
			h.statusMessage = e.Message
		}

	case GameOverEvent:
		h.statusMessage = fmt.Sprintf("💀 GAME OVER! Final: %d/%d (Ante %d)", e.FinalScore, e.Target, e.Ante)

	case VictoryEvent:
		h.statusMessage = "🏆 VICTORY! You conquered all 8 Antes! 🎉"
	}
}

// GetPlayerAction waits for player input from the TUI
func (h *TUIEventHandler) GetPlayerAction(canDiscard bool) (string, []string, bool) {
	responseChan := make(chan PlayerActionResponse)
	request := PlayerActionRequest{
		CanDiscard:   canDiscard,
		ResponseChan: responseChan,
	}

	// Send request to TUI
	select {
	case h.actionChan <- request:
	case <-time.After(100 * time.Millisecond):
		// If the channel is full, return empty action to keep game loop responsive
		return "", nil, false
	}

	// Wait for response
	response := <-responseChan
	return response.Action, response.Params, response.Quit
}

// GetShopAction waits for shop action from the TUI
func (h *TUIEventHandler) GetShopAction() (string, bool) {
	// For now, automatically exit shop in TUI mode since shop UI isn't implemented
	return "exit", false
}

// Close cleans up resources
func (h *TUIEventHandler) Close() {
	close(h.actionChan)
	close(h.shopChan)
}

// Methods for TUI to access current state
func (h *TUIEventHandler) GetGameState() GameStateChangedEvent {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.gameState
}

func (h *TUIEventHandler) GetCards() ([]Card, []int) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.cards, h.displayMap
}

func (h *TUIEventHandler) GetStatusMessage() string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.statusMessage
}

func (h *TUIEventHandler) ClearStatusMessage() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.statusMessage = ""
}

func (h *TUIEventHandler) GetShopInfo() *ShopOpenedEvent {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.shopInfo
}

// Methods for TUI to send actions back to the game
func (h *TUIEventHandler) SendPlayerAction(action string, params []string, quit bool) {
	// Check if there's a pending request
	select {
	case request := <-h.actionChan:
		// Send response
		response := PlayerActionResponse{
			Action: action,
			Params: params,
			Quit:   quit,
		}
		select {
		case request.ResponseChan <- response:
		case <-time.After(100 * time.Millisecond):
			// Response channel full, ignore
		}
	default:
		// No pending request, store the action for later or ignore
	}
}

// Helper method to check if there's a pending action request
func (h *TUIEventHandler) HasPendingActionRequest() bool {
	select {
	case <-h.actionChan:
		// Put it back
		return true
	default:
		return false
	}
}

// Helper method for TUI to handle play action
func (h *TUIEventHandler) HandlePlayAction(selectedCards []int) {
	if len(selectedCards) == 0 {
		return
	}

	// Convert selected indices to string params for game logic
	var params []string
	for _, index := range selectedCards {
		// Convert 0-based TUI index to 1-based display index for game logic
		params = append(params, fmt.Sprintf("%d", index+1))
	}

	h.SendPlayerAction("play", params, false)
}

// Helper method for TUI to handle discard action
func (h *TUIEventHandler) HandleDiscardAction(selectedCards []int) {
	if len(selectedCards) == 0 {
		return
	}

	// Convert selected indices to string params for game logic
	var params []string
	for _, index := range selectedCards {
		// Convert 0-based TUI index to 1-based display index for game logic
		params = append(params, fmt.Sprintf("%d", index+1))
	}

	h.SendPlayerAction("discard", params, false)
}

// Helper method for TUI to handle resort action
func (h *TUIEventHandler) HandleResortAction() {
	h.SendPlayerAction("resort", nil, false)
}

// Helper method for TUI to handle quit action
func (h *TUIEventHandler) HandleQuitAction() {
	h.SendPlayerAction("", nil, true)
}
