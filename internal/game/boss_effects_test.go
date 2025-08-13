package game

import "testing"

func TestApplyBossEffect(t *testing.T) {
	g := &Game{currentTarget: 100, money: 200}

	g.currentBoss = Boss{Effect: DoubleChips}
	g.applyBossEffect()
	if g.currentTarget != 200 {
		t.Errorf("expected target 200, got %d", g.currentTarget)
	}

	g.currentBoss = Boss{Effect: HalveMoney}
	g.applyBossEffect()
	if g.money != 100 {
		t.Errorf("expected money 100, got %d", g.money)
	}
}
