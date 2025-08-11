package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	game "balatno/internal/game"
)

// Bubbletea message types for game events
type gameStartedMsg struct{}
type gameOverMsg game.GameOverEvent
type victoryMsg struct{}
type gameStateChangedMsg game.GameStateChangedEvent
type cardsDealtMsg game.CardsDealtEvent
type handPlayedMsg game.HandPlayedEvent
type cardsDiscardedMsg game.CardsDiscardedEvent
type cardsResortedMsg game.CardsResortedEvent
type blindDefeatedMsg game.BlindDefeatedEvent
type anteCompletedMsg game.AnteCompletedEvent
type newBlindStartedMsg game.NewBlindStartedEvent
type shopOpenedMsg game.ShopOpenedEvent
type shopItemPurchasedMsg game.ShopItemPurchasedEvent
type shopRerolledMsg game.ShopRerolledEvent
type shopClosedMsg struct{}
type invalidActionMsg game.InvalidActionEvent
type messageEventMsg game.MessageEvent
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
	eventLog          []string

	// Timeout configuration
	timeoutDuration time.Duration
	lastActivity    time.Time

	// Game state (updated by messages)
	gameState  game.GameStateChangedEvent
	cards      []game.Card
	displayMap []int
	sortMode   string
	shopInfo   *game.ShopOpenedEvent
	mode       Mode

	// Communication with game
	actionRequestPending *PlayerActionRequest
	program              *tea.Program

	viewport viewport.Model
}

