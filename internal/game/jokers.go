package game

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// JokerEffect represents the type of effect a joker has
type JokerEffect string

const (
	AddMoney    JokerEffect = "AddMoney"
	AddChips    JokerEffect = "AddChips"
	AddMult     JokerEffect = "AddMult"
	AddHandSize JokerEffect = "AddHandSize"
	AddDiscards JokerEffect = "AddDiscards"
	ReplayCard  JokerEffect = "ReplayCard"
)

// HandMatchingRule represents when a joker effect should trigger
type HandMatchingRule string

const (
	None                  HandMatchingRule = "None"
	ContainsPair          HandMatchingRule = "ContainsPair"
	ContainsTwoPair       HandMatchingRule = "ContainsTwoPair"
	ContainsThreeOfAKind  HandMatchingRule = "ContainsThreeOfAKind"
	ContainsStraight      HandMatchingRule = "ContainsStraight"
	ContainsFlush         HandMatchingRule = "ContainsFlush"
	ContainsFullHouse     HandMatchingRule = "ContainsFullHouse"
	ContainsFourOfAKind   HandMatchingRule = "ContainsFourOfAKind"
	ContainsStraightFlush HandMatchingRule = "ContainsStraightFlush"
	ContainsRoyalFlush    HandMatchingRule = "ContainsRoyalFlush"
)

// CardMatchingRule represents a rule for matching individual cards in a played hand
type CardMatchingRule string

const (
	CardNone    CardMatchingRule = "None"
	CardIsAce   CardMatchingRule = "IsAce"
	CardIsSpade CardMatchingRule = "IsSpade"
	CardIsFace  CardMatchingRule = "IsFace"
)

// JokerEffectConfig represents a single effect component of a joker
type JokerEffectConfig struct {
	Effect           JokerEffect      `yaml:"effect"`
	EffectMagnitude  int              `yaml:"effect_magnitude"`
	HandMatchingRule HandMatchingRule `yaml:"hand_matching_rule"`
	CardMatchingRule CardMatchingRule `yaml:"card_matching_rule"`
}

// JokerConfig represents a joker configuration from YAML
type JokerConfig struct {
	Name        string `yaml:"name"`
	Value       int    `yaml:"value"`
	Rarity      string `yaml:"rarity"`
	Description string `yaml:"description"`
	// Legacy single-effect fields for backward compatibility
	Effect           JokerEffect      `yaml:"effect"`
	EffectMagnitude  int              `yaml:"effect_magnitude"`
	HandMatchingRule HandMatchingRule `yaml:"hand_matching_rule"`
	CardMatchingRule CardMatchingRule `yaml:"card_matching_rule"`
	// Composite effects
	Effects []JokerEffectConfig `yaml:"effects"`
}

// JokersYAML represents the root YAML structure
type JokersYAML struct {
	Jokers []JokerConfig `yaml:"jokers"`
}

// Joker represents a joker card that modifies gameplay
type Joker struct {
	Name        string
	Description string
	Price       int
	Effects     []JokerEffectConfig
}

var jokerConfigs []JokerConfig

// LoadJokerConfigs loads joker configurations from YAML file with fallback to defaults
func LoadJokerConfigs() error {
	// Try to load from YAML file
	if err := loadJokersFromYAML(); err != nil {
		fmt.Printf("Warning: Could not load jokers.yaml, using defaults: %v\n", err)
		setDefaultJokerConfigs()
	}

	return nil
}

// loadJokersFromYAML loads joker configurations from YAML file
func loadJokersFromYAML() error {
	file, err := os.Open(filepath.Join("internal", "game", "jokers.yaml"))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var jokersYAML JokersYAML
	err = yaml.Unmarshal(data, &jokersYAML)
	if err != nil {
		return err
	}

	if len(jokersYAML.Jokers) == 0 {
		return fmt.Errorf("jokers.yaml contains no jokers")
	}

	jokerConfigs = jokersYAML.Jokers
	return nil
}

