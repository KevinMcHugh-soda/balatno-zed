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

type ShoppingMode struct {
	selectedItem      *int
	consecutiveEnters int
}

func (ms ShoppingMode) renderContent(m TUIModel) string {
	gameInfo := fmt.Sprintf("%s Ante %d - %s✅\n", "🏪", m.gameState.Ante, m.gameState.Blind) +
		fmt.Sprintf("🎴 Hands: %d | 🗑️ Discards: %d | 💰 Money: $%d | 🎲 Reroll: $%d",
			m.gameState.Hands, m.gameState.Discards, m.gameState.Money, m.shopInfo.RerollCost)

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

func (gm *ShoppingMode) handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd) {
	// Update last activity time on any key press
	m.lastActivity = time.Now()

	switch msg {
	case "1", "2", "3", "4", "5", "6", "7":
		// select a shop item
		i := int(msg[0] - '0')

		if i-1 >= len(m.shopInfo.Items) {
			m.setStatusMessage("That slot is empty!")
			return m, nil
		}
		item := m.shopInfo.Items[i-1]
		if item.Name == "" {
			m.setStatusMessage("That slot is empty!")
			return m, nil
		}
		if !item.CanAfford {
			m.setStatusMessage("Not enough money!")
			return m, nil
		}
		gm.selectedItem = &i
		gm.consecutiveEnters = 0
		return m, nil

	case "enter":
		if gm.selectedItem != nil {
			if *gm.selectedItem-1 >= len(m.shopInfo.Items) {
				m.setStatusMessage("That slot is empty!")
				return m, nil
			}
			item := m.shopInfo.Items[*gm.selectedItem-1]
			if item.Name == "" {
				m.setStatusMessage("That slot is empty!")
				return m, nil
			}
			if !item.CanAfford {
				m.setStatusMessage("Not enough money!")
				return m, nil
			}

			m.sendAction(game.PlayerActionBuy, []string{strconv.Itoa(*gm.selectedItem)})
			m.setStatusMessage(fmt.Sprintf("🛒 Purchased %s!", item.Name))
			gm.consecutiveEnters = 0
			return m, nil
		}

		if gm.consecutiveEnters == 0 {
			gm.consecutiveEnters++
			m.setStatusMessage("Press 'Enter' again to exit shop")
			return m, nil
		}

		gm.consecutiveEnters = 0
		m.sendAction(game.PlayerActionExitShop, nil)
		m.setStatusMessage("🚪 Exiting shop...")
		return m, nil

	case "r":
		// Reroll shop items
		m.sendAction(game.PlayerActionReroll, nil)
		m.setStatusMessage("🎲 Rerolling shop items...")
		gm.consecutiveEnters = 0
		return m, nil

	case "j":
		m.mode = NewJokerOrderMode(gm)
		return m, nil
	}
	gm.consecutiveEnters = 0
	return m, nil
}

func renderJoker(m TUIModel, joker game.ShopItemData) string {
	if joker.Name == "" {
		return ""
	}
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

func (gm ShoppingMode) getControls() string {
	// TODO I do think we'll need the game state to know how many shop items are available
	// but for now hardcode to 4
	return " | 1-4: select item, Enter (with selected): purchase, Enter twice (without selected): exit, C: clear, R: reroll, J: reorder jokers, H: help, ESC: exit, Q: quit"
}

type ShopHelpMode struct{}

// renderHelp renders the help screen
func (gm ShopHelpMode) renderContent(m TUIModel) string {
	return "you're in the shop, brother"
}

func (gm ShopHelpMode) toggleHelp() Mode {
	return &ShoppingMode{}
}

func (gm ShopHelpMode) handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd) {
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

func (gm ShopHelpMode) getControls() string {
	return " | Enter/Esc/H: exit help, Q: quit"
}
