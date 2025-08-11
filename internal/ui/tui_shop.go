package ui

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	game "balatno/internal/game"
)

type ShoppingMode struct {
	selectedItem      *int
	consecutiveEnters int
}

func (ms *ShoppingMode) renderContent(m *TUIModel) string {
	gameInfo := fmt.Sprintf("%s Ante %d - %s✅\n", "🏪", m.gameState.Ante, m.gameState.Blind) +
		fmt.Sprintf("🎴 Hands: %d | 🗑️ Discards: %d | 💰 Money: $%d | 🎲 Reroll: $%d",
			m.gameState.Hands, m.gameState.Discards, m.gameState.Money, m.shopInfo.RerollCost)
	gameInfoBox := gameInfoStyle.
		Height(5).
		Render(gameInfo)

	var jokerViews []string

	for idx, joker := range m.shopInfo.Items {
		jokerStr := renderJoker(*m, joker)

		posNumStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

		if ms.selectedItem != nil && *ms.selectedItem == idx+1 {
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

		item := m.shopInfo.Items[i-1]
		if !item.CanAfford {
			m.setStatusMessage("Not enough money!")
			return m, nil
		}
		gm.selectedItem = &i
		return m, nil

	case "enter":
		if gm.selectedItem != nil {
			idx := *gm.selectedItem - 1
			item := m.shopInfo.Items[idx]
			if !item.CanAfford {
				m.setStatusMessage("Not enough money!")
				return m, nil
			}

			if m.actionRequestPending != nil {
				responseChan := m.actionRequestPending.ResponseChan
				m.actionRequestPending = nil

				go func() {
					responseChan <- PlayerActionResponse{
						Action: game.PlayerActionBuy,
						Params: []string{strconv.Itoa(*gm.selectedItem)},
						Quit:   false,
					}
				}()
			}

			m.setStatusMessage(fmt.Sprintf("🛒 Purchased %s!", item.Name))
			gm.selectedItem = nil
			return m, nil
		}

		if m.actionRequestPending != nil {
			// Capture response channel before clearing the pending request
			responseChan := m.actionRequestPending.ResponseChan
			m.actionRequestPending = nil

			// Send exit shop response
			go func() {
				responseChan <- PlayerActionResponse{
					Action: game.PlayerActionExitShop,
					Params: nil,
					Quit:   false,
				}
			}()

			m.setStatusMessage("🚪 Exiting shop...")
		}
		return m, nil
	case "r":
		// Reroll shop items
		if m.actionRequestPending != nil {
			// Capture response channel before clearing the pending request
			responseChan := m.actionRequestPending.ResponseChan
			m.actionRequestPending = nil

			// Send reroll response
			go func() {
				responseChan <- PlayerActionResponse{
					Action: game.PlayerActionReroll,
					Params: nil,
					Quit:   false,
				}
			}()

			m.setStatusMessage("🎲 Rerolling shop items...")
		}
		return m, nil
	}
	return m, nil
}

func renderJoker(m TUIModel, joker game.ShopItemData) string {
	cost := fmt.Sprintf("%d", joker.Cost)
	if joker.Cost > m.gameState.Money {
		cost = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(cost)
	}
	jokerStr := fmt.Sprintf("%s ($%s): %s\n", joker.Name, cost, joker.Description)

	return jokerStr
}

func (gm *ShoppingMode) toggleHelp() Mode {
	return &ShopHelpMode{}
}

func (gm *ShoppingMode) getControls() string {
	// TODO I do think we'll need the game state to know how many shop items are available
	// but for now hardcode to 4
	return " | 1-4: select item, Enter (with selected): purchase, Enter (without selected): exit, C: clear, R: reroll, H: help, ESC: exit, Q: quit"
}

type ShopHelpMode struct{}

// renderHelp renders the help screen
func (gm ShopHelpMode) renderContent(m *TUIModel) string {
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
