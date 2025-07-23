package main

import (
	"testing"
)

// Mock scanner for testing
type MockScanner struct {
	inputs []string
	index  int
}

func (m *MockScanner) Scan() bool {
	if m.index < len(m.inputs) {
		return true
	}
	return false
}

func (m *MockScanner) Text() string {
	if m.index < len(m.inputs) {
		text := m.inputs[m.index]
		m.index++
		return text
	}
	return ""
}

// Helper function to create a test game with mock scanner
func createTestGame(inputs []string) *Game {
	game := &Game{
		money:      20, // Start with enough money for testing
		rerollCost: 5,
		jokers:     []Joker{},
		scanner:    nil, // We'll handle input differently in tests
	}
	return game
}

// Test helper to get available jokers for shop
func getTestJokers() []Joker {
	return []Joker{
		{Name: "Test Joker 1", Price: 5, Description: "Test joker 1"},
		{Name: "Test Joker 2", Price: 6, Description: "Test joker 2"},
		{Name: "Test Joker 3", Price: 7, Description: "Test joker 3"},
		{Name: "Test Joker 4", Price: 8, Description: "Test joker 4"},
		{Name: "Test Joker 5", Price: 9, Description: "Test joker 5"},
	}
}

// Helper function to simulate shop item selection
func selectShopItems(availableJokers []Joker, playerJokers []Joker, seed int64) []Joker {
	// Filter jokers player doesn't own
	var candidates []Joker
	for _, joker := range availableJokers {
		if !PlayerHasJoker(playerJokers, joker.Name) {
			candidates = append(candidates, joker)
		}
	}

	if len(candidates) == 0 {
		return []Joker{}
	}

	// Simple deterministic selection for testing (not truly random)
	shopItems := make([]Joker, 0, 2)
	if len(candidates) >= 2 {
		// Use seed to get different selections
		first := int(seed) % len(candidates)
		second := (int(seed) + 1) % len(candidates)
		if second == first && len(candidates) > 1 {
			second = (second + 1) % len(candidates)
		}
		shopItems = append(shopItems, candidates[first])
		if second != first {
			shopItems = append(shopItems, candidates[second])
		}
	} else {
		shopItems = candidates
	}

	return shopItems
}

func TestShopItemSelection(t *testing.T) {
	availableJokers := getTestJokers()
	playerJokers := []Joker{}

	// Test that we get 2 items when available
	shopItems := selectShopItems(availableJokers, playerJokers, 0)

	if len(shopItems) != 2 {
		t.Errorf("Expected 2 shop items, got %d", len(shopItems))
	}

	// Test different seed gives different items
	shopItems2 := selectShopItems(availableJokers, playerJokers, 42)

	if len(shopItems2) != 2 {
		t.Errorf("Expected 2 shop items with different seed, got %d", len(shopItems2))
	}

	// Should get different items with different seed
	if shopItems[0].Name == shopItems2[0].Name && shopItems[1].Name == shopItems2[1].Name {
		t.Errorf("Different seeds should produce different shop items")
	}
}

func TestShopItemSelectionWithOwnedJokers(t *testing.T) {
	availableJokers := getTestJokers()
	playerJokers := []Joker{
		{Name: "Test Joker 1", Price: 5},
		{Name: "Test Joker 2", Price: 6},
	}

	shopItems := selectShopItems(availableJokers, playerJokers, 0)

	// Should exclude owned jokers
	for _, item := range shopItems {
		if PlayerHasJoker(playerJokers, item.Name) {
			t.Errorf("Shop should not offer joker %s that player already owns", item.Name)
		}
	}

	// Should still offer 2 items since we have 5 total and player owns 2
	if len(shopItems) != 2 {
		t.Errorf("Expected 2 shop items after filtering owned jokers, got %d", len(shopItems))
	}
}

func TestRerollCostProgression(t *testing.T) {
	game := createTestGame([]string{})

	// Initial reroll cost should be 5
	if game.rerollCost != 5 {
		t.Errorf("Expected initial reroll cost to be 5, got %d", game.rerollCost)
	}

	// Simulate reroll
	if game.money >= game.rerollCost {
		game.money -= game.rerollCost
		game.rerollCost++
	}

	// After first reroll, cost should be 6
	if game.rerollCost != 6 {
		t.Errorf("Expected reroll cost to be 6 after first reroll, got %d", game.rerollCost)
	}

	// Simulate another reroll
	if game.money >= game.rerollCost {
		game.money -= game.rerollCost
		game.rerollCost++
	}

	// After second reroll, cost should be 7
	if game.rerollCost != 7 {
		t.Errorf("Expected reroll cost to be 7 after second reroll, got %d", game.rerollCost)
	}
}

