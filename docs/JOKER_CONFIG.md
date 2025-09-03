# YAML Joker Configuration System

This document explains how to configure jokers in Balatro CLI using the YAML-based system. All jokers are defined in `jokers.yaml` and loaded at runtime - no compilation required!

## 📁 Configuration File

### `jokers.yaml` Structure

```yaml
jokers:
  - name: "Joker Name"
    value: 6                    # Price in shop
    rarity: "Common"            # Currently unused, for future expansion
    description: "Description shown in shop"
    effects:
      - effect: "AddChips"      # Effect type
        effect_magnitude: 30    # Strength of effect
        hand_matching_rule: "ContainsPair"  # When to trigger based on hand type
        card_matching_rule: "IsAce"         # (Optional) bonus per matching card

`hand_matching_rule` and `card_matching_rule` default to `"None"` when omitted.
```

## 🎭 Effect Types

### `AddMoney`
Adds money at the end of each blind (like The Golden Joker).

```yaml
- name: "The Golden Joker"
  effect: "AddMoney"
  effect_magnitude: 4           # Earn $4 per blind
  hand_matching_rule: "None"    # Always triggers
```

### `AddChips` 
Adds base chips when playing matching hands.

```yaml
- name: "Chip Collector"
  effect: "AddChips"
  effect_magnitude: 30          # +30 chips
  hand_matching_rule: "ContainsPair"  # Only for hands containing pairs
```

**Score Calculation**: `(base_score + joker_chips + card_values) × multiplier`

### `AddMult`
Adds multiplier when playing matching hands.

```yaml
- name: "Double Down" 
  effect: "AddMult"
  effect_magnitude: 8           # +8 multiplier
  hand_matching_rule: "ContainsPair"
```

**Score Calculation**: `(base_score + card_values) × (base_mult + joker_mult)`

### `MultiplyMult`
Multiplies the hand's multiplier when conditions are met.

```yaml
- name: "Multiplier"
  effect: "MultiplyMult"
  effect_magnitude: 2           # ×2 multiplier
  hand_matching_rule: "None"
```

**Score Calculation**: `(base_score + card_values) × (base_mult + joker_mult) × joker_mult_factor`

### `ReplayCard`
Replays matching cards so they're scored twice.

```yaml
- name: "Face Dancer"
  effect: "ReplayCard"
  effect_magnitude: 0
  card_matching_rule: "IsFace"
  description: "Face cards are scored twice"
```

**Score Calculation**: Matching cards add their value again and retrigger card-based bonuses.

## 🧩 Composite Jokers

A joker can now include multiple effects via the `effects` array. Each entry follows the same structure as single-effect jokers.

```yaml
- name: "Combo"
  value: 7
  effects:
    - effect: "AddChips"
      effect_magnitude: 10
      hand_matching_rule: "ContainsPair"
    - effect: "AddMult"
      effect_magnitude: 2
      hand_matching_rule: "ContainsPair"
```

The joker above grants both +10 chips and +2 multiplier whenever the played hand contains a pair.

## 🃏 Hand Matching Rules

### `None`
Always triggers (used for money jokers).

```yaml
hand_matching_rule: "None"
```

### `ContainsPair`
Triggers when the played hand contains a pair.

**Matches**: Pair, Two Pair, Three of a Kind, Full House, Four of a Kind

```yaml
hand_matching_rule: "ContainsPair"
```

### `ContainsTwoPair`  
Triggers when the played hand contains two pair.

**Matches**: Two Pair, Full House

```yaml
hand_matching_rule: "ContainsTwoPair"
```

### `ContainsThreeOfAKind`
Triggers when the played hand contains three of a kind.

**Matches**: Three of a Kind, Full House, Four of a Kind

```yaml
hand_matching_rule: "ContainsThreeOfAKind"
```

### `ContainsStraight`
Triggers when the played hand contains a straight.

**Matches**: Straight, Straight Flush, Royal Flush

```yaml
hand_matching_rule: "ContainsStraight"
```

### `ContainsFlush`
Triggers when the played hand contains a flush.

**Matches**: Flush, Straight Flush, Royal Flush

```yaml
hand_matching_rule: "ContainsFlush"
```

### `ContainsFullHouse`
Triggers only on Full House.

```yaml
hand_matching_rule: "ContainsFullHouse"
```

### `ContainsFourOfAKind`
Triggers only on Four of a Kind.

```yaml
hand_matching_rule: "ContainsFourOfAKind"
```

### `ContainsStraightFlush`
Triggers when the played hand contains a straight flush.

**Matches**: Straight Flush, Royal Flush

```yaml
hand_matching_rule: "ContainsStraightFlush"
```

### `ContainsRoyalFlush`
Triggers only on Royal Flush.

```yaml
hand_matching_rule: "ContainsRoyalFlush"
```

## 🂠 Card Matching Rules

Card matching rules award the joker's effect magnitude for each card in the played hand that matches the rule. When a `card_matching_rule` is present, the `hand_matching_rule` is ignored for scoring purposes.

### `IsAce`
Triggers for each Ace in the played hand.

```yaml
card_matching_rule: "IsAce"
```

### `IsSpade`
Triggers for each Spade in the played hand.

```yaml
card_matching_rule: "IsSpade"
```

### `IsFace`
Triggers for each face card (J, Q, K).

```yaml
card_matching_rule: "IsFace"
```

## 📊 Example Configurations

### Early Game Economy Joker
```yaml
- name: "Penny Pincher"
  value: 3
  rarity: "Common"
  effect: "AddMoney"
  effect_magnitude: 2
  hand_matching_rule: "None"
  description: "Earn $2 at the end of each Blind"
```

