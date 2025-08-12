package game

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type saveFile struct {
	SaveVersion   int      `json:"save_version"`
	Seed          int64    `json:"seed"`
	CurrentAnte   int      `json:"current_ante"`
	CurrentBlind  string   `json:"current_blind"`
	CurrentMoney  int      `json:"current_money"`
	CurrentJokers []string `json:"current_jokers"`
}

func parseBlindType(name string) (BlindType, error) {
	switch name {
	case SmallBlind.String():
		return SmallBlind, nil
	case BigBlind.String():
		return BigBlind, nil
	case BossBlind.String():
		return BossBlind, nil
	default:
		return SmallBlind, fmt.Errorf("unknown blind type %q", name)
	}
}

// LoadGameFromFile creates a Game using state from a JSON save file
func LoadGameFromFile(path string, handler EventHandler) (*Game, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var save saveFile
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}

	if save.SaveVersion != 1 {
		return nil, fmt.Errorf("unsupported save version: %d", save.SaveVersion)
	}

	if save.Seed != 0 {
		SetSeed(save.Seed)
	}

	g := NewGame(handler)
	g.currentAnte = save.CurrentAnte
	bt, err := parseBlindType(save.CurrentBlind)
	if err != nil {
		return nil, err
	}
	g.currentBlind = bt
	g.money = save.CurrentMoney

	g.jokers = []Joker{}
	for _, name := range save.CurrentJokers {
		if joker, ok := GetJokerByName(name); ok {
			g.jokers = append(g.jokers, joker)
		} else {
			return nil, fmt.Errorf("unknown joker: %s", name)
		}
	}

	g.currentTarget = GetAnteRequirement(g.currentAnte, g.currentBlind)
	return g, nil
}

// Save writes the current game state to a timestamped JSON file
func (g *Game) Save() (string, error) {
	save := saveFile{
		SaveVersion:   1,
		Seed:          GetSeed(),
		CurrentAnte:   g.currentAnte,
		CurrentBlind:  g.currentBlind.String(),
		CurrentMoney:  g.money,
		CurrentJokers: make([]string, len(g.jokers)),
	}

	for i, joker := range g.jokers {
		save.CurrentJokers[i] = joker.Name
	}

	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll("saves", 0755); err != nil {
		return "", err
	}

	filename := filepath.Join("saves", time.Now().UTC().Format(time.RFC3339)+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", err
	}
	return filename, nil
}
