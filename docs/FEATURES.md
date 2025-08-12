# Balatro CLI - Feature Implementation Guide

## 🎯 Recently Implemented: Money, Shop & Joker System

This document details the complete economic gameplay system added to Balatro CLI, bringing it closer to the authentic Balatro experience.

---

## 💰 Money System

### Core Mechanics
- **Starting Money**: Players begin each run with $4
- **Persistence**: Money carries over between blinds and antes
- **Display**: Always visible in the game status bar
- **Earning**: Primary source is blind completion rewards

### Money Sources

#### 1. Blind Completion Base Rewards
| Blind Type | Base Reward |
|------------|-------------|
| Small Blind | $4 |
| Big Blind | $5 |
| Boss Blind | $6 |

#### 2. Efficiency Bonuses
- **$1 per unused hand** (max $4 if no hands played)
- **$1 per unused discard** (max $3 if no discards used)

#### 3. Joker Bonuses
- **The Golden Joker**: +$4 per blind completion
- *(Framework ready for additional jokers)*

### Example Calculations

**Perfect Small Blind** (0 hands, 0 discards used):
- Base: $4 + Unused: $7 (4 hands + 3 discards) = **$11 total**

**Efficient Big Blind** (2 hands, 1 discard used):
- Base: $5 + Unused: $4 (2 hands + 2 discards) = **$9 total**

**Golden Joker Boss Blind** (4 hands, 3 discards used):
- Base: $6 + Unused: $0 + Joker: $4 = **$10 total**

---

## 🏪 Shop System

### When It Appears
- **Between every blind** (after Small → Big, Big → Boss, Boss → Small of next Ante)
- **Automatic**: No player choice to skip the shop entirely
- **Optional purchases**: Players can choose to skip buying items

### Shop Interface
```
🏪 SHOP 🏪
💰 Your Money: $12

1. The Golden Joker - $6
   Earn $4 at the end of each Blind

🃏 Your Jokers: The Golden Joker

Buy (1) The Golden Joker, or (s)kip shop:
```

### Shop Features
- **Affordability Check**: Only shows purchase options if player can afford them
- **Ownership Check**: Won't offer jokers the player already owns
- **Current Inventory**: Displays owned jokers clearly
- **Simple Input**: Type `1` to buy, anything else to skip
- **Joker Reordering**: Press `j` to reorder owned jokers, `s` to sell the selected joker for half price

---

## 🃏 Joker System

### Architecture
```go
type Joker struct {
    Name        string              // Display name
    Description string              // What it does
    Price       int                 // Shop cost
    OnBlindEnd  func() int         // Money bonus function
}
```

### YAML Joker Configuration System
- **15+ Configurable Jokers**: All defined in `jokers.yaml`
- **Runtime Loading**: No compilation needed for new jokers
- **Three Effect Types**: AddMoney, AddChips, AddMult
- **Hand-Based Triggers**: Effects activate based on played hand types
- **Fallback Safety**: Uses defaults if YAML file missing/invalid

### Effect Types

#### AddMoney Jokers
- **The Golden Joker** ($6): Earn $4 at the end of each Blind
- **Effect**: Triggers after blind completion regardless of hands played
- **ROI**: 66.7% return per blind (pays for itself in 1.5 blinds)

#### AddChips Jokers
- **Chip Collector** ($5): +30 chips if hand contains a Pair
- **Two's Company** ($6): +50 chips if hand contains Two Pair
- **Triple Threat** ($7): +80 chips if hand contains Three of a Kind
- **Straight Shooter** ($8): +100 chips if hand contains a Straight
- **Effect**: Adds to base score before multiplier application

#### AddMult Jokers
- **Double Down** ($4): +8 mult if hand contains a Pair
- **Pair Paradise** ($5): +12 mult if hand contains Two Pair
- **Three's a Charm** ($6): +15 mult if hand contains Three of a Kind
- **Linear Logic** ($7): +20 mult if hand contains a Straight
- **Effect**: Increases final multiplier for explosive scoring

