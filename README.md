# Balatro CLI

A command-line clone of the popular deckbuilding roguelike game Balatro. This implementation focuses on the core poker hand evaluation and scoring mechanics.

## About

Balatro CLI is a simplified version of Balatro that runs in your terminal. Players are dealt 7 cards from a standard 52-card deck and must select up to 5 cards to form the best possible poker hand. Each hand type has a base score and multiplier, and card values are added to create the final score.

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

1. You'll be dealt 7 cards from a shuffled deck
2. Enter the numbers (1-7) of the cards you want to include in your hand
3. You can select 1-5 cards
4. The game will evaluate your poker hand and show your score
5. Type 'quit' to exit

### Example Gameplay
```
Your cards:
1: A♠
2: A♦  
3: K♥
4: Q♣
5: J♠
6: 10♦
7: 9♣

Select cards for your hand: 1 2 3
Your hand: A♠ A♦ K♥
Hand type: Pair
Base Score: 10 | Card Values: 32 | Mult: 2x
Final Score: (10 + 32) × 2 = 84 points
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
- Sometimes a pair of Aces (84 points) might score higher than a straight with low cards
- Face cards and Aces significantly boost your card value total
- Consider both the hand type and the card values when selecting your hand

## Implementation Notes

- Standard 52-card deck
- Poker hand evaluation follows traditional rules
- Ace-low straights (A-2-3-4-5) are supported
- Royal Flush requires A-K-Q-J-10 of the same suit
- Random deck shuffling each game

## Future Enhancements

This is a basic implementation focusing on core mechanics. The full Balatro game includes:
- Jokers that modify scoring
- Multiple rounds and blinds
- Shop system
- Card modifications and enhancements
- Boss blinds with special rules

---

*Built with Go. Inspired by LocalThunk's Balatro.*