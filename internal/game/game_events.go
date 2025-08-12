package game

// Event represents something that happened in the game
type Event interface {
	EventType() string
	// TODO: add a ToMsg() method. Or maybe not - that's specific to the tui
	// ToMsg() Msg
}

type PlayerAction string

const (
	PlayerActionNone      = "none"
	PlayerActionDiscard   = "discard"
	PlayerActionPlay      = "play"
	PlayerActionQuit      = "quit"
	PlayerActionResort    = "resort"
	PlayerActionExitShop  = "exit_shop"
	PlayerActionReroll    = "reroll"
	PlayerActionBuy       = "buy"
	PlayerActionMoveJoker = "move_joker"
)

// EventHandler processes game events and decides how to present them
type EventHandler interface {
	HandleEvent(event Event)
	// Input requests return through channels or direct calls
	GetPlayerAction(canDiscard bool) (action PlayerAction, params []string, quit bool)
	GetShopAction() (action PlayerAction, params []string, quit bool)
	Close()
}

// Game lifecycle events
type GameStartedEvent struct{}

func (e GameStartedEvent) EventType() string { return "game_started" }

type GameOverEvent struct {
	FinalScore int
	Target     int
	Ante       int
}

func (e GameOverEvent) EventType() string { return "game_over" }

type VictoryEvent struct{}

func (e VictoryEvent) EventType() string { return "victory" }

// Game state events
type GameStateChangedEvent struct {
	Ante     int
	Blind    BlindType
	Target   int
	Score    int
	Hands    int
	Discards int
	Money    int
	Jokers   []Joker
}

func (e GameStateChangedEvent) EventType() string { return "game_state_changed" }

type CardsDealtEvent struct {
	Cards          []Card
	DisplayMapping []int
	SortMode       string
}

func (e CardsDealtEvent) EventType() string { return "cards_dealt" }

// Player action events
type HandPlayedEvent struct {
	SelectedCards []Card
	HandType      string
	BaseScore     int
	CardValues    int
	Multiplier    int
	JokerChips    int
	JokerMult     int
	FinalScore    int
	NewTotalScore int
}

func (e HandPlayedEvent) EventType() string { return "hand_played" }

type CardsDiscardedEvent struct {
	DiscardedCards []Card
	NumCards       int
	DiscardsLeft   int
}

func (e CardsDiscardedEvent) EventType() string { return "cards_discarded" }

type CardsResortedEvent struct {
	NewSortMode string
}

func (e CardsResortedEvent) EventType() string { return "cards_resorted" }

// Blind progression events
type BlindDefeatedEvent struct {
	BlindType      BlindType
	Score          int
	Target         int
	BaseReward     int
	BonusReward    int
	JokerReward    int
	TotalReward    int
	NewMoney       int
	UnusedHands    int
	UnusedDiscards int
}

func (e BlindDefeatedEvent) EventType() string { return "blind_defeated" }

type AnteCompletedEvent struct {
	CompletedAnte int
	NewAnte       int
}

func (e AnteCompletedEvent) EventType() string { return "ante_completed" }

type NewBlindStartedEvent struct {
	Ante     int
	Blind    BlindType
	Target   int
	NewCards []Card
}

func (e NewBlindStartedEvent) EventType() string { return "new_blind_started" }

// Shop events
type ShopOpenedEvent struct {
	Money      int
	RerollCost int
	Items      []ShopItemData
}

func (e ShopOpenedEvent) EventType() string { return "shop_opened" }

type ShopItemPurchasedEvent struct {
	Item           ShopItemData
	RemainingMoney int
}

func (e ShopItemPurchasedEvent) EventType() string { return "shop_item_purchased" }

type ShopRerolledEvent struct {
	Cost           int
	NewRerollCost  int
	RemainingMoney int
	NewItems       []ShopItemData
}

func (e ShopRerolledEvent) EventType() string { return "shop_rerolled" }

type ShopClosedEvent struct{}

func (e ShopClosedEvent) EventType() string { return "shop_closed" }

// Error/validation events
type InvalidActionEvent struct {
	Action string
	Reason string
}

func (e InvalidActionEvent) EventType() string { return "invalid_action" }

type MessageEvent struct {
	Message string
	Type    string // "info", "warning", "error", "success"
}

func (e MessageEvent) EventType() string { return "message" }

// Data structures for events
type ShopItemData struct {
	Name        string
	Description string
	Cost        int
	Type        string
	CanAfford   bool
}

// Helper function to create shop item data from joker
func NewShopItemData(joker Joker, money int) ShopItemData {
	return ShopItemData{
		Name:        joker.Name,
		Description: joker.Description,
		Cost:        joker.Price,
		Type:        "joker",
		CanAfford:   money >= joker.Price,
	}
}

// Event bus interface for the game to emit events
type EventEmitter interface {
	EmitEvent(event Event)
	SetEventHandler(handler EventHandler)
}

// Simple event emitter implementation
type SimpleEventEmitter struct {
	handler EventHandler
}

func NewEventEmitter() *SimpleEventEmitter {
	return &SimpleEventEmitter{}
}

func (e *SimpleEventEmitter) SetEventHandler(handler EventHandler) {
	e.handler = handler
}

func (e *SimpleEventEmitter) EmitEvent(event Event) {
	if e.handler != nil {
		e.handler.HandleEvent(event)
	}
}

// Convenience methods for common events
func (e *SimpleEventEmitter) EmitGameStarted() {
	e.EmitEvent(GameStartedEvent{})
}

func (e *SimpleEventEmitter) EmitGameState(g Game) {
	e.EmitEvent(GameStateChangedEvent{
		Ante:     g.currentAnte,
		Blind:    g.currentBlind,
		Target:   g.currentTarget,
		Score:    g.totalScore,
		Hands:    MaxHands,        // TODO: Get the actual value from the game state
		Discards: g.maxDiscards(), // TODO Get the actual value from the game state
		Money:    g.money,
		Jokers:   g.jokers,
	})
}

func (e *SimpleEventEmitter) EmitCardsDealt(cards []Card, displayMapping []int, sortMode SortMode) {
	sortModeStr := "rank"
	if sortMode == SortBySuit {
		sortModeStr = "suit"
	}
	e.EmitEvent(CardsDealtEvent{
		Cards:          cards,
		DisplayMapping: displayMapping,
		SortMode:       sortModeStr,
	})
}

func (e *SimpleEventEmitter) EmitMessage(message, msgType string) {
	e.EmitEvent(MessageEvent{
		Message: message,
		Type:    msgType,
	})
}

func (e *SimpleEventEmitter) EmitError(message string) {
	e.EmitMessage(message, "error")
}

func (e *SimpleEventEmitter) EmitInfo(message string) {
	e.EmitMessage(message, "info")
}

func (e *SimpleEventEmitter) EmitSuccess(message string) {
	e.EmitMessage(message, "success")
}

func (e *SimpleEventEmitter) EmitWarning(message string) {
	e.EmitMessage(message, "warning")
}
