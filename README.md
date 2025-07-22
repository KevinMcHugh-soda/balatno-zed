# Balatro CLI

A command-line clone of the popular deckbuilding roguelike game Balatro. This implementation focuses on the core poker hand evaluation and scoring mechanics.

## About

Balatro CLI is a simplified version of Balatro that runs in your terminal. Players face a challenge to score 300 points using only 4 hands and 3 discards. You're dealt 7 cards from a standard 52-card deck and can either play up to 5 cards to form a poker hand, or discard unwanted cards to get new ones. Each hand type has a base score and multiplier, and card values are added to create the final score.

## How to Run

```bash
# Build the game
go build -o balatro main.go

# Run the game
./balatro
```

Or run directly with Go:
```bash
go run main.go
```

## How to Play

**🎯 CHALLENGE: Score 300 points with 4 hands and 3 discards!**

1. You'll be dealt 7 cards from a shuffled deck
2. Choose your action:
   - **`play <cards>`**: Play 1-5 cards as a poker hand (uses one of your 4 hands)
   - **`discard <cards>`**: Discard unwanted cards and get new ones (uses one of your 3 discards)
3. The game evaluates your poker hand and adds to your total score
4. Win by reaching 300 points before running out of hands
5. Type 'quit' to exit early

### Example Gameplay
```
🎯 Target: 300 | Current Score: 0 | Hands Left: 4 | Discards Left: 3

Your cards:
1: A♠
2: A♦  
3: K♥
4: Q♣
5: J♠
6: 10♦
7: 9♣

Choose action: 'play <cards>' to play hand, 'discard <cards>' to discard, or 'quit': play 1 2 3
Your hand: A♠ A♦ K♥
Hand type: Pair
Base Score: 10 | Card Values: 32 | Mult: 2x
Final Score: (10 + 32) × 2 = 84 points
💰 Total Score: 84/300

Choose action: discard 4 5
Discarded 2 card(s)
New cards dealt!
```

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

## Examples

### High-Value Hands
- **Royal Flush with A♥ K♥ Q♥ J♥ 10♥**: (100 + 51) × 8 = **1,208 points**
- **Four Aces**: (60 + 44) × 7 = **728 points**
- **Straight Flush**: (100 + 35) × 8 = **1,080 points**

### Strategic Considerations
- **Resource Management**: You only get 4 hands and 3 discards - use them wisely!
- **Discard Strategy**: Use discards to get rid of low-value cards and hunt for pairs/straights
- **Scoring Efficiency**: Sometimes a pair of Aces (84 points) scores higher than a straight with low cards
- **Card Values**: Face cards and Aces significantly boost your card value total
- **Risk vs Reward**: Consider both the hand type and card values when deciding to play or discard

## Testing

The game includes comprehensive tests covering all poker hand types and edge cases.

### Running Tests
```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Check test coverage
go test -cover
```

### Test Coverage
- **66%+ statement coverage**
- All 10 poker hand types tested
- Card and deck functionality
- Scoring calculations
- Edge cases (empty hands, single cards, etc.)
- Reproducible gameplay with seeds

### Deterministic Testing
For testing and debugging, you can set a specific seed:
```bash
# Run with a specific seed for reproducible results
./balatro -seed=42

# Same seed always produces the same cards
go test -run TestSetSeed
```

## Code Structure

The codebase is organized into focused, modular files:

- **`main.go`** - Entry point and command-line argument parsing
- **`deck.go`** - Card, Suit, Rank definitions and deck operations
- **`hands.go`** - Interface-based poker hand evaluation system
- **`game.go`** - Main game loop, player interaction, and game state

### Interface-Based Hand System

The hand evaluation uses a clean interface-based approach instead of switch statements:

```go
type HandEvaluator interface {
    Name() string
    BaseScore() int
    Multiplier() int
    Matches(cards []Card) bool
    Priority() int
}
```

Each poker hand type (Pair, Flush, etc.) implements this interface, making the system:
- **Extensible**: Easy to add new hand types
- **Maintainable**: No large switch statements
- **Testable**: Each hand type can be tested independently

## Implementation Notes

- Standard 52-card deck
- Poker hand evaluation follows traditional rules
- Ace-low straights (A-2-3-4-5) are supported
- Royal Flush requires A-K-Q-J-10 of the same suit
- Centralized random source with configurable seeding

## Future Enhancements

This is a basic implementation focusing on core mechanics. The full Balatro game includes:
- Jokers that modify scoring
- Multiple rounds and blinds
- Shop system
- Card modifications and enhancements
- Boss blinds with special rules

---

*Built with Go. Inspired by LocalThunk's Balatro.*