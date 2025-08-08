package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Bubbletea message types for game events
type gameStartedMsg struct{}
type gameOverMsg GameOverEvent
type victoryMsg struct{}
type gameStateChangedMsg GameStateChangedEvent
type cardsDealtMsg CardsDealtEvent
type handPlayedMsg HandPlayedEvent
type cardsDiscardedMsg CardsDiscardedEvent
type cardsResortedMsg CardsResortedEvent
type blindDefeatedMsg BlindDefeatedEvent
type anteCompletedMsg AnteCompletedEvent
type newBlindStartedMsg NewBlindStartedEvent
type shopOpenedMsg ShopOpenedEvent
type shopItemPurchasedMsg ShopItemPurchasedEvent
type shopRerolledMsg ShopRerolledEvent
type shopClosedMsg struct{}
type invalidActionMsg InvalidActionEvent
type messageEventMsg MessageEvent
type playerActionRequestMsg PlayerActionRequest
type tickMsg time.Time

// TUI model holds the application state
type TUIModel struct {
	// UI state
	currentTime       time.Time
	width             int
	height            int
	showHelp          bool
	selectedCards     []int // indices of selected cards (0-based)
	statusMessage     string
	statusMessageTime time.Time

	// Game state (updated by messages)
	gameState  GameStateChangedEvent
	cards      []Card
	displayMap []int
	sortMode   string
	shopInfo   *ShopOpenedEvent
	mode       Mode

	// Communication with game
	actionRequestPending *PlayerActionRequest
	program              *tea.Program
}

// Styles for the UI components
var (
	topBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 1)

	bottomBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	mainContentStyle = lipgloss.NewStyle().
				Padding(1)

	gameInfoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Margin(0, 1)

	handStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")).
			Padding(1).
			Margin(1, 1)

	heartsCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Margin(0, 1)

	diamondsCardStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Margin(0, 1)

	clubsCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("21")).
			Margin(0, 1)

	spadesCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Margin(0, 1)
)

// Init returns the initial command
func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

func (m TUIModel) IsShopping() bool {
	return m.shopInfo != nil
}

