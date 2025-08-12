package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// Global random source for consistent seeding
var rng *rand.Rand
var currentSeed int64

func init() {
	currentSeed = time.Now().UnixNano()
	rng = rand.New(rand.NewSource(currentSeed))
}

// SetSeed allows setting a specific seed for deterministic behavior (useful for testing)
func SetSeed(seed int64) {
	currentSeed = seed
	rng = rand.New(rand.NewSource(seed))
}

// GetSeed returns the current random seed
func GetSeed() int64 {
	return currentSeed
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

type Edition int

const (
	PlainEdition Edition = iota
	Foil
	Holo
	Poly
)

func (e Edition) Emoji() string {
	switch e {
	case PlainEdition:
		return ""
	case Foil:
		return "🪞"
	case Holo:
		return "✨"
	case Poly:
		return "🌈"
	}
	return ""
}

type Sticker int

const (
	NoSticker Sticker = iota
	Red
	Purple
	Blue
)

func (e Sticker) Emoji() string {
	switch e {
	case NoSticker:
		return ""
	case Red:
		return "🔴"
	case Purple:
		return "🟣"
	case Blue:
		return "🔵"
	}
	return ""
}

// I don't actually remember what these are called
type Print int

const (
	PlainPrint Print = iota
	PlusMult
	PlusChips
	Glass
	Steel
	Lucky
)

func (e Print) Emoji() string {
	switch e {
	case PlainPrint:
		return ""
	case PlusMult:
		return "✕"
	case PlusChips:
		return "🍪"
	case Glass:
		return "🪟"
	case Steel:
		return "🧑‍🏭"
	case Lucky:
		return "🍀"
	}
	return ""
}

type Card struct {
	Suit    Suit
	Rank    Rank
	Edition Edition
	Sticker Sticker
	Print   Print
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

// NewDeck creates a standard 52-card deck
func NewDeck() []Card {
	var deck []Card
	for suit := Hearts; suit <= Spades; suit++ {
		for rank := Ace; rank <= King; rank++ {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

// ShuffleDeck shuffles the deck in place using the global random source
func ShuffleDeck(deck []Card) {
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}
