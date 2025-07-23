# Balatro CLI

A command-line clone of the popular deckbuilding roguelike game Balatro. This implementation focuses on the core poker hand evaluation and scoring mechanics with authentic Ante/Blind progression.

## About

Balatro CLI is a faithful recreation of Balatro's core progression system that runs in your terminal. Players must conquer 8 Antes, each containing 3 increasingly difficult Blinds: Small Blind, Big Blind, and Boss Blind. You're dealt 7 cards from a standard 52-card deck and can either play up to 5 cards to form a poker hand, or discard unwanted cards to get new ones. Each hand type has a base score and multiplier, and card values are added to create the final score.

The game features a complete **money and shop system** - earn money by completing blinds and spend it on powerful Jokers that provide ongoing benefits throughout your run.

## How to Run

```bash
# Build the game
go build -o balatro .

# Run the game
./balatro
```

Or run directly with Go:
```bash
go run .
```

For reproducible gameplay:
```bash
# Use a specific seed
go run . -seed 42
```

## How to Play

**🎯 CHALLENGE: Progress through 8 Antes, each with 3 Blinds!**

### Game Structure
- **8 Antes** total to complete the game
- Each Ante contains **3 Blinds** in sequence:
  - 🔸 **Small Blind** - Base difficulty
  - 🔶 **Big Blind** - 1.5x harder than Small Blind  
  - 💀 **Boss Blind** - 2x harder than Small Blind (special rules coming soon!)
- **🏪 Shop** appears between each blind where you can spend money on Jokers

### Each Blind Challenge
1. You get **4 hands** and **3 discards** to reach the target score
2. You'll be dealt 7 cards from a shuffled deck
3. Choose your action:
   - **`play <cards>`**: Play 1-5 cards as a poker hand (uses one of your 4 hands)
   - **`discard <cards>`**: Discard unwanted cards and get new ones (uses one of your 3 discards)
   - **`resort`**: Toggle card sorting between rank and suit
4. The game evaluates your poker hand and adds to your total score
5. Beat the blind by reaching the target score before running out of hands
6. **Earn money** based on blind type and efficiency
7. **Visit the shop** to buy Jokers with your earned money
8. Complete all 3 blinds to advance to the next Ante
9. Type `quit` to exit early

### Example Gameplay
```
🔸 Ante 1 - Small Blind
🎯 Target: 300 | Score: 0 [░░░░░░░░░░░░░░░░░░░░] (0.0%)
🎴 Hands Left: 4 | 🗑️  Discards Left: 3 | 💰 Money: $4

Your cards (sorted by rank):
1: A♠
2: A♦  
3: K♥
4: Q♣
5: J♠
6: 10♦
7: 9♣

(p)lay <cards>, (d)iscard <cards>, (r)esort, or (q)uit: play 1 2 3

Your hand: A♠ A♦ K♥
Hand type: Pair
Base Score: 10 | Card Values: 32 | Mult: 2x
Final Score: (10 + 32) × 2 = 84 points
💰 Total Score: 84/300

💰 REWARD BREAKDOWN:
   Base: $4 + Unused: $4 (2 hands + 3 discards)
   💰 Total Earned: $8 | Your Money: $12

🏪 SHOP 🏪
💰 Your Money: $12

1. The Golden Joker - $6
   Earn $4 at the end of each Blind

Buy (1) The Golden Joker, or (s)kip shop:
```

## Blind Requirements

The difficulty scales progressively through each Ante:

| Ante | Small Blind | Big Blind | Boss Blind |
|------|-------------|-----------|------------|
| 1    | 300         | 450       | 600        |
| 2    | 375         | 562       | 750        |
| 3    | 450         | 675       | 900        |
| 4    | 525         | 787       | 1050       |
| 5    | 600         | 900       | 1200       |
| 6    | 675         | 1012      | 1350       |
| 7    | 750         | 1125      | 1500       |
| 8    | 825         | 1237      | 1650       |

**Formula**: Base requirement increases by 75 points per Ante, with Big Blind = 1.5x Small Blind and Boss Blind = 2x Small Blind.

