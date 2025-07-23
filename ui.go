package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type TerminalUI struct {
	screen   tcell.Screen
	game     *Game
	inputBuf string
	message  string
	quit     bool
}

// Fixed layout positions - NEVER change these
const (
	titleRow       = 0
	anteRow        = 2
	scoreRow       = 3
	moneyRow       = 4
	resourcesRow   = 5
	jokersRow      = 6
	cardsHeaderRow = 8
	cardsRow1      = 9
	cardsRow2      = 10
	cardsRow3      = 11

	// From bottom of screen
	commandHelpOffset = -4
	inputPromptOffset = -2
	messageOffset     = -1
)

func NewTerminalUI(game *Game) (*TerminalUI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	screen.Clear()

	return &TerminalUI{
		screen:  screen,
		game:    game,
		message: "Welcome to Balatro CLI!",
	}, nil
}

func (ui *TerminalUI) Close() {
	ui.screen.Fini()
}

func (ui *TerminalUI) Run() {
	defer ui.Close()

	for !ui.quit && ui.game.currentAnte <= MaxAntes {
		for ui.game.handsPlayed < MaxHands && ui.game.totalScore < ui.game.currentTarget {
			ui.render()
			if ui.handleInput() {
				return
			}
		}

		if ui.quit {
			break
		}

		if ui.game.totalScore >= ui.game.currentTarget {
			ui.game.handleBlindCompletion()
			ui.message = "Blind completed! Moving to next challenge..."
		} else {
			ui.message = "Game Over! Failed to beat the blind."
			ui.render()
			ui.waitForExit()
			break
		}
	}

	if !ui.quit && ui.game.currentAnte > MaxAntes {
		ui.message = "Congratulations! You completed all antes!"
		ui.render()
		ui.waitForExit()
	}
}

func (ui *TerminalUI) render() {
	width, height := ui.screen.Size()

	// Clear entire screen first
	ui.screen.Clear()

	// Render each section in its fixed location
	ui.renderTitle(width)
	ui.renderAnteBlind(width)
	ui.renderScore(width)
	ui.renderMoney(width)
	ui.renderResources(width)
	ui.renderJokers(width)
	ui.renderCards(width)
	ui.renderCommandHelp(width, height)
	ui.renderInputPrompt(width, height)
	ui.renderMessage(width, height)

	ui.screen.Show()
}

func (ui *TerminalUI) renderTitle(width int) {
	text := "🃏 BALATRO CLI 🃏"
	ui.writeTextAtRow(titleRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true))
}

func (ui *TerminalUI) renderAnteBlind(width int) {
	blind := ui.game.currentBlind
	blindEmoji := "🔸"
	switch blind {
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	text := fmt.Sprintf("%s Ante: %d/%d | Blind: %s",
		blindEmoji, ui.game.currentAnte, MaxAntes, blind.String())
	ui.writeTextAtRow(anteRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorBlue))
}

func (ui *TerminalUI) renderScore(width int) {
	progress := float64(ui.game.totalScore) / float64(ui.game.currentTarget)
	if progress > 1.0 {
		progress = 1.0
	}

	progressWidth := 20
	filled := int(progress * float64(progressWidth))

	progressBar := "["
	for i := 0; i < progressWidth; i++ {
		if i < filled {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}
	progressBar += "]"

	text := fmt.Sprintf("🎯 Score: %d/%d %s (%.1f%%)",
		ui.game.totalScore, ui.game.currentTarget, progressBar, progress*100)
	ui.writeTextAtRow(scoreRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorGreen))
}

func (ui *TerminalUI) renderMoney(width int) {
	text := fmt.Sprintf("💰 Money: $%d", ui.game.money)
	ui.writeTextAtRow(moneyRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorYellow))
}

func (ui *TerminalUI) renderResources(width int) {
	text := fmt.Sprintf("🎴 Hands Left: %d/%d | 🗑️ Discards Left: %d/%d",
		MaxHands-ui.game.handsPlayed, MaxHands,
		MaxDiscards-ui.game.discardsUsed, MaxDiscards)
	ui.writeTextAtRow(resourcesRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorBlue))
}

func (ui *TerminalUI) renderJokers(width int) {
	text := "🃏 Jokers: "
	if len(ui.game.jokers) > 0 {
		jokerNames := make([]string, len(ui.game.jokers))
		for i, joker := range ui.game.jokers {
			jokerNames[i] = joker.Name
		}
		text += strings.Join(jokerNames, ", ")
	} else {
		text += "None"
	}
	ui.writeTextAtRow(jokersRow, text, width, tcell.StyleDefault.Foreground(tcell.ColorRed))
}