// Update handles messages and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// if msg.(type) != tea.KeyMsg && msg.(type) != tickMsg{
	// }

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		m.currentTime = time.Time(msg)
		return m, tickCmd()
	}

	switch msg := msg.(type) {
	// Game event messages
	case gameStartedMsg:
		m.setStatusMessage("🎮 Game started! Select cards with 1-7, play with Enter/P, discard with D")
		m.mode = GameMode{}
		return m, nil

	case gameStateChangedMsg:
		m.gameState = GameStateChangedEvent(msg)
		return m, nil

	case cardsDealtMsg:
		event := CardsDealtEvent(msg)
		m.cards = make([]Card, len(event.Cards))
		copy(m.cards, event.Cards)
		m.displayMap = make([]int, len(event.DisplayMapping))
		copy(m.displayMap, event.DisplayMapping)
		m.sortMode = event.SortMode
		return m, nil

	case handPlayedMsg:
		event := HandPlayedEvent(msg)
		var message string
		scoreGained := event.FinalScore

		// Check if this completed the blind
		if event.NewTotalScore >= m.gameState.Target {
			message = fmt.Sprintf("🎉 %s for +%d points! BLIND DEFEATED!", event.HandType, scoreGained)
		} else {
			handsLeft := m.gameState.Hands - 1
			if handsLeft <= 0 {
				message = fmt.Sprintf("💀 %s for +%d points, but Game Over! Final: %d/%d", event.HandType, scoreGained, event.NewTotalScore, m.gameState.Target)
			} else {
				progressPercent := float64(event.NewTotalScore) / float64(m.gameState.Target) * 100
				message = fmt.Sprintf("✅ %s for +%d points! %d/%d (%.0f%%) | %d hands left", event.HandType, scoreGained, event.NewTotalScore, m.gameState.Target, progressPercent, handsLeft)
			}
		}
		m.setStatusMessage(message)
		return m, nil

	case cardsDiscardedMsg:
		event := CardsDiscardedEvent(msg)
		var cardNames []string
		for _, card := range event.DiscardedCards {
			cardNames = append(cardNames, card.String())
		}

		discardedStr := strings.Join(cardNames, ", ")
		if len(discardedStr) > 20 {
			discardedStr = fmt.Sprintf("%d cards", event.NumCards)
		}

		var message string
		if event.DiscardsLeft > 0 {
			message = fmt.Sprintf("🗑️ Discarded %s, dealt new cards | %d discards remaining", discardedStr, event.DiscardsLeft)
		} else {
			message = fmt.Sprintf("🗑️ Discarded %s, dealt new cards | No more discards available!", discardedStr)
		}
		m.setStatusMessage(message)
		return m, nil

	case cardsResortedMsg:
		event := CardsResortedEvent(msg)
		m.setStatusMessage(fmt.Sprintf("🔄 Cards now sorted by %s", event.NewSortMode))
		return m, nil

	case blindDefeatedMsg:
		event := BlindDefeatedEvent(msg)
		var message string
		switch event.BlindType {
		case SmallBlind:
			message = "🔸 SMALL BLIND DEFEATED! Advancing to Big Blind..."
		case BigBlind:
			message = "🔶 BIG BLIND CRUSHED! Prepare for the Boss Blind..."
		case BossBlind:
			message = "💀 BOSS BLIND ANNIHILATED! 💀"
		}
		m.setStatusMessage(message)
		return m, nil

	case anteCompletedMsg:
		event := AnteCompletedEvent(msg)
		m.setStatusMessage(fmt.Sprintf("🎊 ANTE %d COMPLETE! Starting Ante %d", event.CompletedAnte, event.NewAnte))
		return m, nil

	case newBlindStartedMsg:
		event := NewBlindStartedEvent(msg)
		blindEmoji := ""
		switch event.Blind {
		case SmallBlind:
			blindEmoji = "🔸"
		case BigBlind:
			blindEmoji = "🔶"
		case BossBlind:
			blindEmoji = "💀"
		}
		m.setStatusMessage(fmt.Sprintf("%s NOW ENTERING: %s (Ante %d) | Target: %d points", blindEmoji, event.Blind, event.Ante, event.Target))
		return m, nil

	case shopOpenedMsg:
		event := ShopOpenedEvent(msg)
		shopCopy := event
		m.shopInfo = &shopCopy
		m.mode = ShoppingMode{}
		m.setStatusMessage("🛍️ Welcome to the Shop!")
		return m, nil

	case shopItemPurchasedMsg:
		event := ShopItemPurchasedEvent(msg)
		m.setStatusMessage(fmt.Sprintf("✨ Purchased %s! Remaining: $%d", event.Item.Name, event.RemainingMoney))
		return m, nil

	case shopRerolledMsg:
		event := ShopRerolledEvent(msg)
		m.setStatusMessage(fmt.Sprintf("💫 Shop rerolled for $%d! Next reroll: $%d", event.Cost, event.NewRerollCost))
		return m, nil

	case shopClosedMsg:
		m.shopInfo = nil
		m.setStatusMessage("👋 Left the shop")
		m.mode = GameMode{}
		return m, nil

	case invalidActionMsg:
		event := InvalidActionEvent(msg)
		m.setStatusMessage(fmt.Sprintf("❌ %s", event.Reason))
		return m, nil

	case messageEventMsg:
		event := MessageEvent(msg)
		switch event.Type {
		case "error":
			m.setStatusMessage(fmt.Sprintf("❌ %s", event.Message))
		case "warning":
			m.setStatusMessage(fmt.Sprintf("⚠️ %s", event.Message))
		case "success":
			m.setStatusMessage(fmt.Sprintf("✅ %s", event.Message))
		case "info":
			m.setStatusMessage(fmt.Sprintf("ℹ️ %s", event.Message))
		default:
			m.setStatusMessage(event.Message)
		}
		return m, nil

	case gameOverMsg:
		event := GameOverEvent(msg)
		m.setStatusMessage(fmt.Sprintf("💀 GAME OVER! Final: %d/%d (Ante %d)", event.FinalScore, event.Target, event.Ante))
		return m, nil

	case victoryMsg:
		m.setStatusMessage("🏆 VICTORY! You conquered all 8 Antes! 🎉")
		return m, nil

	case playerActionRequestMsg:
		request := PlayerActionRequest(msg)
		m.actionRequestPending = &request
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m TUIModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.actionRequestPending != nil {
			// Capture response channel before clearing the pending request
			responseChan := m.actionRequestPending.ResponseChan
			m.actionRequestPending = nil

			// Send quit response
			go func() {
				responseChan <- PlayerActionResponse{
					Action: PlayerActionNone,
					Params: nil,
					Quit:   true,
				}
			}()
		}
		return m, tea.Quit

	case "h":
		m.showHelp = !m.showHelp
		return m, nil

	case "1", "2", "3", "4", "5", "6", "7":
		if !m.showHelp {
			cardIndex, _ := strconv.Atoi(msg.String())
			if cardIndex <= len(m.cards) {
				m.toggleCardSelection(cardIndex - 1) // Convert to 0-based
			} else {
				m.setStatusMessage(fmt.Sprintf("Invalid card number: %d (only have %d cards)", cardIndex, len(m.cards)))
			}
		}
		return m, nil

	case "enter", "p":
		if !m.showHelp {
			if len(m.selectedCards) > 0 {
				m.handlePlay()
			} else {
				m.setStatusMessage("Select cards first using number keys 1-7")
			}
		}
		return m, nil

	case "d":
		if !m.showHelp {
			if len(m.selectedCards) > 0 {
				m.handleDiscard()
			} else {
				m.setStatusMessage("Select cards first using number keys 1-7")
			}
		}
		return m, nil

	case "r":
		if !m.showHelp {
			m.handleResort()
		}
		return m, nil

	case "escape", "c":
		if !m.showHelp {
			m.selectedCards = []int{}
			m.setStatusMessage("Selection cleared")
		}
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m TUIModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Resize the window please..."
	}

	// Top bar
	topBar := topBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("🃏 Welcome to Balatno"))

	// Status bar (second from bottom)
	statusBar := bottomBarStyle.
		Width(m.width).
		Render(m.getStatusMessage())

	// Bottom bar with time and controls
	timeStr := m.currentTime.Format("15:04:05")
	controls := "⏰ " + timeStr + " | 1-7: select cards, Enter/P: play, D: discard, C: clear, R: resort, H: help, Q: quit"
	bottomBar := bottomBarStyle.
		Width(m.width).
		Render(controls)

	// Main content area - fixed height
	contentHeight := m.height - 3 // Account for top, status, and bottom bars

	var content string
	if m.showHelp {
		m.mode = m.mode.toggleHelp()
	}

	content = m.mode.renderContent(m)

	renderedContent := mainContentStyle.
		Width(m.width).
		Height(contentHeight).
		Render(content)

	// Combine all parts with fixed layout
	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		renderedContent,
		statusBar,
		bottomBar,
	)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type Mode interface {
	renderContent(m TUIModel) string
	toggleHelp() Mode
}

