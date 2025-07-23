package main

// Joker represents a joker card that modifies gameplay
type Joker struct {
	Name        string
	Description string
	Price       int
	OnBlindEnd  func() int // Returns money earned at end of blind
}

// GetGoldenJoker returns The Golden Joker
func GetGoldenJoker() Joker {
	return Joker{
		Name:        "The Golden Joker",
		Description: "Earn $4 at the end of each Blind",
		Price:       6,
		OnBlindEnd: func() int {
			return 4
		},
	}
}

// GetAvailableJokers returns all jokers that can be purchased
func GetAvailableJokers() []Joker {
	return []Joker{
		GetGoldenJoker(),
		// Future jokers will be added here
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
