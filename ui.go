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
		screen: screen,
		game:   game,
	}, nil
}

func (ui *TerminalUI) Close() {
	ui.screen.Fini()
}

func (ui *TerminalUI) Run() {
	defer ui.Close()

	ui.message = "🃏 Welcome to Balatro CLI! 🃏"
	ui.render()

	for !ui.quit && ui.game.currentAnte <= MaxAntes {
		for ui.game.handsPlayed < MaxHands && ui.game.totalScore < ui.game.currentTarget {
			ui.render()
			if ui.handleInput() {
				return // quit requested
			}
		}

		if ui.quit {
			break
		}

		// Check if blind was completed
		if ui.game.totalScore >= ui.game.currentTarget {
			ui.game.handleBlindCompletion()
			ui.message = "✨ Blind completed! Moving to next challenge..."
		} else {
			// Failed to beat the blind
			ui.message = "💀 Game Over! Failed to beat the blind."
			ui.render()
			ui.waitForExit()
			break
		}
	}

	if !ui.quit && ui.game.currentAnte > MaxAntes {
		ui.message = "🎉 Congratulations! You completed all antes! 🎉"
		ui.render()
		ui.waitForExit()
	}
}

func (ui *TerminalUI) render() {
	ui.screen.Clear()
	width, height := ui.screen.Size()

	// Title section
	ui.drawText(0, 0, width, "🃏 BALATRO CLI 🃏", tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true))

	// Game status section
	statusY := 2
	ui.drawGameStatus(0, statusY, width)

	// Cards section
	cardsY := statusY + 8
	ui.drawCards(0, cardsY, width)

	// Input section
	inputY := height - 4
	ui.drawInputSection(0, inputY, width)

	// Message section
	messageY := height - 2
	ui.drawMessage(0, messageY, width)

	ui.screen.Show()
}

func (ui *TerminalUI) drawGameStatus(x, y, width int) {
	blind := ui.game.currentBlind
	blindName := blind.String()

	// Add blind emoji
	blindEmoji := ""
	switch blind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	// Create progress bar for score
	progress := float64(ui.game.totalScore) / float64(ui.game.currentTarget)
	if progress > 1.0 {
		progress = 1.0
	}
	progressWidth := 30
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

	lines := []string{
		fmt.Sprintf("%s Ante: %d/%d | Blind: %s", blindEmoji, ui.game.currentAnte, MaxAntes, blindName),
		fmt.Sprintf("🎯 Score: %d/%d %s (%.1f%%)", ui.game.totalScore, ui.game.currentTarget, progressBar, progress*100),
		fmt.Sprintf("💰 Money: $%d", ui.game.money),
		fmt.Sprintf("🎴 Hands Left: %d/%d | 🗑️ Discards Left: %d/%d", MaxHands-ui.game.handsPlayed, MaxHands, MaxDiscards-ui.game.discardsUsed, MaxDiscards),
	}

	// Add jokers info if any
	if len(ui.game.jokers) > 0 {
		jokerNames := make([]string, len(ui.game.jokers))
		for i, joker := range ui.game.jokers {
			jokerNames[i] = joker.Name
		}
		lines = append(lines, fmt.Sprintf("🃏 Jokers: %s", strings.Join(jokerNames, ", ")))
	}

	for i, line := range lines {
		ui.drawText(x, y+i, width, line, tcell.StyleDefault.Foreground(tcell.ColorBlue))
	}
}

func (ui *TerminalUI) drawCards(x, y, width int) {
	ui.drawText(x, y, width, "Your Cards:", tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true))

	if len(ui.game.playerCards) == 0 {
		ui.drawText(x, y+1, width, "No cards", tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}

	// Create display mapping
	ui.game.displayToOriginal = make(map[string]int)

	// Sort cards based on current sort mode
	var sortedCards []indexedCard
	for i, card := range ui.game.playerCards {
		sortedCards = append(sortedCards, indexedCard{card: card, index: i})
	}

	if ui.game.sortMode == SortByRank {
		// Sort by rank, then by suit
		for i := 0; i < len(sortedCards)-1; i++ {
			for j := i + 1; j < len(sortedCards); j++ {
				if sortedCards[i].card.Rank > sortedCards[j].card.Rank ||
					(sortedCards[i].card.Rank == sortedCards[j].card.Rank && sortedCards[i].card.Suit > sortedCards[j].card.Suit) {
					sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
				}
			}
		}
	} else {
		// Sort by suit, then by rank
		for i := 0; i < len(sortedCards)-1; i++ {
			for j := i + 1; j < len(sortedCards); j++ {
				if sortedCards[i].card.Suit > sortedCards[j].card.Suit ||
					(sortedCards[i].card.Suit == sortedCards[j].card.Suit && sortedCards[i].card.Rank > sortedCards[j].card.Rank) {
					sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
				}
			}
		}
	}

	// Display cards in rows
	cardsPerRow := (width - 2) / 6 // Each card takes about 6 characters
	if cardsPerRow < 1 {
		cardsPerRow = 1
	}

	for i, indexedCard := range sortedCards {
		displayIndex := rune('A' + i)
		if i >= 26 {
			displayIndex = rune('a' + (i - 26))
		}

		ui.game.displayToOriginal[string(displayIndex)] = indexedCard.index

		row := i / cardsPerRow
		col := i % cardsPerRow

		cardX := x + col*6
		cardY := y + 1 + row

		// Color cards by suit
		var style tcell.Style
		switch indexedCard.card.Suit {
		case Spades, Clubs:
			style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
		case Hearts, Diamonds:
			style = tcell.StyleDefault.Foreground(tcell.ColorRed)
		}

		cardStr := fmt.Sprintf("%c:%s", displayIndex, indexedCard.card.String())
		ui.drawText(cardX, cardY, width-cardX, cardStr, style)
	}

	// Show sort mode
	sortMode := "rank"
	if ui.game.sortMode == SortBySuit {
		sortMode = "suit"
	}
	ui.drawText(x, y+1+((len(sortedCards)-1)/cardsPerRow)+1, width,
		fmt.Sprintf("(Sorted by %s - use 'r' to toggle)", sortMode),
		tcell.StyleDefault.Foreground(tcell.ColorGray))
}

