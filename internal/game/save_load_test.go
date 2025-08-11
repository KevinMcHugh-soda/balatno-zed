package game

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadGameFromFile(t *testing.T) {
	save := saveFile{
		SaveVersion:   1,
		Seed:          123,
		CurrentAnte:   2,
		CurrentBlind:  BigBlind.String(),
		CurrentMoney:  10,
		CurrentJokers: []string{"The Golden Joker"},
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

	SetSeed(123)
	expected := NewGame(NewLoggerEventHandler())
	for i, card := range expected.deck {
		if g.deck[i] != card {
			t.Fatalf("deck differs at %d", i)
		}
	}
}
