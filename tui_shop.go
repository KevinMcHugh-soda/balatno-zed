package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type ShoppingMode struct{}

func (ms ShoppingMode) renderContent(m TUIModel) string {
	gameInfo := fmt.Sprintf("%s Ante %d - %s✅\n", "🏪", m.gameState.Ante, m.gameState.Blind) +
		fmt.Sprintf("🎴 Hands: %d | 🗑️ Discards: %d | 💰 Money: $%d",
			m.gameState.Hands, m.gameState.Discards, m.gameState.Money)
	gameInfoBox := gameInfoStyle.
		Height(5).
		Render(gameInfo)

	var jokerViews []string

	for idx, joker := range m.shopInfo.Items {
		jokerStr := renderJoker(m, joker)

		posNumStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

		if m.isCardSelected(idx) {
			posNumStyle = posNumStyle.Foreground(lipgloss.Color("226")).Bold(true)
		}

		positionNum := posNumStyle.Render(fmt.Sprintf("%d", idx+1))

		jokerWithPos := lipgloss.JoinHorizontal(lipgloss.Left, positionNum, jokerStr)
		jokerViews = append(jokerViews, jokerWithPos)
	}

	jokerDisplay := gameInfoStyle.Height(len(jokerViews)).Render(lipgloss.JoinVertical(lipgloss.Top, jokerViews...))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		gameInfoBox,
		jokerDisplay,
	)
}

func renderJoker(m TUIModel, joker ShopItemData) string {
	// TODO - style the cost in red if we can't afford it

	cost := fmt.Sprintf("%d", joker.Cost)
	if joker.Cost > m.gameState.Money {
		cost = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(cost)
	}
	jokerStr := fmt.Sprintf("%s ($%s): %s\n", joker.Name, cost, joker.Description)

	return jokerStr
}

func (gm ShoppingMode) toggleHelp() Mode {
	return &ShopHelpMode{}
}

type ShopHelpMode struct{}

// renderHelp renders the help screen
func (gm ShopHelpMode) renderContent(m TUIModel) string {
	return "you're in the shop, brother"
}

func (gm ShopHelpMode) toggleHelp() Mode {
	return &ShoppingMode{}
}
