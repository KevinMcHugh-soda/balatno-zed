package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	gameInfo := fmt.Sprintf("%s Ante %d - %s\n", blindEmoji, m.gameState.Ante, m.gameState.Blind) +
		fmt.Sprintf("🎯 Target: %d | Current Score: %d [%s] (%.1f%%)\n",
			m.gameState.Target, m.gameState.Score, progressBar, progress*100) +
		fmt.Sprintf("🎴 Hands Left: %d | 🗑️ Discards Left: %d | 💰 Money: $%d",
			m.gameState.Hands, m.gameState.Discards, m.gameState.Money)

	gameInfoBox := gameInfoStyle.
		Height(5).
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
func renderCard(m TUIModel, card Card, isInSelectedArea bool) string {
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

func (gm GameMode) toggleHelp() Mode {
	return &GameHelpMode{}
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
