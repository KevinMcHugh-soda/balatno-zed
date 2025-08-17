package game

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// AnteRequirement represents the score requirements for one ante
type AnteRequirement struct {
	Small int
	Big   int
	Boss  int
}

// HandScore represents the base scores per level and multiplier for a hand type
type HandScore struct {
	Name        string
	LevelScores []int
	Multiplier  int
}

// Config holds all game configuration loaded from CSV files
type Config struct {
	AnteRequirements []AnteRequirement
	HandScores       map[string]HandScore
}

var gameConfig *Config

// LoadConfig loads configuration from CSV files with fallback to defaults
func LoadConfig() error {
	config := &Config{
		HandScores: make(map[string]HandScore),
	}

	// Load ante requirements
	if err := config.loadAnteRequirements(); err != nil {
		fmt.Printf("Warning: Could not load ante_requirements.csv, using defaults: %v\n", err)
		config.setDefaultAnteRequirements()
	}

	// Load hand scores
	if err := config.loadHandScores(); err != nil {
		fmt.Printf("Warning: Could not load hand_scores.csv, using defaults: %v\n", err)
		config.setDefaultHandScores()
	}

	gameConfig = config
	return nil
}

// loadAnteRequirements loads ante requirements from CSV file
func (c *Config) loadAnteRequirements() error {
	file, err := os.Open(filepath.Join("internal", "game", "ante_requirements.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("ante_requirements.csv must have at least a header and one data row")
	}

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 3 {
			return fmt.Errorf("ante_requirements.csv row %d must have exactly 3 columns", i+1)
		}

		small, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("invalid small blind value in row %d: %v", i+1, err)
		}

		big, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("invalid big blind value in row %d: %v", i+1, err)
		}

		boss, err := strconv.Atoi(record[2])
		if err != nil {
			return fmt.Errorf("invalid boss blind value in row %d: %v", i+1, err)
		}

		c.AnteRequirements = append(c.AnteRequirements, AnteRequirement{
			Small: small,
			Big:   big,
			Boss:  boss,
		})
	}

	return nil
}

// loadHandScores loads hand scores from CSV file
func (c *Config) loadHandScores() error {
	file, err := os.Open(filepath.Join("internal", "game", "hand_scores.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("hand_scores.csv must have at least a header and one data row")
	}

	header := records[0]
	if len(header) < 3 {
		return fmt.Errorf("hand_scores.csv must have at least hand, one level, and mult columns")
	}
	levelCount := len(header) - 2 // subtract hand name and multiplier

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != len(header) {
			return fmt.Errorf("hand_scores.csv row %d must have exactly %d columns", i+1, len(header))
		}

		handName := record[0]

		levels := make([]int, levelCount)
		for j := 0; j < levelCount; j++ {
			baseScore, err := strconv.Atoi(record[j+1])
			if err != nil {
				return fmt.Errorf("invalid level %d base score for %s in row %d: %v", j+1, handName, i+1, err)
			}
			levels[j] = baseScore
		}

		multiplier, err := strconv.Atoi(record[len(record)-1])
		if err != nil {
			return fmt.Errorf("invalid multiplier for %s in row %d: %v", handName, i+1, err)
		}

		c.HandScores[handName] = HandScore{
			Name:        handName,
			LevelScores: levels,
			Multiplier:  multiplier,
		}
	}

	return nil
}

// setDefaultAnteRequirements sets hardcoded default ante requirements
func (c *Config) setDefaultAnteRequirements() {
	c.AnteRequirements = []AnteRequirement{
		{300, 450, 600},   // Ante 1
		{375, 562, 750},   // Ante 2
		{450, 675, 900},   // Ante 3
		{525, 787, 1050},  // Ante 4
		{600, 900, 1200},  // Ante 5
		{675, 1012, 1350}, // Ante 6
		{750, 1125, 1500}, // Ante 7
		{825, 1237, 1650}, // Ante 8
	}
}

