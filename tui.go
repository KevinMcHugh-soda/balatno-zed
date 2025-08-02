package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI model holds the application state
type TUIModel struct {
	currentTime time.Time
	width       int
	height      int
	game        *Game
	showHelp    bool
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
	bottomBar := bottomBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("⏰ %s | 📖 Controls: 'q'=quit, 'h'=help | 🎮 Gameplay coming soon!", timeStr))

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

	title := "🃏 Your Hand (7 cards):"

	// Render cards as compact 2-character representations with position numbers
	var cardViews []string
	for i, card := range m.game.playerCards {
		cardStr := fmt.Sprintf("%s%s", card.Rank.String(), card.Suit.String())

		var styledCard string
		switch card.Suit {
		case Hearts:
			styledCard = heartsCardStyle.Render(cardStr)
		case Diamonds:
			styledCard = diamondsCardStyle.Render(cardStr)
		case Clubs:
			styledCard = clubsCardStyle.Render(cardStr)
		case Spades:
			styledCard = spadesCardStyle.Render(cardStr)
		}

		// Add position number below the card
		positionNum := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Margin(0, 1).
			Render(fmt.Sprintf("%d", i+1))

		// Combine card and position number vertically
		cardWithPos := lipgloss.JoinVertical(lipgloss.Center, styledCard, positionNum)
		cardViews = append(cardViews, cardWithPos)
	}

	// Display all cards in a single row since they're now compact
	cardsDisplay := lipgloss.JoinHorizontal(lipgloss.Top, cardViews...)

	handContent := fmt.Sprintf("%s\n\n%s", title, cardsDisplay)
	return handStyle.Render(handContent)
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

🎪 SHOP:
   • Buy jokers to enhance scoring
   • Reroll shop contents (costs money)
   • Skip shop to save money

⌨️  CONTROLS:
   • h - Toggle this help screen
   • q - Quit game

🚧 STATUS: TUI interface under development
   Full gameplay controls coming soon!`

	return helpStyle.Render(helpContent)
}

// RunTUI starts the TUI application
func RunTUI() error {
	game := NewGame()

	m := TUIModel{
		currentTime: time.Now(),
		game:        game,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
