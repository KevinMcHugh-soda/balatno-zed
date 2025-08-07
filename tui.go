package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI model holds the application state
type TUIModel struct {
	currentTime       time.Time
	width             int
	height            int
	eventHandler      *TUIEventHandler
	showHelp          bool
	selectedCards     []int // indices of selected cards (0-based)
	statusMessage     string
	statusMessageTime time.Time
}

// tickMsg is sent every second to update the time
type tickMsg time.Time

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

// Update handles messages and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h":
			m.showHelp = !m.showHelp
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7":
			if !m.showHelp && m.eventHandler != nil {
				cardIndex, _ := strconv.Atoi(msg.String())
				cards, _ := m.eventHandler.GetCards()
				if cardIndex <= len(cards) {
					m.toggleCardSelection(cardIndex - 1) // Convert to 0-based
				} else {
					m.setStatusMessage(fmt.Sprintf("Invalid card number: %d (only have %d cards)", cardIndex, len(cards)))
				}
			}
			return m, nil
		case "enter", "p":
			if !m.showHelp && m.eventHandler != nil {
				if len(m.selectedCards) > 0 {
					m.handlePlay()
				} else {
					m.setStatusMessage("Select cards first using number keys 1-7")
				}
			}
			return m, nil
		case "d":
			if !m.showHelp && m.eventHandler != nil {
				if len(m.selectedCards) > 0 {
					m.handleDiscard()
				} else {
					m.setStatusMessage("Select cards first using number keys 1-7")
				}
			}
			return m, nil
		case "escape", "c":
			if !m.showHelp {
				m.selectedCards = []int{}
				m.setStatusMessage("Selection cleared")
			}
			return m, nil
		}

	case tickMsg:
		m.currentTime = time.Time(msg)

		return m, tickCmd()
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
		Render("🃏 Welcome to Balatno")

	// Status bar (second from bottom)
	statusBar := bottomBarStyle.
		Width(m.width).
		Render(m.getStatusMessage())

	// Bottom bar with time and controls
	timeStr := m.currentTime.Format("15:04:05")
	controls := "⏰ " + timeStr + " | 1-7: select cards, Enter/P: play, D: discard, C: clear, H: help, Q: quit"
	bottomBar := bottomBarStyle.
		Width(m.width).
		Render(controls)

	// Main content area - fixed height
	contentHeight := m.height - 3 // Account for top, status, and bottom bars

	var content string
	if m.showHelp {
		content = m.renderHelp()
	} else if m.eventHandler != nil {
		content = m.renderGameContent()
	} else {
		content = "Loading game..."
	}

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

// renderGameContent renders the main game area
func (m TUIModel) renderGameContent() string {
	if m.eventHandler == nil {
		return "Game not initialized"
	}

	gameState := m.eventHandler.GetGameState()

	// Game status info - fixed height section
	progress := float64(gameState.Score) / float64(gameState.Target)
	if progress > 1.0 {
		progress = 1.0
	}
	progressWidth := 20
	filled := int(progress * float64(progressWidth))

	progressBar := ""
	for i := 0; i < progressWidth; i++ {
		if i < filled {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}

	// Blind type emojis
	blindEmoji := ""
	switch gameState.Blind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	gameInfo := fmt.Sprintf("%s Ante %d - %s\n", blindEmoji, gameState.Ante, gameState.Blind) +
		fmt.Sprintf("🎯 Target: %d | Current Score: %d [%s] (%.1f%%)\n",
			gameState.Target, gameState.Score, progressBar, progress*100) +
		fmt.Sprintf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d",
			gameState.Hands, gameState.Discards, gameState.Money)

	gameInfoBox := gameInfoStyle.
		Height(5).
		Render(gameInfo)

	// Render hand - fixed height section
	hand := m.renderHand()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		gameInfoBox,
		hand,
	)
}

// renderHand renders the player's current hand of cards
func (m TUIModel) renderHand() string {
	if m.eventHandler == nil {
		return handStyle.Height(10).Render("No cards in hand")
	}

	cards, _ := m.eventHandler.GetCards()
	if len(cards) == 0 {
		return handStyle.Height(10).Render("No cards in hand")
	}

	var content strings.Builder

	// Hand cards area - fixed position
	content.WriteString(fmt.Sprintf("🃏 Your Hand (%d cards):\n", len(cards)))
	var cardViews []string
	for i, card := range cards {
		isSelected := m.isCardSelected(i)
		cardStr := m.renderCard(card, false)

		// Add position number below the card
		posNumStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Margin(0, 1)

		if isSelected {
			posNumStyle = posNumStyle.Foreground(lipgloss.Color("226")).Bold(true)
		}

		positionNum := posNumStyle.Render(fmt.Sprintf("%d", i+1))

		// Combine card and position number vertically
		cardWithPos := lipgloss.JoinVertical(lipgloss.Center, cardStr, positionNum)
		cardViews = append(cardViews, cardWithPos)
	}

	// Display all cards in a single row
	cardsDisplay := lipgloss.JoinHorizontal(lipgloss.Top, cardViews...)
	content.WriteString(cardsDisplay)

	return handStyle.Height(10).Render(content.String())
}

