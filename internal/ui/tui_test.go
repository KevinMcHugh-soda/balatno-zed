package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	game "balatno/internal/game"
)

// TestToggleCardSelection verifies selecting and deselecting cards.
func TestToggleCardSelection(t *testing.T) {
	m := TUIModel{cards: []game.Card{{Rank: game.Ace, Suit: game.Spades}}}

	if m.isCardSelected(0) {
		t.Fatalf("card should not be selected initially")
	}

	m.toggleCardSelection(0)
	if !m.isCardSelected(0) {
		t.Fatalf("card should be selected after toggle")
	}

	m.toggleCardSelection(0)
	if m.isCardSelected(0) {
		t.Fatalf("card should be deselected after second toggle")
	}
}

// TestResortPreservesSelection ensures selected cards stay selected after resorting.
func TestResortPreservesSelection(t *testing.T) {
	m := TUIModel{
		cards: []game.Card{
			{Rank: game.Two, Suit: game.Hearts},
			{Rank: game.Three, Suit: game.Spades},
			{Rank: game.Four, Suit: game.Diamonds},
		},
	}

	m.toggleCardSelection(0)
	if len(m.selectedCards) != 1 || m.selectedCards[0] != 0 {
		t.Fatalf("expected card 0 to be selected")
	}

	m.handleResort()

	resorted := []game.Card{m.cards[1], m.cards[2], m.cards[0]}
	event := game.CardsDealtEvent{
		Cards:          resorted,
		DisplayMapping: []int{1, 2, 0},
		SortMode:       "rank",
	}
	model, _ := m.Update(cardsDealtMsg(event))
	m = model.(TUIModel)

	if len(m.selectedCards) != 1 || m.selectedCards[0] != 2 {
		t.Fatalf("expected selection to move to index 2, got %v", m.selectedCards)
	}
}

// TestHelpToggle ensures that pressing 'h' toggles help mode on and off.
func TestHelpToggle(t *testing.T) {
	m := TUIModel{mode: GameMode{}}

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = model.(TUIModel)

	if !m.showHelp {
		t.Fatalf("expected help to be shown")
	}
	if _, ok := m.mode.(*GameHelpMode); !ok {
		t.Fatalf("expected mode to be GameHelpMode")
	}

	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = model.(TUIModel)

	if m.showHelp {
		t.Fatalf("expected help to be hidden")
	}
	if _, ok := m.mode.(*GameMode); !ok {
		t.Fatalf("expected mode to be GameMode after toggling back")
	}
}

// TestShoppingModeActions verifies purchase and reroll flows in the shop.
func TestShoppingModeActions(t *testing.T) {
	// Prepare a model with one affordable item and a pending request
	respChan := make(chan PlayerActionResponse, 1)
	selected := 1
	m := TUIModel{
		gameState:            game.GameStateChangedEvent{Money: 10},
		shopInfo:             &game.ShopOpenedEvent{Money: 10, RerollCost: 5, Items: []game.ShopItemData{{Name: "J1", Cost: 5, Description: "", CanAfford: true}}},
		mode:                 &ShoppingMode{selectedItem: &selected},
		actionRequestPending: &PlayerActionRequest{ResponseChan: respChan},
	}

	// Purchasing selected item
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	resp := <-respChan
	if resp.Action != game.PlayerActionBuy || len(resp.Params) != 1 || resp.Params[0] != "1" {
		t.Fatalf("unexpected purchase response: %+v", resp)
	}

	// Reroll action
	respChan = make(chan PlayerActionResponse, 1)
	mPtr := model.(*TUIModel)
	m = *mPtr
	m.actionRequestPending = &PlayerActionRequest{ResponseChan: respChan}
	m.mode = &ShoppingMode{}
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	resp = <-respChan
	if resp.Action != game.PlayerActionReroll {
		t.Fatalf("unexpected reroll response: %+v", resp)
	}

	// Exit confirmation: first enter prompts, second exits
	respChan = make(chan PlayerActionResponse, 1)
	mPtr = model.(*TUIModel)
	m = *mPtr
	m.actionRequestPending = &PlayerActionRequest{ResponseChan: respChan}
	m.mode = &ShoppingMode{}

	// First enter should ask for confirmation
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	select {
	case resp := <-respChan:
		t.Fatalf("unexpected response after first enter: %+v", resp)
	default:
	}

	// Second enter should exit the shop
	m = *(model.(*TUIModel))
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	resp = <-respChan
	if resp.Action != game.PlayerActionExitShop {
		t.Fatalf("unexpected exit response: %+v", resp)
	}
}