#### ReplayCard Jokers
- **Face Dancer** ($7): Face cards are scored twice
- **Effect**: Replays matching cards, doubling their value and bonuses

### Hand Matching Rules
- **ContainsPair**: Triggers on Pair, Two Pair, Full House, Four of a Kind
- **ContainsTwoPair**: Triggers on Two Pair, Full House
- **ContainsThreeOfAKind**: Triggers on Three of a Kind, Full House, Four of a Kind
- **ContainsStraight**: Triggers on Straight, Straight Flush, Royal Flush
- **ContainsFlush**: Triggers on Flush, Straight Flush, Royal Flush
- **ContainsFullHouse**: Triggers only on Full House
- **ContainsFourOfAKind**: Triggers only on Four of a Kind
- **None**: Always triggers (used for money jokers)

### Implementation Features
- **YAML Configuration**: Easy to add/modify jokers without coding
- **Function-based effects**: Dynamic effect generation from config
- **Modular design**: Each joker is self-contained
- **Helper functions**: `PlayerHasJoker()`, `CalculateJokerRewards()`, `CalculateJokerHandBonus()`
- **Stacking Effects**: Multiple jokers can affect the same hand additively

---

## 📊 Economic Strategy

### Early Game (Antes 1-2)
1. **Priority**: Buy The Golden Joker ASAP
2. **Strategy**: Focus on efficiency over perfect hands
3. **Goal**: Build economic foundation for harder blinds

### Mid Game (Antes 3-5)
1. **Advantage**: Golden Joker provides consistent $4 per blind
2. **Strategy**: Balance efficiency with increasing blind requirements  
3. **Goal**: Maintain money flow while scaling hand strength

### Late Game (Antes 6-8)
1. **Reality**: Money becomes less important than raw scoring power
2. **Strategy**: Use all resources if needed to beat high-requirement blinds
3. **Goal**: Survival over economic optimization

### Golden Joker Value Analysis
- **Break-even**: 1.5 blinds to pay for itself
- **Full run value**: ~$32-40 total return (8 antes × 3 blinds × $4 - $6 cost)
- **Opportunity cost**: Very low - $6 is easily earned in one efficient blind

---

## 🔧 Technical Implementation

### File Structure
```
jokers.go       - Joker definitions and helper functions  
game.go         - Shop integration and money management
```

### Key Functions

#### Money Management
```go
func (g *Game) calculateBlindReward() int
// Calculates total money earned from blind completion

func CalculateJokerRewards(jokers []Joker) int  
// Sums up all joker bonus money
```

#### Shop System
```go
func (g *Game) showShop()
// Displays shop interface and handles purchases

func PlayerHasJoker(playerJokers []Joker, jokerName string) bool
// Checks if player already owns a specific joker
```

#### Reward Display
```go
💰 REWARD BREAKDOWN:
   Base: $4 + Unused: $4 (1 hands + 3 discards) + Jokers: $4
   💰 Total Earned: $12 | Your Money: $24
```

### Integration Points
- **Blind completion**: Triggers reward calculation
- **State transitions**: Shop appears before new blind starts  
- **Status display**: Money shown in game status bar
- **Save state**: Money persists through blind/ante changes

---

## 🎮 Player Experience

### Immediate Feedback
- **Clear cost**: "$6" shown with each item
- **Affordability**: Only show purchasable items
- **Confirmation**: "✨ Purchased The Golden Joker! ✨"
- **Updated status**: Money immediately reflects purchases

### Strategic Depth
- **Risk/Reward**: Spend money now vs save for later
- **Efficiency incentive**: Unused resources = more money
- **Long-term planning**: Golden Joker pays dividends over time
- **Resource management**: Balance between completion and optimization

### Visual Design
- **💰 Money icon**: Consistent throughout UI
- **🏪 Shop icon**: Clear section demarcation  
- **🃏 Joker icon**: Distinct from regular cards
- **Progress tracking**: Money shown in status bar always

---

## 💀 Boss Blind Modifiers

Boss Blinds now apply a random rule to shake up gameplay:

