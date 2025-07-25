package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// JokerEffect represents the type of effect a joker has
type JokerEffect string

const (
	AddMoney JokerEffect = "AddMoney"
	AddChips JokerEffect = "AddChips"
	AddMult  JokerEffect = "AddMult"
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

// JokerConfig represents a joker configuration from YAML
type JokerConfig struct {
	Name             string           `yaml:"name"`
	Value            int              `yaml:"value"`
	Rarity           string           `yaml:"rarity"`
	Effect           JokerEffect      `yaml:"effect"`
	EffectMagnitude  int              `yaml:"effect_magnitude"`
	HandMatchingRule HandMatchingRule `yaml:"hand_matching_rule"`
	Description      string           `yaml:"description"`
}

// JokersYAML represents the root YAML structure
type JokersYAML struct {
	Jokers []JokerConfig `yaml:"jokers"`
}

// Joker represents a joker card that modifies gameplay
type Joker struct {
	Name             string
	Description      string
	Price            int
	Effect           JokerEffect
	EffectMagnitude  int
	HandMatchingRule HandMatchingRule
	OnBlindEnd       func() int              // Returns money earned at end of blind
	OnHandScoring    func(string) (int, int) // Returns (chips, mult) bonus for hand type
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
	file, err := os.Open("jokers.yaml")
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
			Name:             "The Golden Joker",
			Value:            6,
			Rarity:           "Common",
			Effect:           AddMoney,
			EffectMagnitude:  4,
			HandMatchingRule: None,
			Description:      "Earn $4 at the end of each Blind",
		},
		{
			Name:             "Chip Collector",
			Value:            5,
			Rarity:           "Common",
			Effect:           AddChips,
			EffectMagnitude:  30,
			HandMatchingRule: ContainsPair,
			Description:      "+30 Chips if played hand contains a Pair",
		},
		{
			Name:             "Double Down",
			Value:            4,
			Rarity:           "Common",
			Effect:           AddMult,
			EffectMagnitude:  8,
			HandMatchingRule: ContainsPair,
			Description:      "+8 Mult if played hand contains a Pair",
		},
	}
}

// createJokerFromConfig creates a Joker instance from a JokerConfig
func createJokerFromConfig(config JokerConfig) Joker {
	joker := Joker{
		Name:             config.Name,
		Description:      config.Description,
		Price:            config.Value,
		Effect:           config.Effect,
		EffectMagnitude:  config.EffectMagnitude,
		HandMatchingRule: config.HandMatchingRule,
	}

	// Set up effect functions based on effect type
	switch config.Effect {
	case AddMoney:
		joker.OnBlindEnd = func() int {
			return config.EffectMagnitude
		}
	case AddChips, AddMult:
		joker.OnHandScoring = func(handType string) (int, int) {
			if handMatchesRule(handType, config.HandMatchingRule) {
				if config.Effect == AddChips {
					return config.EffectMagnitude, 0
				} else {
					return 0, config.EffectMagnitude
				}
			}
			return 0, 0
		}
	}

	return joker
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
		Effect:      AddMoney,
		OnBlindEnd: func() int {
			return 4
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
		if joker.OnBlindEnd != nil {
			total += joker.OnBlindEnd()
		}
	}
	return total
}

// CalculateJokerHandBonus calculates chips and mult bonus from jokers for a specific hand
func CalculateJokerHandBonus(jokers []Joker, handType string) (int, int) {
	totalChips := 0
	totalMult := 0

	for _, joker := range jokers {
		if joker.OnHandScoring != nil {
			chips, mult := joker.OnHandScoring(handType)
			totalChips += chips
			totalMult += mult
		}
	}

	return totalChips, totalMult
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
		if joker.Effect == effect {
			filtered = append(filtered, joker)
		}
	}
	return filtered
}
