package game

import "math/rand"

// BossRule defines special rules applied during Boss Blinds
// to modify scoring or hand size.
type BossRule int

const (
	BossRuleNone BossRule = iota
	// BossRuleNoHearts disables Hearts card values from scoring
	BossRuleNoHearts
	// BossRuleMinusHand reduces hand size by 1
	BossRuleMinusHand
	// BossRulePlusHand increases hand size by 1
	BossRulePlusHand
)

// randomBossRule returns a random boss rule for Boss Blinds
func randomBossRule() BossRule {
	rules := []BossRule{BossRuleNoHearts, BossRuleMinusHand, BossRulePlusHand}
	return rules[rand.Intn(len(rules))]
}

// Description returns a human-readable description of the boss rule
func (b BossRule) Description() string {
	switch b {
	case BossRuleNoHearts:
		return "Hearts score zero"
	case BossRuleMinusHand:
		return "Hand size reduced by 1"
	case BossRulePlusHand:
		return "Hand size increased by 1"
	default:
		return ""
	}
}
