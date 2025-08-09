package main

import (
	"time"
)

// TUIEventHandler handles events for TUI mode and sends messages to bubbletea
type TUIEventHandler struct {
	// Communication channels
	actionChan chan PlayerActionRequest
	shopChan   chan string

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
	Action PlayerAction
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
	h.tuiModel = model
}

// HandleEvent processes game events and sends them as bubbletea messages
func (h *TUIEventHandler) HandleEvent(event Event) {
	if h.tuiModel == nil {
		return
	}

	switch e := event.(type) {
	case GameStartedEvent:
		h.tuiModel.SendMessage(gameStartedMsg{})

	case GameStateChangedEvent:
		h.tuiModel.SendMessage(gameStateChangedMsg(e))

	case CardsDealtEvent:
		h.tuiModel.SendMessage(cardsDealtMsg(e))

	case HandPlayedEvent:
		h.tuiModel.SendMessage(handPlayedMsg(e))

	case CardsDiscardedEvent:
		h.tuiModel.SendMessage(cardsDiscardedMsg(e))

	case CardsResortedEvent:
		h.tuiModel.SendMessage(cardsResortedMsg(e))

	case BlindDefeatedEvent:
		h.tuiModel.SendMessage(blindDefeatedMsg(e))

	case AnteCompletedEvent:
		h.tuiModel.SendMessage(anteCompletedMsg(e))

	case NewBlindStartedEvent:
		h.tuiModel.SendMessage(newBlindStartedMsg(e))

	case ShopOpenedEvent:
		h.tuiModel.SendMessage(shopOpenedMsg(e))

	case ShopItemPurchasedEvent:
		h.tuiModel.SendMessage(shopItemPurchasedMsg(e))

	case ShopRerolledEvent:
		h.tuiModel.SendMessage(shopRerolledMsg(e))

	case ShopClosedEvent:
		h.tuiModel.SendMessage(shopClosedMsg{})

	case InvalidActionEvent:
		h.tuiModel.SendMessage(invalidActionMsg(e))

	case MessageEvent:
		h.tuiModel.SendMessage(messageEventMsg(e))

	case GameOverEvent:
		h.tuiModel.SendMessage(gameOverMsg(e))

	case VictoryEvent:
		h.tuiModel.SendMessage(victoryMsg{})
	}
}

// GetPlayerAction waits for player input from the TUI
func (h *TUIEventHandler) GetPlayerAction(canDiscard bool) (PlayerAction, []string, bool) {
	responseChan := make(chan PlayerActionResponse)
	request := PlayerActionRequest{
		CanDiscard:   canDiscard,
		ResponseChan: responseChan,
	}

	// Send action request message to TUI
	if h.tuiModel != nil {
		h.tuiModel.SendMessage(playerActionRequestMsg(request))
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		return response.Action, response.Params, response.Quit
	case <-time.After(30 * time.Second):
		// Return empty action to keep game loop responsive
		return PlayerActionNone, nil, false
	}
}

// GetShopAction waits for shop action from the TUI
func (h *TUIEventHandler) GetShopAction() (PlayerAction, []string, bool) {
	// Basically everything is the same, except that an "r" means reroll, not resort.
	action, params, quit := h.GetPlayerAction(false)
	// we should be able to eliminate this soon, given the move to modes
	if action == PlayerActionResort {
		action = PlayerActionReroll
	}
	return action, params, quit
}

// Close cleans up resources
func (h *TUIEventHandler) Close() {
	close(h.actionChan)
	close(h.shopChan)
}

// Helper method to check if there's a pending action request
func (h *TUIEventHandler) HasPendingActionRequest() bool {
	select {
	case request := <-h.actionChan:
		// Put it back
		select {
		case h.actionChan <- request:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
