package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	game "balatno/internal/game"
)

type GameMode struct {
}

// renderContent renders the main game area
func (gm GameMode) renderContent(m TUIModel) string {
	// Game status info - fixed height section
	progress := float64(m.gameState.Score) / float64(m.gameState.Target)
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
	switch m.gameState.Blind {
	case game.SmallBlind:
		blindEmoji = "🔸"
	case game.BigBlind:
		blindEmoji = "🔶"
	case game.BossBlind:
		blindEmoji = "💀"
	}

	gameInfo := fmt.Sprintf("%s Ante %d - %s\n", blindEmoji, m.gameState.Ante, m.gameState.Blind) +
		fmt.Sprintf("🎯 Target: %d | Current Score: %d [%s] (%.1f%%)\n",
			m.gameState.Target, m.gameState.Score, progressBar, progress*100) +
		fmt.Sprintf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d",
			m.gameState.Hands, m.gameState.Discards, m.gameState.Money)

	// Add joker information
	var jokerLines []string
	if len(m.gameState.Jokers) == 0 {
		jokerLines = append(jokerLines, "🃏 Jokers: None")
	} else {
		jokerLines = append(jokerLines, "🃏 Jokers:")
		for _, joker := range m.gameState.Jokers {
			jokerLines = append(jokerLines, renderOwnedJoker(joker))
		}
	}
	gameInfo += "\n" + strings.Join(jokerLines, "\n")

	infoHeight := 3 + len(jokerLines)
	if infoHeight < 5 {
		infoHeight = 5
	}

	gameInfoBox := gameInfoStyle.
		Height(infoHeight).
		Render(gameInfo)

	// Render hand - fixed height section
	hand := renderHand(m)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		gameInfoBox,
		hand,
	)
}

// renderCard renders a single card with appropriate styling
func renderCard(m TUIModel, card game.Card, isInSelectedArea bool) string {
	cardStr := fmt.Sprintf("%s%s", card.Rank.String(), card.Suit.String())

	var style lipgloss.Style
	switch card.Suit {
	case game.Hearts:
		style = heartsCardStyle
	case game.Diamonds:
		style = diamondsCardStyle
	case game.Clubs:
		style = clubsCardStyle
	case game.Spades:
		style = spadesCardStyle
	}

	if isInSelectedArea {
		style = style.Bold(true).Background(lipgloss.Color("235"))
	}

	return style.Render(cardStr)
}

// renderHand renders the player's current hand of cards
func renderHand(m TUIModel) string {
	if len(m.cards) == 0 {
		return handStyle.Height(10).Render("No cards in hand")
	}

	var content strings.Builder

	// Hand cards area - fixed position
	content.WriteString(fmt.Sprintf("🃏 Your Hand (%d cards):\n", len(m.cards)))
	var cardViews []string
	for i, card := range m.cards {
		isSelected := m.isCardSelected(i)
		cardStr := renderCard(m, card, false)

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

// renderOwnedJoker renders a joker the player currently owns
func renderOwnedJoker(joker game.Joker) string {
	return fmt.Sprintf("%s: %s", joker.Name, joker.Description)
}

func (gm GameMode) handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd) {
	// Update last activity time on any key press
	m.lastActivity = time.Now()

	switch msg {
	// there's a bug when you discard and then attempt to play, it won't submit.
	case "1", "2", "3", "4", "5", "6", "7":
		cardIndex, _ := strconv.Atoi(msg)
		if cardIndex <= len(m.cards) {
			// I'd like to move card selection into the mode
			m.toggleCardSelection(cardIndex - 1) // Convert to 0-based
		} else {
			m.setStatusMessage(fmt.Sprintf("Invalid card number: %d (only have %d cards)", cardIndex, len(m.cards)))
		}
		return m, nil

	case "enter", "p":
		if len(m.selectedCards) > 0 {
			m.handlePlay()
		} else {
			m.setStatusMessage("Select cards first using number keys 1-7")
		}
		return m, nil

	case "d":
		if len(m.selectedCards) > 0 {
			m.handleDiscard()
		} else {
			m.setStatusMessage("Select cards first using number keys 1-7")
		}
		return m, nil

	case "r":
		m.handleResort()
		return m, nil

	case "escape", "c":
		m.selectedCards = []int{}
		m.setStatusMessage("Selection cleared")
		return m, nil
	}

	return m, nil
}

func (gm GameMode) toggleHelp() Mode {
	return &GameHelpMode{}
}

func (gm GameMode) getControls() string {
	return " | 1-7: select cards, Enter/P: play, D: discard, C: clear, R: resort, H: help, Q: quit"
}

type GameHelpMode struct{}

// renderHelp renders the help screen
func (gm GameHelpMode) renderContent(m TUIModel) string {
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

func (gm GameHelpMode) toggleHelp() Mode {
	return &GameMode{}
}

func (gm GameHelpMode) handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd) {
	// Update last activity time on any key press
	m.lastActivity = time.Now()
	// fmt.Println(msg)
	m.setStatusMessage(msg)
	if msg == "esc" || msg == "escape" || msg == "enter" {
		m.showHelp = !m.showHelp
		m.mode = gm.toggleHelp()
	}

	return m, nil
}

func (gm GameHelpMode) getControls() string {
	return " | Enter/Esc/H: exit help, Q: quit"
}
