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

### The Golden Joker
- **Name**: "The Golden Joker"
- **Cost**: $6
- **Effect**: Earn $4 at the end of each Blind
- **Description**: "Earn $4 at the end of each Blind"
- **ROI**: 66.7% return per blind (pays for itself in 1.5 blinds)

### Implementation Features
- **Function-based effects**: Easy to add new joker types
- **Modular design**: Each joker is self-contained
- **Helper functions**: `PlayerHasJoker()`, `CalculateJokerRewards()`, etc.
- **Future-ready**: Framework supports multiple joker types

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
2. **Additional Jokers**: Implement scoring-based jokers (not just money)
3. **Tarot Cards**: One-time purchase items with immediate effects
4. **Card Packs**: Purchase additional cards for deck building

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

This implementation brings Balatro CLI significantly closer to the authentic Balatro experience by adding the crucial economic layer that makes the game strategic beyond just poker hand evaluation.