// getStatusMessage returns the current status message or default message
func (m TUIModel) getStatusMessage() string {
	if m.statusMessage != "" {
		return m.statusMessage
	}
	return "Select cards with 1-7, play with Enter/P, discard with D"
}

// setStatusMessage sets a status message with timestamp for auto-clearing
func (m *TUIModel) setStatusMessage(msg string) {
	m.statusMessage = msg
	m.statusMessageTime = time.Now()
}

// isCardSelected checks if a card at the given index is selected
func (m TUIModel) isCardSelected(index int) bool {
	for _, selected := range m.selectedCards {
		if selected == index {
			return true
		}
	}
	return false
}

// toggleCardSelection toggles selection of a card at the given index.
// This
func (m *TUIModel) toggleCardSelection(index int) {
	if index < 0 || index >= len(m.cards) {
		return
	}

	// Check if already selected
	for i, selected := range m.selectedCards {
		if selected == index {
			// Remove from selection
			card := m.cards[index]
			m.selectedCards = append(m.selectedCards[:i], m.selectedCards[i+1:]...)
			remaining := len(m.selectedCards)
			if remaining > 0 {
				m.setStatusMessage(fmt.Sprintf("✗ Deselected %s | %d cards still selected", card.String(), remaining))
			} else {
				m.setStatusMessage(fmt.Sprintf("✗ Deselected %s | No cards selected", card.String()))
			}
			return
		}
	}

	// Add to selection (max 5 cards for poker)
	if len(m.selectedCards) >= 5 {
		m.setStatusMessage("⚠️ Maximum 5 cards can be selected for a poker hand")
		return
	}

	m.selectedCards = append(m.selectedCards, index)
	card := m.cards[index]
	m.setStatusMessage(fmt.Sprintf("✓ Selected %s (card %d) | %d/5 cards selected", card.String(), index+1, len(m.selectedCards)))
}