// renderHelp renders the help screen
func (m TUIModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1).
		Margin(1)

	helpContent := `🎮 BALATNO HELP

🎯 OBJECTIVE:
   Reach the target score by playing poker hands before running out of hands/discards

🃏 GAME ELEMENTS:
   • Ante: Current level (1-8)
   • Blinds: Small 🔸, Big 🔶, Boss 💀
   • Score: Current total vs target score
   • Hands: Number of plays remaining
   • Discards: Number of discards remaining
   • Money: Used for shop purchases
   • Cards: Displayed as compact 2-char format (e.g., A♠, K♥)
     - Hearts ♥: Red, Diamonds ♦: Orange
     - Clubs ♣: Dark Blue, Spades ♠: Gray

🎴 POKER HANDS (from weakest to strongest):
   • High Card      • Pair           • Two Pair
   • Three of Kind  • Straight       • Flush
   • Full House     • Four of Kind   • Straight Flush

💰 SCORING:
   • Base chips + multiplier for hand type
   • Jokers can modify scoring significantly
   • Unused hands/discards give bonus money

⌨️  GAMEPLAY CONTROLS:
   • 1-7: Select/deselect cards by position
   • Enter/P: Play selected cards
   • D: Discard selected cards
   • C/Escape: Clear selection
   • H: Toggle this help screen
   • Q: Quit game

📝 HOW TO PLAY:
   1. Select cards using number keys (1-7)
   2. Selected cards appear above your hand
   3. Press Enter/P to play selected cards as a poker hand
   4. Press D to discard selected cards for new ones
   5. Use C to clear your selection

🎯 GOAL: Beat the target score before running out of hands!`

	return helpStyle.Render(helpContent)
}

// getStatusMessage returns the current status message or default message
func (m TUIModel) getStatusMessage() string {
	// Check for status message from event handler first
	if m.eventHandler != nil {
		if msg := m.eventHandler.GetStatusMessage(); msg != "" {
			return msg
		}
	}

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

// renderCard renders a single card with appropriate styling
func (m TUIModel) renderCard(card Card, isInSelectedArea bool) string {
	cardStr := fmt.Sprintf("%s%s", card.Rank.String(), card.Suit.String())

	var style lipgloss.Style
	switch card.Suit {
	case Hearts:
		style = heartsCardStyle
	case Diamonds:
		style = diamondsCardStyle
	case Clubs:
		style = clubsCardStyle
	case Spades:
		style = spadesCardStyle
	}

	if isInSelectedArea {
		style = style.Bold(true).Background(lipgloss.Color("235"))
	}

	return style.Render(cardStr)
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

// toggleCardSelection toggles selection of a card at the given index
func (m *TUIModel) toggleCardSelection(index int) {
	if m.eventHandler == nil {
		return
	}

	cards, _ := m.eventHandler.GetCards()
	if index < 0 || index >= len(cards) {
		return
	}

	// Check if already selected
	for i, selected := range m.selectedCards {
		if selected == index {
			// Remove from selection
			card := cards[index]
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
	card := cards[index]
	m.setStatusMessage(fmt.Sprintf("✓ Selected %s (card %d) | %d/5 cards selected", card.String(), index+1, len(m.selectedCards)))
}

// handlePlay processes playing the selected cards
func (m *TUIModel) handlePlay() {
	if len(m.selectedCards) == 0 {
		m.setStatusMessage("No cards selected to play")
		return
	}

	gameState := m.eventHandler.GetGameState()
	if gameState.Hands <= 0 {
		m.setStatusMessage("No hands remaining!")
		return
	}

	// Use event handler to handle the play action
	m.eventHandler.HandlePlayAction(m.selectedCards)

	// Clear selection
	m.selectedCards = []int{}

	// Status message will be updated by event handler when HandPlayedEvent is received
}

// handleDiscard processes discarding the selected cards
func (m *TUIModel) handleDiscard() {
	if len(m.selectedCards) == 0 {
		m.setStatusMessage("No cards selected to discard")
		return
	}

	gameState := m.eventHandler.GetGameState()
	if gameState.Discards <= 0 {
		m.setStatusMessage("No discards remaining!")
		return
	}

	// Use event handler to handle the discard action
	m.eventHandler.HandleDiscardAction(m.selectedCards)

	// Clear selection
	m.selectedCards = []int{}

	// Status message will be updated by event handler when CardsDiscardedEvent is received
}

// RunTUI starts the TUI application
func RunTUI() error {
	// Create TUI event handler
	eventHandler := NewTUIEventHandler()

	// Create game with event handler
	game := NewGame(eventHandler)

	// Create TUI model
	m := TUIModel{
		currentTime:   time.Now(),
		eventHandler:  eventHandler,
		selectedCards: []int{},
	}

	// Link the event handler to the TUI model
	eventHandler.SetTUIModel(&m)

	// Start the game in a goroutine
	go game.Run()

	// Set initial status message
	m.setStatusMessage("Welcome! Select cards with 1-7, play with Enter/P, discard with D")

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
