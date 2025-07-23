package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// AnteRequirement represents the score requirements for one ante
type AnteRequirement struct {
	Small int
	Big   int
	Boss  int
}

// HandScore represents the base score and multiplier for a hand type
type HandScore struct {
	Name       string
	BaseScore  int
	Multiplier int
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
	file, err := os.Open("ante_requirements.csv")
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
	file, err := os.Open("hand_scores.csv")
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

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 3 {
			return fmt.Errorf("hand_scores.csv row %d must have exactly 3 columns", i+1)
		}

		handName := record[0]

		baseScore, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("invalid base score for %s in row %d: %v", handName, i+1, err)
		}

		multiplier, err := strconv.Atoi(record[2])
		if err != nil {
			return fmt.Errorf("invalid multiplier for %s in row %d: %v", handName, i+1, err)
		}

		c.HandScores[handName] = HandScore{
			Name:       handName,
			BaseScore:  baseScore,
			Multiplier: multiplier,
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
		{"High Card", 5, 1},
		{"Pair", 10, 2},
		{"Two Pair", 20, 2},
		{"Three of a Kind", 30, 3},
		{"Straight", 30, 4},
		{"Flush", 35, 4},
		{"Full House", 40, 4},
		{"Four of a Kind", 60, 7},
		{"Straight Flush", 100, 8},
		{"Royal Flush", 100, 8},
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

// GetHandScore returns the base score and multiplier for a hand type
func GetHandScore(handName string) (int, int) {
	if gameConfig == nil {
		// Fallback to hardcoded defaults
		defaults := map[string]HandScore{
			"High Card":       {BaseScore: 5, Multiplier: 1},
			"Pair":            {BaseScore: 10, Multiplier: 2},
			"Two Pair":        {BaseScore: 20, Multiplier: 2},
			"Three of a Kind": {BaseScore: 30, Multiplier: 3},
			"Straight":        {BaseScore: 30, Multiplier: 4},
			"Flush":           {BaseScore: 35, Multiplier: 4},
			"Full House":      {BaseScore: 40, Multiplier: 4},
			"Four of a Kind":  {BaseScore: 60, Multiplier: 7},
			"Straight Flush":  {BaseScore: 100, Multiplier: 8},
			"Royal Flush":     {BaseScore: 100, Multiplier: 8},
		}
		if score, exists := defaults[handName]; exists {
			return score.BaseScore, score.Multiplier
		}
		return 5, 1 // Default to High Card values
	}

	if score, exists := gameConfig.HandScores[handName]; exists {
		return score.BaseScore, score.Multiplier
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
