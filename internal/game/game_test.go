package game

import "testing"

// testEventHandler is a minimal EventHandler implementation for testing.
type testEventHandler struct {
	events      []Event
	shopActions []struct {
		action PlayerAction
		params []string
	}
	actionIdx int
}

func (t *testEventHandler) HandleEvent(e Event) {
	t.events = append(t.events, e)
}

func (t *testEventHandler) GetPlayerAction(bool) (PlayerAction, []string, bool) {
	return PlayerActionNone, nil, false
}

func (t *testEventHandler) GetShopAction() (PlayerAction, []string, bool) {
	if t.actionIdx >= len(t.shopActions) {
		return PlayerActionExitShop, nil, false
	}
	a := t.shopActions[t.actionIdx]
	t.actionIdx++
	return a.action, a.params, false
}

func (t *testEventHandler) Close() {}

// TestHandlePlayAction verifies that playing cards updates score, hand count
// and emits a HandPlayedEvent.
func TestHandlePlayAction(t *testing.T) {
	handler := &testEventHandler{}
	deck := NewDeck()
	g := &Game{
		deck:         deck,
		deckIndex:    7,
		playerCards:  []Card{{Rank: Ten, Suit: Hearts}, {Rank: Ten, Suit: Clubs}, {Rank: Three, Suit: Diamonds}, {Rank: Four, Suit: Spades}, {Rank: Five, Suit: Hearts}, {Rank: Six, Suit: Clubs}, {Rank: Seven, Suit: Diamonds}},
		eventEmitter: NewEventEmitter(),
	}
	g.eventEmitter.SetEventHandler(handler)

	g.handlePlayAction([]string{"1", "2"})

	if g.handsPlayed != 1 {
		t.Fatalf("expected hands played to be 1, got %d", g.handsPlayed)
	}
	if g.totalScore <= 0 {
		t.Fatalf("expected positive score, got %d", g.totalScore)
	}
	if len(g.playerCards) != 7 {
		t.Fatalf("expected 7 cards after playing, got %d", len(g.playerCards))
	}

	found := false
	for _, e := range handler.events {
		if _, ok := e.(HandPlayedEvent); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected HandPlayedEvent to be emitted")
	}
}

// TestHandleDiscardAction verifies discarding cards updates discard count and
// emits a CardsDiscardedEvent.
func TestHandleDiscardAction(t *testing.T) {
	handler := &testEventHandler{}
	deck := NewDeck()
	g := &Game{
		deck:         deck,
		deckIndex:    7,
		playerCards:  deck[:7],
		eventEmitter: NewEventEmitter(),
	}
	g.eventEmitter.SetEventHandler(handler)

	g.handleDiscardAction([]string{"1", "2"})

	if g.discardsUsed != 1 {
		t.Fatalf("expected 1 discard used, got %d", g.discardsUsed)
	}
	if len(g.playerCards) != 7 {
		t.Fatalf("expected 7 cards after discard, got %d", len(g.playerCards))
	}
	found := false
	for _, e := range handler.events {
		if _, ok := e.(CardsDiscardedEvent); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected CardsDiscardedEvent to be emitted")
	}
}

// TestShowShopWithItems ensures that purchasing an item deducts money, adds the
// joker and emits the appropriate events.
func TestShowShopWithItems(t *testing.T) {
	handler := &testEventHandler{
		shopActions: []struct {
			action PlayerAction
			params []string
		}{
			{PlayerActionBuy, []string{"1"}},
			{PlayerActionExitShop, nil},
		},
	}

	g := &Game{
		money:        10,
		rerollCost:   5,
		jokers:       []Joker{},
		eventEmitter: NewEventEmitter(),
	}
	g.eventEmitter.SetEventHandler(handler)

	available := []Joker{{Name: "J1", Price: 5, Description: ""}, {Name: "J2", Price: 6, Description: ""}}
	shop := []Joker{available[0], available[1]}

	g.showShopWithItems(available, shop)

	if g.money != 5 {
		t.Fatalf("expected money to be 5 after purchase, got %d", g.money)
	}
	if len(g.jokers) != 1 || g.jokers[0].Name != "J1" {
		t.Fatalf("expected to own J1 after purchase")
	}
	opened, purchased, closed := false, false, false
	for _, e := range handler.events {
		switch e.(type) {
		case ShopOpenedEvent:
			opened = true
		case ShopItemPurchasedEvent:
			purchased = true
		case ShopClosedEvent:
			closed = true
		}
	}
	if !opened {
		t.Fatalf("expected ShopOpenedEvent to be emitted")
	}
	if !purchased {
		t.Fatalf("expected ShopItemPurchasedEvent to be emitted")
	}
	if !closed {
		t.Fatalf("expected ShopClosedEvent to be emitted on exit")
	}
}

// TestHandSizeWithJoker verifies that a joker can increase the hand size.
func TestHandSizeWithJoker(t *testing.T) {
	g := &Game{
		jokers: []Joker{{Effect: AddHandSize, EffectMagnitude: 2}},
	}
	if got := g.handSize(); got != InitialCards+2 {
		t.Fatalf("expected hand size %d, got %d", InitialCards+2, got)
	}
}

// TestDiscardLimitWithJoker verifies that a joker can increase discard count.
func TestDiscardLimitWithJoker(t *testing.T) {
	handler := &testEventHandler{}
	deck := NewDeck()
	g := &Game{
		deck:         deck,
		deckIndex:    InitialCards,
		playerCards:  deck[:InitialCards],
		jokers:       []Joker{{Effect: AddDiscards, EffectMagnitude: 2}},
		eventEmitter: NewEventEmitter(),
	}
	g.eventEmitter.SetEventHandler(handler)

	for i := 0; i < 5; i++ {
		g.handleDiscardAction([]string{"1"})
	}
	if g.discardsUsed != 5 {
		t.Fatalf("expected 5 discards used, got %d", g.discardsUsed)
	}

	// Exceeding the limit should not increase discardsUsed
	g.handleDiscardAction([]string{"1"})
	if g.discardsUsed != 5 {
		t.Fatalf("discard limit not enforced, got %d", g.discardsUsed)
	}
}