func TestRerollCostReset(t *testing.T) {
	game := createTestGame([]string{})

	// Increase reroll cost
	game.rerollCost = 8

	// Simulate starting new blind (reset logic)
	game.rerollCost = 5

	// Should be back to 5
	if game.rerollCost != 5 {
		t.Errorf("Expected reroll cost to reset to 5 for new blind, got %d", game.rerollCost)
	}
}

func TestCanAffordReroll(t *testing.T) {
	game := createTestGame([]string{})

	// Player has $20, reroll costs $5 - should be able to afford
	canAfford := game.money >= game.rerollCost
	if !canAfford {
		t.Errorf("Player with $%d should be able to afford reroll costing $%d", game.money, game.rerollCost)
	}

	// Reduce money below reroll cost
	game.money = 3
	game.rerollCost = 5

	canAfford = game.money >= game.rerollCost
	if canAfford {
		t.Errorf("Player with $%d should NOT be able to afford reroll costing $%d", game.money, game.rerollCost)
	}
}

func TestShopPurchase(t *testing.T) {
	game := createTestGame([]string{})
	initialMoney := game.money

	jokerToBuy := Joker{Name: "Test Purchase", Price: 6, Description: "Test"}

	// Simulate purchase
	if game.money >= jokerToBuy.Price {
		game.money -= jokerToBuy.Price
		game.jokers = append(game.jokers, jokerToBuy)
	}

	// Check money was deducted
	expectedMoney := initialMoney - jokerToBuy.Price
	if game.money != expectedMoney {
		t.Errorf("Expected money to be %d after purchase, got %d", expectedMoney, game.money)
	}

	// Check joker was added
	if len(game.jokers) != 1 {
		t.Errorf("Expected 1 joker after purchase, got %d", len(game.jokers))
	}

	if game.jokers[0].Name != jokerToBuy.Name {
		t.Errorf("Expected purchased joker to be %s, got %s", jokerToBuy.Name, game.jokers[0].Name)
	}
}

func TestPlayerHasJoker(t *testing.T) {
	playerJokers := []Joker{
		{Name: "Owned Joker 1"},
		{Name: "Owned Joker 2"},
	}

	// Test owned joker
	if !PlayerHasJoker(playerJokers, "Owned Joker 1") {
		t.Errorf("PlayerHasJoker should return true for owned joker")
	}

	// Test unowned joker
	if PlayerHasJoker(playerJokers, "Unowned Joker") {
		t.Errorf("PlayerHasJoker should return false for unowned joker")
	}

	// Test empty joker list
	if PlayerHasJoker([]Joker{}, "Any Joker") {
		t.Errorf("PlayerHasJoker should return false for empty joker list")
	}
}

func TestShopWithNoAvailableJokers(t *testing.T) {
	availableJokers := []Joker{
		{Name: "Joker 1", Price: 5},
		{Name: "Joker 2", Price: 6},
	}

	// Player owns all available jokers
	playerJokers := []Joker{
		{Name: "Joker 1", Price: 5},
		{Name: "Joker 2", Price: 6},
	}

	shopItems := selectShopItems(availableJokers, playerJokers, 0)

	// Should return empty shop
	if len(shopItems) != 0 {
		t.Errorf("Expected empty shop when all jokers are owned, got %d items", len(shopItems))
	}
}

func TestShopWithOneAvailableJoker(t *testing.T) {
	availableJokers := []Joker{
		{Name: "Joker 1", Price: 5},
		{Name: "Joker 2", Price: 6},
	}

	// Player owns all but one joker
	playerJokers := []Joker{
		{Name: "Joker 1", Price: 5},
	}

	shopItems := selectShopItems(availableJokers, playerJokers, 0)

	// Should return 1 item
	if len(shopItems) != 1 {
		t.Errorf("Expected 1 shop item when only 1 joker available, got %d items", len(shopItems))
	}

	// Should be the unowned joker
	if shopItems[0].Name != "Joker 2" {
		t.Errorf("Expected shop to offer Joker 2, got %s", shopItems[0].Name)
	}
}