### Pair Synergy Build
```yaml
- name: "Pair Power"
  value: 5
  rarity: "Common"
  effect: "AddChips"
  effect_magnitude: 40
  hand_matching_rule: "ContainsPair"
  description: "+40 Chips if played hand contains a Pair"

- name: "Pair Multiplier"
  value: 6
  rarity: "Common"
  effect: "AddMult"
  effect_magnitude: 10
  hand_matching_rule: "ContainsPair" 
  description: "+10 Mult if played hand contains a Pair"
```

### High-Value Hand Focus
```yaml
- name: "Straight Shooter"
  value: 10
  rarity: "Common"
  effect: "AddChips"
  effect_magnitude: 100
  hand_matching_rule: "ContainsStraight"
  description: "+100 Chips if played hand contains a Straight"

- name: "Flush Fund"
  value: 12
  rarity: "Common"
  effect: "AddMult"
  effect_magnitude: 20
  hand_matching_rule: "ContainsFlush"
  description: "+20 Mult if played hand contains a Flush"
```

### Ultra-Rare Power Joker
```yaml
- name: "Royal Treatment"
  value: 25
  rarity: "Legendary"
  effect: "AddMult"
  effect_magnitude: 50
  hand_matching_rule: "ContainsRoyalFlush"
  description: "+50 Mult if played hand is a Royal Flush"
```

## 🔧 Technical Details

### Loading System
1. Game loads `jokers.yaml` at startup
2. If file missing/invalid, uses hardcoded defaults
3. All jokers become available in shop rotation
4. Players can only buy jokers they don't already own

### Effect Application
- **Money effects**: Applied at end of each blind
- **Scoring effects**: Applied during hand evaluation
- **Multiple jokers**: All applicable effects stack additively

### Score Calculation with Jokers
```
Final Score = (Base Score + Joker Chips + Card Values) × (Base Mult + Joker Mult)
```

**Example**: Pair of Aces with Chip Collector (+30 chips) and Double Down (+8 mult)
- Base: 10, Cards: 22, Base Mult: 2
- With jokers: (10 + 30 + 22) × (2 + 8) = 62 × 10 = **620 points**
- Without jokers: (10 + 22) × 2 = **64 points**

## 🎮 Gameplay Balance Tips

### Pricing Guidelines
- **Early game utility**: $3-6 (affordable after 1-2 blinds)
- **Mid game power**: $7-12 (requires planning/saving)
- **Late game luxury**: $15+ (major investment)

### Effect Magnitude Guidelines
- **Chips**: 20-50 for common hands, 80-150 for rare hands
- **Mult**: 5-15 for common hands, 20-40 for rare hands  
- **Money**: 2-6 per blind

### Rarity Considerations
Currently `rarity` field is unused but reserved for future features:
- Shop appearance rates
- Special visual effects
- Cost scaling
- Unlock requirements

## 🛠️ Adding New Jokers

### Step 1: Design the Joker
- Choose effect type (AddMoney, AddChips, AddMult)
- Select hand matching rule
- Set magnitude and price
- Write descriptive name and description

### Step 2: Add to YAML
```yaml
- name: "Your Joker Name"
  value: 8
  rarity: "Common"
  effect: "AddChips"
  effect_magnitude: 45
  hand_matching_rule: "ContainsThreeOfAKind"
  description: "+45 Chips if played hand contains Three of a Kind"
```

### Step 3: Test
```bash
go run .
# Buy your joker in shop
# Play matching hands to verify effect
```

## 🚨 Troubleshooting

### "Warning: Could not load jokers.yaml"
- **Cause**: File missing or invalid YAML syntax
- **Solution**: Check file exists and YAML is valid
- **Fallback**: Game uses hardcoded defaults

### "No new jokers available" 
- **Cause**: All jokers already owned or unaffordable
- **Solution**: Add more jokers to YAML or increase starting money

### Joker effects not applying
- **Check**: Hand matching rule is correct
- **Check**: Effect type matches expected behavior
- **Check**: Joker was actually purchased and shows in "Your Jokers"

### Invalid YAML format
```yaml
# ❌ Wrong
- name: Broken Joker
  value: "not a number"

# ✅ Correct  
- name: "Fixed Joker"
  value: 5
```

## 📈 Advanced Configurations

### Synergy Sets
Create jokers that work well together:

```yaml
# Pair synergy set
- name: "Pair Starter"
  effect: "AddChips"
  effect_magnitude: 20
  hand_matching_rule: "ContainsPair"

- name: "Pair Booster"  
  effect: "AddMult"
  effect_magnitude: 6
  hand_matching_rule: "ContainsPair"

- name: "Pair Master"
  effect: "AddChips" 
  effect_magnitude: 60
  hand_matching_rule: "ContainsTwoPair"
```

### Progressive Difficulty
```yaml
# Early ante jokers (cheaper, weaker)
- name: "Training Wheels"
  value: 3
  effect: "AddChips"
  effect_magnitude: 15

# Late ante jokers (expensive, powerful)  
- name: "Endgame Engine"
  value: 20
  effect: "AddMult"
  effect_magnitude: 35
```

### Meta Variations
```yaml
# High-risk, high-reward
- name: "Royal Gambit"
  value: 30
  effect: "AddMult" 
  effect_magnitude: 100
  hand_matching_rule: "ContainsRoyalFlush"

# Consistent value
- name: "Steady Eddie"
  value: 8
  effect: "AddChips"
  effect_magnitude: 25
  hand_matching_rule: "None"  # Always triggers
```

## 🔄 Live Editing Workflow

1. **Edit** `jokers.yaml` while game is not running
2. **Start game** - changes load automatically  
3. **Test** new jokers in shop
4. **Iterate** - exit, modify, restart
5. **Share** configurations with other players

The YAML system makes joker experimentation fast, safe, and accessible to non-programmers!

---

*Happy joker crafting! 🃏*