// TestShoppingModeEmptySlotSelection ensures selecting an empty slot is handled gracefully.
func TestShoppingModeEmptySlotSelection(t *testing.T) {
	m := TUIModel{
		gameState: game.GameStateChangedEvent{Money: 10},
		shopInfo:  &game.ShopOpenedEvent{Money: 10, RerollCost: 5, Items: []game.ShopItemData{{}, {Name: "J2", Cost: 5, Description: "", CanAfford: true}}},
		mode:      &ShoppingMode{},
	}

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = *(model.(*TUIModel))
	sm := m.mode.(*ShoppingMode)
	if sm.selectedItem != nil {
		t.Fatalf("expected no selection when choosing empty slot")
	}
	if m.statusMessage != "That slot is empty!" {
		t.Fatalf("unexpected status message: %s", m.statusMessage)
	}
}

// TestShoppingModeSecondItemSelection verifies purchasing the second item when the first is empty.
func TestShoppingModeSecondItemSelection(t *testing.T) {
	respChan := make(chan PlayerActionResponse, 1)
	m := TUIModel{
		gameState:            game.GameStateChangedEvent{Money: 10},
		shopInfo:             &game.ShopOpenedEvent{Money: 10, RerollCost: 5, Items: []game.ShopItemData{{}, {Name: "J2", Cost: 5, Description: "", CanAfford: true}}},
		mode:                 &ShoppingMode{},
		actionRequestPending: &PlayerActionRequest{ResponseChan: respChan},
	}

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = *(model.(*TUIModel))
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	resp := <-respChan
	if resp.Action != game.PlayerActionBuy || len(resp.Params) != 1 || resp.Params[0] != "2" {
		t.Fatalf("unexpected purchase response: %+v", resp)
	}
}

// TestJokerReorder ensures joker reordering sends the correct action and updates state.
func TestJokerReorder(t *testing.T) {
	respChan := make(chan PlayerActionResponse, 1)
	m := TUIModel{
		gameState: game.GameStateChangedEvent{
			Jokers: []game.Joker{{Name: "J1"}, {Name: "J2"}},
		},
		mode:                 GameMode{},
		actionRequestPending: &PlayerActionRequest{ResponseChan: respChan},
	}

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = *(model.(*TUIModel))
	if _, ok := m.mode.(*JokerOrderMode); !ok {
		t.Fatalf("expected mode to be JokerOrderMode")
	}

	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = *(model.(*TUIModel))

	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyUp})
	resp := <-respChan
	if resp.Action != game.PlayerActionMoveJoker || len(resp.Params) != 2 || resp.Params[0] != "2" || resp.Params[1] != "up" {
		t.Fatalf("unexpected response: %+v", resp)
	}

	if m.gameState.Jokers[0].Name != "J2" {
		t.Fatalf("expected J2 to move up, got %v", m.gameState.Jokers)
	}
}

// TestJokerSell ensures selling a joker sends the correct action and removes it from state.
func TestJokerSell(t *testing.T) {
	respChan := make(chan PlayerActionResponse, 1)
	m := TUIModel{
		gameState: game.GameStateChangedEvent{
			Money:  10,
			Jokers: []game.Joker{{Name: "J1", Price: 6}, {Name: "J2", Price: 8}},
		},
		mode:                 GameMode{},
		actionRequestPending: &PlayerActionRequest{ResponseChan: respChan},
	}

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = *(model.(*TUIModel))
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = *(model.(*TUIModel))
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = *(model.(*TUIModel))
	resp := <-respChan
	if resp.Action != game.PlayerActionSellJoker || len(resp.Params) != 1 || resp.Params[0] != "1" {
		t.Fatalf("unexpected response: %+v", resp)
	}

	if len(m.gameState.Jokers) != 1 || m.gameState.Jokers[0].Name != "J2" {
		t.Fatalf("expected remaining joker to be J2, got %v", m.gameState.Jokers)
	}
}