func (ui *TerminalUI) renderCards(width int) {
	// Clear cards area (header + 3 rows)
	ui.clearRow(cardsHeaderRow, width)
	ui.clearRow(cardsRow1, width)
	ui.clearRow(cardsRow2, width)
	ui.clearRow(cardsRow3, width)

	sortMode := "rank"
	if ui.game.sortMode == SortBySuit {
		sortMode = "suit"
	}

	headerText := fmt.Sprintf("Your Cards (sorted by %s):", sortMode)
	ui.writeTextAtRow(cardsHeaderRow, headerText, width, tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true))

	if len(ui.game.playerCards) == 0 {
		ui.writeTextAtRow(cardsRow1, "No cards", width, tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}

	// Sort and display cards
	ui.game.displayToOriginal = make(map[string]int)
	sortedCards := ui.getSortedCards()

	cardsPerRow := 8
	cardTexts := []string{"", "", ""}

	for i, indexedCard := range sortedCards {
		displayIndex := rune('A' + i)
		if i >= 26 {
			displayIndex = rune('a' + (i - 26))
		}

		ui.game.displayToOriginal[string(displayIndex)] = indexedCard.index

		cardStr := fmt.Sprintf("%c:%s", displayIndex, indexedCard.card.String())

		row := i / cardsPerRow
		if row < 3 {
			if cardTexts[row] != "" {
				cardTexts[row] += "  "
			}
			cardTexts[row] += cardStr
		}
	}

	// Write card rows
	if cardTexts[0] != "" {
		ui.writeTextAtRow(cardsRow1, cardTexts[0], width, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}
	if cardTexts[1] != "" {
		ui.writeTextAtRow(cardsRow2, cardTexts[1], width, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}
	if cardTexts[2] != "" {
		ui.writeTextAtRow(cardsRow3, cardTexts[2], width, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}
}

func (ui *TerminalUI) renderCommandHelp(width, height int) {
	row := height + commandHelpOffset
	ui.clearRow(row, width)

	var text string
	if ui.game.discardsUsed >= MaxDiscards {
		text = "Commands: (p)lay <cards>, (r)esort, (q)uit | Example: 'play A B C'"
	} else {
		text = "Commands: (p)lay <cards>, (d)iscard <cards>, (r)esort, (q)uit | Example: 'play A B'"
	}
	ui.writeTextAtRow(row, text, width, tcell.StyleDefault.Foreground(tcell.ColorGray))
}

func (ui *TerminalUI) renderInputPrompt(width, height int) {
	row := height + inputPromptOffset
	ui.clearRow(row, width)

	text := "> " + ui.inputBuf + "_"
	ui.writeTextAtRow(row, text, width, tcell.StyleDefault.Foreground(tcell.ColorYellow))
}

func (ui *TerminalUI) renderMessage(width, height int) {
	row := height + messageOffset
	ui.clearRow(row, width)

	if ui.message != "" {
		ui.writeTextAtRow(row, ui.message, width, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	}
}

func (ui *TerminalUI) writeTextAtRow(row int, text string, maxWidth int, style tcell.Style) {
	// Truncate text if too long
	if len(text) > maxWidth-1 {
		text = text[:maxWidth-1]
	}

	for i, r := range text {
		if i >= maxWidth {
			break
		}
		ui.screen.SetContent(i, row, r, nil, style)
	}
}

func (ui *TerminalUI) clearRow(row, width int) {
	for x := 0; x < width; x++ {
		ui.screen.SetContent(x, row, ' ', nil, tcell.StyleDefault)
	}
}

func (ui *TerminalUI) getSortedCards() []indexedCard {
	var sortedCards []indexedCard
	for i, card := range ui.game.playerCards {
		sortedCards = append(sortedCards, indexedCard{card: card, index: i})
	}

	if ui.game.sortMode == SortByRank {
		for i := 0; i < len(sortedCards)-1; i++ {
			for j := i + 1; j < len(sortedCards); j++ {
				if sortedCards[i].card.Rank > sortedCards[j].card.Rank ||
					(sortedCards[i].card.Rank == sortedCards[j].card.Rank && sortedCards[i].card.Suit > sortedCards[j].card.Suit) {
					sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
				}
			}
		}
	} else {
		for i := 0; i < len(sortedCards)-1; i++ {
			for j := i + 1; j < len(sortedCards); j++ {
				if sortedCards[i].card.Suit > sortedCards[j].card.Suit ||
					(sortedCards[i].card.Suit == sortedCards[j].card.Suit && sortedCards[i].card.Rank > sortedCards[j].card.Rank) {
					sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
				}
			}
		}
	}

	return sortedCards
}

func (ui *TerminalUI) handleInput() bool {
	ev := ui.screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			ui.quit = true
			return true
		case tcell.KeyEnter:
			ui.processCommand()
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(ui.inputBuf) > 0 {
				ui.inputBuf = ui.inputBuf[:len(ui.inputBuf)-1]
			}
		case tcell.KeyRune:
			if unicode.IsPrint(ev.Rune()) {
				ui.inputBuf += string(ev.Rune())
			}
		}
	case *tcell.EventResize:
		ui.screen.Sync()
	}
	return false
}

func (ui *TerminalUI) processCommand() {
	input := strings.TrimSpace(ui.inputBuf)
	ui.inputBuf = ""

	if input == "" {
		ui.message = "Please enter a command"
		return
	}

	parts := strings.Fields(input)
	action := strings.ToLower(parts[0])

	switch action {
	case "p", "play":
		ui.handlePlayAction(parts[1:])
	case "d", "discard":
		ui.handleDiscardAction(parts[1:])
	case "r", "resort":
		ui.handleResortAction()
	case "q", "quit":
		ui.quit = true
	default:
		ui.message = "Unknown command. Use (p)lay, (d)iscard, (r)esort, or (q)uit"
	}
}

func (ui *TerminalUI) handlePlayAction(params []string) {
	if len(params) == 0 {
		ui.message = "Please specify cards to play (e.g., 'play A B C')"
		return
	}

	indices := ui.parseCardSelection(params)
	if len(indices) == 0 {
		ui.message = "Invalid card selection. Use letters (A,B,C) or numbers (1,2,3)"
		return
	}

	oldScore := ui.game.totalScore
	oldMoney := ui.game.money

	ui.game.handlePlayAction(params)

	scoreGained := ui.game.totalScore - oldScore
	moneyGained := ui.game.money - oldMoney

	if scoreGained > 0 {
		ui.message = fmt.Sprintf("Scored %d points! (+$%d)", scoreGained, moneyGained)
	} else {
		ui.message = "Invalid hand - no points scored"
	}
}

func (ui *TerminalUI) handleDiscardAction(params []string) {
	if ui.game.discardsUsed >= MaxDiscards {
		ui.message = "No discards remaining"
		return
	}

	if len(params) == 0 {
		ui.message = "Please specify cards to discard (e.g., 'discard A B')"
		return
	}

	indices := ui.parseCardSelection(params)
	if len(indices) == 0 {
		ui.message = "Invalid card selection. Use letters (A,B,C) or numbers (1,2,3)"
		return
	}

	oldDiscards := ui.game.discardsUsed
	ui.game.handleDiscardAction(params)

	if ui.game.discardsUsed > oldDiscards {
		ui.message = fmt.Sprintf("Discarded %d cards", len(indices))
	} else {
		ui.message = "Failed to discard cards"
	}
}

func (ui *TerminalUI) handleResortAction() {
	ui.game.handleResortAction()
	if ui.game.sortMode == SortByRank {
		ui.message = "Cards sorted by rank"
	} else {
		ui.message = "Cards sorted by suit"
	}
}

func (ui *TerminalUI) parseCardSelection(params []string) []int {
	var indices []int
	seen := make(map[int]bool)

	for _, param := range params {
		var index int
		var found bool

		if len(param) == 1 {
			char := strings.ToUpper(param)
			if originalIndex, exists := ui.game.displayToOriginal[char]; exists {
				index = originalIndex
				found = true
			}
		}

		if !found {
			if num, err := strconv.Atoi(param); err == nil && num >= 1 && num <= len(ui.game.playerCards) {
				index = num - 1
				found = true
			}
		}

		if found && !seen[index] {
			indices = append(indices, index)
			seen[index] = true
		}
	}
	return indices
}

func (ui *TerminalUI) waitForExit() {
	ui.message += " Press any key to exit..."
	ui.render()
	ui.screen.PollEvent()
}

type indexedCard struct {
	card  Card
	index int
}