## Money & Shop System

### Earning Money
You start each run with **$4** and earn money by completing blinds:

| Blind Type | Base Reward | Bonus Rewards |
|------------|-------------|---------------|
| Small Blind | $4 | +$1 per unused hand |
| Big Blind | $5 | +$1 per unused discard |
| Boss Blind | $6 | **Joker bonuses** |

**💡 Example**: Complete Small Blind using only 2 hands and 1 discard = $4 + $2 + $2 = **$8 total**

### The Shop
Between each blind, you visit the **🏪 Shop** where you can:
- Purchase **Jokers** that provide permanent benefits
- View your current money and owned Jokers
- Choose to skip and save money for later

### The Golden Joker
**The Golden Joker** - *$6*
- **Effect**: Earn $4 at the end of each Blind
- **ROI**: Pays for itself in 1.5 blinds  
- **Strategy**: Essential early-game purchase for economic snowballing

### Money Management Tips
- **Efficiency Rewards**: Unused hands/discards = more money
- **Early Investment**: The Golden Joker quickly pays for itself
- **Resource Planning**: Balance between completing blinds and preserving resources
- **Blind Scaling**: Higher blinds give more base money but are harder to complete efficiently

## Scoring System

The final score is calculated using the formula:
**Final Score = (Base Score + Card Values) × Multiplier**

### Card Values
- **Number cards (2-10)**: Face value
- **Face cards (J, Q, K)**: 10 points each
- **Aces**: 11 points each

### Hand Types

| Hand Type | Base Score | Multiplier | Example |
|-----------|------------|------------|---------|
| High Card | 5 | 1x | A♠ K♦ Q♣ |
| Pair | 10 | 2x | 7♥ 7♣ K♠ |
| Two Pair | 20 | 2x | 7♥ 7♣ K♠ K♦ |
| Three of a Kind | 30 | 3x | 7♥ 7♣ 7♠ K♦ |
| Straight | 30 | 4x | 5♠ 6♦ 7♥ 8♣ 9♠ |
| Flush | 35 | 4x | 2♥ 7♥ 9♥ J♥ K♥ |
| Full House | 40 | 4x | 7♥ 7♣ 7♠ K♦ K♠ |
| Four of a Kind | 60 | 7x | 7♥ 7♣ 7♠ 7♦ K♠ |
| Straight Flush | 100 | 8x | 5♥ 6♥ 7♥ 8♥ 9♥ |
| Royal Flush | 100 | 8x | 10♥ J♥ Q♥ K♥ A♥ |

## Victory Celebrations

Each blind type has unique victory celebrations:

- **🔸 Small Blind**: Sparkling celebration, advance to Big Blind
- **🔶 Big Blind**: Lightning celebration, prepare for Boss Blind  
- **💀 Boss Blind**: Epic fireworks celebration, Ante conquered!

Complete all 8 Antes to achieve **ULTIMATE VICTORY** and become a true Balatro master!

## Strategic Considerations

### Early Antes (1-3) - Economic Foundation
- **Money Priority**: Focus on efficient blind completion to maximize unused bonuses
- **Golden Joker**: Buy it ASAP - it pays for itself in 1.5 blinds
- **Card Strategy**: Consistent scoring with pairs and two pairs
- **Resource Efficiency**: Try to save 1-2 hands per blind for bonus money

### Mid Antes (4-6) - Scaling Power
- **Economic Engine**: Golden Joker should be generating $4 per blind
- **Hand Requirements**: Start prioritizing higher-scoring hands (straights, flushes)
- **Money Accumulation**: Build reserves for potential future Jokers
- **Blind Difficulty**: Big Blinds require 800+ point hands

### Late Antes (7-8) - Endgame Strategy  
- **Premium Hands Required**: Boss Blinds demand 1500+ points
- **Resource Scarcity**: Every hand and discard becomes precious
- **All-or-Nothing**: May need Four of a Kind, Straight Flush, or Royal Flush
- **Economic Pressure**: Money becomes less important than raw scoring power

