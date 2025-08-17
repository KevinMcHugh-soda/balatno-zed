package game_test

import (
	"testing"

	game "balatno/internal/game"
)

func TestGetBossForAnte(t *testing.T) {
	if err := game.LoadBossConfigs(); err != nil {
		t.Fatalf("LoadBossConfigs failed: %v", err)
	}

	boss := game.GetBossForAnte(1)
	if boss.Final {
		t.Errorf("expected non-final boss for ante 1, got final boss %s", boss.Name)
	}

	finalBoss := game.GetBossForAnte(8)
	if !finalBoss.Final {
		t.Errorf("expected final boss for ante 8, got non-final boss %s", finalBoss.Name)
	}
}