// getTimeoutDuration reads the BALATRO_TIMEOUT environment variable or returns default 60s
func getTimeoutDuration() time.Duration {
	timeoutStr := os.Getenv("BALATRO_TIMEOUT")
	if timeoutStr == "" {
		return 60 * time.Second // Default 60 seconds
	}

	timeoutSecs, err := strconv.Atoi(timeoutStr)
	if err != nil || timeoutSecs <= 0 {
		fmt.Printf("Warning: Invalid BALATRO_TIMEOUT value '%s', using default 60s\n", timeoutStr)
		return 60 * time.Second
	}

	return time.Duration(timeoutSecs) * time.Second
}

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
		m.viewport = viewport.New(30, m.height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		m.currentTime = time.Time(msg)

		// Check for timeout
		if time.Since(m.lastActivity) > m.timeoutDuration {
			m.setStatusMessage("⏰ Timeout reached - shutting down gracefully")

			// Handle pending action request if any
			if m.actionRequestPending != nil {
				// Capture response channel before clearing the pending request
				responseChan := m.actionRequestPending.ResponseChan
				m.actionRequestPending = nil

				// Send quit response
				go func() {
					responseChan <- PlayerActionResponse{
						Action: game.PlayerActionNone,
						Params: nil,
						Quit:   true,
					}
				}()
			}

			return m, tea.Sequence(
				tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return tea.Quit() }),
				tickCmd(),
			)
		}

		return m, tickCmd()
	}

	switch msg := msg.(type) {
	// Game event messages
	case gameStartedMsg:
		msgStr := "🎮 Game started! Select cards with 1-7, play with Enter/P, discard with D"
		m.setStatusMessage(msgStr)
		m.logEvent("Game started")
		m.mode = GameMode{}
		return m, nil

	case gameStateChangedMsg:
		event := game.GameStateChangedEvent(msg)
		m.gameState = event
		m.logEvent(fmt.Sprintf("Game state: score %d/%d, hands %d, discards %d", event.Score, event.Target, event.Hands, event.Discards))
		return m, nil

	case cardsDealtMsg:
		event := game.CardsDealtEvent(msg)
		m.cards = make([]game.Card, len(event.Cards))
		copy(m.cards, event.Cards)
		m.displayMap = make([]int, len(event.DisplayMapping))
		copy(m.displayMap, event.DisplayMapping)
		m.sortMode = event.SortMode
		m.logEvent(fmt.Sprintf("Dealt %d cards (sorted by %s)", len(event.Cards), event.SortMode))
		return m, nil

	case handPlayedMsg:
		m.lastActivity = time.Now() // User played cards
		event := game.HandPlayedEvent(msg)
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
		m.logEvent(message)
		return m, nil

	case cardsDiscardedMsg:
		m.lastActivity = time.Now() // User discarded cards
		event := game.CardsDiscardedEvent(msg)
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
		m.logEvent(message)
		return m, nil

	case cardsResortedMsg:
		m.lastActivity = time.Now() // User resorted cards
		event := game.CardsResortedEvent(msg)
		msgStr := fmt.Sprintf("🔄 Cards now sorted by %s", event.NewSortMode)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case blindDefeatedMsg:
		event := game.BlindDefeatedEvent(msg)
		// Update money immediately when blind is defeated so the
		// shop shows the correct amount without waiting for the
		// next blind to begin.
		m.gameState.Money = event.NewMoney
		var message string
		switch event.BlindType {
		case game.SmallBlind:
			message = "🔸 SMALL BLIND DEFEATED! Advancing to Big Blind..."
		case game.BigBlind:
			message = "🔶 BIG BLIND CRUSHED! Prepare for the Boss Blind..."
		case game.BossBlind:
			message = "💀 BOSS BLIND ANNIHILATED! 💀"
		}
		m.setStatusMessage(message)
		m.logEvent(message)
		return m, nil

	case anteCompletedMsg:
		event := game.AnteCompletedEvent(msg)
		msgStr := fmt.Sprintf("🎊 ANTE %d COMPLETE! Starting Ante %d", event.CompletedAnte, event.NewAnte)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case newBlindStartedMsg:
		event := game.NewBlindStartedEvent(msg)
		blindEmoji := ""
		switch event.Blind {
		case game.SmallBlind:
			blindEmoji = "🔸"
		case game.BigBlind:
			blindEmoji = "🔶"
		case game.BossBlind:
			blindEmoji = "💀"
		}
		msgStr := fmt.Sprintf("%s NOW ENTERING: %s (Ante %d) | Target: %d points", blindEmoji, event.Blind, event.Ante, event.Target)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case shopOpenedMsg:
		event := game.ShopOpenedEvent(msg)
		shopCopy := event
		m.shopInfo = &shopCopy
		m.gameState.Money = event.Money
		m.mode = &ShoppingMode{}
		m.setStatusMessage("🛍️ Welcome to the Shop!")
		m.logEvent("Entered shop")
		return m, nil

	case shopItemPurchasedMsg:
		m.lastActivity = time.Now() // User purchased item
		event := game.ShopItemPurchasedEvent(msg)
		// Update money after purchase so UI reflects remaining funds
		m.gameState.Money = event.RemainingMoney
		if m.shopInfo != nil {
			m.shopInfo.Money = event.RemainingMoney
		}
		msgStr := fmt.Sprintf("✨ Purchased %s! Remaining: $%d", event.Item.Name, event.RemainingMoney)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case shopRerolledMsg:
		m.lastActivity = time.Now() // User rerolled shop
		event := game.ShopRerolledEvent(msg)
		// Update money after reroll and reflect new reroll cost/items
		m.gameState.Money = event.RemainingMoney
		m.shopInfo.Money = event.RemainingMoney
		m.shopInfo.RerollCost = event.NewRerollCost
		m.shopInfo.Items = event.NewItems
		msgStr := fmt.Sprintf("💫 Shop rerolled for $%d! Next reroll: $%d", event.Cost, event.NewRerollCost)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case shopClosedMsg:
		m.shopInfo = nil
		m.setStatusMessage("👋 Left the shop")
		m.logEvent("Left shop")
		m.mode = GameMode{}
		return m, nil

	case invalidActionMsg:
		event := game.InvalidActionEvent(msg)
		msgStr := fmt.Sprintf("❌ %s", event.Reason)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case messageEventMsg:
		event := game.MessageEvent(msg)
		var msgStr string
		switch event.Type {
		case "error":
			msgStr = fmt.Sprintf("❌ %s", event.Message)
		case "warning":
			msgStr = fmt.Sprintf("⚠️ %s", event.Message)
		case "success":
			msgStr = fmt.Sprintf("✅ %s", event.Message)
		case "info":
			msgStr = fmt.Sprintf("ℹ️ %s", event.Message)
		default:
			msgStr = event.Message
		}
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case gameOverMsg:
		event := game.GameOverEvent(msg)
		msgStr := fmt.Sprintf("💀 GAME OVER! Final: %d/%d (Ante %d)", event.FinalScore, event.Target, event.Ante)
		m.setStatusMessage(msgStr)
		m.logEvent(msgStr)
		return m, nil

	case victoryMsg:
		msgStr := "🏆 VICTORY! You conquered all 8 Antes! 🎉"
		m.setStatusMessage(msgStr)
		m.logEvent("Victory achieved")
		return m, nil

	case playerActionRequestMsg:
		request := PlayerActionRequest(msg)
		m.actionRequestPending = &request
		m.logEvent(fmt.Sprintf("Awaiting player action (discard allowed: %t)", request.CanDiscard))
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m TUIModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Update last activity time on any key press
	m.lastActivity = time.Now()

	// Some keypresses should be handled the same way every time.
	switch msg.String() {
	case "ctrl+c", "q":
		if m.actionRequestPending != nil {
			// Capture response channel before clearing the pending request
			responseChan := m.actionRequestPending.ResponseChan
			m.actionRequestPending = nil

			// Send quit response
			go func() {
				responseChan <- PlayerActionResponse{
					Action: game.PlayerActionNone,
					Params: nil,
					Quit:   true,
				}
			}()
		}
		return m, tea.Quit

	case "h":
		m.showHelp = !m.showHelp
		m.mode = m.mode.toggleHelp()
		return m, nil
	}
	// But mostly what the keypress should do is handled by the mode we're currently in
	return m.mode.handleKeyPress(&m, msg.String())
}

