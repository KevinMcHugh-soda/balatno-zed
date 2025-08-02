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
	currentTime   time.Time
	width         int
	height        int
	game          *Game
	showHelp      bool
	selectedCards []int // indices of selected cards (0-based)
	statusMessage string
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
			if !m.showHelp && m.game != nil {
				cardIndex, _ := strconv.Atoi(msg.String())
				if cardIndex <= len(m.game.playerCards) {
					m.toggleCardSelection(cardIndex - 1) // Convert to 0-based
				} else {
					m.statusMessage = fmt.Sprintf("Invalid card number: %d (only have %d cards)", cardIndex, len(m.game.playerCards))
				}
			}
			return m, nil
		case "enter", "p":
			if !m.showHelp && m.game != nil {
				if len(m.selectedCards) > 0 {
					m.handlePlay()
				} else {
					m.statusMessage = "Select cards first using number keys 1-7"
				}
			}
			return m, nil
		case "d":
			if !m.showHelp && m.game != nil {
				if len(m.selectedCards) > 0 {
					m.handleDiscard()
				} else {
					m.statusMessage = "Select cards first using number keys 1-7"
				}
			}
			return m, nil
		case "escape", "c":
			if !m.showHelp {
				m.selectedCards = []int{}
				m.statusMessage = "Selection cleared"
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
		return "Loading..."
	}

	// Top bar
	topBar := topBarStyle.
		Width(m.width).
		Render("🃏 Welcome to Balatno")

	// Bottom bar with current time and instructions
	timeStr := m.currentTime.Format("15:04:05")
	controls := "⏰ " + timeStr + " | 1-7: select cards, Enter/P: play, D: discard, C: clear, H: help, Q: quit"
	if m.statusMessage != "" {
		controls = "⏰ " + timeStr + " | " + m.statusMessage + " | H: help, Q: quit"
	}
	bottomBar := bottomBarStyle.
		Width(m.width).
		Render(controls)

	// Main content area
	contentHeight := m.height - 2 // Account for top and bottom bars

	var content string
	if m.showHelp {
		content = m.renderHelp()
	} else if m.game != nil {
		content = m.renderGameContent()
	} else {
		content = "Loading game..."
	}

	renderedContent := mainContentStyle.
		Width(m.width).
		Height(contentHeight).
		Render(content)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		renderedContent,
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
	if m.game == nil {
		return "Game not initialized"
	}

	// Game status info
	progress := float64(m.game.totalScore) / float64(m.game.currentTarget)
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
	switch m.game.currentBlind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	gameInfo := fmt.Sprintf("%s Ante %d - %s\n", blindEmoji, m.game.currentAnte, m.game.currentBlind) +
		fmt.Sprintf("🎯 Target: %d | Current Score: %d [%s] (%.1f%%)\n",
			m.game.currentTarget, m.game.totalScore, progressBar, progress*100) +
		fmt.Sprintf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d",
			MaxHands-m.game.handsPlayed, MaxDiscards-m.game.discardsUsed, m.game.money)

	gameInfoBox := gameInfoStyle.Render(gameInfo)

	// Render hand
	hand := m.renderHand()

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		Margin(1, 1).
		Render("🎮 Game controls will be added soon. Currently showing your starting hand of cards.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		gameInfoBox,
		hand,
		instructions,
	)
}

