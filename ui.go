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

// Fixed layout positions
const (
	// Header section
	titleRow    = 0
	gameInfoRow = 2
	scoreRow    = 3
	moneyRow    = 4
	handsRow    = 5
	jokersRow   = 6

	// Cards section
	cardsHeaderRow = 8
	cardsStartRow  = 9
	cardsEndRow    = 12

	// Input section (from bottom)
	messageRow     = -2 // 2 rows from bottom
	inputPromptRow = -4 // 4 rows from bottom
	commandRow     = -5 // 5 rows from bottom
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
		message: "🃏 Welcome to Balatro CLI! 🃏",
	}, nil
}

func (ui *TerminalUI) Close() {
	ui.screen.Fini()
}

func (ui *TerminalUI) Run() {
	defer ui.Close()

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
	_, height := ui.screen.Size()

	// Clear screen
	ui.screen.Clear()

	// Render all static sections
	ui.renderTitle()
	ui.renderGameInfo()
	ui.renderCards()
	ui.renderInput(height)
	ui.renderMessage(height)

	ui.screen.Show()
}

func (ui *TerminalUI) renderTitle() {
	ui.drawTextAt(0, titleRow, "🃏 BALATRO CLI 🃏",
		tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true))
}

func (ui *TerminalUI) renderGameInfo() {
	// Get blind info with emoji
	blind := ui.game.currentBlind
	blindName := blind.String()
	blindEmoji := ""
	switch blind {
	case SmallBlind:
		blindEmoji = "🔸"
	case BigBlind:
		blindEmoji = "🔶"
	case BossBlind:
		blindEmoji = "💀"
	}

	// Ante and Blind info
	ui.drawTextAt(0, gameInfoRow,
		fmt.Sprintf("%s Ante: %d/%d | Blind: %s", blindEmoji, ui.game.currentAnte, MaxAntes, blindName),
		tcell.StyleDefault.Foreground(tcell.ColorBlue))

	// Score with progress bar
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

	ui.drawTextAt(0, scoreRow,
		fmt.Sprintf("🎯 Score: %d/%d %s (%.1f%%)",
			ui.game.totalScore, ui.game.currentTarget, progressBar, progress*100),
		tcell.StyleDefault.Foreground(tcell.ColorGreen))

	// Money
	ui.drawTextAt(0, moneyRow,
		fmt.Sprintf("💰 Money: $%d", ui.game.money),
		tcell.StyleDefault.Foreground(tcell.ColorYellow))

	// Hands and Discards
	ui.drawTextAt(0, handsRow,
		fmt.Sprintf("🎴 Hands Left: %d/%d | 🗑️ Discards Left: %d/%d",
			MaxHands-ui.game.handsPlayed, MaxHands,
			MaxDiscards-ui.game.discardsUsed, MaxDiscards),
		tcell.StyleDefault.Foreground(tcell.ColorBlue))

	// Jokers
	jokersText := "🃏 Jokers: "
	if len(ui.game.jokers) > 0 {
		jokerNames := make([]string, len(ui.game.jokers))
		for i, joker := range ui.game.jokers {
			jokerNames[i] = joker.Name
		}
		jokersText += strings.Join(jokerNames, ", ")
	} else {
		jokersText += "None"
	}
	ui.drawTextAt(0, jokersRow, jokersText, tcell.StyleDefault.Foreground(tcell.ColorRed))
}

func (ui *TerminalUI) renderCards() {
	// Clear cards area first
	for row := cardsHeaderRow; row <= cardsEndRow; row++ {
		ui.clearRow(row)
	}

	// Cards header
	sortMode := "rank"
	if ui.game.sortMode == SortBySuit {
		sortMode = "suit"
	}
	ui.drawTextAt(0, cardsHeaderRow,
		fmt.Sprintf("Your Cards (sorted by %s):", sortMode),
		tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true))

	if len(ui.game.playerCards) == 0 {
		ui.drawTextAt(0, cardsStartRow, "No cards", tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}

	// Create display mapping and sort cards
	ui.game.displayToOriginal = make(map[string]int)
	var sortedCards []indexedCard
	for i, card := range ui.game.playerCards {
		sortedCards = append(sortedCards, indexedCard{card: card, index: i})
	}

	// Sort based on current sort mode
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

	// Display cards in a fixed layout
	cardsPerRow := 8
	currentRow := cardsStartRow

	cardText := ""
	for i, indexedCard := range sortedCards {
		// Use letter for display index
		displayIndex := rune('A' + i)
		if i >= 26 {
			displayIndex = rune('a' + (i - 26))
		}

		ui.game.displayToOriginal[string(displayIndex)] = indexedCard.index

		cardStr := fmt.Sprintf("%c:%s", displayIndex, indexedCard.card.String())

		// Add spacing between cards
		if i > 0 && i%cardsPerRow == 0 {
			// Start new row
			ui.drawTextAt(0, currentRow, cardText, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			currentRow++
			cardText = ""
		}

		if cardText != "" {
			cardText += "  "
		}
		cardText += cardStr
	}

	// Draw remaining cards
	if cardText != "" {
		ui.drawTextAt(0, currentRow, cardText, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}
}

func (ui *TerminalUI) renderInput(screenHeight int) {
	commandRow := screenHeight + commandRow
	promptRow := screenHeight + inputPromptRow

	// Clear input area
	ui.clearRow(commandRow)
	ui.clearRow(promptRow)

	// Command help
	var commandText string
	if ui.game.discardsUsed >= MaxDiscards {
		commandText = "Commands: (p)lay <cards>, (r)esort, (q)uit | Example: 'play A B C'"
	} else {
		commandText = "Commands: (p)lay <cards>, (d)iscard <cards>, (r)esort, (q)uit | Example: 'play A B'"
	}
	ui.drawTextAt(0, commandRow, commandText, tcell.StyleDefault.Foreground(tcell.ColorGray))

	// Input prompt
	ui.drawTextAt(0, promptRow, "> "+ui.inputBuf+"_", tcell.StyleDefault.Foreground(tcell.ColorYellow))
}

func (ui *TerminalUI) renderMessage(screenHeight int) {
	msgRow := screenHeight + messageRow
	ui.clearRow(msgRow)

	if ui.message != "" {
		ui.drawTextAt(0, msgRow, ui.message, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	}
}

func (ui *TerminalUI) drawTextAt(x, y int, text string, style tcell.Style) {
	for i, r := range text {
		ui.screen.SetContent(x+i, y, r, nil, style)
	}
}

func (ui *TerminalUI) clearRow(y int) {
	width, _ := ui.screen.Size()
	for x := 0; x < width; x++ {
		ui.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
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
		ui.message = "Please enter a valid command"
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
		ui.message = fmt.Sprintf("🗑️ Discarded %d cards", len(indices))
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
	seen := make(map[int]bool)

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