// View renders the UI
func (m TUIModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Resize the window please..."
	}

	// Top bar
	topFrame, _ := topBarStyle.GetFrameSize()
	topWidth := m.width - topFrame
	if topWidth < 0 {
		topWidth = 0
	}
	topBar := topBarStyle.
		Width(topWidth).
		Render("🃏 Welcome to Balatno")

	// Status bar (second from bottom)
	barFrame, _ := bottomBarStyle.GetFrameSize()
	barWidth := m.width - barFrame
	if barWidth < 0 {
		barWidth = 0
	}
	statusBar := bottomBarStyle.
		Width(barWidth).
		Render(m.getStatusMessage())

	// Bottom bar with time and controls
	timeStr := time.Now().Format("15:04:05")
	timeoutRemaining := m.timeoutDuration - time.Since(m.lastActivity)
	timeoutStr := fmt.Sprintf("%.0fs", timeoutRemaining.Seconds())
	controls := "⏰ " + timeStr + " | Timeout: " + timeoutStr + m.mode.getControls()
	bottomBar := bottomBarStyle.
		Width(barWidth).
		Render(controls)

	// Main content area - fixed height
	contentHeight := m.height - 3 // Account for top, status, and bottom bars

	var content string

	content = m.mode.renderContent(m)

	logWidth := 30
	// logFrame, _ := eventLogStyle.GetFrameSize()
	// logContentWidth := logWidth - logFrame
	// if logContentWidth < 0 {
	// 	logContentWidth = 0
	// }

	contentAreaWidth := m.width - logWidth
	mainFrame, _ := mainContentStyle.GetFrameSize()
	contentWidth := contentAreaWidth - mainFrame
	if contentWidth < 0 {
		contentWidth = 0
	}

	renderedContent := mainContentStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(content)

	m.viewport.SetContent(m.renderEventLog(contentHeight, logWidth))
	// logView := eventLogStyle.
	// 	Width(logContentWidth).
	// 	Height(contentHeight).
	// 	Render(m.renderEventLog(contentHeight, logContentWidth))

	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, renderedContent, m.viewport.View())

	// Combine all parts with fixed layout
	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		mainArea,
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
	handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd)
	getControls() string
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

// logEvent appends a message to the event log
func (m *TUIModel) logEvent(msg string) {
	m.eventLog = append(m.eventLog, msg)
}

// renderEventLog returns the log contents trimmed to fit the given height
func (m TUIModel) renderEventLog(height, width int) string {
	if len(m.eventLog) == 0 {
		return ""
	}
	start := 0
	if len(m.eventLog) > height {
		start = len(m.eventLog) - height
	}
	truncatedStrs := make([]string, height)
	for _, str := range m.eventLog[start:] {
		if len(str) > width {
			truncatedStrs = append(truncatedStrs, str[:width-3]+"...")
		} else {
			truncatedStrs = append(truncatedStrs, str)
		}
	}
	return strings.Join(m.eventLog[start:], "\n")
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

	// Convert selected indices to string params for game logic
	var params []string
	for _, index := range m.selectedCards {
		// Convert 0-based TUI index to 1-based display index for game logic
		params = append(params, fmt.Sprintf("%d", index+1))
	}

	m.sendAction(game.PlayerActionPlay, params)

	// Clear selection
	m.selectedCards = []int{}
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

		m.sendAction(game.PlayerActionDiscard, params)

		// Clear selection
		m.selectedCards = []int{}
	} else {
		m.setStatusMessage("Cannot discard at this time")
	}
}

// handleResort processes resort action
func (m *TUIModel) handleResort() {
	m.sendAction(game.PlayerActionResort, nil)
}

// sendAction sends a PlayerActionResponse if an action request is pending
func (m *TUIModel) sendAction(action game.PlayerAction, params []string) {
	if m.actionRequestPending == nil {
		return
	}

	responseChan := m.actionRequestPending.ResponseChan
	m.actionRequestPending = nil

	go func() {
		responseChan <- PlayerActionResponse{
			Action: action,
			Params: params,
			Quit:   false,
		}
	}()
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
		timeoutDuration: getTimeoutDuration(),
		lastActivity:    time.Now(),
		currentTime:     time.Now(),
		selectedCards:   []int{},
		eventLog:        []string{},
	}
	// model.Init()

	// Create TUI program
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Set the program reference so we can send messages
	model.SetProgram(program)

	// Create TUI event handler
	eventHandler := NewTUIEventHandler()
	eventHandler.SetTUIModel(&model)

	// Create game with event handler
	game := game.NewGame(eventHandler)

	// Start the game in a goroutine
	go game.Run()

	// Run the TUI
	_, err := program.Run()
	return err
}
