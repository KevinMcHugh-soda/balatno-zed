package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGameFromFile(t *testing.T) {
	save := saveFile{
		SaveVersion:   2,
		Seed:          123,
		CurrentAnte:   2,
		CurrentBlind:  BigBlind.String(),
		CurrentMoney:  10,
		CurrentJokers: []string{"The Golden Joker"},
		HandLevels:    map[string]int{"Pair": 2},
	}

	tmp, err := os.CreateTemp("", "save*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if err := json.NewEncoder(tmp).Encode(save); err != nil {
		t.Fatalf("encoding save: %v", err)
	}
	tmp.Close()

	g, err := LoadGameFromFile(tmp.Name(), NewLoggerEventHandler())
	if err != nil {
		t.Fatalf("LoadGameFromFile returned error: %v", err)
	}

	if g.currentAnte != 2 {
		t.Errorf("currentAnte = %d, want 2", g.currentAnte)
	}
	if g.currentBlind != BigBlind {
		t.Errorf("currentBlind = %v, want %v", g.currentBlind, BigBlind)
	}
	if g.money != 10 {
		t.Errorf("money = %d, want 10", g.money)
	}
	if len(g.jokers) != 1 || g.jokers[0].Name != "The Golden Joker" {
		t.Fatalf("jokers = %#v, want The Golden Joker", g.jokers)
	}
	if g.handLevels["Pair"] != 2 {
		t.Errorf("hand level Pair = %d, want 2", g.handLevels["Pair"])
	}

	SetSeed(123)
	expected := NewGame(NewLoggerEventHandler())
	for i, card := range expected.deck {
		if g.deck[i] != card {
			t.Fatalf("deck differs at %d", i)
		}
	}
}

func TestSaveGameToFile(t *testing.T) {
	SetSeed(456)
	g := NewGame(NewLoggerEventHandler())
	g.LevelUpHand("Pair")

	filename, err := g.Save()
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(filename))

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("reading save file: %v", err)
	}

	var save saveFile
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatalf("unmarshal save: %v", err)
	}

	if save.Seed != 456 {
		t.Errorf("seed = %d, want 456", save.Seed)
	}
	if save.CurrentAnte != g.currentAnte {
		t.Errorf("ante = %d, want %d", save.CurrentAnte, g.currentAnte)
	}
	if save.CurrentBlind != g.currentBlind.String() {
		t.Errorf("blind = %s, want %s", save.CurrentBlind, g.currentBlind.String())
	}
	if save.CurrentMoney != g.money {
		t.Errorf("money = %d, want %d", save.CurrentMoney, g.money)
	}
	if len(save.CurrentJokers) != len(g.jokers) {
		t.Errorf("jokers = %v, want %v", save.CurrentJokers, g.jokers)
	}
	if save.HandLevels["Pair"] != g.handLevels["Pair"] {
		t.Errorf("hand level saved = %d, want %d", save.HandLevels["Pair"], g.handLevels["Pair"])
	}
}