// handlePlay processes playing the selected cards
func (m *TUIModel) handlePlay() {
	if len(m.selectedCards) == 0 {
		m.setStatusMessage("No cards selected to play")
		return
	}

	if m.gameState.Hands <= 0 {
		m.setStatusMessage("No hands remaining!")
		return
	}

	if m.actionRequestPending != nil {
		// Convert selected indices to string params for game logic
		var params []string
		for _, index := range m.selectedCards {
			// Convert 0-based TUI index to 1-based display index for game logic
			params = append(params, fmt.Sprintf("%d", index+1))
		}

		// Capture response channel before clearing the pending request
		responseChan := m.actionRequestPending.ResponseChan
		m.actionRequestPending = nil

		// Send response
		go func() {
			responseChan <- PlayerActionResponse{
				Action: PlayerActionPlay,
				Params: params,
				Quit:   false,
			}
		}()

		// Clear selection
		m.selectedCards = []int{}
	}
}

// handleDiscard processes discarding the selected cards
func (m *TUIModel) handleDiscard() {
	if len(m.selectedCards) == 0 {
		m.setStatusMessage("No cards selected to discard")
		return
	}

	if m.gameState.Discards <= 0 {
		m.setStatusMessage("No discards remaining!")
		return
	}

	if m.actionRequestPending != nil && m.actionRequestPending.CanDiscard {
		// Convert selected indices to string params for game logic
		var params []string
		for _, index := range m.selectedCards {
			// Convert 0-based TUI index to 1-based display index for game logic
			params = append(params, fmt.Sprintf("%d", index+1))
		}

		// Capture response channel before clearing the pending request
		responseChan := m.actionRequestPending.ResponseChan
		m.actionRequestPending = nil

		// Send response
		go func() {
			responseChan <- PlayerActionResponse{
				Action: PlayerActionDiscard,
				Params: params,
				Quit:   false,
			}
		}()

		// Clear selection
		m.selectedCards = []int{}
	} else {
		m.setStatusMessage("Cannot discard at this time")
	}
}

// handleResort processes resort action
func (m *TUIModel) handleResort() {
	if m.actionRequestPending != nil {
		// Capture response channel before clearing the pending request
		responseChan := m.actionRequestPending.ResponseChan
		m.actionRequestPending = nil

		// Send response
		go func() {
			responseChan <- PlayerActionResponse{
				Action: PlayerActionResort,
				Params: nil,
				Quit:   false,
			}
		}()
	}
}

// SetProgram allows the event handler to send messages to this TUI
func (m *TUIModel) SetProgram(program *tea.Program) {
	m.program = program
}

// SendMessage sends a bubbletea message to update the UI
func (m *TUIModel) SendMessage(msg tea.Msg) {
	if m.program != nil {
		m.program.Send(msg)
	}
}

// RunTUI starts the TUI application
func RunTUI() error {
	// Create TUI model
	model := TUIModel{
		currentTime:   time.Now(),
		selectedCards: []int{},
	}

	// Create TUI program
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Set the program reference so we can send messages
	model.SetProgram(program)

	// Create TUI event handler
	eventHandler := NewTUIEventHandler()
	eventHandler.SetTUIModel(&model)

	// Create game with event handler
	game := NewGame(eventHandler)

	// Start the game in a goroutine
	go game.Run()

	// Run the TUI
	_, err := program.Run()
	return err
}