### General Tips
- **Money Management**: Buy The Golden Joker early for economic advantage
- **Resource Efficiency**: Unused hands/discards = more money for Jokers
- **Progressive Scaling**: Each Ante increases both difficulty and rewards
- **Early Game Focus**: Efficient blind completion > perfect hands
- **Late Game Focus**: Raw scoring power > economic efficiency
- **Joker Value**: The Golden Joker provides ~$32-40 over a full run

## Testing

The game includes comprehensive tests covering the Ante/Blind system and poker evaluation.

### Running Tests
```bash
# Run all tests
go test

# Run tests with verbose output  
go test -v

# Run specific test suites
go test -v blind_test.go game.go deck.go hands.go
```

### Test Coverage
- **Ante/Blind progression system**
- **Blind requirement calculations**
- **All 10 poker hand types**
- **Card and deck functionality** 
- **Scoring calculations**
- **State reset between blinds**
- **Victory and defeat conditions**

### Deterministic Testing
```bash
# Reproducible gameplay for testing
./balatro -seed=42

# Demo progression through blinds
./demo_progression.sh
```

## Code Structure

The codebase is organized into focused, modular files:

- **`main.go`** - Entry point and command-line argument parsing
- **`deck.go`** - Card, Suit, Rank definitions and deck operations  
- **`hands.go`** - Interface-based poker hand evaluation system
- **`game.go`** - Ante/Blind progression, game loop, and player interaction
- **`jokers.go`** - Joker system, shop mechanics, and The Golden Joker

### Ante/Blind System

The progression system includes:

```go
type BlindType int
const (
    SmallBlind BlindType = iota
    BigBlind  
    BossBlind
)

// Calculates score requirements
func GetBlindRequirement(ante int, blindType BlindType) int

// Handles blind completion and advancement
func (g *Game) handleBlindCompletion()
```

### Money & Joker System

Economic gameplay mechanics:

```go
type Joker struct {
    Name        string
    Description string
    Price       int
    OnBlindEnd  func() int // Money earned per blind
}

// Reward calculation with bonuses
func (g *Game) calculateBlindReward() int

// Shop system between blinds  
func (g *Game) showShop()
```

### Interface-Based Hand System

Clean interface-based poker evaluation:

```go
type HandEvaluator interface {
    Name() string
    BaseScore() int  
    Multiplier() int
    Matches(cards []Card) bool
    Priority() int
}
```

## Visual Features

- **📊 Progress Bars**: Visual score progress with `█` and `░` characters
- **🎭 Blind Indicators**: Unique emojis for each blind type (🔸🔶💀)
- **🎆 Celebrations**: Escalating victory animations with detailed reward breakdowns
- **💰 Money Tracking**: Always-visible money counter in game status
- **🏪 Shop Interface**: Clean shop display with affordability indicators
- **📍 Clear Status**: Ante, blind type, money, and requirements always visible
- **🎨 Colorful Output**: Rich terminal formatting for better UX

## Implementation Notes

- Standard 52-card deck with proper shuffling
- Authentic Balatro progression scaling with complete economy system
- State resets between blinds (fresh hand, restored resources, money persists)
- Modular Joker system with function-based effects
- Interface-based design for extensibility
- Comprehensive shop and money management
- Comprehensive error handling and input validation
- Centralized random source with configurable seeding

## Future Enhancements

This implementation includes core progression with money/shop/joker systems. The full Balatro experience also includes:

- **Boss Blind Effects**: Special rules and constraints for Boss Blinds *(coming soon!)*
- **Additional Jokers**: Dozens more game-changing modifiers beyond The Golden Joker
- **Advanced Shop Items**: Tarot cards, Planet cards, and card packs
- **Card Enhancements**: Foil, holographic, and other card modifications
- **Vouchers**: Permanent upgrades and rule modifications
- **Stakes**: Higher difficulty modes with additional constraints
- **Endless Mode**: Continue beyond Ante 8 for ultimate challenges

**✅ Currently Implemented**: Ante progression, money system, shop, The Golden Joker

---

*Built with Go. Faithfully recreating LocalThunk's Balatro progression system.*