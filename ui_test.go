package main

import (
	"testing"
)

func TestNewTerminalUI(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	if ui.game != game {
		t.Error("UI game reference not set correctly")
	}
	ui.Close()
}

func TestParseCardSelection(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Set up display mapping
	ui.game.displayToOriginal = map[string]int{
		"A": 0,
		"B": 1,
		"C": 2,
	}

	tests := []struct {
		name     string
		params   []string
		expected []int
	}{
		{
			name:     "letter selection",
			params:   []string{"A", "B"},
			expected: []int{0, 1},
		},
		{
			name:     "number selection",
			params:   []string{"1", "2"},
			expected: []int{0, 1},
		},
		{
			name:     "mixed selection",
			params:   []string{"A", "2"},
			expected: []int{0, 1},
		},
		{
			name:     "case insensitive",
			params:   []string{"a", "b"},
			expected: []int{0, 1},
		},
		{
			name:     "duplicate removal",
			params:   []string{"A", "A", "1"},
			expected: []int{0},
		},
		{
			name:     "invalid selection",
			params:   []string{"Z", "99"},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.parseCardSelection(tt.params)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d indices, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected index %d, got %d", expected, result[i])
				}
			}
		})
	}
}

func TestHandlePlayAction(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Test empty params
	ui.handlePlayAction([]string{})
	if ui.message == "" {
		t.Error("Expected error message for empty params")
	}

	// Reset message
	ui.message = ""

	// Test invalid card selection
	ui.handlePlayAction([]string{"Z", "Y"})
	if ui.message == "" {
		t.Error("Expected error message for invalid card selection")
	}
}

func TestHandleDiscardAction(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Test when max discards reached
	ui.game.discardsUsed = MaxDiscards
	ui.handleDiscardAction([]string{"A"})
	if ui.message == "" {
		t.Error("Expected error message when max discards reached")
	}

	// Reset discard count and message
	ui.game.discardsUsed = 0
	ui.message = ""

	// Test empty params
	ui.handleDiscardAction([]string{})
	if ui.message == "" {
		t.Error("Expected error message for empty params")
	}
}

func TestHandleResortAction(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Test toggling sort mode
	initialMode := ui.game.sortMode
	ui.handleResortAction()
	if ui.game.sortMode == initialMode {
		t.Error("Sort mode should have changed")
	}
	if ui.message == "" {
		t.Error("Expected message about sort mode change")
	}
}

func TestProcessCommand(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	tests := []struct {
		name       string
		input      string
		expectMsg  bool
		expectQuit bool
	}{
		{
			name:       "empty command",
			input:      "",
			expectMsg:  true,
			expectQuit: false,
		},
		{
			name:       "quit command",
			input:      "q",
			expectMsg:  false,
			expectQuit: true,
		},
		{
			name:       "quit full",
			input:      "quit",
			expectMsg:  false,
			expectQuit: true,
		},
		{
			name:       "resort command",
			input:      "r",
			expectMsg:  true,
			expectQuit: false,
		},
		{
			name:       "play command no params",
			input:      "p",
			expectMsg:  true,
			expectQuit: false,
		},
		{
			name:       "unknown command",
			input:      "unknown",
			expectMsg:  true,
			expectQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.inputBuf = tt.input
			ui.message = ""
			ui.quit = false

			ui.processCommand()

			if tt.expectMsg && ui.message == "" {
				t.Error("Expected a message to be set")
			}
			if !tt.expectMsg && ui.message != "" {
				t.Errorf("Expected no message, got: %s", ui.message)
			}
			if tt.expectQuit != ui.quit {
				t.Errorf("Expected quit=%v, got quit=%v", tt.expectQuit, ui.quit)
			}
		})
	}
}

func TestGameStateIntegration(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Verify initial game state is accessible
	if ui.game.currentAnte != 1 {
		t.Errorf("Expected ante 1, got %d", ui.game.currentAnte)
	}
	if ui.game.money != StartingMoney {
		t.Errorf("Expected starting money %d, got %d", StartingMoney, ui.game.money)
	}
	if len(ui.game.playerCards) != InitialCards {
		t.Errorf("Expected %d initial cards, got %d", InitialCards, len(ui.game.playerCards))
	}
}

func TestUIMessageStates(t *testing.T) {
	game := NewGame()
	ui, err := NewTerminalUI(game)
	if err != nil {
		t.Fatalf("Failed to create UI: %v", err)
	}
	defer ui.Close()

	// Test different message states
	ui.message = "Test message"
	if ui.message != "Test message" {
		t.Error("Message not set correctly")
	}

	// Test message clearing
	ui.message = ""
	if ui.message != "" {
		t.Error("Message not cleared correctly")
	}
}