// setDefaultJokerConfigs sets hardcoded default joker configurations
func setDefaultJokerConfigs() {
	jokerConfigs = []JokerConfig{
		{
			Name:   "The Golden Joker",
			Value:  6,
			Rarity: "Common",
			Effects: []JokerEffectConfig{
				{
					Effect:           AddMoney,
					EffectMagnitude:  4,
					HandMatchingRule: None,
					CardMatchingRule: CardNone,
				},
			},
			Description: "Earn $4 at the end of each Blind",
		},
		{
			Name:   "Chip Collector",
			Value:  5,
			Rarity: "Common",
			Effects: []JokerEffectConfig{
				{
					Effect:           AddChips,
					EffectMagnitude:  30,
					HandMatchingRule: ContainsPair,
					CardMatchingRule: CardNone,
				},
			},
			Description: "+30 Chips if played hand contains a Pair",
		},
		{
			Name:   "Double Down",
			Value:  4,
			Rarity: "Common",
			Effects: []JokerEffectConfig{
				{
					Effect:           AddMult,
					EffectMagnitude:  8,
					HandMatchingRule: ContainsPair,
					CardMatchingRule: CardNone,
				},
			},
			Description: "+8 Mult if played hand contains a Pair",
		},
		{
			Name:   "Face Dancer",
			Value:  7,
			Rarity: "Common",
			Effects: []JokerEffectConfig{
				{
					Effect:           ReplayCard,
					EffectMagnitude:  0,
					HandMatchingRule: None,
					CardMatchingRule: CardIsFace,
				},
			},
			Description: "Face cards are scored twice",
		},
	}
}

// createJokerFromConfig creates a Joker instance from a JokerConfig
func createJokerFromConfig(config JokerConfig) Joker {
	var effects []JokerEffectConfig

	if len(config.Effects) > 0 {
		for _, e := range config.Effects {
			cardRule := e.CardMatchingRule
			if cardRule == "" {
				cardRule = CardNone
			}
			effects = append(effects, JokerEffectConfig{
				Effect:           e.Effect,
				EffectMagnitude:  e.EffectMagnitude,
				HandMatchingRule: e.HandMatchingRule,
				CardMatchingRule: cardRule,
			})
		}
	} else if config.Effect != "" {
		cardRule := config.CardMatchingRule
		if cardRule == "" {
			cardRule = CardNone
		}
		effects = append(effects, JokerEffectConfig{
			Effect:           config.Effect,
			EffectMagnitude:  config.EffectMagnitude,
			HandMatchingRule: config.HandMatchingRule,
			CardMatchingRule: cardRule,
		})
	}

	return Joker{
		Name:        config.Name,
		Description: config.Description,
		Price:       config.Value,
		Effects:     effects,
	}
}

// handMatchesRule checks if a hand type matches a given rule
func handMatchesRule(handType string, rule HandMatchingRule) bool {
	switch rule {
	case None:
		return true
	case ContainsPair:
		return containsPair(handType)
	case ContainsTwoPair:
		return containsTwoPair(handType)
	case ContainsThreeOfAKind:
		return containsThreeOfAKind(handType)
	case ContainsStraight:
		return containsStraight(handType)
	case ContainsFlush:
		return containsFlush(handType)
	case ContainsFullHouse:
		return containsFullHouse(handType)
	case ContainsFourOfAKind:
		return containsFourOfAKind(handType)
	case ContainsStraightFlush:
		return containsStraightFlush(handType)
	case ContainsRoyalFlush:
		return containsRoyalFlush(handType)
	default:
		return false
	}
}

// cardMatchesRule checks if a card matches a given card rule
func cardMatchesRule(card Card, rule CardMatchingRule) bool {
	switch rule {
	case CardIsAce:
		return card.Rank == Ace
	case CardIsSpade:
		return card.Suit == Spades
	case CardIsFace:
		return card.Rank == Jack || card.Rank == Queen || card.Rank == King
	default:
		return false
	}
}

// Hand containment checking functions
func containsPair(handType string) bool {
	return handType == "Pair" || handType == "Two Pair" || handType == "Three of a Kind" ||
		handType == "Full House" || handType == "Four of a Kind"
}

func containsTwoPair(handType string) bool {
	return handType == "Two Pair" || handType == "Full House"
}

func containsThreeOfAKind(handType string) bool {
	return handType == "Three of a Kind" || handType == "Full House" || handType == "Four of a Kind"
}

func containsStraight(handType string) bool {
	return handType == "Straight" || handType == "Straight Flush" || handType == "Royal Flush"
}

func containsFlush(handType string) bool {
	return handType == "Flush" || handType == "Straight Flush" || handType == "Royal Flush"
}

func containsFullHouse(handType string) bool {
	return handType == "Full House"
}

