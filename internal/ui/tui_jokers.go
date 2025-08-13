package ui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	game "balatno/internal/game"
)

// JokerOrderMode allows players to reorder owned jokers.
type JokerOrderMode struct {
	prevMode Mode
	selected int // -1 indicates no selection
}

// NewJokerOrderMode returns a JokerOrderMode wrapping the previous mode.
func NewJokerOrderMode(prev Mode) *JokerOrderMode {
	return &JokerOrderMode{prevMode: prev, selected: -1}
}

func (jm JokerOrderMode) renderContent(m TUIModel) string {
	if len(m.gameState.Jokers) == 0 {
		return gameInfoStyle.Render("No jokers to reorder")
	}
	var lines []string
	for i, j := range m.gameState.Jokers {
		line := fmt.Sprintf("%d. %s: %s", i+1, j.Name, j.Description)
		style := lipgloss.NewStyle()
		if jm.selected == i {
			style = style.Foreground(lipgloss.Color("226")).Bold(true)
		}
		lines = append(lines, style.Render(line))
	}
	header := "Reorder Jokers"
	content := lipgloss.JoinVertical(lipgloss.Left, append([]string{header}, lines...)...)
	return gameInfoStyle.Height(len(lines) + 2).Render(content)
}

func (jm *JokerOrderMode) handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd) {
	switch msg {
	case "esc", "enter":
		m.mode = jm.prevMode
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx, _ := strconv.Atoi(msg)
		if idx <= len(m.gameState.Jokers) {
			jm.selected = idx - 1
		} else {
			m.setStatusMessage(fmt.Sprintf("Invalid joker number: %s", msg))
		}
		return m, nil
	case "up", "k":
		if jm.selected == -1 {
			m.setStatusMessage("Select a joker first")
			return m, nil
		}
		if jm.selected > 0 {
			m.sendAction(game.PlayerActionMoveJoker, []string{strconv.Itoa(jm.selected + 1), "up"})
			m.gameState.Jokers[jm.selected-1], m.gameState.Jokers[jm.selected] = m.gameState.Jokers[jm.selected], m.gameState.Jokers[jm.selected-1]
			jm.selected--
		} else {
			m.setStatusMessage("Joker already at top")
		}
		return m, nil
	case "down", "j":
		if jm.selected == -1 {
			m.setStatusMessage("Select a joker first")
			return m, nil
		}
		if jm.selected < len(m.gameState.Jokers)-1 {
			m.sendAction(game.PlayerActionMoveJoker, []string{strconv.Itoa(jm.selected + 1), "down"})
			m.gameState.Jokers[jm.selected], m.gameState.Jokers[jm.selected+1] = m.gameState.Jokers[jm.selected+1], m.gameState.Jokers[jm.selected]
			jm.selected++
		} else {
			m.setStatusMessage("Joker already at bottom")
		}
		return m, nil
	}
	return m, nil
}

func (jm *JokerOrderMode) toggleHelp() Mode {
	return jm
}

func (jm *JokerOrderMode) getControls() string {
	return " | 1-9: select joker, ↑/k: move up, ↓/j: move down, Enter/Esc: back"
}