// setDefaultHandScores sets hardcoded default hand scores
func (c *Config) setDefaultHandScores() {
	defaults := []HandScore{
		{"High Card", []int{5, 10, 15, 20, 25}, 1},
		{"Pair", []int{10, 15, 20, 25, 30}, 2},
		{"Two Pair", []int{20, 25, 30, 35, 40}, 2},
		{"Three of a Kind", []int{30, 35, 40, 45, 50}, 3},
		{"Straight", []int{30, 35, 40, 45, 50}, 4},
		{"Flush", []int{35, 40, 45, 50, 55}, 4},
		{"Full House", []int{40, 45, 50, 55, 60}, 4},
		{"Four of a Kind", []int{60, 65, 70, 75, 80}, 7},
		{"Straight Flush", []int{100, 105, 110, 115, 120}, 8},
		{"Royal Flush", []int{100, 105, 110, 115, 120}, 8},
	}

	for _, handScore := range defaults {
		c.HandScores[handScore.Name] = handScore
	}
}

// GetAnteRequirement returns the score requirements for a specific ante and blind type
func GetAnteRequirement(ante int, blindType BlindType) int {
	if gameConfig == nil || ante < 1 || ante > len(gameConfig.AnteRequirements) {
		// Fallback to original calculation
		base := 300
		requirement := base + (ante-1)*75

		switch blindType {
		case SmallBlind:
			return requirement
		case BigBlind:
			return int(float64(requirement) * 1.5)
		case BossBlind:
			return requirement * 2
		default:
			return requirement
		}
	}

	req := gameConfig.AnteRequirements[ante-1]
	switch blindType {
	case SmallBlind:
		return req.Small
	case BigBlind:
		return req.Big
	case BossBlind:
		return req.Boss
	default:
		return req.Small
	}
}

// GetHandScore returns the base score for a specific level and multiplier for a hand type
func GetHandScore(handName string, level int) (int, int) {
	if level < 1 {
		level = 1
	}
	if gameConfig == nil {
		// Fallback to hardcoded defaults
		defaults := map[string]HandScore{
			"High Card":       {LevelScores: []int{5, 10, 15, 20, 25}, Multiplier: 1},
			"Pair":            {LevelScores: []int{10, 15, 20, 25, 30}, Multiplier: 2},
			"Two Pair":        {LevelScores: []int{20, 25, 30, 35, 40}, Multiplier: 2},
			"Three of a Kind": {LevelScores: []int{30, 35, 40, 45, 50}, Multiplier: 3},
			"Straight":        {LevelScores: []int{30, 35, 40, 45, 50}, Multiplier: 4},
			"Flush":           {LevelScores: []int{35, 40, 45, 50, 55}, Multiplier: 4},
			"Full House":      {LevelScores: []int{40, 45, 50, 55, 60}, Multiplier: 4},
			"Four of a Kind":  {LevelScores: []int{60, 65, 70, 75, 80}, Multiplier: 7},
			"Straight Flush":  {LevelScores: []int{100, 105, 110, 115, 120}, Multiplier: 8},
			"Royal Flush":     {LevelScores: []int{100, 105, 110, 115, 120}, Multiplier: 8},
		}
		if score, exists := defaults[handName]; exists {
			idx := level - 1
			if idx >= len(score.LevelScores) {
				idx = len(score.LevelScores) - 1
			}
			return score.LevelScores[idx], score.Multiplier
		}
		return 5, 1 // Default to High Card values
	}

	if score, exists := gameConfig.HandScores[handName]; exists {
		idx := level - 1
		if idx >= len(score.LevelScores) {
			idx = len(score.LevelScores) - 1
		}
		return score.LevelScores[idx], score.Multiplier
	}

	// Fallback to High Card values
	return 5, 1
}

// GetAllHandScores returns all configured hand scores for display purposes
func GetAllHandScores() map[string]HandScore {
	if gameConfig == nil {
		return make(map[string]HandScore)
	}
	return gameConfig.HandScores
}

// GetAllAnteRequirements returns all configured ante requirements for display purposes
func GetAllAnteRequirements() []AnteRequirement {
	if gameConfig == nil {
		return make([]AnteRequirement, 0)
	}
	return gameConfig.AnteRequirements
}