// renderHand renders the player's current hand of cards
func (m TUIModel) renderHand() string {
	if m.game == nil || len(m.game.playerCards) == 0 {
		return handStyle.Render("No cards in hand")
	}

	var content strings.Builder

	// Render selected cards first (if any)
	if len(m.selectedCards) > 0 {
		content.WriteString("🎯 Selected Cards:\n")
		var selectedViews []string
		for _, index := range m.selectedCards {
			if index >= 0 && index < len(m.game.playerCards) {
				card := m.game.playerCards[index]
				cardStr := m.renderCard(card, true)
				selectedViews = append(selectedViews, cardStr)
			}
		}
		selectedDisplay := lipgloss.JoinHorizontal(lipgloss.Top, selectedViews...)
		content.WriteString(selectedDisplay)
		content.WriteString("\n\n")
	}

	// Render all cards in hand
	content.WriteString("🃏 Your Hand (7 cards):\n")
	var cardViews []string
	for i, card := range m.game.playerCards {
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

	return handStyle.Render(content.String())
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
	if index < 0 || index >= len(m.game.playerCards) {
		return
	}

	// Check if already selected
	for i, selected := range m.selectedCards {
		if selected == index {
			// Remove from selection
			m.selectedCards = append(m.selectedCards[:i], m.selectedCards[i+1:]...)
			m.statusMessage = fmt.Sprintf("Card %d deselected", index+1)
			return
		}
	}

	// Add to selection (max 5 cards for poker)
	if len(m.selectedCards) >= 5 {
		m.statusMessage = "Maximum 5 cards can be selected"
		return
	}

	m.selectedCards = append(m.selectedCards, index)
	m.statusMessage = fmt.Sprintf("Card %d selected (%d/5)", index+1, len(m.selectedCards))
}

// handlePlay processes playing the selected cards
func (m *TUIModel) handlePlay() {
	if len(m.selectedCards) == 0 {
		m.statusMessage = "No cards selected to play"
		return
	}

	if m.game.handsPlayed >= MaxHands {
		m.statusMessage = "No hands remaining!"
		return
	}

	// Store previous state to check for changes
	prevScore := m.game.totalScore

	// Convert selected indices to string params for existing game logic
	var params []string
	for _, index := range m.selectedCards {
		// Convert 0-based TUI index to 1-based display index for game logic
		params = append(params, fmt.Sprintf("%d", index+1))
	}

	// Use existing game logic
	m.game.handlePlayAction(params)

	// Update display-to-original mapping after cards are removed
	m.game.displayToOriginal = make([]int, len(m.game.playerCards))
	for i := range m.game.playerCards {
		m.game.displayToOriginal[i] = i
	}

	// Create detailed status message
	scoreGained := m.game.totalScore - prevScore
	m.selectedCards = []int{}

	if m.game.totalScore >= m.game.currentTarget {
		m.statusMessage = fmt.Sprintf("🎉 BLIND DEFEATED! +%d points (%d total) - Advancing to next blind!", scoreGained, m.game.totalScore)
	} else if m.game.handsPlayed >= MaxHands {
		m.statusMessage = fmt.Sprintf("💀 Game Over! No hands left. Final score: %d/%d", m.game.totalScore, m.game.currentTarget)
	} else {
		handsLeft := MaxHands - m.game.handsPlayed
		m.statusMessage = fmt.Sprintf("✅ +%d points! Score: %d/%d | %d hands left", scoreGained, m.game.totalScore, m.game.currentTarget, handsLeft)
	}
}

// handleDiscard processes discarding the selected cards
func (m *TUIModel) handleDiscard() {
	if len(m.selectedCards) == 0 {
		m.statusMessage = "No cards selected to discard"
		return
	}

	if m.game.discardsUsed >= MaxDiscards {
		m.statusMessage = "No discards remaining!"
		return
	}

	numCards := len(m.selectedCards)

	// Convert selected indices to string params for existing game logic
	var params []string
	for _, index := range m.selectedCards {
		// Convert 0-based TUI index to 1-based display index for game logic
		params = append(params, fmt.Sprintf("%d", index+1))
	}

	// Use existing game logic
	m.game.handleDiscardAction(params)

	// Update display-to-original mapping after cards are removed
	m.game.displayToOriginal = make([]int, len(m.game.playerCards))
	for i := range m.game.playerCards {
		m.game.displayToOriginal[i] = i
	}

	// Create detailed status message
	discardsLeft := MaxDiscards - m.game.discardsUsed
	m.selectedCards = []int{}

	if discardsLeft > 0 {
		m.statusMessage = fmt.Sprintf("🗑️ Discarded %d cards, dealt new ones | %d discards left", numCards, discardsLeft)
	} else {
		m.statusMessage = fmt.Sprintf("🗑️ Discarded %d cards, dealt new ones | No discards remaining!", numCards)
	}
}

// RunTUI starts the TUI application
func RunTUI() error {
	game := NewGame()

	// Initialize display-to-original mapping for TUI (1:1 since we don't sort)
	game.displayToOriginal = make([]int, len(game.playerCards))
	for i := range game.playerCards {
		game.displayToOriginal[i] = i
	}

	m := TUIModel{
		currentTime:   time.Now(),
		game:          game,
		selectedCards: []int{},
		statusMessage: "Select cards with 1-7, play with Enter/P, discard with D",
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