- **Hearts score zero** – any heart card contributes no value
- **Hand size reduced by 1** – start the blind with one fewer card
- **Hand size increased by 1** – begin with an extra card for more options

These modifiers are announced at the start of each Boss Blind.

---

## 🚀 Future Expansion

### Ready Framework
The current implementation provides a solid foundation for:

1. **Multiple Jokers**: Easy to add new joker types with different effects
2. **Complex Effects**: Framework supports any `func() int` bonus structure
3. **Shop Expansion**: Can add more item types (tarot cards, planet cards, etc.)
4. **Dynamic Pricing**: Joker prices could scale with ante or other factors
5. **Rarity System**: Could implement common/uncommon/rare jokers
6. **Conditional Effects**: Jokers could have requirements or triggers

### Next Logical Steps
1. **Boss Blind Constraints**: Add special rules that make Boss Blinds unique
2. **Extended Joker Effects**: Conditional triggers, card-specific bonuses, deck modifications
3. **Tarot Cards**: One-time purchase items with immediate effects
4. **Card Packs**: Purchase additional cards for deck building
5. **Joker Rarity System**: Implement rarity-based shop appearance and effects

---

## ✅ Testing & Quality

### Test Coverage
- **Money calculations**: All reward formulas tested
- **Joker functions**: Golden Joker behavior verified  
- **Shop logic**: Purchase/skip flows validated
- **Integration**: Full game flow tested
- **Edge cases**: Empty joker lists, insufficient funds, etc.

### Performance
- **Efficient calculations**: O(n) complexity for joker rewards
- **Minimal memory**: Jokers stored as simple structs
- **Fast UI**: Shop appears instantly between blinds

---

## 🃏 YAML Joker System Implementation

### Configuration Architecture
```yaml
jokers:
  - name: "Joker Name"
    value: 6                    # Shop price
    rarity: "Common"            # Future expansion
    effect: "AddChips"          # Effect type  
    effect_magnitude: 30        # Effect strength
    hand_matching_rule: "ContainsPair"  # Trigger condition
    description: "Effect description"
```

### Technical Implementation
```go
type JokerEffect string
const (
    AddMoney JokerEffect = "AddMoney"
    AddChips JokerEffect = "AddChips" 
    AddMult  JokerEffect = "AddMult"
)

type HandMatchingRule string
const (
    ContainsPair HandMatchingRule = "ContainsPair"
    ContainsStraight HandMatchingRule = "ContainsStraight"
    // ... more rules
)
```

### Scoring Integration
- **Hand Evaluation**: Jokers checked during `EvaluateHand()`
- **Effect Application**: `CalculateJokerHandBonus()` returns chips/mult bonuses
- **Score Calculation**: `(base + joker_chips + cards) × (base_mult + joker_mult)`
- **Visual Feedback**: Detailed breakdown shows joker contributions

### Strategic Impact
The YAML joker system transforms gameplay from simple optimization to complex strategic decisions:

**Before**: "Play highest scoring hand possible"
**After**: "Build joker synergies that multiply scoring potential over multiple blinds"

**Example Decision Tree**:
- Buy Chip Collector early for consistent pair bonuses
- Save for Double Down to multiply pair effectiveness  
- Consider Straight Shooter for high-stakes late game power
- Balance economic jokers vs scoring jokers based on ante progression

### Balance Configurability
- **No Recompilation**: Edit `jokers.yaml` → restart game → changes active
- **Rapid Iteration**: Test balance changes in seconds
- **Community Sharing**: Players can share custom joker configs
- **A/B Testing**: Easy to compare different joker power levels

This implementation brings Balatro CLI significantly closer to the authentic Balatro experience by adding both the crucial economic layer AND the strategic depth of configurable joker synergies that make the game endlessly replayable.

## 💾 Save & Load

- **Load from JSON**: Resume a run using `-load <file>`
- **Auto-save**: Game state written to timestamped JSON when quitting or timing out
- **Save format**: JSON with `save_version`, `seed`, `current_ante`, `current_blind`, `current_money`, and `current_jokers`