func (ui *TerminalUI) drawInputSection(x, y, width int) {
	var prompt string
	if ui.game.discardsUsed >= MaxDiscards {
		prompt = "Commands: (p)lay <cards>, (r)esort, (q)uit | Example: 'play A B C' or 'play 1 2 3'"
	} else {
		prompt = "Commands: (p)lay <cards>, (d)iscard <cards>, (r)esort, (q)uit | Example: 'play A B' or 'discard 1 2'"
	}

	ui.drawText(x, y, width, prompt, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	ui.drawText(x, y+1, width, "> "+ui.inputBuf+"_", tcell.StyleDefault.Foreground(tcell.ColorYellow))
}

func (ui *TerminalUI) drawMessage(x, y, width int) {
	if ui.message != "" {
		ui.drawText(x, y, width, ui.message, tcell.StyleDefault.Foreground(tcell.ColorPurple))
	}
}

func (ui *TerminalUI) drawText(x, y, maxWidth int, text string, style tcell.Style) {
	for i, r := range text {
		if i >= maxWidth {
			break
		}
		ui.screen.SetContent(x+i, y, r, nil, style)
	}
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
	if len(parts) < 1 {
		ui.message = "Please enter 'play <cards>' or 'discard <cards>'"
		return
	}

	action := strings.ToLower(parts[0])

	// Support abbreviated commands
	switch action {
	case "p":
		action = "play"
	case "d":
		action = "discard"
	case "r":
		action = "resort"
	case "q", "quit":
		ui.quit = true
		return
	}

	var params []string
	if len(parts) > 1 {
		params = parts[1:]
	}

	switch action {
	case "play":
		ui.handlePlayAction(params)
	case "discard":
		ui.handleDiscardAction(params)
	case "resort":
		ui.handleResortAction()
	default:
		ui.message = "Unknown command. Use (p)lay, (d)iscard, (r)esort, or (q)uit"
	}
}

func (ui *TerminalUI) handlePlayAction(params []string) {
	if len(params) == 0 {
		ui.message = "Please specify cards to play (e.g., 'play A B C' or 'play 1 2 3')"
		return
	}

	indices := ui.parseCardSelection(params)
	if len(indices) == 0 {
		ui.message = "Invalid card selection. Use letters (A,B,C) or numbers (1,2,3)"
		return
	}

	// Use the existing game logic
	oldScore := ui.game.totalScore
	oldMoney := ui.game.money

	ui.game.handlePlayAction(params)

	scoreGained := ui.game.totalScore - oldScore
	moneyGained := ui.game.money - oldMoney

	if scoreGained > 0 {
		ui.message = fmt.Sprintf("✨ Scored %d points! (+$%d)", scoreGained, moneyGained)
	} else {
		ui.message = "❌ Invalid hand - no points scored"
	}
}

func (ui *TerminalUI) handleDiscardAction(params []string) {
	if ui.game.discardsUsed >= MaxDiscards {
		ui.message = "❌ No discards remaining"
		return
	}

	if len(params) == 0 {
		ui.message = "Please specify cards to discard (e.g., 'discard A B' or 'discard 1 2')"
		return
	}

	indices := ui.parseCardSelection(params)
	if len(indices) == 0 {
		ui.message = "Invalid card selection. Use letters (A,B,C) or numbers (1,2,3)"
		return
	}

	if len(indices) > len(ui.game.playerCards) {
		ui.message = "Cannot discard more cards than you have"
		return
	}

	oldDiscards := ui.game.discardsUsed
	ui.game.handleDiscardAction(params)

	if ui.game.discardsUsed > oldDiscards {
		ui.message = fmt.Sprintf("🗑️ Discarded %d cards [%d/%d discards used]", len(indices), ui.game.discardsUsed, MaxDiscards)
	} else {
		ui.message = "❌ Failed to discard cards"
	}
}

func (ui *TerminalUI) handleResortAction() {
	ui.game.handleResortAction()
	if ui.game.sortMode == SortByRank {
		ui.message = "📊 Cards sorted by rank"
	} else {
		ui.message = "🃏 Cards sorted by suit"
	}
}

func (ui *TerminalUI) parseCardSelection(params []string) []int {
	var indices []int
	seen := make(map[int]bool) // Prevent duplicate selections

	for _, param := range params {
		var index int
		var found bool

		// Try character-based selection first (A, B, C, etc.)
		if len(param) == 1 {
			char := strings.ToUpper(param)
			if originalIndex, exists := ui.game.displayToOriginal[char]; exists {
				index = originalIndex
				found = true
			}
		}

		// Try numeric selection (1, 2, 3, etc.)
		if !found {
			if num, err := strconv.Atoi(param); err == nil && num >= 1 && num <= len(ui.game.playerCards) {
				index = num - 1
				found = true
			}
		}

		// Add to selection if valid and not already selected
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

// Helper type for card sorting
type indexedCard struct {
	card  Card
	index int
}