func containsFourOfAKind(handType string) bool {
	return handType == "Four of a Kind"
}

func containsStraightFlush(handType string) bool {
	return handType == "Straight Flush" || handType == "Royal Flush"
}

func containsRoyalFlush(handType string) bool {
	return handType == "Royal Flush"
}

// GetAvailableJokers returns all jokers that can be purchased
func GetAvailableJokers() []Joker {
	var jokers []Joker
	for _, config := range jokerConfigs {
		jokers = append(jokers, createJokerFromConfig(config))
	}
	return jokers
}

// GetJokerByName returns a Joker by its name if it exists
func GetJokerByName(name string) (Joker, bool) {
	for _, config := range jokerConfigs {
		if config.Name == name {
			return createJokerFromConfig(config), true
		}
	}
	return Joker{}, false
}

// GetGoldenJoker returns The Golden Joker (for backward compatibility)
func GetGoldenJoker() Joker {
	for _, config := range jokerConfigs {
		if config.Name == "The Golden Joker" {
			return createJokerFromConfig(config)
		}
	}
	// Fallback if not found in config
	return Joker{
		Name:        "The Golden Joker",
		Description: "Earn $4 at the end of each Blind",
		Price:       6,
		Effects: []JokerEffectConfig{
			{
				Effect:           AddMoney,
				EffectMagnitude:  4,
				HandMatchingRule: None,
				CardMatchingRule: CardNone,
			},
		},
	}
}

// PlayerHasJoker checks if the player already owns a specific joker
func PlayerHasJoker(playerJokers []Joker, jokerName string) bool {
	for _, joker := range playerJokers {
		if joker.Name == jokerName {
			return true
		}
	}
	return false
}

// CalculateJokerRewards calculates total money earned from all jokers at blind end
func CalculateJokerRewards(jokers []Joker) int {
	total := 0
	for _, joker := range jokers {
		for _, eff := range joker.Effects {
			if eff.Effect == AddMoney {
				total += eff.EffectMagnitude
			}
		}
	}
	return total
}

// CalculateJokerHandBonus calculates chips and mult bonus from jokers for a specific hand
func CalculateJokerHandBonus(jokers []Joker, handType string, cards []Card) (int, int) {
	totalChips := 0
	totalMult := 0

	for _, joker := range jokers {
		for _, eff := range joker.Effects {
			switch eff.Effect {
			case AddChips, AddMult:
				matches := 0
				if eff.CardMatchingRule != CardNone {
					for _, c := range cards {
						if cardMatchesRule(c, eff.CardMatchingRule) {
							matches++
						}
					}
				} else if handMatchesRule(handType, eff.HandMatchingRule) {
					matches = 1
				}
				if matches == 0 {
					continue
				}
				bonus := eff.EffectMagnitude * matches
				if eff.Effect == AddChips {
					totalChips += bonus
				} else {
					totalMult += bonus
				}
			}
		}
	}

	return totalChips, totalMult
}

// ApplyReplayCardEffects duplicates cards that match any ReplayCard joker
// and returns the extended card slice along with additional card value
// contributed by the replays.
func ApplyReplayCardEffects(jokers []Joker, cards []Card) ([]Card, int) {
	var replayed []Card
	extraValue := 0
	for _, joker := range jokers {
		for _, eff := range joker.Effects {
			if eff.Effect == ReplayCard && eff.CardMatchingRule != CardNone {
				for _, c := range cards {
					if cardMatchesRule(c, eff.CardMatchingRule) {
						replayed = append(replayed, c)
						extraValue += c.Rank.Value()
					}
				}
			}
		}
	}
	combined := append([]Card{}, cards...)
	combined = append(combined, replayed...)
	return combined, extraValue
}

// FormatJokersList returns a formatted string of player's jokers
func FormatJokersList(jokers []Joker) string {
	if len(jokers) == 0 {
		return "No jokers"
	}

	result := ""
	for i, joker := range jokers {
		if i > 0 {
			result += ", "
		}
		result += joker.Name
	}
	return result
}

// GetJokersByEffect returns jokers that have a specific effect type
func GetJokersByEffect(jokers []Joker, effect JokerEffect) []Joker {
	var filtered []Joker
	for _, joker := range jokers {
		for _, eff := range joker.Effects {
			if eff.Effect == effect {
				filtered = append(filtered, joker)
				break
			}
		}
	}
	return filtered
}
