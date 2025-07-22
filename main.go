package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Global random source for consistent seeding
var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// SetSeed allows setting a specific seed for deterministic behavior (useful for testing)
func SetSeed(seed int64) {
	rng = rand.New(rand.NewSource(seed))
}

type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

func (s Suit) String() string {
	switch s {
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	case Spades:
		return "♠"
	default:
		return "?"
	}
}

type Rank int

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return strconv.Itoa(int(r))
	}
}

func (r Rank) Value() int {
	switch r {
	case Ace:
		return 11
	case Jack, Queen, King:
		return 10
	default:
		return int(r)
	}
}

type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

type HandType int

const (
	HighCard HandType = iota
	Pair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (ht HandType) String() string {
	switch ht {
	case HighCard:
		return "High Card"
	case Pair:
		return "Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

func (ht HandType) BaseScore() int {
	switch ht {
	case HighCard:
		return 5
	case Pair:
		return 10
	case TwoPair:
		return 20
	case ThreeOfAKind:
		return 30
	case Straight:
		return 30
	case Flush:
		return 35
	case FullHouse:
		return 40
	case FourOfAKind:
		return 60
	case StraightFlush:
		return 100
	case RoyalFlush:
		return 100
	default:
		return 0
	}
}

func (ht HandType) Mult() int {
	switch ht {
	case HighCard:
		return 1
	case Pair:
		return 2
	case TwoPair:
		return 2
	case ThreeOfAKind:
		return 3
	case Straight:
		return 4
	case Flush:
		return 4
	case FullHouse:
		return 4
	case FourOfAKind:
		return 7
	case StraightFlush:
		return 8
	case RoyalFlush:
		return 8
	default:
		return 1
	}
}

type Hand struct {
	Cards []Card
}

func (h Hand) String() string {
	var cards []string
	for _, card := range h.Cards {
		cards = append(cards, card.String())
	}
	return strings.Join(cards, " ")
}

func NewDeck() []Card {
	var deck []Card
	for suit := Hearts; suit <= Spades; suit++ {
		for rank := Ace; rank <= King; rank++ {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

func ShuffleDeck(deck []Card) {
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func EvaluateHand(hand Hand) (HandType, int, int, int) {
	if len(hand.Cards) == 0 {
		return HighCard, 0, 0, 0
	}

	// Sort cards by rank for easier evaluation
	cards := make([]Card, len(hand.Cards))
	copy(cards, hand.Cards)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Rank < cards[j].Rank
	})

	// Count ranks and suits
	rankCounts := make(map[Rank]int)
	suitCounts := make(map[Suit]int)

	for _, card := range cards {
		rankCounts[card.Rank]++
		suitCounts[card.Suit]++
	}

	// Check for flush
	isFlush := len(suitCounts) == 1 && len(cards) == 5

	// Check for straight
	isStraight := false
	if len(cards) == 5 {
		isStraight = true
		for i := 1; i < len(cards); i++ {
			if cards[i].Rank != cards[i-1].Rank+1 {
				isStraight = false
				break
			}
		}

		// Special case: A-2-3-4-5 straight (wheel)
		if !isStraight && cards[0].Rank == Ace && cards[1].Rank == Two &&
			cards[2].Rank == Three && cards[3].Rank == Four && cards[4].Rank == Five {
			isStraight = true
		}

		// Special case: A-10-J-Q-K straight (high ace) - Royal Flush possibility
		if !isStraight && len(cards) == 5 && cards[0].Rank == Ace && cards[1].Rank == Ten &&
			cards[2].Rank == Jack && cards[3].Rank == Queen && cards[4].Rank == King {
			isStraight = true
		}
	}

	// Calculate total card value
	totalValue := 0
	for _, card := range cards {
		totalValue += card.Rank.Value()
	}

	// Determine hand type
	var handType HandType

	// Royal Flush: A-10-J-Q-K all same suit (cards are sorted, so A comes first)
	if isFlush && isStraight && len(cards) == 5 && cards[0].Rank == Ace && cards[1].Rank == Ten &&
		cards[2].Rank == Jack && cards[3].Rank == Queen && cards[4].Rank == King {
		handType = RoyalFlush
	} else if isFlush && isStraight {
		handType = StraightFlush
	} else if isFlush {
		handType = Flush
	} else if isStraight {
		handType = Straight
	} else {
		// Check for pairs, three of a kind, etc.
		var counts []int
		for _, count := range rankCounts {
			counts = append(counts, count)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(counts)))

		if counts[0] == 4 {
			handType = FourOfAKind
		} else if counts[0] == 3 && len(counts) > 1 && counts[1] == 2 {
			handType = FullHouse
		} else if counts[0] == 3 {
			handType = ThreeOfAKind
		} else if counts[0] == 2 && len(counts) > 1 && counts[1] == 2 {
			handType = TwoPair
		} else if counts[0] == 2 {
			handType = Pair
		} else {
			handType = HighCard
		}
	}

	// Calculate final score: (base score + card values) * mult
	baseScore := handType.BaseScore()
	mult := handType.Mult()
	finalScore := (baseScore + totalValue) * mult

	return handType, finalScore, totalValue, baseScore
}

func main() {
	// Parse command line flags
	seed := flag.Int64("seed", 0, "Set random seed for reproducible gameplay (0 for random)")
	flag.Parse()

	// Set seed if provided
	if *seed != 0 {
		SetSeed(*seed)
		fmt.Printf("Using seed: %d\n", *seed)
	}

	fmt.Println("🃏 Welcome to Balatro CLI! 🃏")
	fmt.Println("🎯 CHALLENGE: Score 300 points with 4 hands and 3 discards!")
	fmt.Println("Face cards (J, Q, K) = 10 points, Aces = 11 points")
	fmt.Println()

	// Game state
	const targetScore = 300
	const maxHands = 4
	const maxDiscards = 3

	totalScore := 0
	handsPlayed := 0
	discardsUsed := 0

	// Create and shuffle deck
	deck := NewDeck()
	ShuffleDeck(deck)
	deckIndex := 0

	// Deal initial hand (7 cards)
	handSize := 7
	playerCards := make([]Card, handSize)
	copy(playerCards, deck[deckIndex:deckIndex+handSize])
	deckIndex += handSize

	scanner := bufio.NewScanner(os.Stdin)

	for handsPlayed < maxHands && totalScore < targetScore {
		// Show game status
		fmt.Printf("🎯 Target: %d | Current Score: %d | Hands Left: %d | Discards Left: %d\n",
			targetScore, totalScore, maxHands-handsPlayed, maxDiscards-discardsUsed)
		fmt.Println()

		fmt.Println("Your cards:")
		for i, card := range playerCards {
			fmt.Printf("%d: %s\n", i+1, card)
		}
		fmt.Println()

		fmt.Print("Choose action: 'play <cards>' to play hand, 'discard <cards>' to discard, or 'quit': ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			}
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if strings.ToLower(input) == "quit" {
			fmt.Println("Thanks for playing!")
			break
		}

		if input == "" {
			fmt.Println("Please enter an action")
			continue
		}

		parts := strings.Fields(input)
		if len(parts) < 1 {
			fmt.Println("Please enter 'play <cards>' or 'discard <cards>'")
			continue
		}

		action := strings.ToLower(parts[0])

		if action == "play" {
			if len(parts) < 2 {
				fmt.Println("Please specify cards to play: 'play 1 2 3'")
				continue
			}

			// Parse selected card indices
			selections := parts[1:]
			if len(selections) > 5 {
				fmt.Println("You can only play up to 5 cards!")
				continue
			}

			var selectedCards []Card
			var selectedIndices []int
			valid := true

			for _, sel := range selections {
				index, err := strconv.Atoi(sel)
				if err != nil || index < 1 || index > len(playerCards) {
					fmt.Printf("Invalid card number: %s\n", sel)
					valid = false
					break
				}
				selectedCards = append(selectedCards, playerCards[index-1])
				selectedIndices = append(selectedIndices, index-1)
			}

			if !valid {
				continue
			}

			if len(selectedCards) == 0 {
				fmt.Println("Please select at least one card!")
				continue
			}

			// Evaluate the hand
			hand := Hand{Cards: selectedCards}
			handType, score, cardValues, baseScore := EvaluateHand(hand)

			fmt.Println()
			fmt.Printf("Your hand: %s\n", hand)
			fmt.Printf("Hand type: %s\n", handType)
			fmt.Printf("Base Score: %d | Card Values: %d | Mult: %dx\n", baseScore, cardValues, handType.Mult())
			fmt.Printf("Final Score: (%d + %d) × %d = %d points\n", baseScore, cardValues, handType.Mult(), score)

			totalScore += score
			handsPlayed++

			fmt.Printf("💰 Total Score: %d/%d\n", totalScore, targetScore)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			// Remove played cards and deal new ones
			playerCards = removeCards(playerCards, selectedIndices)
			newCardsNeeded := len(selectedCards)

			// Deal new cards if available
			if deckIndex+newCardsNeeded <= len(deck) {
				for i := 0; i < newCardsNeeded; i++ {
					playerCards = append(playerCards, deck[deckIndex])
					deckIndex++
				}
			}

		} else if action == "discard" {
			if discardsUsed >= maxDiscards {
				fmt.Println("No discards remaining!")
				continue
			}

			if len(parts) < 2 {
				fmt.Println("Please specify cards to discard: 'discard 1 2'")
				continue
			}

			// Parse selected card indices
			selections := parts[1:]
			var selectedIndices []int
			valid := true

			for _, sel := range selections {
				index, err := strconv.Atoi(sel)
				if err != nil || index < 1 || index > len(playerCards) {
					fmt.Printf("Invalid card number: %s\n", sel)
					valid = false
					break
				}
				selectedIndices = append(selectedIndices, index-1)
			}

			if !valid {
				continue
			}

			if len(selectedIndices) == 0 {
				fmt.Println("Please select at least one card!")
				continue
			}

			fmt.Printf("Discarded %d card(s)\n", len(selectedIndices))
			discardsUsed++

			// Remove discarded cards and deal new ones
			playerCards = removeCards(playerCards, selectedIndices)
			newCardsNeeded := len(selectedIndices)

			// Deal new cards if available
			if deckIndex+newCardsNeeded <= len(deck) {
				for i := 0; i < newCardsNeeded; i++ {
					playerCards = append(playerCards, deck[deckIndex])
					deckIndex++
				}
			}

			fmt.Println("New cards dealt!")
			fmt.Println()

		} else {
			fmt.Println("Invalid action. Use 'play <cards>' or 'discard <cards>'")
			continue
		}
	}

	// Game over - show results
	fmt.Println(strings.Repeat("=", 50))
	if totalScore >= targetScore {
		fmt.Println("🎉 VICTORY! You reached the target score!")
	} else {
		fmt.Println("💀 DEFEAT! You ran out of hands.")
	}
	fmt.Printf("Final Score: %d/%d\n", totalScore, targetScore)
	fmt.Printf("Hands Played: %d/%d\n", handsPlayed, maxHands)
	fmt.Printf("Discards Used: %d/%d\n", discardsUsed, maxDiscards)
	fmt.Println(strings.Repeat("=", 50))
}

// removeCards removes cards at specified indices and returns the new slice
func removeCards(cards []Card, indices []int) []Card {
	// Sort indices in descending order to remove from end first
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	result := make([]Card, len(cards))
	copy(result, cards)

	for _, index := range indices {
		if index >= 0 && index < len(result) {
			result = append(result[:index], result[index+1:]...)
		}
	}

	return result
}
